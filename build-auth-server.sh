#!/bin/bash

app_tag="registry.blacksmithlabs.dev/webauthn-server:alpha"
migration_tag="registry.blacksmithlabs.dev/webauthn-server-migrations:alpha"

pushd src

echo "Building the auth image..."
docker buildx build --platform linux/amd64,linux/arm64 --target=run -f Dockerfile.auth -t "$app_tag" --push .

if [ $? -ne 0 ]; then
    echo "Build auth failed. Exiting..."
    exit 1
fi

popd
pushd database

echo "Building the migration image..."
docker buildx build --platform linux/amd64,linux/arm64 -f Dockerfile.migrations -t "$migration_tag" --push .

if [ $? -ne 0 ]; then
    echo "Build app failed. Exiting..."
    exit 1
fi

popd
