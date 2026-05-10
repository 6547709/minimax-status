package main

import (
	"fmt"
	"html/template"
	"log"
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

	// Get billing records for stats (all pages)
	records, err := client.GetAllBillingRecords()
	if err != nil {
		log.Printf("Billing records error: %v", err)
	}
	stats := client.CalculateUsageStats(records, time.Now().AddDate(0, -1, 0))
	log.Printf("Got %d billing records, stats: LastDay=%d, Weekly=%d, Monthly=%d", len(records), stats.LastDay, stats.Weekly, stats.Monthly)

	// Get subscription details for expiry date
	subscription, subErr := client.GetSubscriptionDetails()
	if subErr != nil {
		log.Printf("Subscription error: %v", subErr)
	}
	log.Printf("Subscription response: %+v", subscription)
	var expiryDate string
	var expiryDays int
	if subscription != nil && subscription.CurrentSubscribe.CurrentSubscribeEndTime != "" {
		expiryDate = subscription.CurrentSubscribe.CurrentSubscribeEndTime
		expiryTime, _ := time.Parse("2006-01-02 15:04:05", expiryDate)
		if expiryTime.IsZero() {
			expiryTime, _ = time.Parse(time.RFC3339, expiryDate)
		}
		if !expiryTime.IsZero() {
			expiryDays = int(math.Ceil(time.Until(expiryTime).Hours() / 24))
		}
		log.Printf("Expiry: date=%s, days=%d", expiryDate, expiryDays)
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

		pageData.ExpiryDate = expiryDate
		pageData.ExpiryDays = expiryDays
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