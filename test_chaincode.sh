#!/bin/bash
set -e

echo "Running Chaincode Tests in Docker (golang:1.21)..."

docker run --rm \
    -v $(pwd)/chaincode:/opt/chaincode \
    -w /opt/chaincode \
    golang:1.21 \
    /bin/bash -c "go mod tidy && go test ./... -v"

if [ $? -eq 0 ]; then
    echo "✅ Tests Passed!"
else
    echo "❌ Tests Failed."
fi
