build:
	go build

install:
	go install

test:
	go test ./...

node1: build
	./gloster 127.0.0.1:12000 127.0.0.1:12001,127.0.0.1:12002 node1.db

node2: build
	./gloster 127.0.0.1:12001 127.0.0.1:12000,127.0.0.1:12002 node2.db

node3: build
	./gloster 127.0.0.1:12002 127.0.0.1:12000,127.0.0.1:12001 node3.db
