# Chaincode Documentation

This document describes the Smart Contracts (Chaincode) for the Asset Management System (AMS-2).

## Data Models

The world state is composed of three main entity types, distinguished by their logical purpose and usage.

### 1. Asset
The core entity representing a physical or digital asset.

```go
type Asset struct {
    ID        string   `json:"id"`        // Unique Identifier
    Name      string   `json:"name"`      // Display Name
    Owner     string   `json:"owner"`     // Owner's Wallet/MSP ID
    MetaUrl   string   `json:"metaUrl"`   // IPFS CID or External URL
    Status    string   `json:"status"`    // "Active", "Locked", "Deleted"
    Views     []string `json:"views"`     // Access Control: ["public"] or [UserID...]
    Sequence  uint64   `json:"sequence"`  // Version control
    LastTxid  string   `json:"lastTxid"`  // Last Transaction ID that modified this asset
    UpdatedAt int64    `json:"updatedAt"` // Unix timestamp
    UpdatedBy string   `json:"updatedBy"` // User ID who performed update
}
```

### 2. User
On-chain representation of a system user, primarily for role and status management.

```go
type User struct {
    ID        string `json:"id"`
    WalletID  string `json:"walletId"` // Linked Fabric Identity
    Role      string `json:"role"`     // e.g., "Manager", "Admin"
    Status    string `json:"status"`   // "Active", "Suspended", "Deleted"
    Sequence  uint64 `json:"sequence"`
    UpdatedAt int64  `json:"updatedAt"`
}
```

### 3. TransferRequest
Managing the **Two-Phase Commit** state for safe asset transfers.

```go
type TransferRequest struct {
    AssetID  string `json:"assetId"`
    SellerID string `json:"sellerId"`
    BuyerID  string `json:"buyerId"`
    Status   string `json:"status"` // "Pending", "Completed", "Rejected", "Cancelled"
}
```

---

## Smart Contract Methods

The chaincode exposes the following transaction functions via `smartcontract.go`.

### Asset Operations

| Method | Inputs | Description | Events Emitted |
| :--- | :--- | :--- | :--- |
| **`CreateAsset`** | `id`, `name`, `metaUrl`, `views` | Creates a new asset. Owner is set to caller. | `AssetCreated` |
| **`ReadAsset`** | `id` | Returns asset details. | - |
| **`UpdateAsset`** | `id`, `name`, `metaUrl`, `status`, `views` | Updates metadata. Fails if asset is Locked. | `AssetUpdated` |
| **`DeleteAsset`** | `id` | Removes asset from state. | `AssetDeleted` |
| **`GetAllAssets`** | - | Returns all assets (range query). | - |
| **`AssetExists`** | `id` | Boolean check. | - |

### Transfer Operations (Two-Phase Commit)

This workflow ensures atomic and verifiable ownership transfer.

1.  **`InitiateTransfer`**
    *   **Inputs**: `assetID`, `buyerWalletID`
    *   **Action**: Locks the Asset (`Status="Locked"`) and creates a `TransferRequest` (`Status="Pending"`).
    *   **Event**: `TransferInitiated`
2.  **`CompleteTransfer`**
    *   **Inputs**: `assetID`
    *   **Action**: Verifies request. Updates Asset Owner to Buyer. Unlocks Asset (`Status="Active"`). Marks Request as "Completed".
    *   **Event**: `AssetTransferred`
3.  **`RejectTransfer`**
    *   **Inputs**: `assetID`
    *   **Action**: Receiver rejects. Asset Unlocked. Request marked "Rejected".
    *   **Event**: `TransferRejected`
4.  **`CancelTransfer`**
    *   **Inputs**: `assetID`
    *   **Action**: Seller cancels. Asset Unlocked. Request marked "Cancelled".
    *   **Event**: `TransferCancelled`

### User Operations

| Method | Inputs | Description | Events Emitted |
| :--- | :--- | :--- | :--- |
| **`CreateUser`** | `id`, `name`, `role` | registers a new user. | - |
| **`ReadUser`** | `id` | Returns user details. | - |
| **`UserExists`** | `id` | Boolean check. | - |

---

## Events for Eventual Consistency

The system relies on Chaincode Events to synchronize data to the off-chain PostgreSQL database. The backend API must subscribe to these events:

*   `AssetCreated`
*   `AssetUpdated`
*   `AssetDeleted`
*   `TransferInitiated`
*   `AssetTransferred`
*   `TransferRejected`
*   `TransferCancelled`
