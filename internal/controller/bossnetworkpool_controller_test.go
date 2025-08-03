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
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bossnetiov1 "github.com/boss-net/bossnet-operator/api/v1"
)

var _ = Describe("BossnetWorkPool Controller", func() {
	var (
		ctx             context.Context
		namespace       *corev1.Namespace
		namespaceName   string
		name            types.NamespacedName
		bossnetworkpool *bossnetiov1.BossnetWorkPool
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespaceName = fmt.Sprintf("any-ns-%d", time.Now().UnixNano())

		namespace = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: namespaceName},
		}
		Expect(k8sClient.Create(ctx, namespace)).To(Succeed())
	})

	It("should ignore removed BossnetWorkPools", func() {
		serverList := &bossnetiov1.BossnetWorkPoolList{}
		err := k8sClient.List(ctx, serverList, &client.ListOptions{Namespace: namespaceName})
		Expect(err).NotTo(HaveOccurred())
		Expect(serverList.Items).To(HaveLen(0))

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespaceName,
				Name:      "nonexistant-work-pool",
			},
		})
		Expect(err).NotTo(HaveOccurred())
	})

	It("should allow specifying a full image name", func() {
		bossnetworkpool = &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Image: ptr.To("bossnethq/bossnet:custom-bossnet-image"),
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			}, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Image).To(Equal("bossnethq/bossnet:custom-bossnet-image"))
	})

	It("should allow specifying a Bossnet version", func() {
		bossnetworkpool = &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Version: ptr.To("3.3.3.3.3.3.3.3"),
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			},
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			}, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Image).To(Equal("bossnethq/bossnet:3.3.3.3.3.3.3.3-python3.12"))
	})

	Context("when creating a work pool", func() {
		var deployment *appsv1.Deployment

		BeforeEach(func() {
			name = types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			}

			bossnetworkpool = &bossnetiov1.BossnetWorkPool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespaceName,
					Name:      "example-work-pool",
				},
				Spec: bossnetiov1.BossnetWorkPoolSpec{
					Version: ptr.To("3.0.0"),
					Type:    "kubernetes",
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
					DeploymentLabels: map[string]string{
						"some":    "additional-label",
						"another": "extra-label",
					},
				},
			}
			Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

			controllerReconciler := &BossnetWorkPoolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: name,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Get(ctx, name, bossnetworkpool)).To(Succeed())

			deployment = &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: namespaceName,
					Name:      "example-work-pool",
				}, deployment)
			}).Should(Succeed())
		})

		Describe("the BossnetWorkPool", func() {
			It("should have the DeploymentReconciled condition", func() {
				deploymentReconciled := meta.FindStatusCondition(bossnetworkpool.Status.Conditions, "DeploymentReconciled")
				Expect(deploymentReconciled).NotTo(BeNil())
				Expect(deploymentReconciled.Status).To(Equal(metav1.ConditionTrue))
				Expect(deploymentReconciled.Reason).To(Equal("DeploymentCreated"))
				Expect(deploymentReconciled.Message).To(Equal("Deployment was created"))
			})
		})

		Describe("the Deployment", func() {
			It("should be owned by the BossnetWorkPool", func() {
				Expect(deployment.OwnerReferences).To(ContainElement(
					metav1.OwnerReference{
						APIVersion:         "bossnet.io/v1",
						Kind:               "BossnetWorkPool",
						Name:               "example-work-pool",
						UID:                bossnetworkpool.UID,
						Controller:         ptr.To(true),
						BlockOwnerDeletion: ptr.To(true),
					},
				))
			})

			It("should have appropriate labels", func() {
				Expect(deployment.Spec.Selector.MatchLabels).To(Equal(map[string]string{
					"bossnet.io/worker": "example-work-pool",
					"some":              "additional-label",
					"another":           "extra-label",
				}))
				Expect(deployment.Spec.Template.Labels).To(Equal(map[string]string{
					"bossnet.io/worker": "example-work-pool",
					"some":              "additional-label",
					"another":           "extra-label",
				}))
			})

			It("should have a worker container with the right image and command", func() {
				Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := deployment.Spec.Template.Spec.Containers[0]

				Expect(container.Name).To(Equal("bossnet-worker"))
				Expect(container.Image).To(Equal("bossnethq/bossnet:3.0.0-python3.12-kubernetes"))
				Expect(container.Command).To(BeNil())
				Expect(container.Args).To(Equal([]string{
					"bossnet", "worker", "start",
					"--pool", "example-work-pool", "--type", "kubernetes",
					"--with-healthcheck",
				}))
			})

			It("should have an environment with BOSSNET_HOME set", func() {
				Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := deployment.Spec.Template.Spec.Containers[0]

				Expect(container.Env).To(ContainElements([]corev1.EnvVar{
					{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				}))
			})

			It("should not expose any ports", func() {
				Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := deployment.Spec.Template.Spec.Containers[0]

				Expect(container.Ports).To(BeEmpty())
			})

			It("should have the specified resource requirements", func() {
				Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := deployment.Spec.Template.Spec.Containers[0]

				Expect(container.Resources.Requests).To(Equal(corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				}))
				Expect(container.Resources.Limits).To(Equal(corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				}))
			})

			It("should have the correct startup, readiness, and liveness probes", func() {
				Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
				container := deployment.Spec.Template.Spec.Containers[0]

				Expect(container.Env).To(ContainElement(corev1.EnvVar{
					Name:  "BOSSNET_WORKER_WEBSERVER_PORT",
					Value: "8080",
				}))

				Expect(container.StartupProbe).To(Equal(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/health",
							Port:   intstr.FromInt(8080),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 10,
					PeriodSeconds:       5,
					TimeoutSeconds:      5,
					SuccessThreshold:    1,
					FailureThreshold:    30,
				}))

				Expect(container.ReadinessProbe).To(Equal(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/health",
							Port:   intstr.FromInt(8080),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 10,
					PeriodSeconds:       5,
					TimeoutSeconds:      5,
					SuccessThreshold:    1,
					FailureThreshold:    30,
				}))

				Expect(container.LivenessProbe).To(Equal(&corev1.Probe{
					ProbeHandler: corev1.ProbeHandler{
						HTTPGet: &corev1.HTTPGetAction{
							Path:   "/health",
							Port:   intstr.FromInt(8080),
							Scheme: corev1.URISchemeHTTP,
						},
					},
					InitialDelaySeconds: 120,
					PeriodSeconds:       10,
					TimeoutSeconds:      5,
					SuccessThreshold:    1,
					FailureThreshold:    2,
				}))
			})
		})
	})

	Context("When updating a work pool", func() {
		BeforeEach(func() {
			name = types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool",
			}

			bossnetworkpool = &bossnetiov1.BossnetWorkPool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespaceName,
					Name:      "example-work-pool",
				},
				Spec: bossnetiov1.BossnetWorkPoolSpec{
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("128Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("500m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

			controllerReconciler := &BossnetWorkPoolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Reconcile once to create the work pool
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: name,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(k8sClient.Get(ctx, name, bossnetworkpool)).To(Succeed())

			bossnetworkpool.Spec.Settings = []corev1.EnvVar{
				{Name: "BOSSNET_SOME_SETTING", Value: "some-value"},
			}
			bossnetworkpool.Spec.Resources = corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("200m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("1"),
					corev1.ResourceMemory: resource.MustParse("1Gi"),
				},
			}
			bossnetworkpool.Spec.ExtraContainers = []corev1.Container{
				{
					Name:  "extra-container",
					Image: "extra-image",
				},
			}
			Expect(k8sClient.Update(ctx, bossnetworkpool)).To(Succeed())

			// Reconcile again to update the work pool
			_, err = controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: name,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update the Deployment with the new setting", func() {
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: namespaceName,
					Name:      "example-work-pool",
				}, deployment)
			}).Should(Succeed())

			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(2))
			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container.Env).To(ContainElement(corev1.EnvVar{
				Name:  "BOSSNET_SOME_SETTING",
				Value: "some-value",
			}))
		})

		It("should update the Deployment with new resource requirements", func() {
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: namespaceName,
					Name:      "example-work-pool",
				}, deployment)
			}).Should(Succeed())

			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(2))
			container := deployment.Spec.Template.Spec.Containers[0]
			Expect(container.Resources.Requests).To(Equal(corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			}))
			Expect(container.Resources.Limits).To(Equal(corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			}))
		})

		It("should update the Deployment with the extra container", func() {
			deployment := &appsv1.Deployment{}
			Eventually(func() error {
				return k8sClient.Get(ctx, types.NamespacedName{
					Namespace: namespaceName,
					Name:      "example-work-pool",
				}, deployment)
			}).Should(Succeed())

			Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(2))
			container := deployment.Spec.Template.Spec.Containers[1]
			Expect(container.Name).To(Equal("extra-container"))
		})
	})

	Context("When evaluating changes with a work pool", func() {
		BeforeEach(func() {
			name = types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool-no-changes",
			}

			bossnetworkpool = &bossnetiov1.BossnetWorkPool{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespaceName,
					Name:      "example-work-pool-no-changes",
				},
			}
			Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

			controllerReconciler := &BossnetWorkPoolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			// Reconcile once to create the server
			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: name,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should not change a deployment if nothing has changed", func() {
			before := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool-no-changes",
			}, before)).To(Succeed())

			controllerReconciler := &BossnetWorkPoolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: name,
			})
			Expect(err).NotTo(HaveOccurred())

			after := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Namespace: namespaceName,
				Name:      "example-work-pool-no-changes",
			}, after)).To(Succeed())

			Expect(after.Generation).To(Equal(before.Generation))
			Expect(after).To(Equal(before))
		})
	})

	Context("WorkPool Status Updates", func() {
		var (
			workPool   *bossnetiov1.BossnetWorkPool
			deployment *appsv1.Deployment
			reconciler *BossnetWorkPoolReconciler
			name       types.NamespacedName
		)

		BeforeEach(func() {
			name = types.NamespacedName{
				Namespace: "test-" + uuid.New().String(),
				Name:      "test-workpool",
			}
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name.Namespace}}
			Expect(k8sClient.Create(ctx, ns)).To(Succeed())

			workPool = &bossnetiov1.BossnetWorkPool{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name.Name,
					Namespace: name.Namespace,
				},
				Spec: bossnetiov1.BossnetWorkPoolSpec{
					Workers: 3,
				},
			}
			Expect(k8sClient.Create(ctx, workPool)).To(Succeed())

			reconciler = &BossnetWorkPoolReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: name,
			})
			Expect(err).NotTo(HaveOccurred())

			deployment = &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, name, deployment)).To(Succeed())

			// Update the replicas to 3, which the Deployment controller would do
			deployment.Status.Replicas = 3
			Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())
		})

		It("should have default values initially", func() {
			result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: name})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			updatedWorkPool := &bossnetiov1.BossnetWorkPool{}
			Expect(k8sClient.Get(ctx, name, updatedWorkPool)).To(Succeed())
			Expect(updatedWorkPool.Status.ReadyWorkers).To(Equal(int32(0)))
			Expect(updatedWorkPool.Status.Ready).To(Equal(false))
		})

		It("should update status when becoming ready", func() {
			deployment.Status.ReadyReplicas = 3
			Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())

			result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: name})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			updatedWorkPool := &bossnetiov1.BossnetWorkPool{}
			Expect(k8sClient.Get(ctx, name, updatedWorkPool)).To(Succeed())
			Expect(updatedWorkPool.Status.ReadyWorkers).To(Equal(int32(3)))
			Expect(updatedWorkPool.Status.Ready).To(Equal(true))
		})

		It("should update status when becoming unready", func() {
			deployment.Status.ReadyReplicas = 0
			Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())

			result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: name})
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(reconcile.Result{}))

			updatedWorkPool := &bossnetiov1.BossnetWorkPool{}
			Expect(k8sClient.Get(ctx, name, updatedWorkPool)).To(Succeed())
			Expect(updatedWorkPool.Status.ReadyWorkers).To(Equal(int32(0)))
			Expect(updatedWorkPool.Status.Ready).To(Equal(false))
		})

		It("should toggle status correctly", func() {
			// First, make it ready
			deployment.Status.ReadyReplicas = 3
			Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())
			_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: name})
			Expect(err).NotTo(HaveOccurred())

			updatedWorkPool := &bossnetiov1.BossnetWorkPool{}
			Expect(k8sClient.Get(ctx, name, updatedWorkPool)).To(Succeed())
			Expect(updatedWorkPool.Status.ReadyWorkers).To(Equal(int32(3)))
			Expect(updatedWorkPool.Status.Ready).To(Equal(true))

			// Then, make it unready
			deployment.Status.ReadyReplicas = 0
			Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: name})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, name, updatedWorkPool)).To(Succeed())
			Expect(updatedWorkPool.Status.ReadyWorkers).To(Equal(int32(0)))
			Expect(updatedWorkPool.Status.Ready).To(Equal(false))

			// Finally, make it ready again
			deployment.Status.ReadyReplicas = 2
			Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())
			_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: name})
			Expect(err).NotTo(HaveOccurred())

			Expect(k8sClient.Get(ctx, name, updatedWorkPool)).To(Succeed())
			Expect(updatedWorkPool.Status.ReadyWorkers).To(Equal(int32(2)))
			Expect(updatedWorkPool.Status.Ready).To(Equal(true))
		})
	})

	It("should set BOSSNET_API_URL when provided", func() {
		name := types.NamespacedName{
			Namespace: namespaceName,
			Name:      "example-work-pool-with-api-key",
		}

		bossnetworkpool := &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name.Namespace,
				Name:      name.Name,
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Server: bossnetiov1.BossnetServerReference{
					Name:         "test-server",
					Namespace:    name.Namespace,
					RemoteAPIURL: ptr.To("https://some-server.example.com/api"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, name, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name:  "BOSSNET_API_URL",
			Value: "https://some-server.example.com/api",
		}))
	})

	It("should ensure BOSSNET_API_URL ends with /api when provided", func() {
		name := types.NamespacedName{
			Namespace: namespaceName,
			Name:      "example-work-pool-with-api-key",
		}

		bossnetworkpool := &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name.Namespace,
				Name:      name.Name,
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Server: bossnetiov1.BossnetServerReference{
					Name:         "test-server",
					Namespace:    name.Namespace,
					RemoteAPIURL: ptr.To("https://some-server.example.com"),
				},
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, name, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name:  "BOSSNET_API_URL",
			Value: "https://some-server.example.com/api",
		}))
	})

	It("should set BOSSNET_API_KEY and a remote BOSSNET_API_URL when apiKey.value is provided", func() {
		name := types.NamespacedName{
			Namespace: namespaceName,
			Name:      "example-work-pool-with-api-key",
		}

		bossnetworkpool := &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name.Namespace,
				Name:      name.Name,
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Server: bossnetiov1.BossnetServerReference{
					Name:         "test-server",
					Namespace:    name.Namespace,
					RemoteAPIURL: ptr.To("https://remote.bossnet.cloud/api"),
					APIKey: &bossnetiov1.APIKeySpec{
						Value: ptr.To("test-api-key"),
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, name, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name:  "BOSSNET_API_KEY",
			Value: "test-api-key",
		}))
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name:  "BOSSNET_API_URL",
			Value: "https://remote.bossnet.cloud/api",
		}))
	})

	It("should set BOSSNET_API_KEY with valueFrom and a remote BOSSNET_API_URL when apiKey.valueFrom is provided", func() {
		name := types.NamespacedName{
			Namespace: namespaceName,
			Name:      "example-work-pool-with-api-key-from",
		}

		bossnetworkpool := &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name.Namespace,
				Name:      name.Name,
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Server: bossnetiov1.BossnetServerReference{
					Name:         "test-server",
					Namespace:    name.Namespace,
					RemoteAPIURL: ptr.To("https://remote.bossnet.cloud"),
					APIKey: &bossnetiov1.APIKeySpec{
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "api-key-secret",
								},
								Key: "api-key",
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, name, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name: "BOSSNET_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "api-key-secret",
					},
					Key: "api-key",
				},
			},
		}))
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name:  "BOSSNET_API_URL",
			Value: "https://remote.bossnet.cloud/api",
		}))
	})

	It("should set correct BOSSNET_API_URL with accountID and workspaceID", func() {
		name := types.NamespacedName{
			Namespace: namespaceName,
			Name:      "workpool-with-account-workspace",
		}

		accountID := uuid.New().String()
		workspaceID := uuid.New().String()

		bossnetworkpool := &bossnetiov1.BossnetWorkPool{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: name.Namespace,
				Name:      name.Name,
			},
			Spec: bossnetiov1.BossnetWorkPoolSpec{
				Server: bossnetiov1.BossnetServerReference{
					RemoteAPIURL: ptr.To("https://api.bossnet.cloud"),
					AccountID:    ptr.To(accountID),
					WorkspaceID:  ptr.To(workspaceID),
					APIKey: &bossnetiov1.APIKeySpec{
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "api-key-secret",
								},
								Key: "api-key",
							},
						},
					},
				},
			},
		}
		Expect(k8sClient.Create(ctx, bossnetworkpool)).To(Succeed())

		controllerReconciler := &BossnetWorkPoolReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}

		_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
			NamespacedName: name,
		})
		Expect(err).NotTo(HaveOccurred())

		deployment := &appsv1.Deployment{}
		Eventually(func() error {
			return k8sClient.Get(ctx, name, deployment)
		}).Should(Succeed())

		Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
		container := deployment.Spec.Template.Spec.Containers[0]

		expectedAPIURL := fmt.Sprintf("https://api.bossnet.cloud/api/accounts/%s/workspaces/%s", accountID, workspaceID)
		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name:  "BOSSNET_API_URL",
			Value: expectedAPIURL,
		}))

		Expect(container.Env).To(ContainElement(corev1.EnvVar{
			Name: "BOSSNET_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "api-key-secret",
					},
					Key: "api-key",
				},
			},
		}))
	})
})
