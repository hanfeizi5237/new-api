package dto

// UserUsageOverview 用户用量概览
type UserUsageOverview struct {
	UserID      int                `json:"user_id"`
	Username    string             `json:"username"`
	DisplayName string             `json:"display_name"`
	TotalCount  int                `json:"total_count"`
	TotalQuota  int                `json:"total_quota"`
	TotalTokens int                `json:"total_tokens"`
	ErrorCount  int                `json:"error_count"`
	TimeSeries  []TimeSeriesItem   `json:"time_series"`
}

// TimeSeriesItem 时间序列数据项
type TimeSeriesItem struct {
	Timestamp int64 `json:"timestamp"`
	Count     int   `json:"count"`
	Quota     int   `json:"quota"`
	Tokens    int   `json:"tokens"`
	AvgUseMs  int   `json:"avg_use_ms"`
}

// UserUsageDetail 用户用量详情
type UserUsageDetail struct {
	Summary           UserUsageSummary    `json:"summary"`
	ModelDistribution []ModelDistribution `json:"model_distribution"`
	TimeDistribution  []TimeSeriesItem    `json:"time_distribution"`
	ErrorDistribution []ErrorDistribution `json:"error_distribution"`
}

// UserUsageSummary 用量汇总
type UserUsageSummary struct {
	UserID       int    `json:"user_id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	TotalCount   int    `json:"total_count"`
	TotalQuota   int    `json:"total_quota"`
	TotalTokens  int    `json:"total_tokens"`
	ErrorCount   int    `json:"error_count"`
	AvgUseTimeMs int    `json:"avg_use_time_ms"`
}

// ModelDistribution 模型分布
type ModelDistribution struct {
	ModelName        string `json:"model_name"`
	Count            int    `json:"count"`
	Quota            int    `json:"quota"`
	PromptTokens     int    `json:"prompt_tokens"`
	CompletionTokens int    `json:"completion_tokens"`
	ErrorCount       int    `json:"error_count"`
}

// ErrorDistribution 错误分布
type ErrorDistribution struct {
	ModelName    string `json:"model_name"`
	ErrorContent string `json:"error_content"`
	Count        int    `json:"count"`
	LatestAt     int64  `json:"latest_at"`
}
