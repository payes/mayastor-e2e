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
