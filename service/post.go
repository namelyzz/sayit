package service

import (
    "github.com/namelyzz/sayit/dao/mysql"
    "github.com/namelyzz/sayit/models"
    "github.com/namelyzz/sayit/utils/snowflake"
)

func CreatePost(p *models.Post) (err error) {
    // 使用雪花算法为帖子生成一个 ID
    p.PostID = snowflake.GenID()
    err = mysql.CreatePost(p)
    if err != nil {
        return err
    }
    return
}
