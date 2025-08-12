package assets

import (
	"fmt"
	"math/big"
	"sync"
)

type AssetService struct {
	// userID -> (assetEnum -> *Asset)
	userAssets sync.Map // key: int64, val: *sync.Map
}

// get the whole assets map of a certain user
func (a *AssetService) getUserAssets(userID int64) *sync.Map {
	m, _ := a.userAssets.Load(userID)
	if m == nil {
		return nil
	}
	return m.(*sync.Map) // assetEnum -> *Asset
}

// get a certain kind of asset of the user
// if it doesn't exist, create it
func (a *AssetService) getOrCreateAssetType(userID int64, kind AssetsEnum) *Asset {
	m := a.getUserAssets(userID)
	if m == nil {
		return nil
	}

	if v, ok := m.Load(kind); ok {
		return v.(*Asset)
	}

	newAssetKind := NewAsset()
	if old, loaded := m.LoadOrStore(kind, newAssetKind); loaded {
		return old.(*Asset)
	}
	return newAssetKind
}

func (a *AssetService) tryTransfer(
	transferType TransferType,
	fromUser,
	toUser int64,
	assetKind AssetsEnum,
	amount *big.Float,
	checkBalance bool,
) (bool, error) {
	if amount == nil {
		return false, fmt.Errorf("param:amount is nil")
	}
	switch amount.Sign() {
	case 0:
		return true, nil
	case -1:
		return false, fmt.Errorf("param:amount is negative")
	}

	fromAsset := a.getOrCreateAssetType(fromUser, assetKind)
	toAsset := a.getOrCreateAssetType(toUser, assetKind)

	fromAsset.mu.Lock()
	toAsset.mu.Lock()
	defer fromAsset.mu.Unlock()
	defer toAsset.mu.Unlock()

	switch transferType {
	case AvailableToAvailable:
		if checkBalance && fromAsset.available.Cmp(amount) < 0 {
			return false, nil
		}
		fromAsset.available = new(big.Float).Sub(fromAsset.available, amount)
		toAsset.available = new(big.Float).Add(toAsset.available, amount)
		return true, nil
	case AvailableToFrozen:
		if checkBalance && fromAsset.available.Cmp(amount) < 0 {
			return false, nil
		}
		fromAsset.available = new(big.Float).Sub(fromAsset.available, amount)
		toAsset.frozen = new(big.Float).Add(toAsset.frozen, amount)
		return true, nil
	case FrozenToAvailable:
		if checkBalance && fromAsset.frozen.Cmp(amount) < 0 {
			return false, nil
		}
		fromAsset.frozen = new(big.Float).Sub(fromAsset.frozen, amount)
		toAsset.available = new(big.Float).Add(toAsset.available, amount)
		return true, nil

	default:
		return false, fmt.Errorf("invalid transfer type: %v", transferType)
	}

}
