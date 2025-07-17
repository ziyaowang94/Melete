#  A Privacy-Preserving Layered Blockchain for Cross-Organization Collaboration

## Introduction

This repository contains the source code of the Melete prototype system and implementations of other consensus protocols for comparison.

## Folder structure

```sh
├── build # the output binaries file
└── source # source code of system
```

## Prerequisites

The code is based on golang and the test is running on Linux. There are a few dependencies to run the code. The major libraries are listed as follows:

* rocksdb
* proto
* golang

### Install rocksdb (need for both test and build)

1. preparing the environment.

   ```bash
   sudo apt install build-essential
   sudo apt-get install libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev libgflags-dev
   ```

2. build rocksdb lib from source.

   ```bash
   wget https://github.com/facebook/rocksdb/archive/v6.28.2.zip
   
   unzip v6.28.2.zip && cd rocksdb-6.28.2
   
   make shared_lib
   make static_lib
   ```

3. install lib to system path.

   ```bash
   sudo make install-shared
   sudo make install-static
   
   echo "/usr/local/lib" | sudo tee /etc/ld.so.conf.d/rocksdb-x86_64.conf
   sudo ldconfig -v
   ```

### Install golang (need for build)

Follow the offical manual [https://go.dev/doc/install](https://go.dev/doc/install).

### Install proto-go (need for build)

1. `protoc` is already provided and placed in the `source` directory.

2. In order to compile the code, you need to install `protoc-gen-go` and `protoc-gen-go-grpc`.

    ```sh
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest 
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```
## How to Build

Install all prerequisites, then enter the `source` folder and run `make build-all`, all the binaries will be palced in the `build` folder.
