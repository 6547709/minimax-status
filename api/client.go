package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Config struct {
	APIKey string
	APIURL string
}

type Client struct {
	Config Config
	HTTP   *http.Client
}

type ModelRemain struct {
	ModelName                    string  `json:"model_name"`
	StartTime                    int64   `json:"start_time"`
	EndTime                      int64   `json:"end_time"`
	RemainsTime                  int64   `json:"remains_time"`
	CurrentIntervalUsageCount    int     `json:"current_interval_usage_count"`
	CurrentIntervalTotalCount    int     `json:"current_interval_total_count"`
	CurrentWeeklyUsageCount      int     `json:"current_weekly_usage_count"`
	CurrentWeeklyTotalCount      int     `json:"current_weekly_total_count"`
	WeeklyRemainsTime            int64   `json:"weekly_remains_time"`
}

type TokenPlanResponse struct {
	ModelRemains []ModelRemain `json:"model_remains"`
}

type BillingRecord struct {
	ConsumeToken int   `json:"consume_token"`
	CreatedAt    int64 `json:"created_at"`
}

type BillingResponse struct {
	ChargeRecords []BillingRecord `json:"charge_records"`
}

type UsageStats struct {
	LastDay   int64
	Weekly    int64
	Monthly   int64
}

type SubscriptionDetails struct {
	CurrentSubscribe struct {
		CurrentSubscribeEndTime string `json:"current_subscribe_end_time"`
	} `json:"current_subscribe"`
}

func NewClient(cfg Config) *Client {
	return &Client{
		Config: cfg,
		HTTP: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetTokenPlan() (*TokenPlanResponse, error) {
	url := c.Config.APIURL + "/v1/token_plan/remains"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("invalid API key")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result TokenPlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetBillingRecords(page, limit int) (*BillingResponse, error) {
	url := c.Config.APIURL + "/account/amount"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	q := req.URL.Query()
	q.Add("page", fmt.Sprintf("%d", page))
	q.Add("limit", fmt.Sprintf("%d", limit))
	q.Add("aggregate", "false")
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result BillingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (c *Client) GetAllBillingRecords() ([]BillingRecord, error) {
	var allRecords []BillingRecord
	for page := 1; page <= 100; page++ {
		resp, err := c.GetBillingRecords(page, 100)
		if err != nil {
			break
		}
		records := resp.ChargeRecords
		if len(records) == 0 {
			break
		}
		allRecords = append(allRecords, records...)
		if len(records) < 100 {
			break
		}
	}
	return allRecords, nil
}

func (c *Client) GetAllBillingRecordsRaw() (interface{}, error) {
	url := c.Config.APIURL + "/account/amount"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	q := req.URL.Query()
	q.Add("page", "1")
	q.Add("limit", "100")
	q.Add("aggregate", "false")
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

func (c *Client) GetSubscriptionDetails() (*SubscriptionDetails, error) {
	url := c.Config.APIURL + "/v1/api/openplatform/charge/combo/cycle_audio_resource_package"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.Config.APIKey)
	q := req.URL.Query()
	q.Add("biz_line", "2")
	q.Add("cycle_type", "1")
	q.Add("resource_package_type", "7")
	req.URL.RawQuery = q.Encode()

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var result SubscriptionDetails
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &result, nil
}

func (c *Client) CalculateUsageStats(records []BillingRecord, since time.Time) UsageStats {
	stats := UsageStats{}
	monthStart := time.Date(since.Year(), since.Month(), 1, 0, 0, 0, 0, since.Location())
	yesterdayStart := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	weekStart := time.Now().AddDate(0, 0, -7)

	for _, r := range records {
		t := time.Unix(r.CreatedAt, 0)
		if t.Before(since) {
			continue
		}
		if t.After(yesterdayStart) && t.Before(yesterdayStart.Add(24*time.Hour)) {
			stats.LastDay += int64(r.ConsumeToken)
		}
		if t.After(weekStart) {
			stats.Weekly += int64(r.ConsumeToken)
		}
		if t.After(monthStart) {
			stats.Monthly += int64(r.ConsumeToken)
		}
	}
	return stats
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func GetConfig() Config {
	return Config{
		APIKey: getEnv("MINIMAX_API_KEY", ""),
		APIURL: getEnv("MINIMAX_API_URL", "https://www.minimaxi.com"),
	}
}
