/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bossnetiov1 "github.com/boss-net/bossnet-operator/api/v1"
	"github.com/boss-net/bossnet-operator/internal/bossnet"
	"github.com/boss-net/bossnet-operator/internal/utils"
)

const (
	// BossnetDeploymentFinalizer is the finalizer used to ensure cleanup of Bossnet deployments
	BossnetDeploymentFinalizer = "bossnet.io/deployment-cleanup"

	// BossnetDeploymentConditionReady indicates the deployment is ready
	BossnetDeploymentConditionReady = "Ready"

	// BossnetDeploymentConditionSynced indicates the deployment is synced with Bossnet API
	BossnetDeploymentConditionSynced = "Synced"

	// RequeueIntervalReady is the interval for requeuing when deployment is ready
	RequeueIntervalReady = 5 * time.Minute

	// RequeueIntervalError is the interval for requeuing on errors
	RequeueIntervalError = 30 * time.Second

	// RequeueIntervalSync is the interval for requeuing during sync operations
	RequeueIntervalSync = 10 * time.Second
)

// BossnetDeploymentReconciler reconciles a BossnetDeployment object
type BossnetDeploymentReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	BossnetClient bossnet.BossnetClient
}

//+kubebuilder:rbac:groups=bossnet.io,resources=bossnetdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bossnet.io,resources=bossnetdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bossnet.io,resources=bossnetdeployments/finalizers,verbs=update

// Reconcile handles the reconciliation of a BossnetDeployment
func (r *BossnetDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(1).Info("Reconciling BossnetDeployment", "request", req)

	var deployment bossnetiov1.BossnetDeployment
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		if apierrors.IsNotFound(err) {
			log.V(1).Info("BossnetDeployment not found, ignoring", "request", req)
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get BossnetDeployment", "request", req)
		return ctrl.Result{}, err
	}

	// Handle deletion
	if deployment.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &deployment)
	}

	// Ensure finalizer is present
	if !controllerutil.ContainsFinalizer(&deployment, BossnetDeploymentFinalizer) {
		controllerutil.AddFinalizer(&deployment, BossnetDeploymentFinalizer)
		if err := r.Update(ctx, &deployment); err != nil {
			log.Error(err, "Failed to add finalizer", "deployment", deployment.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	specHash, err := utils.Hash(deployment.Spec, 16)
	if err != nil {
		log.Error(err, "Failed to calculate spec hash", "deployment", deployment.Name)
		return ctrl.Result{}, err
	}

	if r.needsSync(&deployment, specHash) {
		log.Info("Starting sync with Bossnet API", "deployment", deployment.Name)
		result, err := r.syncWithBossnet(ctx, &deployment)
		if err != nil {
			return result, err
		}
		return result, nil
	}

	return ctrl.Result{RequeueAfter: RequeueIntervalReady}, nil
}

// needsSync determines if the deployment needs to be synced with Bossnet API
func (r *BossnetDeploymentReconciler) needsSync(deployment *bossnetiov1.BossnetDeployment, currentSpecHash string) bool {
	if deployment.Status.Id == nil || *deployment.Status.Id == "" {
		return true
	}

	if deployment.Status.SpecHash != currentSpecHash {
		return true
	}

	if deployment.Status.ObservedGeneration < deployment.Generation {
		return true
	}

	// Drift detection: sync if last sync was too long ago
	if deployment.Status.LastSyncTime == nil {
		return true
	}

	timeSinceLastSync := time.Since(deployment.Status.LastSyncTime.Time)
	return timeSinceLastSync > 10*time.Minute
}

// syncWithBossnet syncs the deployment with the Bossnet API
func (r *BossnetDeploymentReconciler) syncWithBossnet(ctx context.Context, deployment *bossnetiov1.BossnetDeployment) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Use injected client if available (for testing)
	bossnetClient := r.BossnetClient
	if bossnetClient == nil {
		var err error
		bossnetClient, err = bossnet.NewClientFromK8s(ctx, &deployment.Spec.Server, r.Client, deployment.Namespace, log)
		if err != nil {
			log.Error(err, "Failed to create Bossnet client", "deployment", deployment.Name)
			return ctrl.Result{}, err
		}
	}

	flowID, err := bossnet.GetFlowIDFromDeployment(ctx, bossnetClient, deployment)
	if err != nil {
		log.Error(err, "Failed to get flow ID", "deployment", deployment.Name)
		r.setCondition(deployment, BossnetDeploymentConditionSynced, metav1.ConditionFalse, "FlowIDError", err.Error())
		return ctrl.Result{}, err
	}

	deploymentSpec, err := bossnet.ConvertToDeploymentSpec(deployment, flowID)
	if err != nil {
		log.Error(err, "Failed to convert deployment spec", "deployment", deployment.Name)
		r.setCondition(deployment, BossnetDeploymentConditionSynced, metav1.ConditionFalse, "ConversionError", err.Error())
		return ctrl.Result{}, err
	}

	bossnetDeployment, err := bossnetClient.CreateOrUpdateDeployment(ctx, deploymentSpec)
	if err != nil {
		log.Error(err, "Failed to create or update deployment in Bossnet", "deployment", deployment.Name)
		r.setCondition(deployment, BossnetDeploymentConditionSynced, metav1.ConditionFalse, "SyncError", err.Error())
		return ctrl.Result{}, err
	}

	bossnet.UpdateDeploymentStatus(deployment, bossnetDeployment)

	specHash, err := utils.Hash(deployment.Spec, 16)
	if err != nil {
		log.Error(err, "Failed to calculate spec hash", "deployment", deployment.Name)
		return ctrl.Result{}, err
	}
	deployment.Status.SpecHash = specHash
	deployment.Status.ObservedGeneration = deployment.Generation

	r.setCondition(deployment, BossnetDeploymentConditionSynced, metav1.ConditionTrue, "SyncSuccessful", "Deployment successfully synced with Bossnet API")
	r.setCondition(deployment, BossnetDeploymentConditionReady, metav1.ConditionTrue, "DeploymentReady", "Deployment is ready and operational")

	if err := r.Status().Update(ctx, deployment); err != nil {
		log.Error(err, "Failed to update deployment status", "deployment", deployment.Name)
		return ctrl.Result{}, err
	}

	log.Info("Successfully synced deployment with Bossnet", "deploymentId", bossnetDeployment.ID)
	return ctrl.Result{RequeueAfter: RequeueIntervalReady}, nil
}

// setCondition sets a condition on the deployment status
func (r *BossnetDeploymentReconciler) setCondition(deployment *bossnetiov1.BossnetDeployment, conditionType string, status metav1.ConditionStatus, reason, message string) {
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             reason,
		Message:            message,
	}

	meta.SetStatusCondition(&deployment.Status.Conditions, condition)
}

