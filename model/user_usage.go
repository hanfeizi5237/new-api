package model

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

// getDBType 返回当前数据库类型标识
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

// getTimestampNormExpr 获取按粒度规范化时间戳的 SQL 表达式
func getTimestampNormExpr(granularity string) string {
	dbType := getDBType()
	switch dbType {
	case "mysql":
		switch granularity {
		case "day":
			return "UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(created_at)))"
		case "week":
			return "UNIX_TIMESTAMP(DATE_SUB(FROM_UNIXTIME(created_at), INTERVAL WEEKDAY(FROM_UNIXTIME(created_at)) DAY))"
		case "month":
			return "UNIX_TIMESTAMP(DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-01'))"
		default:
			return "UNIX_TIMESTAMP(DATE(FROM_UNIXTIME(created_at)))"
		}
	case "postgres":
		switch granularity {
		case "day":
			return "EXTRACT(EPOCH FROM DATE(TO_TIMESTAMP(created_at)))"
		case "week":
			return "EXTRACT(EPOCH FROM DATE_TRUNC('week', TO_TIMESTAMP(created_at)))"
		case "month":
			return "EXTRACT(EPOCH FROM DATE_TRUNC('month', TO_TIMESTAMP(created_at)))"
		default:
			return "EXTRACT(EPOCH FROM DATE(TO_TIMESTAMP(created_at)))"
		}
	case "sqlite":
		switch granularity {
		case "day":
			return "strftime('%s', DATE(created_at, 'unixepoch'))"
		case "week":
			return "strftime('%s', DATE(created_at - (CAST(created_at AS INTEGER) % 604800), 'unixepoch'))"
		case "month":
			return "strftime('%s', strftime('%Y-%m-01', created_at, 'unixepoch'))"
		default:
			return "strftime('%s', DATE(created_at, 'unixepoch'))"
		}
	default:
		return "created_at - (created_at % 86400)"
	}
}

// NormalizeTimestamp 根据聚合粒度规范化时间戳（用于 Go 层计算）
func NormalizeTimestamp(ts int64, granularity string) int64 {
	switch granularity {
	case "day":
		return ts - (ts % 86400)
	case "week":
		return ts - (ts % 604800)
	case "month":
		t := time.Unix(ts, 0)
		return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location()).Unix()
	default:
		return ts - (ts % 86400)
	}
}

// GetUserUsageOverview 获取所有用户的用量概览
func GetUserUsageOverview(startTimestamp, endTimestamp int64, granularity string) ([]dto.UserUsageOverview, error) {
	// 1. 从 quota_data 获取用户维度的用量汇总
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

	// 收集所有用户 ID
	userIDs := make([]int, 0, len(userStats))
	userStatMap := make(map[int]*userStat)
	for i := range userStats {
		userIDs = append(userIDs, userStats[i].UserID)
		userStatMap[userStats[i].UserID] = &userStats[i]
	}

	// 2. 从 logs 表获取错误统计（type=5）
	type errorStat struct {
		UserID     int `gorm:"column:user_id"`
		ErrorCount int `gorm:"column:error_count"`
	}
	var errorStats []errorStat
	err = DB.Table("logs").
		Select("user_id, COUNT(*) as error_count").
		Where("user_id IN ? AND type = 5 AND created_at >= ? AND created_at <= ?", userIDs, startTimestamp, endTimestamp).
		Group("user_id").
		Find(&errorStats).Error
	if err != nil {
		common.SysError("查询用户错误统计失败: " + err.Error())
	}

	errorMap := make(map[int]int)
	for _, es := range errorStats {
		errorMap[es.UserID] = es.ErrorCount
	}

	// 3. 构建返回数据
	overviews := make([]dto.UserUsageOverview, 0, len(userStats))
	for _, us := range userStats {
		overview := dto.UserUsageOverview{
			UserID:      us.UserID,
			Username:    us.Username,
			TotalCount:  us.TotalCount,
			TotalQuota:  us.TotalQuota,
			TotalTokens: us.TotalTokens,
			ErrorCount:  errorMap[us.UserID],
		}
		overviews = append(overviews, overview)
	}

	return overviews, nil
}

