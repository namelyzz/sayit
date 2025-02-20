package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/utils/errors"
	"strconv"
)

const CtxUserIDKey = "userID"

// getCurrentUserID 获取当前登录的用户ID
func getCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(CtxUserIDKey)
	if !ok {
		err = errors.ErrorUserNotLogin
		return
	}
	userID, ok = uid.(int64)
	if !ok {
		err = errors.ErrorUserNotLogin
		return
	}
	return
}

func getPageInfo(c *gin.Context) (int64, int64) {
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	var (
		page int64
		size int64
		err  error
	)

	page, err = strconv.ParseInt(pageStr, 10, 64)
	if err != nil {
		page = 1
	}
	size, err = strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		size = 10
	}
	return page, size
}
