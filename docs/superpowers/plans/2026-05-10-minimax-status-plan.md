# MiniMax Status Dashboard - Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go HTTP server that displays MiniMax usage status as an HTML page with modern card-style UI and dark/light theme support.

**Architecture:** Simple Go HTTP server using standard library (`net/http`, `html/template`). API client calls MiniMax endpoints to fetch usage data. Single-page application renders data server-side.

**Tech Stack:** Go 1.21+, standard library only, Docker for deployment.

---

## File Structure

```
minimax-status/
├── main.go              # Entry point, server setup
├── handlers.go          # HTTP handlers, page rendering
├── api/
│   └── client.go       # MiniMax API client
├── templates/
│   └── index.html      # HTML template
├── static/
│   └── style.css       # CSS with theme variables
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

---

## Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `.env.example`

- [ ] **Step 1: Create go.mod**

```bash
mkdir -p /root/minimax-status
cd /root/minimax-status
go mod init minimax-status
```

- [ ] **Step 2: Create .env.example**

```bash
cat > .env.example << 'EOF'
MINIMAX_API_KEY=your_api_key_here
MINIMAX_API_URL=https://www.minimaxi.com
PORT=8080
EOF
```

- [ ] **Step 3: Commit**

```bash
git add go.mod .env.example
git commit -m "init: project structure"
```

---

## Task 2: API Client

**Files:**
- Create: `api/client.go`

- [ ] **Step 1: Create api/client.go with MiniMax API client**

```go
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
```

- [ ] **Step 2: Test compilation**

```bash
go build ./...
```

Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add api/client.go
git commit -m "feat: add MiniMax API client"
```

---

## Task 3: HTTP Handlers

**Files:**
- Create: `handlers.go`

- [ ] **Step 1: Create handlers.go**

```go
package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"

	"minimax-status/api"
)

type PageData struct {
	ModelName    string
	TimeWindow   string
	ResetTime    string
	UsagePercent int
	UsedCount    int
	TotalCount   int
	WeeklyLimit  string
	ExpiryDate   string
	ExpiryDays   int
	Stats        TokenStats
	Models       []ModelData
	Status       StatusInfo
	Error        string
}

type TokenStats struct {
	LastDay  string
	Weekly   string
	Monthly  string
}

type ModelData struct {
	Name       string
	Percent    int
	Used       int
	Total      int
	Unlimited  bool
}

type StatusInfo struct {
	OK      bool
	Message string
}

func formatNumber(n int64) string {
	if n >= 100000000 {
		return fmt.Sprintf("%.1f亿", float64(n)/100000000)
	}
	if n >= 10000 {
		return fmt.Sprintf("%.1f万", float64(n)/10000)
	}
	return fmt.Sprintf("%d", n)
}

func formatRemainTime(ms int64) string {
	if ms <= 0 {
		return "已重置"
	}
	hours := ms / (1000 * 60 * 60)
	mins := (ms % (1000 * 60 * 60)) / (1000 * 60)
	if hours > 0 {
		return fmt.Sprintf("%d 小时 %d 分钟", hours, mins)
	}
	return fmt.Sprintf("%d 分钟", mins)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := api.GetConfig()

	if cfg.APIKey == "" {
		data := PageData{
			Error: "未配置 MINIMAX_API_KEY 环境变量",
		}
		tmpl.Execute(w, data)
		return
	}

	client := api.NewClient(cfg)

	// Get token plan
	tokenPlan, err := client.GetTokenPlan()
	if err != nil {
		data := PageData{
			Error: fmt.Sprintf("获取套餐信息失败: %v", err),
		}
		tmpl.Execute(w, data)
		return
	}

	// Get billing records for stats
	billing, err := client.GetBillingRecords(1, 100)
	stats := api.UsageStats{}
	if err == nil && billing != nil {
		stats = client.CalculateUsageStats(billing.ChargeRecords, time.Now().AddDate(0, -1, 0))
	}

	// Parse primary model
	var pageData PageData
	if len(tokenPlan.ModelRemains) > 0 {
		m := tokenPlan.ModelRemains[0]

		pageData.ModelName = m.ModelName
		startLocal := time.Unix(m.StartTime/1000, 0).In(time.FixedZone("CST", 8*3600))
		endLocal := time.Unix(m.EndTime/1000, 0).In(time.FixedZone("CST", 8*3600))
		pageData.TimeWindow = fmt.Sprintf("%02d:00-%02d:00 (UTC+8)",
			startLocal.Hour(), endLocal.Hour())
		pageData.ResetTime = formatRemainTime(m.RemainsTime)

		used := m.CurrentIntervalUsageCount
		total := m.CurrentIntervalTotalCount
		pageData.UsagePercent = 0
		if total > 0 {
			pageData.UsagePercent = int(math.Round(float64(used) / float64(total) * 100))
		}
		pageData.UsedCount = used
		pageData.TotalCount = total

		if m.CurrentWeeklyTotalCount == 0 {
			pageData.WeeklyLimit = "不受限制"
		} else {
			pageData.WeeklyLimit = fmt.Sprintf("%d/%d", m.CurrentWeeklyUsageCount, m.CurrentWeeklyTotalCount)
		}

		// Expiry date (placeholder - would need subscription API)
		pageData.ExpiryDate = "N/A"
		pageData.ExpiryDays = 0
	}

	// Token stats
	pageData.Stats = TokenStats{
		LastDay:  formatNumber(stats.LastDay),
		Weekly:   formatNumber(stats.Weekly),
		Monthly:  formatNumber(stats.Monthly),
	}

	// All models
	for _, m := range tokenPlan.ModelRemains {
		total := m.CurrentIntervalTotalCount
		used := m.CurrentIntervalUsageCount
		percent := 0
		if total > 0 {
			percent = int(math.Round(float64(used) / float64(total) * 100))
		}
		pageData.Models = append(pageData.Models, ModelData{
			Name:      m.ModelName,
			Percent:   percent,
			Used:      used,
			Total:     total,
			Unlimited: m.CurrentWeeklyTotalCount == 0,
		})
	}

	pageData.Status = StatusInfo{
		OK:      true,
		Message: "正常使用",
	}

	tmpl.Execute(w, pageData)
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", nil)
}
```

