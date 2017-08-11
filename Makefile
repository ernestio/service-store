install:
	go install -v

build:
	go build -v ./...

deps:
	go get gopkg.in/r3labs/graph.v2
	go get github.com/jinzhu/gorm
	go get github.com/nats-io/nats
	go get github.com/lib/pq
	go get github.com/r3labs/natsdb
	go get github.com/ernestio/ernest-config-client

dev-deps: deps
	go get github.com/smartystreets/goconvey/convey
	go get github.com/stretchr/testify/suite
	go get github.com/alecthomas/gometalinter
	gometalinter --install

test:
	go test -v ./...

lint:
	gometalinter --config .linter.conf
