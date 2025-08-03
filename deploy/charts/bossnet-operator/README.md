# bossnet-operator

## Installing the Chart

### Prerequisites

1. Add the Bossnet Helm repository to your Helm client:

    ```bash
    helm repo add bossnet-operator https://bossnethq.github.io/bossnet-operator
    helm repo update
    ```

2. Create a new namespace in your Kubernetes cluster to deploy the Bossnet Operator in:

    ```bash
    kubectl create namespace bossnet-system
    ```

### [Optional] Verify the Chart

We use a PGP key to sign the Helm chart as recommended by Helm.

If you would like to verify the charts signature before installing it, you can do so by following the instructions below:

```bash
# Pull Bossnets public PGP key from keybase
curl https://keybase.io/bossnethq/pgp_keys.asc | gpg --dearmor > .gnupg/pubring.gpg
# Run `helm fetch` to validate the chart
helm fetch --verify bossnet-operator/bossnet-operator --version 2024.9.15203739 --keyring .gnupg/pubring.gpg
```

### Install the Chart

1. Install the Bossnet Operator using Helm

    ```bash
    helm install bossnet-operator bossnet-operator/bossnet-operator --namespace=bossnet-system -f values.yaml
    ```

2. Verify the deployment

    Check the status of your Bossnet Operatr deployment:

    ```bash
    kubectl get pods -n bossnet-system

    NAME                                READY   STATUS    RESTARTS       AGE
    bossnet-operator-69874bdc54-lc9vk   2/2     Running   0              25m
    ```

    You should see the Bossnet Operator pod running

## Uninstalling the Chart

To uninstall/delete the bossnet-operator deployment:

```bash
helm delete bossnet-operator -n bossnet-system
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## FAQ
tbd

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| jamiezieziula | <jamie@bossnet.io> |  |
| jimid27 | <jimi@bossnet.io> |  |
| parkedwards | <edward@bossnet.io> |  |
| mitchnielsen | <mitchell@bossnet.io> |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | common | 2.31.3 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| commonAnnotations | object | `{}` | annotations to add to all deployed objects |
| commonLabels | object | `{"app.kubernetes.io/component":"operator"}` | labels to add to all deployed objects |
| fullnameOverride | string | `"bossnet-operator"` | fully override common.names.fullname |
| nameOverride | string | `""` | partially overrides common.names.name |
| namespaceOverride | string | `""` | fully override common.names.namespace |
| operator.affinity | object | `{}` | affinity for operator pods assignment |
| operator.containerSecurityContext.allowPrivilegeEscalation | bool | `false` | set operator containers' security context allowPrivilegeEscalation |
| operator.containerSecurityContext.capabilities | object | `{"drop":["ALL"]}` | set operator container's security context capabilities |
| operator.extraEnvVars | list | `[]` | array with environment variables to add to operator container |
| operator.image.pullPolicy | string | `"IfNotPresent"` | operator image pull policy |
| operator.image.pullSecrets | list | `[]` | operator image pull secrets |
| operator.image.repository | string | `"boss-net/bossnet-operator"` | operator image repository |
| operator.image.tag | string | `"latest"` | operator image tag (immutable tags are recommended) |
| operator.livenessProbe.config.initialDelaySeconds | int | `15` | The number of seconds to wait before starting the first probe. |
| operator.livenessProbe.config.periodSeconds | int | `20` | The number of seconds to wait between consecutive probes. |
| operator.livenessProbe.enabled | bool | `true` |  |
| operator.nodeSelector | object | `{}` | node labels for operator pods assignment |
| operator.podAnnotations | object | `{}` | extra annotations for operator pod |
| operator.podLabels | object | `{}` | extra labels for operator pod |
| operator.podSecurityContext.runAsNonRoot | bool | `true` | set operator pod's security context runAsNonRoot |
| operator.priorityClassName | string | `""` | priority class name to use for the operator pods; if the priority class is empty or doesn't exist, the operator pods are scheduled without a priority class |
| operator.readinessProbe.config.initialDelaySeconds | int | `5` | The number of seconds to wait before starting the first probe. |
| operator.readinessProbe.config.periodSeconds | int | `10` | The number of seconds to wait between consecutive probes. |
| operator.readinessProbe.enabled | bool | `true` |  |
| operator.replicaCount | int | `1` | number of operator replicas to deploy |
| operator.resources.limits | object | `{"cpu":"500m","memory":"128Mi"}` | the requested limits for the operator container |
| operator.resources.requests | object | `{"cpu":"10m","memory":"64Mi"}` | the requested resources for the operator container |
| operator.terminationGracePeriodSeconds | int | `10` | seconds operator pod needs to terminate gracefully |
| operator.tolerations | list | `[]` | tolerations for operator pods assignment |
| operator.topologySpreadConstraints | list | `[]` | topology spread constraints for operator pod assignment spread across your cluster among failure-domains |
| rbac.operator.create | bool | `true` | specifies whether the operator role & role binding should be created |
| rbac.userRoles.bossnetServer.editor.create | bool | `true` | specifies whether the server editor role should be created |
| rbac.userRoles.bossnetServer.viewer.create | bool | `true` | specifies whether the server viewer role should be created |
| rbac.userRoles.bossnetWorkpool.editor.create | bool | `true` | specifies whether the workpool editor role should be created |
| rbac.userRoles.bossnetWorkpool.viewer.create | bool | `true` | specifies whether the workpool viewer role should be created |
| serviceAccount.annotations | object | `{}` | additional service account annotations (evaluated as a template) |
| serviceAccount.create | bool | `true` | specifies whether a ServiceAccount should be created |
| serviceAccount.name | string | `""` | the name of the ServiceAccount to use. if not set and create is true, a name is generated using the common.names.fullname template |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
