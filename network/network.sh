#!/bin/bash

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=${PWD}
export VERBOSE=false

# Global Variables
CHANNEL_NAME="mychannel"
CLI_DELAY=3
MAX_RETRY=5
COMPOSE_FILE_BASE=docker/docker-compose.yaml

function printHelp() {
  echo "Usage: "
  echo "  network.sh <mode> [options]"
  echo "    <mode> - one of 'up', 'down', 'createChannel'"
  echo "      - 'up' - bring up the network with docker-compose up"
  echo "      - 'down' - clear the network with docker-compose down"
  echo "      - 'createChannel' - create and join a channel"
}

function checkPrereqs() {
    peer version > /dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        echo "Fabric binaries not found in PATH. Will use Docker container for generation steps."
        USE_DOCKER_TOOLS=true
    else
        USE_DOCKER_TOOLS=false
    fi
}

function runTool() {
    if [ "$USE_DOCKER_TOOLS" == "true" ]; then
        docker run --rm -v $(pwd):/data -w /data -e FABRIC_CFG_PATH=/data hyperledger/fabric-tools:2.5.9 "$@"
    else
        "$@"
    fi
}

function createGenesisBlock() {
  echo "#########  Generating Orderer Genesis block #########"
  set -x
  runTool configtxgen -profile SampleMultiNodeEtcdRaft -channelID system-channel -outputBlock ./system-genesis-block/genesis.block
  res=$?
  set +x
  if [ $res -ne 0 ]; then
    echo "Failed to generate orderer genesis block..."
    exit 1
  fi
}

function createConsortium() {
    echo "Consortium creation step skipped"
}

function generateCerts() {
  if [ -d "crypto-config" ]; then
    rm -Rf crypto-config
  fi

  echo "##### Generate certificates using cryptogen tool #########"
  set -x
  runTool cryptogen generate --config=./crypto-config.yaml
  res=$?
  set +x
  if [ $res -ne 0 ]; then
    echo "Failed to generate certificates..."
    exit 1
  fi
}

function networkUp() {
  checkPrereqs
  
  # Generate artifacts if they don't exist
  if [ ! -d "crypto-config" ]; then
    generateCerts
  fi
  
  if [ ! -f "./system-genesis-block/genesis.block" ]; then
     mkdir -p system-genesis-block
     createGenesisBlock
  fi

  echo "Starting network..."
  IMAGE_TAG=latest docker-compose -f ${COMPOSE_FILE_BASE} up -d 2>&1
  docker ps -a
  if [ $? -ne 0 ]; then
    echo "ERROR !!!! Unable to start network"
    exit 1
  fi
}

function networkDown() {
  docker-compose -f ${COMPOSE_FILE_BASE} down --volumes --remove-orphans
  
  # Cleanup artifacts
  if [ "$1" == "clean" ]; then
      rm -rf crypto-config system-genesis-block
  fi
}

function createChannel() {
    checkPrereqs
    # Create Channel Transaction
    set -x
    runTool configtxgen -profile OneOrgChannel -outputCreateChannelTx ./channel-artifacts/${CHANNEL_NAME}.tx -channelID $CHANNEL_NAME
    res=$?
    set +x
    if [ $res -ne 0 ]; then
        echo "Failed to generate channel configuration transaction..."
        exit 1
    fi

    # Create Channel
    # We use CLI container to run peer commands
    echo "Creating channel ${CHANNEL_NAME}..."
    docker exec cli peer channel create -o orderer0.example.com:7050 -c $CHANNEL_NAME -f /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/${CHANNEL_NAME}.tx --outputBlock /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/${CHANNEL_NAME}.block --tls --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer0.example.com/msp/tlscacerts/tlsca.example.com-cert.pem

    # Join Peers
    for peer in peer0 peer1 peer2; do
        if [ "$peer" == "peer0" ]; then PORT=7051; fi
        if [ "$peer" == "peer1" ]; then PORT=8051; fi
        if [ "$peer" == "peer2" ]; then PORT=9051; fi

        echo "Joining $peer to channel on port $PORT..."
        docker exec -e CORE_PEER_ADDRESS=${peer}.org1.example.com:${PORT} -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/${peer}.org1.example.com/tls/ca.crt cli peer channel join -b /opt/gopath/src/github.com/hyperledger/fabric/peer/channel-artifacts/${CHANNEL_NAME}.block
        
        # Wait a bit
        sleep 2
    done
    
    # Update Anchor Peers (Optional but good practice)
    # Skipping for brevity in initial setup, but essential for cross-org comms (not strictly needed for 1 org but good hygiene)
}

# Directories
mkdir -p channel-artifacts system-genesis-block

# Mode selector
if [ "$1" == "up" ]; then
  networkUp
elif [ "$1" == "down" ]; then
  networkDown
elif [ "$1" == "createChannel" ]; then
  createChannel
else
  printHelp
  exit 1
fi
