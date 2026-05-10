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