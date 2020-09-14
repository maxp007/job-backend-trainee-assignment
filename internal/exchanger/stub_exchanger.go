package exchanger

import (
	"context"
	"github.com/shopspring/decimal"
)

type StubExchanger struct{}

func (se *StubExchanger) GetAmountInCurrency(ctx context.Context,
	amount decimal.Decimal,
	targetCurrencyName string) (amountInCurrency *decimal.Decimal, err error) {
	if targetCurrencyName == "UNKNOWN_CURRENCY" {
		return nil, ErrTargetCurrencyNameNotFound
	}
	r := decimal.NewFromFloat(750.0)
	return &r, nil
}
