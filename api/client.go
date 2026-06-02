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

// ModelRemain 字段名直接对齐官方接口返回的 snake_case
// 当前接口不再返回 total/usage 实际数值（恒为 0），
// 新的数据源是 remaining_percent / status / boost_permille
type ModelRemain struct {
	StartTime                       int64  `json:"start_time"`
	EndTime                         int64  `json:"end_time"`
	RemainsTime                     int64  `json:"remains_time"`
	CurrentIntervalTotalCount       int    `json:"current_interval_total_count"`
	CurrentIntervalUsageCount       int    `json:"current_interval_usage_count"`
	CurrentIntervalRemainingPercent int    `json:"current_interval_remaining_percent"`
	CurrentIntervalStatus           int    `json:"current_interval_status"`
	ModelName                       string `json:"model_name"`
	CurrentWeeklyTotalCount         int    `json:"current_weekly_total_count"`
	CurrentWeeklyUsageCount         int    `json:"current_weekly_usage_count"`
	CurrentWeeklyRemainingPercent   int    `json:"current_weekly_remaining_percent"`
	CurrentWeeklyStatus             int    `json:"current_weekly_status"`
	WeeklyStartTime                 int64  `json:"weekly_start_time"`
	WeeklyEndTime                   int64  `json:"weekly_end_time"`
	WeeklyRemainsTime               int64  `json:"weekly_remains_time"`
	IntervalBoostPermille           int    `json:"interval_boost_permille"`
	WeeklyBoostPermille             int    `json:"weekly_boost_permille"`
}

type TokenPlanResponse struct {
	ModelRemains []ModelRemain `json:"model_remains"`
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
	req.Header.Set("Content-Type", "application/json")

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
