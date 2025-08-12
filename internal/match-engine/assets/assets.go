package assets

import (
	"math/big"
	"sync"
)

type TransferType int

const (
	AvailableToAvailable TransferType = iota
	AvailableToFrozen
	FrozenToAvailable
)

func (t TransferType) String() string {
	switch t {
	case AvailableToAvailable:
		return "AVAILABLE_TO_AVAILABLE"
	case AvailableToFrozen:
		return "AVAILABLE_TO_FROZEN"
	case FrozenToAvailable:
		return "FROZEN_TO_AVAILABLE"
	default:
		return "UNKNOWN"
	}
}

type AssetsEnum int

const (
	USD AssetsEnum = iota
)

type Asset struct {
	mu        sync.Mutex
	available *big.Float
	frozen    *big.Float
}

func NewAsset() *Asset {
	return NewAssetWithValues(big.NewFloat(0), big.NewFloat(0))
}

func NewAssetWithValues(available, frozen *big.Float) *Asset {
	return &Asset{
		available: available,
		frozen:    frozen,
	}
}

func (a *Asset) GetAvailable() *big.Float {
	return a.available
}

func (a *Asset) GetFrozen() *big.Float {
	return a.frozen
}
