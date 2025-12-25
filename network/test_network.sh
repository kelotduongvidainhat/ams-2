#!/bin/bash

# Test Script for AMS-2 Network Infrastructure
# Verifies that:
# 1. All expected containers are running.
# 2. Channel 'mychannel' has been created.
# 3. All peers have joined 'mychannel'.

echo "=== Starting Network Verification ==="

# 1. Check Containers
echo "[1/3] Checking Container Status..."
EXPECTED_CONTAINERS=(
  "orderer0.example.com" "orderer1.example.com" "orderer2.example.com"
  "peer0.org1.example.com" "peer1.org1.example.com" "peer2.org1.example.com"
  "couchdb0" "couchdb1" "couchdb2"
  "ca_org1" "cli"
)

ALL_RUNNING=true
for container in "${EXPECTED_CONTAINERS[@]}"; do
  if [ "$(docker inspect -f '{{.State.Running}}' ${container} 2>/dev/null)" == "true" ]; then
    echo "  [OK] $container is running."
  else
    echo "  [FAIL] $container is NOT running."
    ALL_RUNNING=false
  fi
done

if [ "$ALL_RUNNING" = false ]; then
  echo "Critical: Some containers are not running. Aborting test."
  exit 1
fi

# 2. Check Channel Block
echo "[2/3] Checking Channel Creation..."
if [ -f "./channel-artifacts/mychannel.block" ] || [ -f "network/channel-artifacts/mychannel.block" ]; then
    echo "  [OK] Channel block found."
else
    echo "  [FAIL] Channel block not found. Did you run './network.sh createChannel'?"
    exit 1
fi

# 3. Check Peer Membership
echo "[3/3] Checking Peer Channel Membership..."
PEERS=("peer0.org1.example.com" "peer1.org1.example.com" "peer2.org1.example.com")
CHANNEL_NAME="mychannel"

ALL_JOINED=true
for peer in "${PEERS[@]}"; do
  echo "  Checking $peer..."
  CHANNELS=$(docker exec $peer peer channel list 2>&1)
  if [[ $CHANNELS == *"$CHANNEL_NAME"* ]]; then
    echo "    [OK] $peer has joined $CHANNEL_NAME."
  else
    echo "    [FAIL] $peer has NOT joined $CHANNEL_NAME."
    ALL_JOINED=false
  fi
done

echo "=== Verification Result ==="
if [ "$ALL_JOINED" = true ]; then
    echo "SUCCESS: Network is healthy and fully configured."
    exit 0
else
    echo "FAILURE: Network issues detected."
    exit 1
fi
