GOPATH = $(shell go env GOPATH)

check: error
	golangci-lint run

check-clean-cache:
	golangci-lint cache clean

protoc-setup:
	cd meshes
	wget https://raw.githubusercontent.com/layer5io/meshery/master/meshes/meshops.proto

proto:	
	protoc -I meshes/ meshes/meshops.proto --go_out=plugins=grpc:./meshes/

docker:
	docker build -t meshery/meshery-nsm .

docker-run:
	(docker rm -f meshery-nsm) || true
	docker run --name meshery-nsm -d \
	-p 10004:10004 \
	-e DEBUG=true \
	meshery/meshery-nsm

run:
	DEBUG=true go run main.go

.PHONY: error
error:
	go run github.com/layer5io/meshkit/cmd/errorutil -d . analyze -i ./helpers -o ./helpers