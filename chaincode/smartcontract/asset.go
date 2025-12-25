package smartcontract

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Owner     string   `json:"owner"`
	MetaUrl   string   `json:"metaUrl"`
	Status    string   `json:"status"` // Active, Locked, Deleted
	Views     []string `json:"views"`  // ["public"] or list of UserIDs
	Sequence  uint64   `json:"sequence"`
	LastTxid  string   `json:"lastTxid"`
	UpdatedAt int64    `json:"updatedAt"`
	UpdatedBy string   `json:"updatedBy"`
}
