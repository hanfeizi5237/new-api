package model

import (
	"fmt"
	"sort"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

func getDBType() string {
	if common.UsingMySQL {
		return "mysql"
	}
	if common.UsingPostgreSQL {
		return "postgres"
	}
	if common.UsingSQLite {
		return "sqlite"
	}
	return "sqlite"
}

func normalizeModelName(name string) string {
	if name == "" {
		return "未知模型"
	}
	return name
}

// 失败调用口径：
// 1. 真实错误日志 type=5
// 2. 记录为消费日志(type=2)但 quota/token 都为 0 且有错误内容的零扣费异常
func failureConditionSQL() string {
	return `(type = 5 OR (type = 2 AND quota = 0 AND prompt_tokens = 0 AND completion_tokens = 0 AND content IS NOT NULL AND content != ''))`
}

func getTimestampNormExpr(granularity string) string {
	dbType := getDBType()
	switch dbType {
	case "mysql":
		switch granularity {
		case "day", "week":
			return "UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(created_at)))"
		case "month":
			return "UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-01'))"
		default:
			return "UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(created_at)))"
		}
	case "postgres":
		switch granularity {
		case "day", "week":
			return "EXTRACT(EPOCH FROM DATE(TO_TIMESTAMP(created_at)))"
		case "month":
			return "EXTRACT(EPOCH FROM DATE_TRUNC('month', TO_TIMESTAMP(created_at)))"
		default:
			return "EXTRACT(EPOCH FROM DATE(TO_TIMESTAMP(created_at)))"
		}
	case "sqlite":
		switch granularity {
		case "day", "week":
			return "strftime('%s', DATE(created_at, 'unixepoch'))"
		case "month":
			return "strftime('%s', strftime('%Y-%m-01', created_at, 'unixepoch'))"
		default:
			return "strftime('%s', DATE(created_at, 'unixepoch'))"
		}
	default:
		return "created_at - (created_at % 86400)"
	}
}

func NormalizeTimestamp(ts int64, granularity string) int64 {
	switch granularity {
	case "day", "week":
		return ts - (ts % 86400)
	case "month":
		t := time.Unix(ts, 0)
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).Unix()
	default:
		return ts - (ts % 86400)
	}
}

func GetUserUsageOverview(startTimestamp, endTimestamp int64, granularity string) ([]dto.UserUsageOverview, error) {
	type userStat struct {
		UserID      int    `gorm:"column:user_id"`
		Username    string `gorm:"column:username"`
		TotalCount  int    `gorm:"column:total_count"`
		TotalQuota  int    `gorm:"column:total_quota"`
		TotalTokens int    `gorm:"column:total_tokens"`
	}

	var userStats []userStat
	err := DB.Table("quota_data").
		Select("user_id, username, SUM(count) as total_count, SUM(quota) as total_quota, SUM(token_used) as total_tokens").
		Where("created_at >= ? AND created_at <= ?", startTimestamp, endTimestamp).
		Group("user_id, username").
		Order("total_quota DESC").
		Find(&userStats).Error
	if err != nil {
		return nil, fmt.Errorf("查询用户用量汇总失败: %w", err)
	}
	if len(userStats) == 0 {
		return []dto.UserUsageOverview{}, nil
	}

	overviewMap := make(map[int]*dto.UserUsageOverview, len(userStats))
	userIDs := make([]int, 0, len(userStats))
	for _, us := range userStats {
		item := &dto.UserUsageOverview{
			UserID:      us.UserID,
			Username:    us.Username,
			TotalCount:  us.TotalCount,
			TotalQuota:  us.TotalQuota,
			TotalTokens: us.TotalTokens,
			TimeSeries:  []dto.TimeSeriesItem{},
		}
		overviewMap[us.UserID] = item
		userIDs = append(userIDs, us.UserID)
	}

	type errorStat struct {
		UserID     int `gorm:"column:user_id"`
		ErrorCount int `gorm:"column:error_count"`
	}
	var errorStats []errorStat
	err = DB.Table("logs").
		Select("user_id, COUNT(*) as error_count").
		Where("user_id IN ? AND created_at >= ? AND created_at <= ? AND "+failureConditionSQL(), userIDs, startTimestamp, endTimestamp).
		Group("user_id").
		Find(&errorStats).Error
	if err == nil {
		for _, es := range errorStats {
			if item, ok := overviewMap[es.UserID]; ok {
				item.ErrorCount = es.ErrorCount
			}
		}
	}

	type seriesRow struct {
		UserID    int   `gorm:"column:user_id"`
		Timestamp int64 `gorm:"column:timestamp"`
		Count     int   `gorm:"column:count"`
		Quota     int   `gorm:"column:quota"`
		Tokens    int   `gorm:"column:tokens"`
	}
	normExpr := getTimestampNormExpr(granularity)
	seriesSQL := fmt.Sprintf(`
		SELECT user_id, %s as timestamp, SUM(count) as count, SUM(quota) as quota, SUM(token_used) as tokens
		FROM quota_data
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY user_id, %s
		ORDER BY timestamp ASC
	`, normExpr, normExpr)
	var seriesRows []seriesRow
	err = DB.Raw(seriesSQL, startTimestamp, endTimestamp).Scan(&seriesRows).Error
	if err == nil {
		for _, row := range seriesRows {
			if item, ok := overviewMap[row.UserID]; ok {
				item.TimeSeries = append(item.TimeSeries, dto.TimeSeriesItem{Timestamp: row.Timestamp, Count: row.Count, Quota: row.Quota, Tokens: row.Tokens})
			}
		}
	}

	overviews := make([]dto.UserUsageOverview, 0, len(userStats))
	for _, us := range userStats {
		if item, ok := overviewMap[us.UserID]; ok {
			overviews = append(overviews, *item)
		}
	}
	return overviews, nil
}

