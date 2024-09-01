#!/bin/bash

cd src

tag="192.168.13.3:32000/webauthn-server:alpha"

echo "Building the image..."
docker buildx build --platform linux/amd64,linux/arm64 --target=run -f Dockerfile.app -t "$tag" --output=type=registry,registry.insecure=true --push .

if [ $? -ne 0 ]; then
    echo "Build failed. Exiting..."
    exit 1
fi

echo "Uploading the image..."
docker push "$tag"
