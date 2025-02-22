package service

import (
	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/models"
)

func GetCommunityList() ([]*models.Community, error) {
	return mysql.GetCommunityList()
}

func GetCommunityDetailByID(id int64) (*models.CommunityDetail, error) {
	return mysql.GetCommunityDetailByID(id)
}
