# 🛠️ bossnet-operator

> ⚙️ Kubernetes Operator for Managing Bossnet Server & WorkPool Resources

[![License](https://img.shields.io/github/license/boss-net/bossnet-operator?color=blue)](./LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.23+-blue.svg)](https://golang.org)
[![Kubernetes](https://img.shields.io/badge/kubernetes-operator-informational?logo=kubernetes)](https://kubernetes.io)
[![Docker](https://img.shields.io/badge/docker-compatible-blue?logo=docker)](https://docker.com)

---

## 📘 Overview

The **Bossnet Operator** is a declarative Kubernetes-native controller that automates deployment and lifecycle management of Bossnet resources inside a Kubernetes cluster.
It uses the [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/) for seamless integration, scaling, and automation.

---

## 📚 Custom Resource Definitions (CRDs)

| 💼 Kind           | 📄 Description                         | 🔗 Docs                      |
| ----------------- | -------------------------------------- | ---------------------------- |
| `BossnetServer`   | Manages instances of Bossnet server    | [View](./BossnetServer.md)   |
| `BossnetWorkPool` | Coordinates worker pool configurations | [View](./BossnetWorkPool.md) |

---

## ⚙️ Getting Started

### 🧰 Prerequisites

| Tool                           | Version     | Install Guide                                        |      |
| ------------------------------ | ----------- | ---------------------------------------------------- | ---- |
| Go                             | `v1.23.0+`  | [go.dev](https://go.dev/doc/install)                 |      |
| Docker                         | `v17.03+`   | [docker.com](https://docs.docker.com/get-docker/)    |      |
| kubectl                        | `v1.11.3+`  | [k8s tools](https://kubernetes.io/docs/tasks/tools/) |      |
| Kubernetes                     | `v1.11.3+`+ | Cluster access                                       |      |
| [`mise`](https://mise.jdx.dev) | ✔️ Required | \`curl [https://mise.run](https://mise.run)          | sh\` |
| `pipx`                         | Python deps | `pip install --user pipx`                            |      |

> ⚠️ Python tools like `yamllint` require `pipx`. See [issue](https://github.com/jdx/mise/issues/2536)

### 📦 Setup Dev Environment

```bash
mise trust
mise install
```

> These are run automatically as part of some `make` targets.

---

## 🚀 Build & Deploy Instructions

### 🔨 Step 1: Build & Push Image

```bash
make docker-build docker-push IMG=<your-registry>/bossnet-operator:<tag>
```

### 📐 Step 2: Install CRDs into Cluster

```bash
make install
```

### 🚢 Step 3: Deploy Operator Manager

```bash
make deploy IMG=<your-registry>/bossnet-operator:<tag>
```

> 🛑 If RBAC errors occur, ensure you're logged in with cluster-admin privileges.

### 🧪 Step 4: Apply Sample Custom Resources

```bash
kubectl apply -k deploy/samples/
```

> ✔️ Samples should have appropriate default values.

---

## 🧼 Uninstallation

| 🔧 Action             | 💻 Command                          |
| --------------------- | ----------------------------------- |
| 🗑️ Delete Sample CRs | `kubectl delete -k deploy/samples/` |
| ❌ Remove CRDs         | `make uninstall`                    |
| ⛔ Remove Controller   | `make undeploy`                     |

---

## 📦 Distribution for End Users

### 🏗️ Build Installer YAML

```bash
make build-installer IMG=<your-registry>/bossnet-operator:<tag>
```

> Generates `dist/install.yaml` using Kustomize.

### 🌐 Remote Install via URL

```bash
kubectl apply -f https://raw.githubusercontent.com/<org>/bossnet-operator/<tag>/dist/install.yaml
```

> ✅ Perfect for public GitHub releases.

---

## 🤝 Contributing

* 🔍 Explore available tasks with:

  ```bash
  make help
  ```

* 📖 Refer to [Kubebuilder Docs](https://book.kubebuilder.io/introduction.html) to extend functionality.

---

## 📄 License

```
Apache License 2.0

Licensed under the Apache License, Version 2.0 (the "License");
http://www.apache.org/licenses/LICENSE-2.0
```

---

## 🧭 Quick Reference

| 🏷️ Label               | 🔧 Command                              |
| ----------------------- | --------------------------------------- |
| 🛠️ Build & Push        | `make docker-build docker-push IMG=...` |
| 📐 Install CRDs         | `make install`                          |
| 🚀 Deploy to Cluster    | `make deploy IMG=...`                   |
| 🧪 Apply Sample CRs     | `kubectl apply -k deploy/samples/`      |
| 🧹 Full Cleanup         | `make undeploy && make uninstall`       |
| 📦 Build Installer YAML | `make build-installer IMG=...`          |