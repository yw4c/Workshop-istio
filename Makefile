NAMESPACE ?= ares

gen_pb:
	# go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
	# Generate gRPC stub (.pb.go)
	protoc -I ./ --go_out=plugins=grpc:. ./pb/pingpong.proto
	# Generate reverse-proxy (.pb.gw.go)
	#protoc -I ./ --grpc-gateway_out=logtostderr=true:. ./pb/pingpong.proto
	cp ./pb/pingpong.pb.go ./ws001-api/pb/pingpong.pb.go
	cp ./pb/pingpong.pb.go ./ws002-pingpong/pb/pingpong.pb.go

api.upgrade:
	docker build -t yw4code/ws001-api:latest ./ws001-api
	docker image push yw4code/ws001-api:latest
	kubectl set image deployment/ws001-api-blue ws001-api=yw4code/ws001-api:latest -n ${NAMESPACE}
	kubectl rollout restart deployment/ws001-api-blue -n ${NAMESPACE}
	kubectl rollout status deployment ws001-api-blue -n ${NAMESPACE}


pingpong.upgrade:
	docker build -t yw4code/ws002-pingpong:latest ./ws002-pingpong
	docker push yw4code/ws002-pingpong:latest
	kubectl set image deployment/ws002-pingpong-blue ws002-pingpong=yw4code/ws002-pingpong:latest  -n ${NAMESPACE}
	kubectl rollout restart deployment/ws002-pingpong-blue -n ${NAMESPACE}
	kubectl rollout status deployment ws002-pingpong-blue -n ${NAMESPACE}

auth.upgrade:
	docker build -t yw4code/ws003-auth:latest ./ws003-auth
	docker image push yw4code/ws003-auth:latest
	kubectl set image deployment/ws003-auth ws003-auth=yw4code/ws003-auth:latest -n ${NAMESPACE}
	kubectl rollout restart deployment/ws003-auth -n ${NAMESPACE}
	kubectl rollout status deployment ws003-auth -n ${NAMESPACE}

reload-ordering:
	#kubectl delete -f ./deployment/ordering -n ${NAMESPACE}
	kubectl apply -f ./deployment/http-filter/030-auth-filter.yaml -n ${NAMESPACE}
	kubectl apply -f ./deployment/http-filter/040-accesslog-filter.yaml -n ${NAMESPACE}