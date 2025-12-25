package smartcontract

// User represents a participant in the network
type User struct {
	ID        string `json:"id"`
	WalletID  string `json:"walletId"`
	Role      string `json:"role"`
	Status    string `json:"status"` // Active, Suspended, Deleted
	Sequence  uint64 `json:"sequence"`
	UpdatedAt int64  `json:"updatedAt"`
}
