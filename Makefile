.PHONY: protos clean client agent 
protos:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	protos/protobuf.proto

agent:
	go build -o agent ./cmd/agent	
client:
	go build -o client ./cmd/client

clean-db:
	rm -rf docker/data && rm -rf docker/jobs
agent-up: clean-db
	docker-compose -f docker/docker-compose.yml up -d --build
agent-down:
	docker-compose -f docker/docker-compose.yml down
agent-shell:
	docker exec  -it rc-agent  /bin/bash
