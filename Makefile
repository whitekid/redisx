TARGET=bin/redisx
SRC=$(shell find . -type f -name '*.go' -not -path "*_test.go")
PROTO_SRC=$(shell find . -type f -name '*.proto')
PROTO_GO_SRC=$(shell find . -type f -name '*.pb.go')
BUILD_FLAGS?=-v -ldflags="-s -w"

.PHONY: clean test dep tidy

all: build
build: $(TARGET)

$(TARGET): $(SRC) $(PROTO_GO_SRC)
	@mkdir -p bin
	go build -o bin/ ${BUILD_FLAGS} ./cmd/...

clean:
	@rm -f ${TARGET}

test:
	go test -v ./...

# update modules & tidy
dep:
	@rm -f go.mod go.sum
	@go mod init github.com/whitekid/redisx

	@$(MAKE) tidy

tidy:
	@go mod tidy -v
