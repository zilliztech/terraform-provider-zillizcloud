default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

unit-test:
	go test ./... -v 

install:
	go install -v ./...

doc:
	go generate ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...
