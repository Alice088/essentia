package llm_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/pkg/currency"
)

type BalancePolicy struct {
	soft currency.USD
	max  currency.USD
}

func NewBalancePolicy(config config.LLMManager) BalancePolicy {
	return BalancePolicy{
		soft: config.SoftBalanceLimit,
		max:  config.MaxBalanceLimit,
	}
}

func (bp *BalancePolicy) IsMaxReached(balance *BalanceManager) bool {
	return bp.max > 0 && balance.Current() <= bp.max
}

func (bp *BalancePolicy) IsSoftReached(balance *BalanceManager) bool {
	return bp.soft > 0 && balance.Current() <= bp.soft
}
