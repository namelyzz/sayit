package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/service"
	"github.com/namelyzz/sayit/utils/api"
	"go.uber.org/zap"
	"strconv"
)

// SearchSuggestHandler 搜索建议接口
// 根据类型返回社区或用户的模糊搜索建议
// GET /api/v1/search/suggest?q=xxx&type=community|user&limit=10
func SearchSuggestHandler(c *gin.Context) {
	keyword := c.Query("q")
	if keyword == "" {
		api.ResponseSuccess(c, []interface{}{})
		return
	}

	suggestType := c.DefaultQuery("type", "community")
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	var data interface{}

	switch suggestType {
	case "community":
		data, err = service.SearchCommunitiesByName(keyword, limit)
	case "user":
		data, err = service.SearchUsersByName(keyword, limit)
	default:
		api.ResponseError(c, api.CodeInvalidParam)
		return
	}

	if err != nil {
		zap.L().Error("search suggest failed",
			zap.String("type", suggestType),
			zap.String("keyword", keyword),
			zap.Error(err))
		api.ResponseError(c, api.CodeServerBusy)
		return
	}

	api.ResponseSuccess(c, data)
}