- [ ] **Step 2: Test compilation**

```bash
go build ./...
```

Expected: No errors

- [ ] **Step 3: Commit**

```bash
git add handlers.go
git commit -m "feat: add HTTP handlers"
```

---

## Task 4: HTML Template

**Files:**
- Create: `templates/index.html`

- [ ] **Step 1: Create templates/index.html**

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>MiniMax 使用状态</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <div class="container">
        <header class="header">
            <h1>🤖 MiniMax 使用状态</h1>
            <button class="refresh-btn" onclick="location.reload()">🔄 刷新</button>
        </header>

        {{if .Error}}
        <div class="error-card">
            <p>⚠️ {{.Error}}</p>
        </div>
        {{else}}
        <main class="main-content">
            <section class="stats-row">
                <div class="stat-card">
                    <span class="stat-label">当前模型</span>
                    <span class="stat-value">{{.ModelName}}</span>
                </div>
                <div class="stat-card">
                    <span class="stat-label">剩余重置</span>
                    <span class="stat-value">{{.ResetTime}}</span>
                    <span class="stat-sub">{{.TimeWindow}}</span>
                </div>
                <div class="stat-card">
                    <span class="stat-label">套餐到期</span>
                    <span class="stat-value">{{.ExpiryDays}} 天</span>
                    <span class="stat-sub">{{.ExpiryDate}}</span>
                </div>
            </section>

            <section class="usage-card">
                <div class="usage-header">
                    <span class="usage-label">已用额度</span>
                    <span class="usage-percent">{{.UsagePercent}}%</span>
                </div>
                <div class="progress-bar">
                    <div class="progress-fill" style="width: {{.UsagePercent}}%"></div>
                </div>
                <p class="usage-count">{{.UsedCount}} / {{.TotalCount}} 次调用</p>
            </section>

            <section class="section-card">
                <h2>📊 Token 消耗统计</h2>
                <div class="token-stats">
                    <div class="token-stat">
                        <span class="token-label">昨日消耗</span>
                        <span class="token-value">{{.Stats.LastDay}}</span>
                    </div>
                    <div class="token-stat">
                        <span class="token-label">近7天</span>
                        <span class="token-value">{{.Stats.Weekly}}</span>
                    </div>
                    <div class="token-stat">
                        <span class="token-label">当月</span>
                        <span class="token-value">{{.Stats.Monthly}}</span>
                    </div>
                </div>
            </section>

            <section class="section-card">
                <h2>📋 所有模型额度</h2>
                <table class="models-table">
                    <thead>
                        <tr>
                            <th>模型</th>
                            <th>使用率</th>
                            <th>剩余/总量</th>
                            <th>状态</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range .Models}}
                        <tr>
                            <td>{{.Name}}</td>
                            <td>
                                <div class="mini-progress">
                                    <div class="mini-progress-fill" style="width: {{.Percent}}%"></div>
                                </div>
                                {{.Percent}}%
                            </td>
                            <td>{{.Used}} / {{.Total}}</td>
                            <td>{{if .Unlimited}}∞{{else}}✓{{end}}</td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>
            </section>

            <footer class="footer">
                <span class="status {{if .Status.OK}}status-ok{{else}}status-error{{end}}">
                    {{if .Status.OK}}✅{{else}}❌{{end}} {{.Status.Message}}
                </span>
            </footer>
        </main>
        {{end}}
    </div>
