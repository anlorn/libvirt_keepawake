binary_name := "libvirt-keepawake"

all: lint test run

.PHONY: build
build:
	go build -o ${binary_name} .

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	go test -v -buildvcs -o /tmp/test_binaries/ ./...

.PHONY: run
run: build
	./${binary_name}

clean:
	rm ${binary_name}