// GetUserUsageDetail 获取指定用户的用量详情
func GetUserUsageDetail(userID int, startTimestamp, endTimestamp int64, granularity string) (*dto.UserUsageDetail, error) {
	detail := &dto.UserUsageDetail{}

	// 1. 获取用户基本信息
	user, err := GetUserById(userID, false)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	// 2. 从 logs 表聚合消费日志 (type=2) 和错误日志
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
			SUM(CASE WHEN type != 2 THEN 1 ELSE 0 END) as error_count,
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

	detail.Summary = dto.UserUsageSummary{
		UserID:       userID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		TotalCount:   summary.TotalCount,
		TotalQuota:   summary.TotalQuota,
		TotalTokens:  summary.TotalTokens,
		ErrorCount:   summary.ErrorCount,
		AvgUseTimeMs: avgUseMs,
	}

	// 3. 模型分布
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
			SUM(CASE WHEN type != 2 THEN 1 ELSE 0 END) as error_count
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
		detail.ModelDistribution = append(detail.ModelDistribution, dto.ModelDistribution{
			ModelName:        ms.ModelName,
			Count:            ms.Count,
			Quota:            ms.Quota,
			PromptTokens:     ms.PromptTokens,
			CompletionTokens: ms.CompletionTokens,
			ErrorCount:       ms.ErrorCount,
		})
	}

	// 4. 时间分布（按粒度聚合）
	normExpr := getTimestampNormExpr(granularity)
	type timeStat struct {
		Timestamp    int64 `gorm:"column:timestamp"`
		Count        int   `gorm:"column:count"`
		Quota        int   `gorm:"column:quota"`
		Tokens       int   `gorm:"column:tokens"`
		TotalUseTime int   `gorm:"column:total_use_time"`
	}

	timeSQL := fmt.Sprintf(`
		SELECT
			%s as timestamp,
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
		detail.TimeDistribution = append(detail.TimeDistribution, dto.TimeSeriesItem{
			Timestamp: ts.Timestamp,
			Count:     ts.Count,
			Quota:     ts.Quota,
			Tokens:    ts.Tokens,
			AvgUseMs:  avgMs,
		})
	}

	// 5. 错误分布（所有非 type=2 的记录）
	dbType := getDBType()
	subFunc := "SUBSTRING"
	if dbType == "sqlite" {
		subFunc = "substr"
	}
	// PostgreSQL 用 LEFT 替代 SUBSTRING
	if dbType == "postgres" {
		subFunc = "LEFT"
	}

	type errorStat struct {
		ModelName    string `gorm:"column:model_name"`
		ErrorContent string `gorm:"column:error_content"`
		Count        int    `gorm:"column:count"`
		LatestAt     int64  `gorm:"column:latest_at"`
	}

	var errorSQL string
	if dbType == "postgres" {
		// PostgreSQL 用 LEFT(content, 200)
		errorSQL = fmt.Sprintf(`
			SELECT
				model_name,
				LEFT(content, 200) as error_content,
				COUNT(*) as count,
				MAX(created_at) as latest_at
			FROM logs
			WHERE user_id = ? AND type != 2 AND created_at >= ? AND created_at <= ?
			GROUP BY model_name, LEFT(content, 200)
			ORDER BY count DESC
			LIMIT 50
		`)
	} else {
		errorSQL = fmt.Sprintf(`
			SELECT
				model_name,
				%s(content, 1, 200) as error_content,
				COUNT(*) as count,
				MAX(created_at) as latest_at
			FROM logs
			WHERE user_id = ? AND type != 2 AND created_at >= ? AND created_at <= ?
			GROUP BY model_name, %s(content, 1, 200)
			ORDER BY count DESC
			LIMIT 50
		`, subFunc, subFunc)
	}

	var errorStats []errorStat
	err = DB.Raw(errorSQL, userID, startTimestamp, endTimestamp).Scan(&errorStats).Error
	if err != nil {
		common.SysError("查询错误分布失败: " + err.Error())
		errorStats = []errorStat{}
	}

	// 聚合相同 model_name + 截断内容的错误
	errorMap := make(map[string]*dto.ErrorDistribution)
	for _, es := range errorStats {
		key := es.ModelName + "|" + es.ErrorContent
		if existing, ok := errorMap[key]; ok {
			existing.Count += es.Count
			if es.LatestAt > existing.LatestAt {
				existing.LatestAt = es.LatestAt
			}
		} else {
			errorMap[key] = &dto.ErrorDistribution{
				ModelName:    es.ModelName,
				ErrorContent: es.ErrorContent,
				Count:        es.Count,
				LatestAt:     es.LatestAt,
			}
		}
	}

	detail.ErrorDistribution = make([]dto.ErrorDistribution, 0, len(errorMap))
	for _, ed := range errorMap {
		detail.ErrorDistribution = append(detail.ErrorDistribution, *ed)
	}

	// 按错误次数降序排序
	for i := 0; i < len(detail.ErrorDistribution); i++ {
		for j := i + 1; j < len(detail.ErrorDistribution); j++ {
			if detail.ErrorDistribution[j].Count > detail.ErrorDistribution[i].Count {
				detail.ErrorDistribution[i], detail.ErrorDistribution[j] = detail.ErrorDistribution[j], detail.ErrorDistribution[i]
			}
		}
	}

	return detail, nil
}

// GetTimeSeriesData 获取全局时间序列数据（用于主看板趋势图）
func GetTimeSeriesData(startTimestamp, endTimestamp int64, granularity string) ([]dto.TimeSeriesItem, error) {
	normExpr := getTimestampNormExpr(granularity)

	type timeStat struct {
		Timestamp    int64 `gorm:"column:timestamp"`
		Count        int   `gorm:"column:count"`
		Quota        int   `gorm:"column:quota"`
		Tokens       int   `gorm:"column:tokens"`
		TotalUseTime int   `gorm:"column:total_use_time"`
	}

	rawSQL := fmt.Sprintf(`
		SELECT
			%s as timestamp,
			SUM(CASE WHEN type = 2 THEN 1 ELSE 0 END) as count,
			SUM(CASE WHEN type = 2 THEN quota ELSE 0 END) as quota,
			SUM(CASE WHEN type = 2 THEN prompt_tokens + completion_tokens ELSE 0 END) as tokens,
			SUM(CASE WHEN type = 2 THEN use_time ELSE 0 END) as total_use_time
		FROM logs
		WHERE created_at >= ? AND created_at <= ?
		GROUP BY %s
		ORDER BY timestamp ASC
	`, normExpr, normExpr)

	var timeStats []timeStat
	err := DB.Raw(rawSQL, startTimestamp, endTimestamp).Scan(&timeStats).Error
	if err != nil {
		return nil, fmt.Errorf("查询时间序列数据失败: %w", err)
	}

	result := make([]dto.TimeSeriesItem, 0, len(timeStats))
	for _, ts := range timeStats {
		avgMs := 0
		if ts.Count > 0 {
			avgMs = (ts.TotalUseTime * 1000) / ts.Count
		}
		result = append(result, dto.TimeSeriesItem{
			Timestamp: ts.Timestamp,
			Count:     ts.Count,
			Quota:     ts.Quota,
			Tokens:    ts.Tokens,
			AvgUseMs:  avgMs,
		})
	}

	return result, nil
}