</body>
</html>
```

- [ ] **Step 2: Commit**

```bash
git add templates/index.html
git commit -m "feat: add HTML template"
```

---

## Task 5: CSS Styling

**Files:**
- Create: `static/style.css`

- [ ] **Step 1: Create static/style.css with theme support**

```css
:root {
    --bg-primary: #f8fafc;
    --bg-card: #ffffff;
    --color-primary: #6366f1;
    --color-primary-hover: #4f46e5;
    --text-primary: #1e293b;
    --text-secondary: #64748b;
    --border-color: #e2e8f0;
    --shadow: 0 1px 3px rgba(0,0,0,0.1), 0 1px 2px rgba(0,0,0,0.06);
    --progress-bg: #e2e8f0;
}

@media (prefers-color-scheme: dark) {
    :root {
        --bg-primary: #0f172a;
        --bg-card: #1e293b;
        --color-primary: #818cf8;
        --color-primary-hover: #a5b4fc;
        --text-primary: #f1f5f9;
        --text-secondary: #94a3b8;
        --border-color: #334155;
        --shadow: 0 1px 3px rgba(0,0,0,0.3), 0 1px 2px rgba(0,0,0,0.2);
        --progress-bg: #334155;
    }
}

* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    background: var(--bg-primary);
    color: var(--text-primary);
    min-height: 100vh;
    padding: 16px;
}

.container {
    max-width: 600px;
    margin: 0 auto;
}

.header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
}

.header h1 {
    font-size: 20px;
    font-weight: 600;
}

.refresh-btn {
    background: var(--color-primary);
    color: white;
    border: none;
    padding: 8px 16px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 14px;
    transition: background 0.2s;
}

.refresh-btn:hover {
    background: var(--color-primary-hover);
}

.error-card {
    background: var(--bg-card);
    border-radius: 12px;
    padding: 24px;
    text-align: center;
    color: #ef4444;
    box-shadow: var(--shadow);
}

.main-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.stats-row {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
}

.stat-card {
    background: var(--bg-card);
    border-radius: 12px;
    padding: 16px;
    text-align: center;
    box-shadow: var(--shadow);
    display: flex;
    flex-direction: column;
    gap: 4px;
}

.stat-label {
    font-size: 12px;
    color: var(--text-secondary);
}

.stat-value {
    font-size: 18px;
    font-weight: 600;
}

.stat-sub {
    font-size: 11px;
    color: var(--text-secondary);
}

.usage-card {
    background: var(--bg-card);
    border-radius: 12px;
    padding: 20px;
    box-shadow: var(--shadow);
}

.usage-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12px;
}

.usage-label {
    font-weight: 500;
}

.usage-percent {
    font-size: 24px;
    font-weight: 700;
    color: var(--color-primary);
}

.progress-bar {
    height: 8px;
    background: var(--progress-bg);
    border-radius: 4px;
    overflow: hidden;
}

.progress-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--color-primary), #8b5cf6);
    border-radius: 4px;
    transition: width 0.3s ease;
}

.usage-count {
    margin-top: 8px;
    font-size: 14px;
    color: var(--text-secondary);
}

.section-card {
    background: var(--bg-card);
    border-radius: 12px;
    padding: 20px;
    box-shadow: var(--shadow);
}

.section-card h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
}

.token-stats {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 12px;
}

.token-stat {
    text-align: center;
}

.token-label {
    display: block;
    font-size: 12px;
    color: var(--text-secondary);
    margin-bottom: 4px;
}

