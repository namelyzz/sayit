package mysql

import (
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/namelyzz/sayit/utils/security"
	"gorm.io/gorm"
)

// CheckUserExist 检查用户名是否已存在
// 通过查询数据库中该用户名的记录数来判断
// 返回 nil 表示用户名可用，返回 api.ErrorUserExist 表示已存在
func CheckUserExist(username string) (err error) {
	var count int64
	if err = db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return api.ErrorUserExist
	}
	return nil
}

// InsertUser 将新用户插入数据库
// 在插入前对密码进行 SHA256 加盐哈希，数据库中只存储哈希值
func InsertUser(user *models.User) (err error) {
	// 对明文密码进行哈希加密后替换，确保数据库中不存储明文密码
	user.Password = security.HashPassword(user.Password)
	res := db.Omit("CreateTime", "UpdateTime").Create(user)
	return res.Error
}

// Login 验证用户登录
// 流程: 根据用户名查询用户 -> 用暂存的明文密码与数据库中的哈希密码比对
// 注意: 此方法会修改传入的 user 对象，将数据库中的用户信息覆盖到 user 中
func Login(user *models.User) (err error) {
	// 暂存用户输入的明文密码，因为后续查询会覆盖 user.Password
	userPwd := user.Password

	// 根据用户名查询数据库，查询结果（含哈希密码）会覆盖 user 对象的字段
	err = db.Where("username = ?", user.Username).First(user).Error
	if err == gorm.ErrRecordNotFound {
		// 用户名不存在
		return api.ErrorUserNotExist
	}
	if err != nil {
		return err
	}

	// 将用户输入的明文密码与数据库中的哈希密码进行比对
	if !security.VerifyPassword(userPwd, user.Password) {
		// 密码不匹配
		return api.ErrorInvalidLogin
	}

	return nil
}

// GetUserByID 根据用户ID查询用户信息（仅返回 user_id 和 username，不返回密码）
// 用于其他模块获取用户公开信息
func GetUserByID(userID int64) (user *models.User, err error) {
	user = new(models.User)
	res := db.Model(&models.User{}).
		Select("user_id", "username").
		Where("user_id = ?", userID).
		First(user)

	if res.Error != nil {
		return nil, res.Error
	}
	return user, nil
}

// GetUserProfileByID 根据用户ID查询用户公开资料，不返回密码等敏感字段。
func GetUserProfileByID(userID int64) (user *models.User, err error) {
	user = new(models.User)
	res := db.Model(&models.User{}).
		Select("user_id", "username", "signature", "create_time").
		Where("user_id = ?", userID).
		First(user)

	if res.Error == gorm.ErrRecordNotFound {
		return nil, api.ErrorUserNotExist
	}
	if res.Error != nil {
		return nil, res.Error
	}
	return user, nil
}

// UpdateUserSignature 更新用户个性签名。
func UpdateUserSignature(userID int64, signature string) error {
	res := db.Model(&models.User{}).
		Where("user_id = ?", userID).
		Update("signature", signature)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return api.ErrorUserNotExist
	}
	return nil
}
