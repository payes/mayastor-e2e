The file `mayastor.proto` is copied from the mayastor repo.
The client code in generated, using `protoc-gen-go-grpc`, this can be installed using the following command
```
go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

To generate the client code 
1. Edit `mayastor.proto` and add the line 
    * `option go_package = "/grpc";`
2. Use the command
    * `protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative mayastor.proto`
3. Remove the line that was added in step 1

To generate the client code using nix shell
1. run below command to update dependency
	nix-shell ci.nix

2. install package (if not present already)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

5. Run below command to generate server and client code
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative protobuf/mayastor.proto