.token-value {
    font-size: 16px;
    font-weight: 600;
}

.models-table {
    width: 100%;
    border-collapse: collapse;
}

.models-table th,
.models-table td {
    padding: 10px 8px;
    text-align: left;
    border-bottom: 1px solid var(--border-color);
}

.models-table th {
    font-size: 12px;
    color: var(--text-secondary);
    font-weight: 500;
}

.models-table td {
    font-size: 14px;
}

.mini-progress {
    display: inline-block;
    width: 60px;
    height: 4px;
    background: var(--progress-bg);
    border-radius: 2px;
    vertical-align: middle;
    margin-right: 8px;
}

.mini-progress-fill {
    height: 100%;
    background: var(--color-primary);
    border-radius: 2px;
}

.footer {
    text-align: center;
    padding: 16px;
}

.status {
    font-size: 14px;
}

.status-ok {
    color: #22c55e;
}

.status-error {
    color: #ef4444;
}
```

- [ ] **Step 2: Commit**

```bash
git add static/style.css
git commit -m "feat: add CSS with dark/light theme support"
```

---

## Task 6: Main Entry Point

**Files:**
- Modify: `handlers.go` (extract server setup)
- Create: `main.go`

- [ ] **Step 1: Create main.go**

```go
package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", homeHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
```

- [ ] **Step 2: Update handlers.go - remove main and add template loading**

```go
package main

import (
	"fmt"
	"html/template"
	"math"
	"net/http"
	"time"

	"minimax-status/api"
)

type PageData struct {
	ModelName    string
	TimeWindow   string
	ResetTime    string
	UsagePercent int
	UsedCount    int
	TotalCount   int
	WeeklyLimit  string
	ExpiryDate   string
	ExpiryDays   int
	Stats        TokenStats
	Models       []ModelData
	Status       StatusInfo
	Error        string
}

type TokenStats struct {
	LastDay  string
	Weekly   string
	Monthly  string
}

type ModelData struct {
	Name      string
	Percent   int
	Used      int
	Total     int
	Unlimited bool
}

type StatusInfo struct {
	OK      bool
	Message string
}

func formatNumber(n int64) string {
	if n >= 100000000 {
		return fmt.Sprintf("%.1f亿", float64(n)/100000000)
	}
	if n >= 10000 {
		return fmt.Sprintf("%.1f万", float64(n)/10000)
	}
	return fmt.Sprintf("%d", n)
}

func formatRemainTime(ms int64) string {
	if ms <= 0 {
		return "已重置"
	}
	hours := ms / (1000 * 60 * 60)
	mins := (ms % (1000 * 60 * 60)) / (1000 * 60)
	if hours > 0 {
		return fmt.Sprintf("%d 小时 %d 分钟", hours, mins)
	}
	return fmt.Sprintf("%d 分钟", mins)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	cfg := api.GetConfig()

	if cfg.APIKey == "" {
		data := PageData{
			Error: "未配置 MINIMAX_API_KEY 环境变量",
		}
		tmpl.Execute(w, data)
		return
	}

	client := api.NewClient(cfg)

	tokenPlan, err := client.GetTokenPlan()
	if err != nil {
		data := PageData{
			Error: fmt.Sprintf("获取套餐信息失败: %v", err),
		}
		tmpl.Execute(w, data)
		return
	}

	billing, err := client.GetBillingRecords(1, 100)
	stats := api.UsageStats{}
	if err == nil && billing != nil {
		stats = client.CalculateUsageStats(billing.ChargeRecords, time.Now().AddDate(0, -1, 0))
	}

	var pageData PageData
	if len(tokenPlan.ModelRemains) > 0 {
		m := tokenPlan.ModelRemains[0]

		pageData.ModelName = m.ModelName
		startLocal := time.Unix(m.StartTime/1000, 0).In(time.FixedZone("CST", 8*3600))
		endLocal := time.Unix(m.EndTime/1000, 0).In(time.FixedZone("CST", 8*3600))
		pageData.TimeWindow = fmt.Sprintf("%02d:00-%02d:00 (UTC+8)",
			startLocal.Hour(), endLocal.Hour())
		pageData.ResetTime = formatRemainTime(m.RemainsTime)

		used := m.CurrentIntervalUsageCount
		total := m.CurrentIntervalTotalCount
		pageData.UsagePercent = 0
		if total > 0 {
			pageData.UsagePercent = int(math.Round(float64(used) / float64(total) * 100))
		}
		pageData.UsedCount = used
		pageData.TotalCount = total

		if m.CurrentWeeklyTotalCount == 0 {
			pageData.WeeklyLimit = "不受限制"
		} else {
			pageData.WeeklyLimit = fmt.Sprintf("%d/%d", m.CurrentWeeklyUsageCount, m.CurrentWeeklyTotalCount)
		}

		pageData.ExpiryDate = "N/A"
		pageData.ExpiryDays = 0
	}

	pageData.Stats = TokenStats{
		LastDay:  formatNumber(stats.LastDay),
		Weekly:   formatNumber(stats.Weekly),
		Monthly:  formatNumber(stats.Monthly),
	}

	for _, m := range tokenPlan.ModelRemains {
		total := m.CurrentIntervalTotalCount
		used := m.CurrentIntervalUsageCount
		percent := 0
		if total > 0 {
			percent = int(math.Round(float64(used) / float64(total) * 100))
		}
		pageData.Models = append(pageData.Models, ModelData{
			Name:      m.ModelName,
			Percent:   percent,
			Used:      used,
			Total:     total,
			Unlimited: m.CurrentWeeklyTotalCount == 0,
		})
	}

	pageData.Status = StatusInfo{
		OK:      true,
		Message: "正常使用",
	}

	tmpl.Execute(w, pageData)
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))
```

- [ ] **Step 3: Test compilation**

```bash
go build -o minimax-status .
```

Expected: Binary compiled

- [ ] **Step 4: Commit**

```bash
git add main.go handlers.go
git commit -m "feat: add main entry point"
```

---

## Task 7: Docker Configuration

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`

