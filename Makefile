GOLINT := golangci-lint

.PHONY: build test

build: dep
	#$(GOPATH)/pkg/mod/github.com/ibmdb/go_ibm_db@0.4.2/installer/setenv.sh
	go build -o conduit-connector-db2 cmd/db2/main.go

test:
	go install github.com/ibmdb/go_ibm_db/installer@v0.4.2
	go run ./test/setup.go
	docker run -itd --name mydb2 --privileged=true -p 50000:50000 -e LICENSE=accept -e DB2INST1_PASSWORD=pwd -e DBNAME=testdb -v vol:/database ibmcom/db2
	go test $(GOTEST_FLAGS) -race -gcflags=all=-d=checkptr=0 ./...

lint:
	$(GOLINT) run --timeout=5m -c .golangci.yml

mockgen:
	mockgen -package mock -source destination/interface.go -destination destination/mock/destination.go
	mockgen -package mock -source source/interface.go -destination source/mock/iterator.go

