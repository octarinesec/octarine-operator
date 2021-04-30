# Please install protobuf C executable from https://github.com/google/protobuf/releases/tag/v3.5.1 for your platform
# Also - install the go pugin : go get -u github.com/golang/protobuf/protoc-gen-go
# export PATH=$PATH;~/go/bin

protoc --go_out=plugins=grpc:. *.proto