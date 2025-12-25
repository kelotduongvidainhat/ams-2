package smartcontract

// TransferRequest manages the state of a Two-Phase Commit transfer
type TransferRequest struct {
	AssetID  string `json:"assetId"`
	SellerID string `json:"sellerId"`
	BuyerID  string `json:"buyerId"`
	Status   string `json:"status"` // Pending, Completed, Rejected, Cancelled
}
