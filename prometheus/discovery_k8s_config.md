# prometheus监控外部kubernetes集群配置

作为监控系统，`prometheus`首先要解决的就是"`要监控谁`"的问题。静态配置比较适合`targets`基本稳定不变的场景，云原生体系下所有的被监控target都在动态的变化，如果使用静态配置就相形见绌了(*注：在[001.prometheus简述](001.prometheus简述.md)章节已经做了简述*）。本章主要讨论**服务发现**。

最新版本的`prometheus`(`v2.53`) 支持39种服务发现协议。配置关键字如下：  
- azure_sd_config
- consul_sd_config
- digitalocean_sd_config
- docker_sd_config
- dockerswarm_sd_config
- dns_sd_config
- ec2_sd_config
- openstack_sd_config
- ovhcloud_sd_config
- puppetdb_sd_config
- file_sd_config
- gce_sd_config
- hetzner_sd_config
- http_sd_config
- ionos_sd_config
- **kubernetes_sd_config**
- kuma_sd_config
- lightsail_sd_config
- linode_sd_config
- marathon_sd_config
- nerve_sd_config
- nomad_sd_config
- serverset_sd_config
- triton_sd_config
- eureka_sd_config
- scaleway_sd_config
- uyuni_sd_config
- vultr_sd_config
- **static_config**

由上述服务发现配置关键字，可以发现：
- 服务发现的配置都是以 `xxxx_sd_config`。因为`_sd_config`后缀是在代码中`hard-code`写死的，所以二开新的服务发现协议，需要遵守此命名规则。
- **static_config** 是一种特殊的服务发现。可以理解为target始终不变的服务发现。也是最简单的服务发现形式  

<br/>
服务发现的配置详见: https://prometheus.io/docs/prometheus/2.53/configuration/configuration/  

<br/>  

本章分别介绍 最简单的**static_config**和最常用的**kubernetes_sd_config** 服务发现。其他的服务发现实现逻辑是相同的，不再讲述。  


# prometheus监控kubernetes配置

静态配置不再赘述，可见[prometheus简述:静态文件配置](./prometheus简述.md#静态文件配置)

为了方便调试`prometheus`代码，`prometheus`在本地运行、debug，监控`kubernetes`集群。 针对`kubernetes`集群外部部署(或运行的)`prometheus` 首先需要创建`token` ,这样`prometheus`能够访问`kubernetes`集群的`apiserver`获取监控数据。

## 步骤1：创建`token`,即创建`prometheus`访问权限

### 1.1 在`kubernetes`集群`master`节点上创建文件 [`rbac-setup.yaml`](https://github.com/prometheus/prometheus/blob/v2.53.0/documentation/examples/rbac-setup.yml) 
  
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: prometheus
rules:
  - apiGroups: [""]
    resources:
      - nodes
      - nodes/metrics
      - services
      - endpoints
      - pods
    verbs: ["get", "list", "watch"]
  - apiGroups:
      - extensions
      - networking.k8s.io
    resources:
      - ingresses
    verbs: ["get", "list", "watch"]
  - nonResourceURLs: ["/metrics", "/metrics/cadvisor"]
    verbs: ["get"]
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: prometheus
  namespace: default
---
apiVersion: v1
kind: Secret
metadata:
  name: prometheus-sa-token
  namespace: default
  annotations:
    kubernetes.io/service-account.name: prometheus
type: kubernetes.io/service-account-token
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: prometheus
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: prometheus
subjects:
  - kind: ServiceAccount
    name: prometheus
    namespace: default

```

### 1.2 `kubernetes`集群`master`节点上执行`kubectl  apply  -f  rbac-setup.yaml`  
  
```shell
$ kubectl  apply  -f rbac-setup.yaml

clusterrole.rbac.authorization.k8s.io/prometheus created
serviceaccount/prometheus created
secret/prometheus-sa-token created
clusterrolebinding.rbac.authorization.k8s.io/prometheus created
```

## 步骤2: 获取`token`

[步骤1](#步骤1创建token即创建prometheus访问权限)创建`rbac`时，是在`namespace:default`下进行的。作用可以直接省略 `-n default` 参数

### 2.1  获取`token`名称 

`kubernetes`集群`master`节点上执行`kubectl get secret`

```shell
$  kubectl get secret

NAME                  TYPE                                  DATA   AGE
prometheus-sa-token   kubernetes.io/service-account-token   3      46s
```

`prometheus-sa-token` 就是刚才创建的`token`名称。
<br>

### 2.2 获取`token`内容

 `kubernetes`集群`master`节点上,执行`kubectl  describe secrets  prometheus-sa-token` 
  
```shell
$ kubectl  describe secrets  prometheus-sa-token

Name:         prometheus-sa-token
Namespace:    default
Labels:       <none>
Annotations:  kubernetes.io/service-account.name: prometheus
              kubernetes.io/service-account.uid: 516bf408-a6ba-4ec3-b242-9e46118951a8

Type:  kubernetes.io/service-account-token

Data
====
ca.crt:     1107 bytes
namespace:  7 bytes
token:      eyJhbGciOiJSUzI1NiIsImtpZCI6InhJaVZJS1g1aTgtZ2JGY0Z3dWNoR0FhV3hOMzVIX0J6NXdCY3RVbWM4MDgifQ.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6InByb21ldGhldXMtc2EtdG9rZW4iLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC5uYW1lIjoicHJvbWV0aGV1cyIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50LnVpZCI6IjUxNmJmNDA4LWE2YmEtNGVjMy1iMjQyLTllNDYxMTg5NTFhOCIsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDpkZWZhdWx0OnByb21ldGhldXMifQ.eqNYzgSTUPqabbO-mEKHPMOGVkgLcmHFcuYgHjPS5nwFnf_1TBOJ_9roSGfs9RHE2JXPLj3t4e0lcoMRnX_i32oEbI2qSOoQ6L-2sZ2MGmYsWmSE6WtYyTFRxsIFfLNNcKxPwKgwXAxdo5QNpxDQ4VUjaMBdqeth2Z1uGXhN1tf295rH9e-DaoZbgY78_gh4GWwYvMv5F7gEP6O6a5oczbEApLwtPnunrZdQ2YeyNQYsSsfQSBko1iIdFm0TEXgZi2-Zp17Wz9UE8x0HYjwvR95P-rvCAh1x3WaTux6Mddm8xO9QtYBhht_gdWElWzkSQFY1yUYm0ts6PYRBOpmW7w
```
<br>

### 2.3 将`token`保存本地文件内 

将`token`保存本地文件内，文件名字任意。作者将其保存在prometheus项目文件中`documentation/examples/k8s.token`(绝对路径:`/Users/tyltr/opencode/prometheus/documentation/examples/k8s.token`)

## 步骤3：获取`kubernetes`集群的`api server`的地址  

在`kubernetes`集群`master`节点上,执行`cat ~/.kube/config | grep server` 

```shell
$ cat ~/.kube/config | grep server

server: https://192.168.0.107:6443
```

注：`https://192.168.0.107:6443` 就是 `kubernetes`集群的`api server`的地址


## 步骤4: `prometheus`配置文件

`prometheus`配置文件,作者命名为`k8s-prometheus.yaml`,内容如下：

```yaml
global:
  keep_dropped_targets: 100

scrape_configs:
  - job_name: "kubernetes-apiservers"
    scheme: https
    kubernetes_sd_configs:
      - api_server: https://192.168.0.107:6443
        role: endpoints
        namespaces:
          names: ["default"]
        bearer_token_file: /Users/ollie/opencode/prometheus/documentation/examples/k8s.token
        tls_config:
          insecure_skip_verify: true
    bearer_token_file:   /Users/ollie/opencode/prometheus/documentation/examples/k8s.token
    tls_config:
      insecure_skip_verify: true
    relabel_configs:
      - source_labels:
          [
            __meta_kubernetes_namespace,
            __meta_kubernetes_service_name,
            __meta_kubernetes_endpoint_port_name,
          ]
        action: keep
        regex: default;kubernetes;https

# node
  - job_name: "kubernetes-nodes"
    scheme: https
    kubernetes_sd_configs:
      - api_server: https://192.168.0.107:6443
        role: node
        namespaces:
          names: ["default"]
        bearer_token_file: /Users/ollie/opencode/prometheus/documentation/examples/k8s.token
        tls_config:
          insecure_skip_verify: true
    bearer_token_file:   /Users/ollie/opencode/prometheus/documentation/examples/k8s.token
    tls_config:
      insecure_skip_verify: true
    relabel_configs:
      - action: labelmap
        regex: __meta_kubernetes_node_label_(.+)

```

注：`bearer_token_file` 为 [2.3`token`存放的目录](#23-将token保存本地文件内)

### 步骤5：运行

```shell
./prometheus  --config.file=/your/path/k8s-prometheus.yml
```

即可在`web ui`上观察到`target`列表  

![k8s-demo](src/k8s-prometheus-demo.png)
