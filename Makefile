binary_name := "libvirt_keepawake"

all: run

.PHONY: build
build:
	go build -o ${binary_name} ./cmd/libvirt_keepawake

.PHONY: test
test:
	go test -v -buildvcs -o /tmp/test_binaries/ ./...

.PHONY: run
run: build
	./${binary_name}

clean:
	rm ${binary_name}