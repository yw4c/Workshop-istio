# Work shop - istio
## required
* kubectl 
````
gcloud auth login
gcloud config set project silkrode-golang
gcloud container clusters get-credentials istio --region us-central1-c
````
* istioctl
````
export PATH=$PWD/bin:$PATH
istioctl version
````

## Cluster
* Istio 1.4.6 on GKE

## clone me
````
git submodule update --init --recursive \
git submodule foreach git pull origin master \
git submodule foreach git checkout master 
````


## 部署微服務
* 為你的專案命名 namespace
````
export NAMESPACE=<your name>
````

* 部署 pingpong 服務
````
kubectl create namespace ${NAMESPACE} \
kubectl apply -f deployment/workshop.yaml -n ${NAMESPACE}
````

## 觀察下 grpc 流向

* 透果 proxy 從本地訪問 ws001-api
````
  nohup kubectl port-forward service/ws001-api 8080:7001 -n ${NAMESPACE} &
````

* 查看 ws002-pingpong 每個 pod log, 看看是否被調用

* 監控 pod 1
````
 kubectl logs -f $(kubectl get pods --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -n ${NAMESPACE} -l \
    app=ws002-pingpong | sed -n 1p) -c ws002-pingpong --tail=10 -n ${NAMESPACE}
````

* 打打看 > http://localhost:8080/api/pingpong

* 監控 pod 2 
````
 kubectl logs -f $(kubectl get pods --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -n ${NAMESPACE} -l \
    app=ws002-pingpong | sed -n 2p) -c ws002-pingpong --tail=10 -n ${NAMESPACE}
````

* 使用 service, 我們發現 grpc 無法實現負載平衡. 
* kube-proxy 只有在連線建立的當下，才成功做了負載均衡，之後的每一次 RPC 請求，都會利用原本的連線。

## 注入 sidecar
注入條件
* 带有 metadata.labels.app  标签（label） 的 Deployment
* 带有 spec.ports.name  的 Service
````
    # 新產生的 pod 會自動套用
    kubectl label namespace $NAMESPACE  istio-injection=enabled

    # 直接套用在現有的 pod
    kubectl apply -f <(istioctl kube-inject -f deployment/workshop.yaml) -n ${NAMESPACE}
````

## Gateway 路由管理
1. 在 deployment/gateway.yaml 更換專屬你的 domain
1. 取得叢集 istio-gateway 的 host ip
    ````
        export INGRESS_HOST=$(kubectl -n istio-system get service istio-ingressgateway -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
    ````
1. 我們對 host name, prefix path 套用規則。試著請求看看
    ````
        [O] curl -HHost:<你的專屬 domain> http://$INGRESS_HOST/api/pingpong 
        [X] curl   http://$INGRESS_HOST/api/pingpong 
        [X] curl -HHost:<你的專屬 domain> http://$INGRESS_HOST/
     ````
   
## 故障注入
````

````

## 觀察指標(metrics)
* 为了监控服务行为，Istio 为服务网格中所有出入的服务流量都生成了指标。这些指标提供了关于行为的信息，例如总流量数、错误率和请求响应时间。
````
 kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') 9090:9090 &
````

* 觀測 ws002-pingpong 的被請求紀錄與流量，打開 http://127.0.0.1:9090/ ， 在 query 搜尋
````
istio_requests_total{destination_service="ws002-pingpong.< 你的 NAMESPACE >.svc.cluster.local"}

````

* result
````
{
   "connection_security_policy="   "none",
   "destination_app="   "ws002-pingpong",
   "destination_principal="   "unknown",
   "destination_service="   "ws002-pingpong.ares.svc.cluster.local",
   "destination_service_name="   "ws002-pingpong",
   "destination_service_namespace="   "ares",
   "destination_version="   "unknown",
   "destination_workload="   "ws002-pingpong",
   "destination_workload_namespace="   "ares",
   "instance="   "10.12.0.5:42422",
   "job="   "istio-mesh",
   "permissive_response_code="   "none",
   "permissive_response_policyid="   "none",
   "reporter="   "destination",
   "request_protocol="   "grpc",
   "response_code="   "200",
   "response_flags="   "-",
   "source_app="   "ws001-api",
   "source_principal="   "unknown",
   "source_version="   "unknown",
   "source_workload="   "ws001-api",
   "source_workload_namespace="   "ares"
}
````

## 錯誤熔斷



todo: 
retry
load balance advance
how grpc loadbalance 