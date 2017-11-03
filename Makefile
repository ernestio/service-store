install:
	go install -v

build:
	go build -v ./...

deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

dev-deps: deps
	go get github.com/stretchr/testify/suite
	go get github.com/alecthomas/gometalinter
	gometalinter --install

test:
	go test -v ./...

lint:
	gometalinter --config .linter.conf
