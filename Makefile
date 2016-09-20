all: build/logdna-stdin

build/logdna-stdin: cmds/logdna-stdin/main.go
	cd cmds/logdna-stdin && go build -v \
		-o ../../build/logdna-stdin

test:
	go test

clean:
	rm -f build/logdna-stdin

install: build/logdna-stdin
	cp build/logdna-stdin ${GOPATH}/bin/logdna-stdin

.PHONY: all test clean install