// handleDeletion handles the cleanup of a BossnetDeployment that is being deleted
func (r *BossnetDeploymentReconciler) handleDeletion(ctx context.Context, deployment *bossnetiov1.BossnetDeployment) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Handling deletion of BossnetDeployment", "deployment", deployment.Name)

	// If finalizer is not present, nothing to do
	if !controllerutil.ContainsFinalizer(deployment, BossnetDeploymentFinalizer) {
		return ctrl.Result{}, nil
	}

	// Only attempt cleanup if we have a deployment ID in Bossnet
	if deployment.Status.Id != nil && *deployment.Status.Id != "" {
		// Create Bossnet client for cleanup
		bossnetClient := r.BossnetClient
		if bossnetClient == nil {
			var err error
			bossnetClient, err = bossnet.NewClientFromK8s(ctx, &deployment.Spec.Server, r.Client, deployment.Namespace, log)
			if err != nil {
				log.Error(err, "Failed to create Bossnet client for deletion", "deployment", deployment.Name)
				// Continue with finalizer removal even if client creation fails
				// to avoid blocking deletion indefinitely
			} else {
				// Attempt to delete from Bossnet API
				if err := bossnetClient.DeleteDeployment(ctx, *deployment.Status.Id); err != nil {
					log.Error(err, "Failed to delete deployment from Bossnet API", "deployment", deployment.Name, "bossnetId", *deployment.Status.Id)
					// Continue with finalizer removal even if Bossnet deletion fails
					// to avoid blocking Kubernetes deletion indefinitely
				} else {
					log.Info("Successfully deleted deployment from Bossnet API", "deployment", deployment.Name, "bossnetId", *deployment.Status.Id)
				}
			}
		}
	}

	// Remove finalizer to allow Kubernetes to complete deletion
	controllerutil.RemoveFinalizer(deployment, BossnetDeploymentFinalizer)
	if err := r.Update(ctx, deployment); err != nil {
		log.Error(err, "Failed to remove finalizer", "deployment", deployment.Name)
		return ctrl.Result{}, err
	}

	log.Info("Finalizer removed, deletion will proceed", "deployment", deployment.Name)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BossnetDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bossnetiov1.BossnetDeployment{}).
		Complete(r)
}
