# Work shop - istio
## required
    * kubectl 

## 部署服務
````
export NAMESPACE=<your name>
make deploy
````

## 觀察下 grpc 流向

````
 # 找出 pingpong 的 pod 
 kubectl get pod -l app=ws002-pingpong -n ${NAMESPACE}
 # 監控 pod 1 
 kubectl logs -f <pod-name1> --all-containers=true --tail=10 -n  ${NAMESPACE}
 # 監控 pod 2 
 kubectl logs -f <pod-name2> --all-containers=true --tail=10 -n  ${NAMESPACE}

 # 打打看 > http://localhost:8080/api/pingpong
  kubectl port-forward service/ws001-api 8080:7001 -n ${NAMESPACE}
````
* 使用 service, 我們發現 grpc 無法實現負載平衡. 
* kube-proxy 只有在連線建立的當下，才成功做了負載均衡，之後的每一次 RPC 請求，都會利用原本的連線。

## 注入 sidecar
注入條件
* 带有 metadata.labels.app  标签（label） 的 Deployment
* 带有 metadata.labels.app  标签（label） 的 Deployment
````
    # 新產生的 pod 會自動套用
    kubectl label namespace $NAMESPACE  istio-injection=enabled

    # 直接套用在現有的 pod
    kubectl apply -f <(istioctl kube-inject -f deployment/workshop.yaml) -n ${NAMESPACE}
````

