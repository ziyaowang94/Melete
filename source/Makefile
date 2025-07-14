
TARGET_DIR := "../build"

install-dependencies:
	@ echo "install dependencies"
	@ go mod tidy
.PHONY: install-dependencies

proto-gen:
	@ echo "generate proto file"
	@ find ./proto -type f -name "*.proto" -exec ./protoc --proto_path=. --go_out=. --go_opt=paths=source_relative {} \;
.PHONY: proto-gen


proto-clean:
	@ echo "clean proto file"
	@ find ./proto -type f -name "*.pb.go" -delete
.PHONY: proto-clean

build-melete:
	@ echo "Building melete..."
	@ cd melete/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv melete/main/main $(TARGET_DIR)/melete
.PHONY: build-melete

build-melete-latancy:
	@ echo "Building melete latency..."
	@ cd melete/abci/minibank/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv melete/abci/minibank/main/main $(TARGET_DIR)/melete-latency
.PHONY: build-melete-latancy

build-block-logger:
	@ echo "Building block-logger..."
	@ cd logger/blocklogger/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv logger/blocklogger/main/main $(TARGET_DIR)/logger
.PHONY: build-block-logger


build-pyramid:
	@ echo "Building pyramid..."
	@ cd pyramid/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv pyramid/main/main $(TARGET_DIR)/pyramid
.PHONY: build-pyramid

build-pyramid-latancy:
	@ echo "Building pyramid latency..."
	@ cd pyramid/abci/minibank/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv pyramid/abci/minibank/main/main $(TARGET_DIR)/pyramid-latency
.PHONY: build-pyramid-latancy


build-rapid:
	@ echo "Building rapid..."
	@ cd rapid/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv rapid/main/main $(TARGET_DIR)/rapid
.PHONY: build-rapid

build-rapid-latancy:
	@ echo "Building rapid latency..."
	@ cd rapid/abci/minibank/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv rapid/abci/minibank/main/main $(TARGET_DIR)/rapid-latency
.PHONY: build-rapid-latancy

build-store:
	@ echo "Building store..."
	@ cd utils/store/main && go build
	@ echo "move to $(TARGET_DIR)"
	@ mv utils/store/main/main $(TARGET_DIR)/store
.PHONY: build-store

.PHONY: build-all

build-all: 
	make install-dependencies
	make proto-gen
	make build-melete
	make build-melete-latancy
	make build-block-logger
	make build-pyramid
	make build-pyramid-latancy
	make build-store
	make build-rapid
	make build-rapid-latancy