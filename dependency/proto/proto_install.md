# proto install manual

> We use protobuf for node data exchange.

1. `protoc` is already provided and placed in the `source` directory.

2. In order to compile the code, you need to install `protoc-gen-go` and `protoc-gen-go-grpc`.

    ```sh
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest 
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```
