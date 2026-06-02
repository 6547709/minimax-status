package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"minimax-status/api"
)

type PageData struct {
	ModelName     string
	ModelSubtitle string
	TimeWindow    string
	ResetTime     string

	IntervalUsedPercent   int
	IntervalRemainingText string

	WeeklyUsedPercent   int
	WeeklyRemainingText string
	WeeklyResetText     string

	Status StatusInfo
	Error  string
}

type StatusInfo struct {
	OK      bool
	Message string
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

// mapModelName 把 API 返回的分类名（general/video/...）映射成中文 UI 名
func mapModelName(name string) (display, subtitle string) {
	lower := strings.ToLower(name)
	switch {
	case lower == "general":
		return "MiniMax 语言模型", "对话、写作、推理"
	case lower == "video":
		return "MiniMax 视频生成", "文生视频/图生视频"
	case lower == "music":
		return "MiniMax 音乐生成", "歌曲/纯音乐生成"
	case lower == "image":
		return "MiniMax 图像生成", "文生图/图生图"
	case strings.HasPrefix(lower, "speech"):
		return "MiniMax 语音", "同步/异步语音合成"
	case strings.HasPrefix(lower, "MiniMax-m") || strings.HasPrefix(lower, "MiniMax"):
		return name, ""
	default:
		if name == "" {
			return "未知模型", ""
		}
		return name, ""
	}
}

// selectPrimaryModel 优先选 "general"（即 MiniMax 语言模型），否则选第一个
func selectPrimaryModel(models []api.ModelRemain) *api.ModelRemain {
	for i := range models {
		if strings.EqualFold(models[i].ModelName, "general") {
			return &models[i]
		}
	}
	if len(models) > 0 {
		return &models[0]
	}
	return nil
}

// usedPercent 把 API 返回的「剩余百分比」转成「已用百分比（相对正常额度）」
//
// 关键点：remaining_percent 是基于 boost 后的总额度算的，不是正常额度。
//
// 设 N = 正常额度（=1.0），boost 后总额度 = N × boost_permille/1000 = N × B
// 设 U = 实际已用量（相对正常额度的比例，0~B 之间）
// 剩余% = (总额度 - U) / 总额度 × 100 = (B - U) / B × 100
// 反推 U = B × (100 - remaining) / 100
// 相对于正常额度的已用% = U / 1.0 × 100 = B × (100 - remaining) / 1000
//
// 实测：5h (B=2.0, remaining=98) → 2×2/1000 = 4%；weekly (B=3.0, remaining=98) → 3×2/1000 = 6%
func usedPercent(remainingPercent int, boostPermille int) int {
	if boostPermille <= 0 {
		// 没有 boost 信息，回退到朴素算法
		if remainingPercent <= 0 {
			return 100
		}
		if remainingPercent >= 100 {
			return 0
		}
		return 100 - remainingPercent
	}
	used := boostPermille * (100 - remainingPercent) / 1000
	if used < 0 {
		return 0
	}
	if used > 100 {
		return 100
	}
	return used
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

	pageData := PageData{}

	primary := selectPrimaryModel(tokenPlan.ModelRemains)
	if primary != nil {
		display, subtitle := mapModelName(primary.ModelName)
		pageData.ModelName = display
		pageData.ModelSubtitle = subtitle

		startLocal := time.Unix(primary.StartTime/1000, 0).In(time.FixedZone("CST", 8*3600))
		endLocal := time.Unix(primary.EndTime/1000, 0).In(time.FixedZone("CST", 8*3600))
		pageData.TimeWindow = fmt.Sprintf("%02d:00-%02d:00 (UTC+8)",
			startLocal.Hour(), endLocal.Hour())
		pageData.ResetTime = formatRemainTime(primary.RemainsTime)

		pageData.IntervalUsedPercent = usedPercent(primary.CurrentIntervalRemainingPercent, primary.IntervalBoostPermille)
		pageData.IntervalRemainingText = fmt.Sprintf("剩余 %d%%", 100-pageData.IntervalUsedPercent)

		pageData.WeeklyUsedPercent = usedPercent(primary.CurrentWeeklyRemainingPercent, primary.WeeklyBoostPermille)
		pageData.WeeklyRemainingText = fmt.Sprintf("剩余 %d%%", 100-pageData.WeeklyUsedPercent)
		pageData.WeeklyResetText = formatRemainTime(primary.WeeklyRemainsTime)
	} else {
		pageData.Error = "API 未返回模型额度信息"
	}

	pageData.Status = StatusInfo{
		OK:      true,
		Message: "正常使用",
	}

	tmpl.Execute(w, pageData)
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))
