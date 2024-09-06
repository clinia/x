#!/bin/bash

# Variables
REPO_URL="https://github.com/triton-inference-server/common.git"
FOLDER_PATH="protobuf"
DESTINATION_FOLDER="tritonx/proto"

# Clone the repository with sparse-checkout
git clone --depth 1 --filter=blob:none --sparse $REPO_URL temp_repo
cd temp_repo
git sparse-checkout set $FOLDER_PATH

# Move the folder to the destination
mkdir -p ../$DESTINATION_FOLDER
mv $FOLDER_PATH/* ../$DESTINATION_FOLDER/

# Clean up
cd ..
rm -rf temp_repo

echo "Folder downloaded and moved to $DESTINATION_FOLDER"