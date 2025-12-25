# Network Infrastructure Documentation

## Overview
The Asset Management System (AMS) runs on a centralized Hyperledger Fabric network designed for high availability and fault tolerance.

### Topology
- **Organization**: `Org1`
- **Consensus**: Raft (EtcdRaft)
- **Orderer Service**: 3-node Raft Cluster
    - `orderer0.example.com`
    - `orderer1.example.com`
    - `orderer2.example.com`
- **Peers**: 3 Peers for Org1 (High Availability)
    - `peer0.org1.example.com`
    - `peer1.org1.example.com`
    - `peer2.org1.example.com`
- **State Database**: CouchDB (1 per peer)
    - `couchdb0`, `couchdb1`, `couchdb2`
- **Certificate Authority**: Fabric CA (`ca_org1`)

## Prerequisites
- Docker Engine v24.0+
- Docker Compose v2.20+
- Hyperledger Fabric binaries (optional, script uses Docker image fallback)

## Network Operations

All network operations are managed via the `network/network.sh` script.

### 1. Start Network
This command generates crypto material, creates the genesis block, and starts all containers.
```bash
cd network
./network.sh up
```

### 2. Create Channel
Creates the channel `mychannel` and joins all 3 peers to it.
```bash
./network.sh createChannel
```

### 3. Verification
Verify that peers have joined the channel:
```bash
docker exec peer0.org1.example.com peer channel list
```

### 4. Stop Network
Stop containers and remove volumes (preserves crypto-config):
```bash
./network.sh down
```

To clean up everything (including generated certificates):
```bash
./network.sh down clean
```

## Configuration Files
- **`crypto-config.yaml`**: Defines identity specs for Orderers and Peers.
- **`configtx.yaml`**: Defines the `SampleMultiNodeEtcdRaft` profile and `OneOrgChannel`.
- **`docker/docker-compose.yaml`**: Defines container services, ports, and volume mounts.
