package portfolio

import (
	"math"
	"testing"
)

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-6
}

func TestRebalance_AlreadyBalanced(t *testing.T) {
	meta := NewStock("META", 100)
	aapl := NewStock("AAPL", 200)

	p := NewPortfolio()
	p.AddHolding(meta, 4)
	p.AddHolding(aapl, 3)
	if err := p.SetAllocation(map[*Stock]float64{meta: 0.4, aapl: 0.6}); err != nil {
		t.Fatal(err)
	}

	orders, err := p.Rebalance()
	if err != nil {
		t.Fatal(err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected no orders, got %+v", orders)
	}
}

func TestRebalance_SellOverweightBuyUnderweight(t *testing.T) {
	meta := NewStock("META", 100)
	aapl := NewStock("AAPL", 200)

	p := NewPortfolio()
	p.AddHolding(meta, 7)
	p.AddHolding(aapl, 1.5)
	p.SetAllocation(map[*Stock]float64{meta: 0.4, aapl: 0.6})

	orders, err := p.Rebalance()
	if err != nil {
		t.Fatal(err)
	}
	want := []Order{
		{Symbol: "META", Side: Sell, Shares: 3, Amount: 300},
		{Symbol: "AAPL", Side: Buy, Shares: 1.5, Amount: 300},
	}
	assertOrders(t, orders, want)
}

func TestRebalance_PriceChangeUnbalances(t *testing.T) {
	meta := NewStock("META", 100)
	aapl := NewStock("AAPL", 100)

	p := NewPortfolio()
	p.AddHolding(meta, 5)
	p.AddHolding(aapl, 5)
	p.SetAllocation(map[*Stock]float64{meta: 0.5, aapl: 0.5})

	// META doubles, so the portfolio drifts away from 50/50.
	meta.SetPrice(200)

	orders, err := p.Rebalance()
	if err != nil {
		t.Fatal(err)
	}
	want := []Order{
		{Symbol: "META", Side: Sell, Shares: 1.25, Amount: 250},
		{Symbol: "AAPL", Side: Buy, Shares: 2.5, Amount: 250},
	}
	assertOrders(t, orders, want)
}

func TestRebalance_SellsStockNotInAllocation(t *testing.T) {
	meta := NewStock("META", 100)
	tsla := NewStock("TSLA", 50)

	p := NewPortfolio()
	p.AddHolding(meta, 5)
	p.AddHolding(tsla, 10)
	p.SetAllocation(map[*Stock]float64{meta: 1})

	orders, err := p.Rebalance()
	if err != nil {
		t.Fatal(err)
	}
	want := []Order{
		{Symbol: "TSLA", Side: Sell, Shares: 10, Amount: 500},
		{Symbol: "META", Side: Buy, Shares: 5, Amount: 500},
	}
	assertOrders(t, orders, want)
}

func TestRebalance_BuysAllocatedStockNotHeld(t *testing.T) {
	meta := NewStock("META", 100)
	aapl := NewStock("AAPL", 250)

	p := NewPortfolio()
	p.AddHolding(meta, 10)
	p.SetAllocation(map[*Stock]float64{meta: 0.5, aapl: 0.5})

	orders, err := p.Rebalance()
	if err != nil {
		t.Fatal(err)
	}
	want := []Order{
		{Symbol: "META", Side: Sell, Shares: 5, Amount: 500},
		{Symbol: "AAPL", Side: Buy, Shares: 2, Amount: 500},
	}
	assertOrders(t, orders, want)
}

func TestSetAllocation_Invalid(t *testing.T) {
	meta := NewStock("META", 100)
	aapl := NewStock("AAPL", 100)
	p := NewPortfolio()

	if err := p.SetAllocation(map[*Stock]float64{meta: 0.4, aapl: 0.4}); err == nil {
		t.Error("expected error when weights do not add up to 1")
	}
	if err := p.SetAllocation(map[*Stock]float64{meta: 1.2, aapl: -0.2}); err == nil {
		t.Error("expected error for negative weight")
	}
}

func TestRebalance_Errors(t *testing.T) {
	meta := NewStock("META", 100)

	p := NewPortfolio()
	p.AddHolding(meta, 1)
	if _, err := p.Rebalance(); err == nil {
		t.Error("expected error when there is no allocation")
	}

	p.SetAllocation(map[*Stock]float64{meta: 1})
	meta.SetPrice(0)
	if _, err := p.Rebalance(); err == nil {
		t.Error("expected error when a price is 0")
	}
}

func TestAddHolding_NegativeShares(t *testing.T) {
	p := NewPortfolio()
	if err := p.AddHolding(NewStock("META", 100), -1); err == nil {
		t.Error("expected error for negative shares")
	}
}

func assertOrders(t *testing.T, got, want []Order) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected %d orders, got %d: %+v", len(want), len(got), got)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Symbol != w.Symbol || g.Side != w.Side ||
			!almostEqual(g.Shares, w.Shares) || !almostEqual(g.Amount, w.Amount) {
			t.Errorf("order %d: expected %+v, got %+v", i, w, g)
		}
	}
}
