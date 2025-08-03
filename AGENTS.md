# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with the codebase in this repository.

---

## 🛠 Development Commands

### Prerequisites Setup

```bash
mise trust
mise install     # Installs all dependencies: Go, Docker, kubectl, etc.
```

### Core Development Workflow

```bash
make help        # Show all available make targets
make build       # Build the operator manager binary
make run         # Run the operator controller locally
make test        # Run unit tests with coverage reporting
make test-e2e    # Run end-to-end tests (requires Kubernetes cluster)
make watch       # Watch for changes and run tests continuously
make lint        # Run golangci-lint and yamllint for code quality
make lint-fix    # Run linters and apply automatic fixes where possible
```

> After building once, individual tests can also be run directly using `ginkgo`. See the `Makefile` for details.

---

### Code Generation

```bash
make manifests   # Generate CRDs, webhooks, ClusterRoles manifests
make generate    # Generate DeepCopy methods and other boilerplate code
make docs        # Generate CRD documentation files (BossnetServer.md, BossnetWorkPool.md)
```

---

### Kubernetes Operations

```bash
make install     # Install CRDs into your Kubernetes cluster
make uninstall   # Uninstall CRDs from the cluster
make deploy      # Deploy the operator using Helm
make undeploy    # Remove the operator deployment
kubectl apply -k deploy/samples/  # Apply sample CR instances
```

---

### Docker Operations

```bash
make docker-build IMG=<registry>/bossnet-operator:tag   # Build operator container image
make docker-push IMG=<registry>/bossnet-operator:tag    # Push container image to registry
```

---

## 🏗️ Architecture Overview

This project is a Kubernetes operator built with [Kubebuilder](https://book.kubebuilder.io) that manages Bossnet infrastructure components.

### Core Custom Resources (CRDs)

* **BossnetServer**
  Manages Bossnet server deployments with configurable storage backends:

  * Ephemeral (in-memory)
  * SQLite (persistent PVC)
  * PostgreSQL (external DB)
    Also supports optional Redis messaging.

* **BossnetWorkPool**
  Manages pools of workers for different execution environments such as Kubernetes pods, local processes, or external hosts.

* **BossnetDeployment**
  Manages deployments of Python flow executions (distinct from Kubernetes deployments).

---

### Key Components

* **Controllers** (`internal/controller/`):
  Reconciliation logic for CRDs and syncing actual state with desired state.

* **API Types** (`api/v1/`):
  Custom resource definitions and helper functions.

* **Utils** (`internal/utils/`):
  Hash utilities and other helpers for detecting resource changes.

* **Conditions** (`internal/conditions/`):
  Standardized status condition management.

* **Constants** (`internal/constants/`):
  Shared constant definitions across the project.

---

### Deployment Architecture

* The operator runs inside the cluster as a Deployment.
* Uses [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) for Kubernetes API interactions.
* Supports **namespace-scoped** operation via `WATCH_NAMESPACES` environment variable.
* Provides health checks on port `8081` and metrics on port `8080`.
* Supports leader election for high availability in multi-instance setups.

---

### Storage Backend Configuration for BossnetServer

Supported storage backends:

1. **Ephemeral**: In-memory SQLite database (non-persistent).
2. **SQLite**: Persistent SQLite database backed by Kubernetes PersistentVolumeClaim.
3. **PostgreSQL**: External PostgreSQL database with configurable connection parameters.

Each backend configures environment variables for the BossnetServer container accordingly.

---

### Testing Strategy

* **Unit tests:**
  Written with Ginkgo/Gomega; use envtest to simulate Kubernetes API.

* **E2E tests:**
  Require a real Kubernetes cluster (e.g., `minikube`, `kind`).

* **Test coverage:**
  Covers API types, controllers, utilities, and other core packages.

> Typically, tests are run against Kubernetes v1.29.0 via envtest. For local development, running the operator locally against a `minikube` or remote cluster is common, using port-forwarding when needed.

---

### Code Generation Workflow

This operator relies on [controller-gen](https://github.com/kubernetes-sigs/controller-tools) to generate:

* CRD manifests from Go API types → `deploy/charts/bossnet-operator/crds/`
* DeepCopy method implementations for API types
* RBAC manifests for operator permissions

> **Important:** Always run `make manifests generate` after editing or adding API types in `api/v1/` to keep generated code and manifests up to date.