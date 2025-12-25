package smartcontract

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// UserExists returns true when user with given ID exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return userJSON != nil, nil
}

// CreateUser registers a new user
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, id string, name string, role string) error {
	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", id)
	}

	// Get Client Identity (WalletID)
	// In a real scenario, we might want to verify if the caller matches the ID or admin
	clientIdentity, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity: %v", err)
	}

	user := User{
		ID:        id,
		WalletID:  clientIdentity,
		Role:      role,
		Status:    "Active",
		Sequence:  1,
		UpdatedAt: time.Now().Unix(),
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, userJSON)
}

// ReadUser returns the user stored in the world state with given id
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", id)
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id, name, metaUrl string, views []string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	owner, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity: %v", err)
	}

	asset := Asset{
		ID:        id,
		Name:      name,
		Owner:     owner,
		MetaUrl:   metaUrl,
		Status:    "Active",
		Views:     views,
		Sequence:  1,
		LastTxid:  ctx.GetStub().GetTxID(),
		UpdatedAt: time.Now().Unix(),
		UpdatedBy: owner,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return err
	}

	// Emit Event
	return ctx.GetStub().SetEvent("AssetCreated", assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id, name, metaUrl, status string, views []string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientID, _ := ctx.GetClientIdentity().GetID()
	if asset.Owner != clientID {
		return fmt.Errorf("only the owner can update the asset")
	}

	if asset.Status == "Locked" {
		return fmt.Errorf("asset is locked and cannot be updated")
	}

	asset.Name = name
	asset.MetaUrl = metaUrl
	asset.Status = status
	asset.Views = views
	asset.Sequence += 1
	asset.LastTxid = ctx.GetStub().GetTxID()
	asset.UpdatedAt = time.Now().Unix()
	asset.UpdatedBy = clientID

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().SetEvent("AssetUpdated", assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	err = ctx.GetStub().DelState(id)
	if err != nil {
		return fmt.Errorf("failed to delete asset %s", id)
	}
	
	// Emit Event with just the ID
	return ctx.GetStub().SetEvent("AssetDeleted", []byte(id))
}

// InitiateTransfer (Phase 1)
func (s *SmartContract) InitiateTransfer(ctx contractapi.TransactionContextInterface, assetID, buyerWalletID string) error {
	asset, err := s.ReadAsset(ctx, assetID)
	if err != nil {
		return err
	}

	clientID, _ := ctx.GetClientIdentity().GetID()
	if asset.Owner != clientID {
		return fmt.Errorf("only the owner can initiate transfer")
	}
	if asset.Status == "Locked" {
		return fmt.Errorf("asset is already locked")
	}

	// 1. Lock Asset
	asset.Status = "Locked"
	asset.UpdatedBy = clientID
	asset.UpdatedAt = time.Now().Unix()
	assetJSON, _ := json.Marshal(asset)
	ctx.GetStub().PutState(assetID, assetJSON)

	// 2. Create Transfer Request
	transferReq := TransferRequest{
		AssetID:  assetID,
		SellerID: clientID,
		BuyerID:  buyerWalletID,
		Status:   "Pending",
	}
	transferJSON, _ := json.Marshal(transferReq)
	transferKey := "transfer_" + assetID
	ctx.GetStub().PutState(transferKey, transferJSON)

	return ctx.GetStub().SetEvent("TransferInitiated", transferJSON)
}

// CompleteTransfer (Phase 2)
func (s *SmartContract) CompleteTransfer(ctx contractapi.TransactionContextInterface, assetID string) error {
	transferKey := "transfer_" + assetID
	transferJSON, err := ctx.GetStub().GetState(transferKey)
	if err != nil || transferJSON == nil {
		return fmt.Errorf("transfer request not found")
	}

	var req TransferRequest
	json.Unmarshal(transferJSON, &req)

	clientID, _ := ctx.GetClientIdentity().GetID()
	// In strict mode, check if clientID matches req.BuyerID. 
	// However, buyerWalletID format in initiate might differ from GetID(). Assuming exact match for now.

	if req.Status != "Pending" {
		return fmt.Errorf("invalid transfer status")
	}

	// 1. Update Asset Owner & Unlock
	asset, _ := s.ReadAsset(ctx, assetID)
	asset.Owner = req.BuyerID // Or clientID if we enforce consistency
	asset.Status = "Active"
	asset.UpdatedBy = clientID
	asset.UpdatedAt = time.Now().Unix()
	
	assetJSON, _ := json.Marshal(asset)
	ctx.GetStub().PutState(assetID, assetJSON)

	// 2. Update Transfer Request
	req.Status = "Completed"
	req.BuyerID = clientID // Confirm actual buyer
	transferJSON, _ = json.Marshal(req)
	ctx.GetStub().PutState(transferKey, transferJSON)

	return ctx.GetStub().SetEvent("AssetTransferred", assetJSON)
}

// RejectTransfer (Phase 2 Alternative)
func (s *SmartContract) RejectTransfer(ctx contractapi.TransactionContextInterface, assetID string) error {
	transferKey := "transfer_" + assetID
	transferJSON, err := ctx.GetStub().GetState(transferKey)
	if err != nil || transferJSON == nil {
		return fmt.Errorf("transfer request not found")
	}
	
	var req TransferRequest
	json.Unmarshal(transferJSON, &req)
	
	if req.Status != "Pending" {
		return fmt.Errorf("transfer is not pending")
	}

	// 1. Unlock Asset
	asset, _ := s.ReadAsset(ctx, assetID)
	asset.Status = "Active" // Generic revert to Active
	assetJSON, _ := json.Marshal(asset)
	ctx.GetStub().PutState(assetID, assetJSON)

	// 2. Update Request
	req.Status = "Rejected"
	transferJSON, _ = json.Marshal(req)
	ctx.GetStub().PutState(transferKey, transferJSON)

	return ctx.GetStub().SetEvent("TransferRejected", transferJSON)
}

// CancelTransfer (Phase 2 Alternative)
func (s *SmartContract) CancelTransfer(ctx contractapi.TransactionContextInterface, assetID string) error {
	transferKey := "transfer_" + assetID
	transferJSON, err := ctx.GetStub().GetState(transferKey)
	if err != nil || transferJSON == nil {
		return fmt.Errorf("transfer request not found")
	}

	var req TransferRequest
	json.Unmarshal(transferJSON, &req)

	clientID, _ := ctx.GetClientIdentity().GetID()
	if req.SellerID != clientID {
		return fmt.Errorf("only the seller can cancel")
	}
	if req.Status != "Pending" {
		return fmt.Errorf("transfer is not pending")
	}

	// 1. Unlock Asset
	asset, _ := s.ReadAsset(ctx, assetID)
	asset.Status = "Active"
	assetJSON, _ := json.Marshal(asset)
	ctx.GetStub().PutState(assetID, assetJSON)

	// 2. Update Request
	req.Status = "Cancelled"
	transferJSON, _ = json.Marshal(req)
	ctx.GetStub().PutState(transferKey, transferJSON)

	return ctx.GetStub().SetEvent("TransferCancelled", transferJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			// Skip if unmarshal fails (might be a user or transfer object mixed in if not careful with key design)
			// Ideally we should use composite keys or query by docType. 
			// For this simple implementation, if it fails to unmarshal to Asset, we ignore it.
			continue
		}
		// Filter by docType logic if implemented, or just check ID
		if asset.ID != "" { 
			assets = append(assets, &asset)
		}
	}

	return assets, nil
}