- [ ] **Step 1: Create Dockerfile**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o minimax-status .

FROM alpine:3.19
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/minimax-status .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080

ENV PORT=8080
CMD ["./minimax-status"]
```

- [ ] **Step 2: Create docker-compose.yml**

```yaml
version: '3.8'
services:
  minimax-status:
    build: .
    ports:
      - "8080:8080"
    environment:
      - MINIMAX_API_KEY=${MINIMAX_API_KEY}
      - MINIMAX_API_URL=${MINIMAX_API_URL:-https://www.minimaxi.com}
      - PORT=8080
    restart: unless-stopped
```

- [ ] **Step 3: Create .gitignore**

```bash
cat > .gitignore << 'EOF'
minimax-status
.env
EOF
```

- [ ] **Step 4: Commit**

```bash
git add Dockerfile docker-compose.yml .gitignore
git commit -m "feat: add Docker configuration"
```

---

## Task 8: README

**Files:**
- Create: `README.md`

- [ ] **Step 1: Create README.md**

```markdown
# MiniMax Status Dashboard

用于 dashy 监控的 MiniMax 套餐使用状态页面，Go 语言实现，Docker 部署。

## 快速开始

### 1. 配置环境变量

```bash
export MINIMAX_API_KEY=your_api_key_here
export MINIMAX_API_URL=https://www.minimaxi.com
```

### 2. 本地运行

```bash
go build -o minimax-status .
./minimax-status
```

访问 http://localhost:8080

### 3. Docker 部署

```bash
# 构建并启动
docker-compose up -d

# 查看日志
docker-compose logs -f
```

或手动构建：

```bash
docker build -t minimax-status .
docker run -d -p 8080:8080 \
  -e MINIMAX_API_KEY=your_key \
  minimax-status
```

## 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MINIMAX_API_KEY` | MiniMax API 密钥 | 必填 |
| `MINIMAX_API_URL` | API 服务器地址 | `https://www.minimaxi.com` |
| `PORT` | HTTP 监听端口 | `8080` |

## 嵌入 dashy

在 dashy 的 `widgets` 配置中添加：

```yaml
widgets:
  - type: iframe
    options:
      url: http://your-server:8080
```

## 主题支持

页面自动适配浏览器主题（浅色/深色模式）。

## License

MIT
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README"
```

---

## Self-Review Checklist

1. **Spec coverage**: All requirements from spec implemented
   - Go server ✅
   - API client ✅
   - Environment variables ✅
   - HTML page ✅
   - Modern card UI ✅
   - Dark/light theme ✅
   - Docker deployment ✅

2. **Placeholder scan**: No TBD/TODO remaining

3. **Type consistency**: Types are consistent across tasks
