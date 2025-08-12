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
	m, _ := a.userAssets.LoadOrStore(userID, &sync.Map{})
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

func lockPair(fromUser, toUser int64, kind AssetsEnum, from, to *Asset) func() {
	if from == to {
		from.mu.Lock()
		return func() { from.mu.Unlock() }
	}
	// stable order by (userID, kind)
	type key struct{ user int64; k AssetsEnum }
	k1, k2 := key{fromUser, kind}, key{toUser, kind}
	less := func(a, b key) bool { // total order
		if a.user != b.user {
			return a.user < b.user
		}
		return a.k < b.k
	}
	if less(k1, k2) {
		from.mu.Lock(); to.mu.Lock()
		return func() { to.mu.Unlock(); from.mu.Unlock() }
	}
	to.mu.Lock(); from.mu.Lock()
	return func() { from.mu.Unlock(); to.mu.Unlock() }
}

func (a *AssetService) tryTransfer(
	transferType TransferType,
	fromUser, toUser int64,
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

	unlock := lockPair(fromUser, toUser, assetKind, fromAsset, toAsset)
	defer unlock()

	switch transferType {
	case AvailableToAvailable:
		if checkBalance && fromAsset.available.Cmp(amount) < 0 {
			return false, nil
		}
		fromAsset.available.Sub(fromAsset.available, amount)
		toAsset.available.Add(toAsset.available, amount)
		return true, nil
	case AvailableToFrozen:
		if checkBalance && fromAsset.available.Cmp(amount) < 0 {
			return false, nil
		}
		fromAsset.available.Sub(fromAsset.available, amount)
		toAsset.frozen.Add(toAsset.frozen, amount)
		return true, nil
	case FrozenToAvailable:
		if checkBalance && fromAsset.frozen.Cmp(amount) < 0 {
			return false, nil
		}
		fromAsset.frozen.Sub(fromAsset.frozen, amount)
		toAsset.available.Add(toAsset.available, amount)
		return true, nil

	default:
		return false, fmt.Errorf("invalid transfer type: %v", transferType)
	}
}
