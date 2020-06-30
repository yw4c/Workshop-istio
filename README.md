# Quick Start
1. deploy micro services
    ````
        kubectl apply -f deployment/microservices -n ${NAMESPACE}
    ````
1. deploy ordering istio CRDs, 這裡的 CRDs 有順序性, 勿使用 apply -f 路徑， 若順序有異動要全砍重跑
    ````
        make reload-ordering
    ````
1. test accesslog-filter
    ````
        // get
        curl -HHost:"ares.workshop.com" -H Authorization:1234 http://10.20.0.164:31380/private/api/auth-info\?foo\=bar -L -v
        // post
        curl -HHost:"ares.workshop.com" -H Authorization:1234 -X POST --data "foo=bar" http://10.20.0.164:31380/private/api/auth-info 
    ````

# Workshop - istio
## Required
* gcloud, kubectl 
````
gcloud auth login
gcloud config set project silkrode-golang
gcloud container clusters get-credentials istio --region us-central1-c
kubectl get node
````
* istioctl
````
curl https://storage.googleapis.com/istio-build/dev/latest | xargs -I {} curl https://storage.googleapis.com/istio-build/dev/\{\}/istioctl-\{\}-osx.tar.gz | tar xvz
./istioctl version
cp ./istioctl /usr/local/bin

````

## Cluster
* Istio 1.7 alpha

## Clone me
````
git clone https://gitlab.silkrode.com.tw/team_golang/workshop-istio.git &&
cd workshop-istio &&
git submodule update --init --recursive &&
git submodule foreach git pull origin master &&
git submodule foreach git checkout master 
````


## 部署微服務
* 為你的專案命名 namespace
````
export NAMESPACE=<your name>
````

* 部署 pingpong 服務
````
kubectl create namespace ${NAMESPACE} &&
kubectl apply -f deployment/workshop.yaml -n ${NAMESPACE}
````

## 觀察下 grpc 流向

1. 透果 proxy 從本地訪問 ws001-api
    ````
      nohup kubectl port-forward service/ws001-api 8080:7001 -n ${NAMESPACE} &
    ````

1. 查看 ws002-pingpong 每個 pod log, 看看是否被調用

1. 監控 pod 1
    ````
     kubectl logs -f $(kubectl get pods --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}' -n ${NAMESPACE} -l \
        app=ws002-pingpong | sed -n 1p) -c ws002-pingpong --tail=10 -n ${NAMESPACE}
    ````

1. 打打看 > curl http://localhost:8080/api/pingpong

1. 監控 pod 2 
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

開始注入
````
    # 新產生的 pod 會自動套用
    kubectl label namespace $NAMESPACE  istio-injection=enabled

    # 直接套用在現有的 pod
    kubectl apply -f <(istioctl kube-inject -f deployment/workshop.yaml) -n ${NAMESPACE}
````

* 回上一步試看看，使否是已預設輪循方式分流

## Gateway 路由管理
1. 在 deployment/gateway.yaml 更換專屬你的 domain
1. 部署 gateway 和路由規則
    ````
        kubectl apply -f deployment/gateway.yaml -n ${NAMESPACE}
    ````
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
kubectl apply -f deployment/fault-inject.yaml -n ${NAMESPACE}  
# 訪問看看注入效果
curl -HHost:<你的專屬 domain> http://$INGRESS_HOST/api/pingpong
kubectl delete -f deployment/fault-inject.yaml -n ${NAMESPACE}  
````

## 觀察指標(metrics)
* 为了监控服务行为，Istio 为服务网格中所有出入的服务流量都生成了指标。这些指标提供了关于行为的信息，例如总流量数、错误率和请求响应时间。
````
 nohup kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=prometheus -o jsonpath='{.items[0].metadata.name}') 9090:9090 &
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

## 鏈路追蹤
````
    nohup kubectl -n istio-system port-forward $(kubectl -n istio-system get pod -l app=jaeger -o jsonpath='{.items[0].metadata.name}') 15032:16686 &
````

再戳看看
````
curl -HHost:<你的專屬 domain> http://$INGRESS_HOST/api/pingpong
````
前往 http://localhost:15032/ 查看你的 service.namespace


## envoy filter
* ws003 負責用戶 token 驗證, 我們預設 Authorization:1234 就放行
![private request flow](https://www.websequencediagrams.com/cgi-bin/cdraw?lz=dGl0bGUgUHJpdmF0ZSBSZXF1ZXN0IEZsb3cKCktPL0tNIFxuIEZyb250ZW5kLT7lhaXlj6PntrLpl5wgXG4gaXN0aW8taW5ncmVzc2dhdGV3YXk6CgACJC0-SFRUUAA7DyAAPAcgQ1JEOiBbT10gaHR0cDovL2tiYy5iYWNrZW5kLmNvbSBcbltYABQJNjY2ABAFCgAzHy0-U2VydmljZSBQcm94eQBsClZpcnR1YWwAFggAdAkvcHVibGljL2FwaSBcbgAMBgCCJAYvYXBpCgAlKS0-QXV0aCBNaWRkbGV3YXJlAIFnCkVudm95RmlsdGVyAIFvBiBsYWJlbACCVwVkZW50aXR5LXZhbGlkYXRpb246ZW5hYmxlZAoAKCgtPkkAPwcgAIF1BwCBUggAg2sHAIJDBgCBYgcKAB0QACwT6amX6K2JIHRva2VuAB0TAIE_Kk9LIQCBHCtDZXJ0YWluAIE9CQoAAg8tPgCFHRE6Cg&s=napkin)
* 我們部署一 EnvoyFilter 中間層
````
kubectl apply -f deployment/auth-filter.yaml -n ${NAMESPACE}
````

戳看看: 
````
curl -HHost:<你的專屬 domain> -H Authorization:6666  http://$INGRESS_HOST/api/pingpong
// response: {Code: "40300", Msg: "identity deny"}
curl -HHost:<你的專屬 domain> -H Authorization:1234  http://$INGRESS_HOST/api/pingpong
// response {"msg": "we got auth info {user-id:1}"}
````