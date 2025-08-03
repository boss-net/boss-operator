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

package v1

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// BossnetWorkPoolSpec defines the desired state of BossnetWorkPool
type BossnetWorkPoolSpec struct {
	// Version defines the version of the Bossnet Server to deploy
	Version *string `json:"version,omitempty"`

	// Image defines the exact image to deploy for the Bossnet Server, overriding Version
	Image *string `json:"image,omitempty"`

	// Server defines which Bossnet Server to connect to
	Server BossnetServerReference `json:"server,omitempty"`

	// The type of the work pool, such as "kubernetes" or "process"
	Type string `json:"type,omitempty"`

	// Workers defines the number of workers to run in the Work Pool
	Workers int32 `json:"workers,omitempty"`

	// Resources defines the CPU and memory resources for each worker in the Work Pool
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// ExtraContainers defines additional containers to add to each worker in the Work Pool
	ExtraContainers []corev1.Container `json:"extraContainers,omitempty"`

	// A list of environment variables to set on the Bossnet Worker
	Settings []corev1.EnvVar `json:"settings,omitempty"`

	// DeploymentLabels defines additional labels to add to the Bossnet Server Deployment
	DeploymentLabels map[string]string `json:"deploymentLabels,omitempty"`
}

// BossnetWorkPoolStatus defines the observed state of BossnetWorkPool
type BossnetWorkPoolStatus struct {
	// Version is the version of the Bossnet Worker that is currently running
	Version string `json:"version"`

	// ReadyWorkers is the number of workers that are currently ready
	ReadyWorkers int32 `json:"readyWorkers"`

	// Ready is true if the work pool is ready to accept work
	Ready bool `json:"ready"`

	// Conditions store the status conditions of the BossnetWorkPool instances
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:path="bossnetworkpools",singular="bossnetworkpool",shortName="pwp",scope="Namespaced"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="The type of this work pool"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version",description="The version of this work pool"
// +kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready",description="Whether the work pool is ready"
// +kubebuilder:printcolumn:name="Desired Workers",type="integer",JSONPath=".spec.workers",description="How many workers are desired"
// +kubebuilder:printcolumn:name="Ready Workers",type="integer",JSONPath=".status.readyWorkers",description="How many workers are ready"
// BossnetWorkPool is the Schema for the bossnetworkpools API
type BossnetWorkPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BossnetWorkPoolSpec   `json:"spec,omitempty"`
	Status BossnetWorkPoolStatus `json:"status,omitempty"`
}

func (s *BossnetWorkPool) WorkerLabels() map[string]string {
	labels := map[string]string{
		"bossnet.io/worker": s.Name,
	}

	for k, v := range s.Spec.DeploymentLabels {
		labels[k] = v
	}

	return labels
}

func (s *BossnetWorkPool) Image() string {
	suffix := ""
	if s.Spec.Type == "kubernetes" {
		suffix = "-kubernetes"
	}

	if s.Spec.Image != nil && *s.Spec.Image != "" {
		return *s.Spec.Image
	}
	if s.Spec.Version != nil && *s.Spec.Version != "" {
		return "bossnethq/bossnet:" + *s.Spec.Version + "-python3.12" + suffix
	}
	return DEFAULT_BOSSNET_IMAGE + suffix
}

func (s *BossnetWorkPool) EntrypointArguments() []string {
	workPoolName := s.Name
	if strings.HasPrefix(workPoolName, "bossnet") {
		workPoolName = "pool-" + workPoolName
	}
	return []string{
		"bossnet", "worker", "start",
		"--pool", workPoolName, "--type", s.Spec.Type,
		"--with-healthcheck",
	}
}

// BossnetAPIURL returns the API URL for the Bossnet Server, either from the RemoteAPIURL or
// from the in-cluster server
func (s *BossnetWorkPool) BossnetAPIURL() string {
	return s.Spec.Server.GetAPIURL(s.Namespace)
}

func (s *BossnetWorkPool) ToEnvVars() []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name:  "BOSSNET_HOME",
			Value: "/var/lib/bossnet/",
		},
		{
			Name:  "BOSSNET_API_URL",
			Value: s.BossnetAPIURL(),
		},
		{
			Name:  "BOSSNET_WORKER_WEBSERVER_PORT",
			Value: "8080",
		},
	}

	// If the API key is specified, add it to the environment variables.
	// If both are set, we favor ValueFrom > Value as it is more secure.
	if s.Spec.Server.APIKey != nil && s.Spec.Server.APIKey.ValueFrom != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:      "BOSSNET_API_KEY",
			ValueFrom: s.Spec.Server.APIKey.ValueFrom,
		})
	} else if s.Spec.Server.APIKey != nil && s.Spec.Server.APIKey.Value != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "BOSSNET_API_KEY",
			Value: *s.Spec.Server.APIKey.Value,
		})
	}

	return envVars
}

func (s *BossnetWorkPool) HealthProbe() corev1.ProbeHandler {
	return corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   "/health",
			Port:   intstr.FromInt(8080),
			Scheme: corev1.URISchemeHTTP,
		},
	}
}

func (s *BossnetWorkPool) StartupProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        s.HealthProbe(),
		InitialDelaySeconds: 10,
		PeriodSeconds:       5,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    30,
	}
}
func (s *BossnetWorkPool) ReadinessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        s.HealthProbe(),
		InitialDelaySeconds: 10,
		PeriodSeconds:       5,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    30,
	}
}
func (s *BossnetWorkPool) LivenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler:        s.HealthProbe(),
		InitialDelaySeconds: 120,
		PeriodSeconds:       10,
		TimeoutSeconds:      5,
		SuccessThreshold:    1,
		FailureThreshold:    2,
	}
}

//+kubebuilder:object:root=true

// BossnetWorkPoolList contains a list of BossnetWorkPool
type BossnetWorkPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BossnetWorkPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BossnetWorkPool{}, &BossnetWorkPoolList{})
}
