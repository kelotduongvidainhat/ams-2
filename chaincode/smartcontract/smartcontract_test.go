package smartcontract_test

import (
	"testing"

	"chaincode/smartcontract"
	"github.com/stretchr/testify/assert"
)

func TestCreateAssetStruct(t *testing.T) {
	// Simple test to verify struct fields are accessible
	asset := smartcontract.Asset{
		ID:     "asset1",
		Name:   "Test Asset",
		Status: "Active",
	}
	assert.Equal(t, "asset1", asset.ID)
	assert.Equal(t, "Test Asset", asset.Name)
	assert.Equal(t, "Active", asset.Status)
}