func GetUserUsageDetail(userID int, startTimestamp, endTimestamp int64, granularity string) (*dto.UserUsageDetail, error) {
	detail := &dto.UserUsageDetail{}
	user, err := GetUserById(userID, false)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	type summaryStat struct {
		TotalCount   int `gorm:"column:total_count"`
		TotalQuota   int `gorm:"column:total_quota"`
		TotalTokens  int `gorm:"column:total_tokens"`
		ErrorCount   int `gorm:"column:error_count"`
		TotalUseTime int `gorm:"column:total_use_time"`
	}
	var summary summaryStat
	err = DB.Table("logs").
		Select(`
			SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) as total_count,
			SUM(CASE WHEN type = 2 THEN quota ELSE 0 END) as total_quota,
			SUM(CASE WHEN type = 2 THEN prompt_tokens + completion_tokens ELSE 0 END) as total_tokens,
			SUM(CASE WHEN ` + failureConditionSQL() + ` THEN 1 ELSE 0 END) as error_count,
			SUM(CASE WHEN type = 2 THEN use_time ELSE 0 END) as total_use_time
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startTimestamp, endTimestamp).
		Scan(&summary).Error
	if err != nil {
		return nil, fmt.Errorf("查询用量汇总失败: %w", err)
	}
	avgUseMs := 0
	if summary.TotalCount > 0 {
		avgUseMs = (summary.TotalUseTime * 1000) / summary.TotalCount
	}
	detail.Summary = dto.UserUsageSummary{UserID: userID, Username: user.Username, DisplayName: user.DisplayName, TotalCount: summary.TotalCount, TotalQuota: summary.TotalQuota, TotalTokens: summary.TotalTokens, ErrorCount: summary.ErrorCount, AvgUseTimeMs: avgUseMs}

	type modelStat struct {
		ModelName        string `gorm:"column:model_name"`
		Count            int    `gorm:"column:count"`
		Quota            int    `gorm:"column:quota"`
		PromptTokens     int    `gorm:"column:prompt_tokens"`
		CompletionTokens int    `gorm:"column:completion_tokens"`
		ErrorCount       int    `gorm:"column:error_count"`
	}
	var modelStats []modelStat
	err = DB.Table("logs").
		Select(`
			model_name,
			SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) as count,
			SUM(CASE WHEN type = 2 THEN quota ELSE 0 END) as quota,
			SUM(CASE WHEN type = 2 THEN prompt_tokens ELSE 0 END) as prompt_tokens,
			SUM(CASE WHEN type = 2 THEN completion_tokens ELSE 0 END) as completion_tokens,
			SUM(CASE WHEN ` + failureConditionSQL() + ` THEN 1 ELSE 0 END) as error_count
		`).
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startTimestamp, endTimestamp).
		Group("model_name").
		Order("quota DESC").
		Find(&modelStats).Error
	if err != nil {
		common.SysError("查询模型分布失败: " + err.Error())
	}
	detail.ModelDistribution = make([]dto.ModelDistribution, 0, len(modelStats))
	for _, ms := range modelStats {
		if ms.Count == 0 && ms.ErrorCount == 0 && ms.Quota == 0 && ms.PromptTokens == 0 && ms.CompletionTokens == 0 {
			continue
		}
		detail.ModelDistribution = append(detail.ModelDistribution, dto.ModelDistribution{ModelName: normalizeModelName(ms.ModelName), Count: ms.Count, Quota: ms.Quota, PromptTokens: ms.PromptTokens, CompletionTokens: ms.CompletionTokens, ErrorCount: ms.ErrorCount})
	}

	normExpr := getTimestampNormExpr(granularity)
	type timeStat struct {
		Timestamp    int64 `gorm:"column:timestamp"`
		Count        int   `gorm:"column:count"`
		Quota        int   `gorm:"column:quota"`
		Tokens       int   `gorm:"column:tokens"`
		TotalUseTime int   `gorm:"column:total_use_time"`
	}
	timeSQL := fmt.Sprintf(`
		SELECT %s as timestamp,
			SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) as count,
			SUM(CASE WHEN type = 2 THEN quota ELSE 0 END) as quota,
			SUM(CASE WHEN type = 2 THEN prompt_tokens + completion_tokens ELSE 0 END) as tokens,
			SUM(CASE WHEN type = 2 THEN use_time ELSE 0 END) as total_use_time
		FROM logs
		WHERE user_id = ? AND created_at >= ? AND created_at <= ?
		GROUP BY %s
		ORDER BY timestamp ASC
	`, normExpr, normExpr)
	var timeStatsRaw []timeStat
	err = DB.Raw(timeSQL, userID, startTimestamp, endTimestamp).Scan(&timeStatsRaw).Error
	if err != nil {
		common.SysError("查询时间分布失败: " + err.Error())
		timeStatsRaw = []timeStat{}
	}
	detail.TimeDistribution = make([]dto.TimeSeriesItem, 0, len(timeStatsRaw))
	for _, ts := range timeStatsRaw {
		avgMs := 0
		if ts.Count > 0 {
			avgMs = (ts.TotalUseTime * 1000) / ts.Count
		}
		detail.TimeDistribution = append(detail.TimeDistribution, dto.TimeSeriesItem{Timestamp: ts.Timestamp, Count: ts.Count, Quota: ts.Quota, Tokens: ts.Tokens, AvgUseMs: avgMs})
	}

	dbType := getDBType()
	subFunc := "SUBSTRING"
	if dbType == "sqlite" { subFunc = "substr" }
	if dbType == "postgres" { subFunc = "LEFT" }
	type errorStat struct {
		ModelName string `gorm:"column:model_name"`
		ErrorContent string `gorm:"column:error_content"`
		Count int `gorm:"column:count"`
		LatestAt int64 `gorm:"column:latest_at"`
	}
	var errorSQL string
	if dbType == "postgres" {
		errorSQL = `SELECT model_name, LEFT(content, 200) as error_content, COUNT(*) as count, MAX(created_at) as latest_at FROM logs WHERE user_id = ? AND created_at >= ? AND created_at <= ? AND ` + failureConditionSQL() + ` GROUP BY model_name, LEFT(content, 200) ORDER BY count DESC LIMIT 50`
	} else {
		errorSQL = fmt.Sprintf(`SELECT model_name, %s(content, 1, 200) as error_content, COUNT(*) as count, MAX(created_at) as latest_at FROM logs WHERE user_id = ? AND created_at >= ? AND created_at <= ? AND %s GROUP BY model_name, %s(content, 1, 200) ORDER BY count DESC LIMIT 50`, subFunc, failureConditionSQL(), subFunc)
	}
	var errorStats []errorStat
	err = DB.Raw(errorSQL, userID, startTimestamp, endTimestamp).Scan(&errorStats).Error
	if err != nil {
		common.SysError("查询错误分布失败: " + err.Error())
		errorStats = []errorStat{}
	}
	detail.ErrorDistribution = make([]dto.ErrorDistribution, 0, len(errorStats))
	for _, es := range errorStats {
		detail.ErrorDistribution = append(detail.ErrorDistribution, dto.ErrorDistribution{ModelName: normalizeModelName(es.ModelName), ErrorContent: es.ErrorContent, Count: es.Count, LatestAt: es.LatestAt})
	}
	sort.Slice(detail.ErrorDistribution, func(i, j int) bool { return detail.ErrorDistribution[i].Count > detail.ErrorDistribution[j].Count })
	return detail, nil
}

func GetTimeSeriesData(startTimestamp, endTimestamp int64, granularity string) ([]dto.TimeSeriesItem, error) {
	normExpr := getTimestampNormExpr(granularity)
	type timeStat struct {
		Timestamp int64 `gorm:"column:timestamp"`
		Count int `gorm:"column:count"`
		Quota int `gorm:"column:quota"`
		Tokens int `gorm:"column:tokens"`
		TotalUseTime int `gorm:"column:total_use_time"`
	}
	rawSQL := fmt.Sprintf(`SELECT %s as timestamp, SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) as count, SUM(CASE WHEN type = 2 THEN quota ELSE 0 END) as quota, SUM(CASE WHEN type = 2 THEN prompt_tokens + completion_tokens ELSE 0 END) as tokens, SUM(CASE WHEN type = 2 THEN use_time ELSE 0 END) as total_use_time FROM logs WHERE created_at >= ? AND created_at <= ? GROUP BY %s ORDER BY timestamp ASC`, normExpr, normExpr)
	var timeStats []timeStat
	err := DB.Raw(rawSQL, startTimestamp, endTimestamp).Scan(&timeStats).Error
	if err != nil {
		return nil, fmt.Errorf("查询时间序列数据失败: %w", err)
	}
	result := make([]dto.TimeSeriesItem, 0, len(timeStats))
	for _, ts := range timeStats {
		avgMs := 0
		if ts.Count > 0 { avgMs = (ts.TotalUseTime * 1000) / ts.Count }
		result = append(result, dto.TimeSeriesItem{Timestamp: ts.Timestamp, Count: ts.Count, Quota: ts.Quota, Tokens: ts.Tokens, AvgUseMs: avgMs})
	}
	return result, nil
}
