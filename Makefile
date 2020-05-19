NAMESPACE ?= workshop
deploy:
	kubectl create namespace ${NAMESPACE}
	kubectl apply -f deployment/workshop.yaml -n ${NAMESPACE}

gen_pb:
	protoc -I ./ --go_out=plugins=grpc:. ./pb/pingpong.proto

api.upgrade:
	docker build -t gcr.io/silkrode-golang/ws001-api ./ws001-api
	docker push gcr.io/silkrode-golang/ws001-api:latest
	kubectl set image deployment/ws001-api ws001-api=gcr.io/silkrode-golang/ws001-api:latest -n ${NAMESPACE}

pingpong.upgrade:
	docker build -t gcr.io/silkrode-golang/ws002-pingpong ./ws002-pingpong
	docker push gcr.io/silkrode-golang/ws002-pingpong:latest
	kubectl set image deployment/ws002-pingpong ws002-pingpong=gcr.io/silkrode-golang/ws002-pingpong:latest  -n ${NAMESPACE}

