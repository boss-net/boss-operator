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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("BossnetServer type", func() {
	It("can be deep copied", func() {
		original := &BossnetServer{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
			},
			Spec: BossnetServerSpec{
				Version: ptr.To("0.0.1"),
				Image:   ptr.To("bossnethq/bossnet:0.0.1"),
				SQLite: &SQLiteConfiguration{
					StorageClassName: "standard",
					Size:             resource.MustParse("1Gi"),
				},
			},
		}

		copied := original.DeepCopy()

		Expect(copied).To(Equal(original))
		Expect(copied).NotTo(BeIdenticalTo(original))
	})

	Context("Server environment variables", func() {
		It("should generate correct environment variables for PostgreSQL with direct values", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Postgres: &PostgresConfiguration{
						Host:     ptr.To("postgres.example.com"),
						Port:     ptr.To(5432),
						User:     ptr.To("bossnet"),
						Password: ptr.To("secret123"),
						Database: ptr.To("bossnet"),
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "postgresql+asyncpg"},
				{Name: "BOSSNET_API_DATABASE_HOST", Value: "postgres.example.com"},
				{Name: "BOSSNET_API_DATABASE_PORT", Value: "5432"},
				{Name: "BOSSNET_API_DATABASE_USER", Value: "bossnet"},
				{Name: "BOSSNET_API_DATABASE_PASSWORD", Value: "secret123"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "bossnet"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "False"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should generate correct environment variables for PostgreSQL with environment variable sources", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Postgres: &PostgresConfiguration{
						HostFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-config"},
								Key:                  "host",
							},
						},
						PortFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-config"},
								Key:                  "port",
							},
						},
						UserFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-secret"},
								Key:                  "username",
							},
						},
						PasswordFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-secret"},
								Key:                  "password",
							},
						},
						DatabaseFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-config"},
								Key:                  "database",
							},
						},
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "postgresql+asyncpg"},
				{
					Name: "BOSSNET_API_DATABASE_HOST",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-config"},
							Key:                  "host",
						},
					},
				},
				{
					Name: "BOSSNET_API_DATABASE_PORT",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-config"},
							Key:                  "port",
						},
					},
				},
				{
					Name: "BOSSNET_API_DATABASE_USER",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-secret"},
							Key:                  "username",
						},
					},
				},
				{
					Name: "BOSSNET_API_DATABASE_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-secret"},
							Key:                  "password",
						},
					},
				},
				{
					Name: "BOSSNET_API_DATABASE_NAME",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "postgres-config"},
							Key:                  "database",
						},
					},
				},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "False"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should generate correct environment variables for SQLite", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					SQLite: &SQLiteConfiguration{
						StorageClassName: "standard",
						Size:             resource.MustParse("1Gi"),
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "sqlite+aiosqlite"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "/var/lib/bossnet/bossnet.db"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "True"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should generate correct environment variables for ephemeral storage", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Ephemeral: &EphemeralConfiguration{},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "sqlite+aiosqlite"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "/var/lib/bossnet/bossnet.db"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "True"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should include additional settings in environment variables", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Ephemeral: &EphemeralConfiguration{},
					Settings: []corev1.EnvVar{
						{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
						{Name: "BOSSNET_ANOTHER_SETTING", Value: "another-value"},
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "sqlite+aiosqlite"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "/var/lib/bossnet/bossnet.db"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "True"},
				{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
				{Name: "BOSSNET_ANOTHER_SETTING", Value: "another-value"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should combine database and Redis configuration", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Postgres: &PostgresConfiguration{
						Host:     ptr.To("postgres.example.com"),
						Port:     ptr.To(5432),
						User:     ptr.To("bossnet"),
						Password: ptr.To("secret123"),
						Database: ptr.To("bossnet"),
					},
					Redis: &RedisConfiguration{
						Host:     ptr.To("redis.example.com"),
						Port:     ptr.To(6379),
						Database: ptr.To(0),
					},
					Settings: []corev1.EnvVar{
						{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				// Postgres vars
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "postgresql+asyncpg"},
				{Name: "BOSSNET_API_DATABASE_HOST", Value: "postgres.example.com"},
				{Name: "BOSSNET_API_DATABASE_PORT", Value: "5432"},
				{Name: "BOSSNET_API_DATABASE_USER", Value: "bossnet"},
				{Name: "BOSSNET_API_DATABASE_PASSWORD", Value: "secret123"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "bossnet"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "False"},
				// Redis vars
				{Name: "BOSSNET_MESSAGING_BROKER", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_MESSAGING_CACHE", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_REDIS_MESSAGING_HOST", Value: "redis.example.com"},
				{Name: "BOSSNET_REDIS_MESSAGING_PORT", Value: "6379"},
				{Name: "BOSSNET_REDIS_MESSAGING_DB", Value: "0"},
				// Extra settings
				{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})
	})

	Context("Redis configuration", func() {
		It("should generate correct environment variables with direct values", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Redis: &RedisConfiguration{
						Host:     ptr.To("redis.example.com"),
						Port:     ptr.To(6379),
						Database: ptr.To(0),
						Username: ptr.To("bossnet"),
						Password: ptr.To("secret123"),
					},
				},
			}

			envVars := server.Spec.Redis.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_MESSAGING_BROKER", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_MESSAGING_CACHE", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_REDIS_MESSAGING_HOST", Value: "redis.example.com"},
				{Name: "BOSSNET_REDIS_MESSAGING_PORT", Value: "6379"},
				{Name: "BOSSNET_REDIS_MESSAGING_DB", Value: "0"},
				{Name: "BOSSNET_REDIS_MESSAGING_USERNAME", Value: "bossnet"},
				{Name: "BOSSNET_REDIS_MESSAGING_PASSWORD", Value: "secret123"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should generate correct environment variables with environment variable sources", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Redis: &RedisConfiguration{
						HostFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config"},
								Key:                  "host",
							},
						},
						PortFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config"},
								Key:                  "port",
							},
						},
						DatabaseFrom: &corev1.EnvVarSource{
							ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config"},
								Key:                  "database",
							},
						},
						UsernameFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "redis-secret"},
								Key:                  "username",
							},
						},
						PasswordFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{Name: "redis-secret"},
								Key:                  "password",
							},
						},
					},
				},
			}

			envVars := server.Spec.Redis.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_MESSAGING_BROKER", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_MESSAGING_CACHE", Value: "bossnet_redis.messaging"},
				{
					Name: "BOSSNET_REDIS_MESSAGING_HOST",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config"},
							Key:                  "host",
						},
					},
				},
				{
					Name: "BOSSNET_REDIS_MESSAGING_PORT",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config"},
							Key:                  "port",
						},
					},
				},
				{
					Name: "BOSSNET_REDIS_MESSAGING_DB",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "redis-config"},
							Key:                  "database",
						},
					},
				},
				{
					Name: "BOSSNET_REDIS_MESSAGING_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "redis-secret"},
							Key:                  "username",
						},
					},
				},
				{
					Name: "BOSSNET_REDIS_MESSAGING_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "redis-secret"},
							Key:                  "password",
						},
					},
				},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should include Redis environment variables in server configuration", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Redis: &RedisConfiguration{
						Host:     ptr.To("redis.example.com"),
						Port:     ptr.To(6379),
						Database: ptr.To(0),
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedRedisEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_MESSAGING_BROKER", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_MESSAGING_CACHE", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_REDIS_MESSAGING_HOST", Value: "redis.example.com"},
				{Name: "BOSSNET_REDIS_MESSAGING_PORT", Value: "6379"},
				{Name: "BOSSNET_REDIS_MESSAGING_DB", Value: "0"},
			}

			for _, expected := range expectedRedisEnvVars {
				Expect(envVars).To(ContainElement(expected))
			}
		})

		It("should handle partial Redis configuration", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Redis: &RedisConfiguration{
						Host: ptr.To("redis.example.com"),
						// Only specifying host, other fields left empty
					},
				},
			}

			envVars := server.Spec.Redis.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_MESSAGING_BROKER", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_MESSAGING_CACHE", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_REDIS_MESSAGING_HOST", Value: "redis.example.com"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})
	})

	Context("Additional settings", func() {
		It("should include settings with direct values", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Ephemeral: &EphemeralConfiguration{},
					Settings: []corev1.EnvVar{
						{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
						{Name: "BOSSNET_ANOTHER_SETTING", Value: "another-value"},
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "sqlite+aiosqlite"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "/var/lib/bossnet/bossnet.db"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "True"},
				{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
				{Name: "BOSSNET_ANOTHER_SETTING", Value: "another-value"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should include settings with environment variable sources", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Ephemeral: &EphemeralConfiguration{},
					Settings: []corev1.EnvVar{
						{
							Name: "BOSSNET_EXTRA_SETTING",
							ValueFrom: &corev1.EnvVarSource{
								ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "bossnet-config"},
									Key:                  "extra-setting",
								},
							},
						},
						{
							Name: "BOSSNET_SECRET_SETTING",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "bossnet-secret"},
									Key:                  "secret-setting",
								},
							},
						},
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "sqlite+aiosqlite"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "/var/lib/bossnet/bossnet.db"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "True"},
				{
					Name: "BOSSNET_EXTRA_SETTING",
					ValueFrom: &corev1.EnvVarSource{
						ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "bossnet-config"},
							Key:                  "extra-setting",
						},
					},
				},
				{
					Name: "BOSSNET_SECRET_SETTING",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "bossnet-secret"},
							Key:                  "secret-setting",
						},
					},
				},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should merge settings with all configurations", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Postgres: &PostgresConfiguration{
						Host:     ptr.To("postgres.example.com"),
						Port:     ptr.To(5432),
						User:     ptr.To("bossnet"),
						Password: ptr.To("secret123"),
						Database: ptr.To("bossnet"),
					},
					Redis: &RedisConfiguration{
						Host:     ptr.To("redis.example.com"),
						Port:     ptr.To(6379),
						Database: ptr.To(0),
					},
					Settings: []corev1.EnvVar{
						{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
						{
							Name: "BOSSNET_SECRET_SETTING",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{Name: "bossnet-secret"},
									Key:                  "secret-setting",
								},
							},
						},
					},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				// Postgres vars
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "postgresql+asyncpg"},
				{Name: "BOSSNET_API_DATABASE_HOST", Value: "postgres.example.com"},
				{Name: "BOSSNET_API_DATABASE_PORT", Value: "5432"},
				{Name: "BOSSNET_API_DATABASE_USER", Value: "bossnet"},
				{Name: "BOSSNET_API_DATABASE_PASSWORD", Value: "secret123"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "bossnet"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "False"},
				// Redis vars
				{Name: "BOSSNET_MESSAGING_BROKER", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_MESSAGING_CACHE", Value: "bossnet_redis.messaging"},
				{Name: "BOSSNET_REDIS_MESSAGING_HOST", Value: "redis.example.com"},
				{Name: "BOSSNET_REDIS_MESSAGING_PORT", Value: "6379"},
				{Name: "BOSSNET_REDIS_MESSAGING_DB", Value: "0"},
				// Extra settings
				{Name: "BOSSNET_EXTRA_SETTING", Value: "extra-value"},
				{
					Name: "BOSSNET_SECRET_SETTING",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "bossnet-secret"},
							Key:                  "secret-setting",
						},
					},
				},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})

		It("should handle empty settings", func() {
			server := &BossnetServer{
				Spec: BossnetServerSpec{
					Ephemeral: &EphemeralConfiguration{},
					Settings:  []corev1.EnvVar{},
				},
			}

			envVars := server.ToEnvVars()

			expectedEnvVars := []corev1.EnvVar{
				{Name: "BOSSNET_HOME", Value: "/var/lib/bossnet/"},
				{Name: "BOSSNET_API_DATABASE_DRIVER", Value: "sqlite+aiosqlite"},
				{Name: "BOSSNET_API_DATABASE_NAME", Value: "/var/lib/bossnet/bossnet.db"},
				{Name: "BOSSNET_API_DATABASE_MIGRATE_ON_START", Value: "True"},
			}

			Expect(envVars).To(ConsistOf(expectedEnvVars))
		})
	})
})
