# Asset Management System (AMS-2)

Dự án quản lý tài sản sử dụng nền tảng Blockchain Hyperledger Fabric.

## Technology Stack

### Network Infrastructure
- **Hyperledger Fabric**: v2.5.9
- **Fabric CA**: v1.5.12
- **Database (World State)**: CouchDB v3.3.3
- **Container Runtime**: Docker Engine v24.0+ & Docker Compose v2.20+

### Chaincode (Smart Contracts)
- **Language**: Go (Golang) v1.21
- **API**: Fabric Contract API (`github.com/hyperledger/fabric-contract-api-go` v1.2.2)

## Documentation
- [Network Infrastructure](./docs/network.md): Detailed guide on the 3-node Orderer/Peer topology and management scripts.
- [Chaincode Documentation](./docs/chaincode.md): Data models, smart contract methods, and transfer workflows.

## Project Structure
- `/documents`: Project documentation.
- `/network`: Fabric network configuration and scripts.
- `/chaincode`: Smart Contracts (Go).
- `/application`: Backend API & Client SDK.