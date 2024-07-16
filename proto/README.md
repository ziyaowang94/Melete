Protobuf Installation and Usage (Lunix)


1. Add the 'protoc' executable to the user bin directory in the install folder, then set the execution permission with 'chmod a+x protoc', and finally confirm the installation with 'protoc --version'.

* if it is a Windows user, please go to https://github.com/protocolbuffers/protobuf/releases/tag/v26.1 and select the appropriate version download, extracted from the bin folder for the executable program, And add it to your path directory. The latest release is currently 1.26*

Enter the following command:

` ` `
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
` ` `

3. go back to the project root and run 'make proto-gen' to generate the go text of the proto file and 'make proto-delete' to delete all the go text.

4. proto file writing conventions:

(1) option go_package: emulator/proto/proto file relative path, for example. / proto/ours/abci/types. The proto file, fill in the corresponding ` "emulator/proto/ours/abci `".

(2) package: naming domain, in order to avoid the bug, unified use file definition the relative path, for example. / proto/ours/abci/types. The proto file, fill in the corresponding ` "ours. Abci `".

(3) the import: import other proto file, fill in the relative path to the proto file, for example. / proto/ours/abci/types. The proto file should use ` "ours/abci/types. Proto ` for import. message types in the same namespace (same folder) can be used directly, message types in different namespace must be imported using namespace. For Example, the Example type in '"ours/abci/types.proto"' above should be applied as ours.abci.example.