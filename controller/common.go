package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/utils/api"
)

func getCurrentUserID(c *gin.Context) (userID int64, err error) {
	uid, ok := c.Get(api.CtxUserIDKey)
	if !ok {
		err = api.ErrorUserNotLogin
		return 0, nil
	}

	userID, ok = uid.(int64)
	if !ok {
		err = api.ErrorUserNotLogin
		return
	}

	return
}
