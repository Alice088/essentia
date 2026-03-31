package llm_provider

import (
	"Alice088/essentia/pkg/currency"
	"Alice088/essentia/pkg/retry"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

// DeepSeekProvider implements BalanceProvider for DeepSeek API.
type DeepSeekProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewDeepSeekProvider creates a new DeepSeekProvider.
func NewDeepSeekProvider(apiKey, baseURL string) *DeepSeekProvider {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	return &DeepSeekProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetBalance retrieves current account balance from DeepSeek API.
func (d *DeepSeekProvider) GetBalance(ctx context.Context) (currency.USD, error) {
	req, err := http.NewRequest(http.MethodGet, d.baseURL+"/user/balance", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	var res *http.Response
	err = retry.Exponential(ctx, retry.ExponentialOpts{
		Seconds: 2,
		Tries:   3,
		Fn: func(ctx context.Context) error {
			res, err = d.client.Do(req)
			return err
		},
	})
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var balanceRes balanceResponse
	err = json.Unmarshal(body, &balanceRes)
	if err != nil {
		return 0, err
	}

	if !balanceRes.IsAvailable {
		return 0, errors.New("balance not available")
	}

	if len(balanceRes.BalanceInfos) == 0 {
		return 0, errors.New("balance info is empty")
	}

	info := balanceRes.BalanceInfos[0]

	if info.Currency != "USD" {
		return 0, errors.New("incorrect currency")
	}

	f, err := strconv.ParseFloat(info.TotalBalance, 64)
	if err != nil {
		return 0, err
	}

	return f, nil
}

type balanceResponse struct {
	IsAvailable  bool `json:"is_available"`
	BalanceInfos []struct {
		Currency     string `json:"currency"`
		TotalBalance string `json:"total_balance"`
	} `json:"balance_infos"`
}
