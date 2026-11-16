# Portfolio rebalance

Fintual pre-task, in Go.

```bash
go test -v ./...
```

## Approach

For each stock that is held or part of the allocation:

```
diff = total_value * target_weight - current_value
```

Positive means buy, negative means sell, and `|diff| / price` is the number of shares.
Sells come first in the result because their money pays for the buys.

`Rebalance()` only returns the orders. It doesn't change the portfolio.

## Assumptions

- Fractional shares are allowed.
- No cash comes in or out: total sells equal total buys.
- A stock held but not in the allocation is sold completely.
- Weights must add up to 1 and prices must be greater than 0. Otherwise it returns an error.
- Money uses `float64`. In production I'd use integer cents or a decimal type.

Fees, minimum order sizes and drift thresholds are out of scope.

## LLM usage

I used Claude to help with this. The conversation is shared with the application.
