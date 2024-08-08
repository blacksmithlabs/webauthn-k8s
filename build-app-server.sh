#!/bin/bash

app_tag="192.168.13.3:32000/webauthn-server:alpha"
migration_tag="192.168.13.3:32000/webauthn-server-migrations:alpha"

pushd src

echo "Building the app image..."
docker buildx build --platform linux/amd64,linux/arm64 --target=run -f Dockerfile.app -t "$app_tag" --output=type=registry,registry.insecure=true --push .

if [ $? -ne 0 ]; then
    echo "Build app failed. Exiting..."
    exit 1
fi

popd
pushd database

echo "Building the migration image..."
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile.migrations -t "$migration_tag" --output=type=registry,registry.insecure=true --push .

if [ $? -ne 0 ]; then
    echo "Build app failed. Exiting..."
    exit 1
fi

popd
