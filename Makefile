init:
	go get github.com/golang/dep/cmd/dep
.PHONY: init

deps:
	dep ensure
.PHONY: deps

test:
	go test -v $(go list ./...)
.PHONY: test