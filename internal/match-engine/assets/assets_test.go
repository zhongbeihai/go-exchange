package assets

import (
	"errors"
	"math/big"
	"testing"
	"time"

	"go-exchange/common/enum"
	"go-exchange/common/serror"
)

// ---- stub registry 实现 enum.AssetRegistry ----

type stubRegistry struct {
	allowed map[enum.AssetID]bool
}

func (r *stubRegistry) Exists(id enum.AssetID) bool {
	return r != nil && r.allowed[id]
}

// ---- 辅助函数 ----

func setAvail(t *testing.T, s *AssetService, user int64, id enum.AssetID, v string) {
	t.Helper()
	a := s.getOrCreateAssetType(user, id)
	a.mu.Lock()
	defer a.mu.Unlock()
	if f, ok := new(big.Float).SetString(v); ok {
		a.available = f
	} else {
		t.Fatalf("bad float: %s", v)
	}
}

func setFrozen(t *testing.T, s *AssetService, user int64, id enum.AssetID, v string) {
	t.Helper()
	a := s.getOrCreateAssetType(user, id)
	a.mu.Lock()
	defer a.mu.Unlock()
	if f, ok := new(big.Float).SetString(v); ok {
		a.frozen = f
	} else {
		t.Fatalf("bad float: %s", v)
	}
}

func getAvail(s *AssetService, user int64, id enum.AssetID) string {
	a := s.getOrCreateAssetType(user, id)
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.available.Text('f', -1)
}

func getFrozen(s *AssetService, user int64, id enum.AssetID) string {
	a := s.getOrCreateAssetType(user, id)
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.frozen.Text('f', -1)
}

func bf(s string) *big.Float {
	f, _ := new(big.Float).SetString(s)
	return f
}

// ---- 测试用例 ----

func TestTransfer_AvailableToAvailable_Sufficient(t *testing.T) {
	reg := enum.NewMemoryRegistry()
	svc := NewAssetService(reg)
	id := enum.AssetID("USD")

	setAvail(t, svc, 1, id, "100")
	setAvail(t, svc, 2, id, "0")

	ok, err := svc.Transfer(AvailableToAvailable, 1, 2, id, bf("30"), true)
	if err != nil || !ok {
		t.Fatalf("transfer failed: ok=%v err=%v", ok, err)
	}

	if got := getAvail(svc, 1, id); got != "70" {
		t.Fatalf("from.available = %s, want 70", got)
	}
	if got := getAvail(svc, 2, id); got != "30" {
		t.Fatalf("to.available = %s, want 30", got)
	}
	if got := getFrozen(svc, 1, id); got != "0" {
		t.Fatalf("from.frozen = %s, want 0", got)
	}
	if got := getFrozen(svc, 2, id); got != "0" {
		t.Fatalf("to.frozen = %s, want 0", got)
	}
}

func TestTransfer_AvailableToAvailable_Insufficient(t *testing.T) {
	reg := enum.NewMemoryRegistry()
	svc := NewAssetService(reg)
	id := enum.AssetID("USD")

	setAvail(t, svc, 1, id, "10")
	setAvail(t, svc, 2, id, "0")

	ok, err := svc.Transfer(AvailableToAvailable, 1, 2, id, bf("20"), true)
	if ok || !errors.Is(err, serror.ErrInsufficientAvailable) {
		t.Fatalf("want insufficient available, got ok=%v err=%v", ok, err)
	}
}

func TestFreeze_And_Unfreeze(t *testing.T) {
	reg := enum.NewMemoryRegistry()
	svc := NewAssetService(reg)
	id := enum.AssetID("USD")

	setAvail(t, svc, 1001, id, "50")
	setFrozen(t, svc, 1001, id, "0")

	// Freeze 30
	ok, err := svc.Freeze(1001, id, bf("30"))
	if err != nil || !ok {
		t.Fatalf("freeze failed: ok=%v err=%v", ok, err)
	}
	if got := getAvail(svc, 1001, id); got != "20" {
		t.Fatalf("available=%s, want 20", got)
	}
	if got := getFrozen(svc, 1001, id); got != "30" {
		t.Fatalf("frozen=%s, want 30", got)
	}

	// Unfreeze 20
	if _, err := svc.Unfreeze(1001, id, bf("20")); err != nil {
		t.Fatalf("unfreeze failed: %v", err)
	}
	if got := getAvail(svc, 1001, id); got != "40" {
		t.Fatalf("available=%s, want 40", got)
	}
	if got := getFrozen(svc, 1001, id); got != "10" {
		t.Fatalf("frozen=%s, want 10", got)
	}

	// Unfreeze 超过剩余冻结 -> 期望 ErrInsufficientFrozen
	_, err = svc.Unfreeze(1001, id, bf("20"))
	if !errors.Is(err, serror.ErrInsufficientFrozen) {
		t.Fatalf("want ErrInsufficientFrozen, got %v", err)
	}
}

func TestUnknownAssetRejected(t *testing.T) {
	reg := enum.NewMemoryRegistry()
	svc := NewAssetService(reg)

	ok, err := svc.Transfer(AvailableToAvailable, 1, 2, enum.AssetID("NEW"), bf("1"), true)
	if ok || !errors.Is(err, serror.ErrUnknownAsset) {
		t.Fatalf("want unknown asset error, got ok=%v err=%v", ok, err)
	}
}

func TestZeroAndNilAmount(t *testing.T) {
	reg := enum.NewMemoryRegistry()
	svc := NewAssetService(reg)
	id := enum.AssetID("USD")

	// 0 金额：应该直接成功且不变
	setAvail(t, svc, 1, id, "5")
	ok, err := svc.Transfer(AvailableToAvailable, 1, 2, id, bf("0"), true)
	if !ok || err != nil {
		t.Fatalf("zero amount should succeed: ok=%v err=%v", ok, err)
	}
	if got := getAvail(svc, 1, id); got != "5" {
		t.Fatalf("available changed: %s", got)
	}

	// nil 金额：应返回错误（你当前实现返回 fmt.Errorf）
	ok, err = svc.Transfer(AvailableToAvailable, 1, 2, id, nil, true)
	if ok || err == nil {
		t.Fatalf("nil amount should error, got ok=%v err=%v", ok, err)
	}
}

func TestConcurrentOppositeTransfers_NoDeadlock(t *testing.T) {
	reg := enum.NewMemoryRegistry()
	svc := NewAssetService(reg)
	id := enum.AssetID("USD")

	setAvail(t, svc, 1, id, "100")
	setAvail(t, svc, 2, id, "100")

	const N = 200 // 次数多一些，放大死锁概率

	done := make(chan struct{})
	go func() {
		for i := 0; i < N; i++ {
			// 1 -> 2
			_, _ = svc.Transfer(AvailableToAvailable, 1, 2, id, bf("1"), true)
			// 2 -> 1
			_, _ = svc.Transfer(AvailableToAvailable, 2, 1, id, bf("1"), true)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("potential deadlock detected (timeout)")
	}

	if got := getAvail(svc, 1, id); got != "100" {
		t.Fatalf("user1 available=%s, want 100", got)
	}
	if got := getAvail(svc, 2, id); got != "100" {
		t.Fatalf("user2 available=%s, want 100", got)
	}
}
