package portfolio

import (
	"errors"
	"fmt"
	"math"
	"sort"
)

type Stock struct {
	Symbol string
	price  float64
}

func NewStock(symbol string, price float64) *Stock {
	return &Stock{Symbol: symbol, price: price}
}

func (s *Stock) CurrentPrice() float64 { return s.price }

func (s *Stock) SetPrice(price float64) { s.price = price }

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type Order struct {
	Symbol string
	Side   Side
	Shares float64
	Amount float64
}

type Portfolio struct {
	stocks     map[string]*Stock
	holdings   map[string]float64 // symbol -> shares
	allocation map[string]float64 // symbol -> target weight
}

// Differences under one cent are float noise, not real orders.
const minOrderAmount = 0.01

func NewPortfolio() *Portfolio {
	return &Portfolio{
		stocks:     map[string]*Stock{},
		holdings:   map[string]float64{},
		allocation: map[string]float64{},
	}
}

func (p *Portfolio) AddHolding(stock *Stock, shares float64) error {
	if shares < 0 {
		return fmt.Errorf("%s: shares cannot be negative", stock.Symbol)
	}
	p.stocks[stock.Symbol] = stock
	p.holdings[stock.Symbol] += shares
	return nil
}

// SetAllocation receives *Stock instead of symbols because the portfolio
// needs the price of stocks it doesn't own yet.
func (p *Portfolio) SetAllocation(weights map[*Stock]float64) error {
	total := 0.0
	for stock, w := range weights {
		if w < 0 {
			return fmt.Errorf("%s: weight cannot be negative", stock.Symbol)
		}
		total += w
	}
	if math.Abs(total-1) > 1e-9 {
		return fmt.Errorf("weights must add up to 1, got %.4f", total)
	}

	p.allocation = map[string]float64{}
	for stock, w := range weights {
		p.stocks[stock.Symbol] = stock
		p.allocation[stock.Symbol] = w
	}
	return nil
}

func (p *Portfolio) TotalValue() float64 {
	total := 0.0
	for symbol, shares := range p.holdings {
		total += shares * p.stocks[symbol].CurrentPrice()
	}
	return total
}

// Rebalance returns the orders needed to reach the target allocation using
// only the money already invested. It doesn't modify the portfolio.
// Stocks held but not in the allocation have a target of 0, so they're sold.
func (p *Portfolio) Rebalance() ([]Order, error) {
	if len(p.allocation) == 0 {
		return nil, errors.New("portfolio has no target allocation")
	}
	for symbol, stock := range p.stocks {
		if stock.CurrentPrice() <= 0 {
			return nil, fmt.Errorf("%s: price must be greater than 0", symbol)
		}
	}

	total := p.TotalValue()
	var orders []Order

	for symbol, stock := range p.stocks {
		price := stock.CurrentPrice()
		diff := total*p.allocation[symbol] - p.holdings[symbol]*price
		if math.Abs(diff) < minOrderAmount {
			continue
		}

		side := Buy
		if diff < 0 {
			side = Sell
		}
		orders = append(orders, Order{
			Symbol: symbol,
			Side:   side,
			Shares: math.Abs(diff) / price,
			Amount: math.Abs(diff),
		})
	}

	// Sells first, since that money pays for the buys.
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Side != orders[j].Side {
			return orders[i].Side == Sell
		}
		return orders[i].Symbol < orders[j].Symbol
	})
	return orders, nil
}
