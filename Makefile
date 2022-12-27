.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest
	go install github.com/bradleyjkemp/grpc-tools/...@latest
	go install sigs.k8s.io/kind@latest


.PHONY: proto
proto:
	mkdir -p pkg/api
	rm -fr pb
	mkdir -p pb

	protoc -I api/proto --go_out=pb --go_opt=paths=source_relative \
           --go-grpc_out=pb --go-grpc_opt=require_unimplemented_servers=false,paths=source_relative \
           --grpc-gateway_out=pb  --grpc-gateway_opt=logtostderr=true,paths=source_relative,generate_unbound_methods=true \
           --openapiv2_out cmd/server/playground --openapiv2_opt logtostderr=true --openapiv2_opt allow_merge=true,merge_file_name=currency --openapiv2_opt generate_unbound_methods=true \
            currency.proto

.PHONY: tls
tls:
	mkdir -p tls
	openssl req -x509 -nodes -days 36500 -newkey rsa:2048 -subj "/CN=grpc" -keyout cmd/server/server.key -out cmd/server/server.crt

.PHONY: dump
dump:
	grpc-dump --key=cmd/server/server.key --cert=cmd/server/server.crt --port=8081 -destination localhost:8080 -proto_roots api

.PHONY: dock
dock:
	docker build  -t localhost:5001/mybuildimage --file=cmd/server/Dockerfile .
	docker push localhost:5001/mybuildimage

.PHONY: dockrun
dockrun:
	docker run --rm -p8080:8080 -it localhost:5001/mybuildimage
