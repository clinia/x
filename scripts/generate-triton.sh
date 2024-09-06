#!/bin/bash

REPO_URL="https://github.com/triton-inference-server/common.git"
PROTO_FOLDER="protobuf"
PROTO_DEST="tritonx/proto"
GRPC_DEST="tritonx/grpc"
PACKAGE="github.com/clinia/tritonx/tritongrpc"

download_protos() {
    git clone --depth 1 --filter=blob:none --sparse $REPO_URL temp_repo
    cd temp_repo
    git sparse-checkout set $PROTO_FOLDER
    mkdir -p ../$PROTO_DEST
    mv $PROTO_FOLDER/* ../$PROTO_DEST/
    cd ..
    rm -rf temp_repo
    rm -rf $PROTO_DEST/health.proto
    echo "Folder downloaded and moved to $PROTO_DEST"
}

generate_protos() {
    for i in ${PROTO_DEST}/*.proto
    do
        echo "option go_package = \"${PACKAGE}\";" >> $i
    done
    protoc --proto_path="${PROTO_DEST}" --go-grpc_out="${GRPC_DEST}" --go-grpc_opt=paths=source_relative --go_out="${GRPC_DEST}" --go_opt=paths=source_relative ${PROTO_DEST}/*.proto
    echo "Protos generated in $GRPC_DEST"
}

download_protos
generate_protos