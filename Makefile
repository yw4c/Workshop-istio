NAMESPACE ?= workshop
deploy:
	kubectl create namespace ${NAMESPACE}
	kubectl apply -f deployment/workshop.yaml -n ${NAMESPACE}

gen_pb:
	# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	# Generate gRPC stub (.pb.go)
	protoc -I ./ --go_out=plugins=grpc:. ./pb/pingpong.proto
	# Generate reverse-proxy (.pb.gw.go)
	protoc -I ./ --grpc-gateway_out=logtostderr=true:. ./pb/pingpong.proto
	cp ./pb/*.go ./ws001-api/pb/
	cp ./pb/*.go ./ws002-pingpong/pb/

api.upgrade:
	docker build -t gcr.io/silkrode-golang/ws001-api ./ws001-api
	docker image push gcr.io/silkrode-golang/ws001-api:latest
	kubectl set image deployment/ws001-api ws001-api=gcr.io/silkrode-golang/ws001-api:latest -n ${NAMESPACE}
	kubectl rollout restart deployment/ws001-api -n ${NAMESPACE}
	kubectl rollout status deployment ws001-api -n ${NAMESPACE}


pingpong.upgrade:
	docker build -t gcr.io/silkrode-golang/ws002-pingpong ./ws002-pingpong
	docker push gcr.io/silkrode-golang/ws002-pingpong:latest
	kubectl set image deployment/ws002-pingpong ws002-pingpong=gcr.io/silkrode-golang/ws002-pingpong:latest  -n ${NAMESPACE}
	kubectl rollout restart deployment/ws002-pingpong -n ${NAMESPACE}
	kubectl rollout status deployment ws002-pingpong -n ${NAMESPACE}

