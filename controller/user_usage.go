package controller

import (
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// GetUserUsageOverview 获取用户用量概览
func GetUserUsageOverview(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	granularity := c.DefaultQuery("granularity", "day")

	// 验证必填参数
	if startTimestamp == 0 || endTimestamp == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "start_timestamp 和 end_timestamp 为必填参数",
		})
		return
	}

	// 验证聚合粒度
	if granularity != "day" && granularity != "week" && granularity != "month" {
		granularity = "day"
	}

	data, err := model.GetUserUsageOverview(startTimestamp, endTimestamp, granularity)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 补充 display_name
	userMap := make(map[int]string)
	for _, item := range data {
		if item.UserID > 0 && userMap[item.UserID] == "" {
			userMap[item.UserID] = ""
		}
	}
	// 批量获取用户显示名
	for userID := range userMap {
		user, err := model.GetUserById(userID, false)
		if err == nil {
			userMap[userID] = user.DisplayName
		}
	}
	for i := range data {
		data[i].DisplayName = userMap[data[i].UserID]
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}

// GetUserUsageDetail 获取用户用量详情
func GetUserUsageDetail(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Query("user_id"))
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	granularity := c.DefaultQuery("granularity", "day")

	if userID == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "user_id 为必填参数",
		})
		return
	}

	if startTimestamp == 0 || endTimestamp == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "start_timestamp 和 end_timestamp 为必填参数",
		})
		return
	}

	// 限制最大时间跨度 31 天
	if endTimestamp-startTimestamp > 31*86400 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "时间跨度不能超过 31 天",
		})
		return
	}

	if granularity != "day" && granularity != "week" && granularity != "month" {
		granularity = "day"
	}

	data, err := model.GetUserUsageDetail(userID, startTimestamp, endTimestamp, granularity)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}

// GetGlobalTimeSeries 获取全局时间序列数据
func GetGlobalTimeSeries(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	granularity := c.DefaultQuery("granularity", "day")

	if startTimestamp == 0 || endTimestamp == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "start_timestamp 和 end_timestamp 为必填参数",
		})
		return
	}

	if granularity != "day" && granularity != "week" && granularity != "month" {
		granularity = "day"
	}

	data, err := model.GetTimeSeriesData(startTimestamp, endTimestamp, granularity)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    data,
	})
}
