# Uranus: A Layered and Integrated Sharding Blockchain System

## Introduction

This repository contains the source code of the Uranus prototype system and implementations of other consensus protocols for comparison.

## Folder structure

```sh
├── build # pre-compiled binaries
├── dependency # Prerequisites dependency
├── scripts # scripts for running test
└── source # source code of system
```

## Prerequisites

The code is based on golang and the test is running on Linux. There are a few dependencies to run the code. The major libraries are listed as follows:

* rocksdb
* proto
* golang

### Install rocksdb (need for both test and build)

Follow the manual [rocks_install.md](./dependency/rocksdb/rocks_install.md).

### Install golang (need for build)

Follow the offical manual [https://go.dev/doc/install](https://go.dev/doc/install).

## Install proto-go (need for build)

Follow the manual [proto_install.md](./dependency/proto/proto_install.md).

## How to run

We provide pre-compiled binaries to run the tests, you can also build from source.

### run with binaries (only need rocksdb)

How to run Uranus and experiment can be found in [run.md](/scripts/run.md), all the need scripts is placed in `/scripts`.

### build from source (need all Prerequisites)

Enter the `source` directory and run `make build-all`, all the binaries will be palced in the `build` directory.
