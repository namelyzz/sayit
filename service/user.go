package service

import (
	"context"
	"strconv"
	"strings"

	"github.com/namelyzz/sayit/dao/mysql"
	"github.com/namelyzz/sayit/dao/redis"
	"github.com/namelyzz/sayit/models"
	"github.com/namelyzz/sayit/utils/api"
	"github.com/namelyzz/sayit/utils/jwt"
	"github.com/namelyzz/sayit/utils/snowflake"
)

const defaultUserSignature = "这个人很懒，还没有留下签名。"

const maxUserFollowListSize = 50

// SignUp 用户注册业务逻辑
// 步骤: 检查用户名是否已存在 -> 生成雪花ID -> 构造用户对象 -> 密码加密后入库
func SignUp(p *models.ParamSignUp) (err error) {
	// 1. 先判断用户是否存在（用户名唯一性校验）
	if err = mysql.CheckUserExist(p.Username); err != nil {
		return err
	}

	// 2. 通过雪花算法生成全局唯一用户 ID，然后构造用户数据
	// 雪花ID相比自增ID的优势: 全局唯一、趋势递增、不暴露用户数量、支持分布式
	userID := snowflake.GenID()
	user := &models.User{
		UserID:   models.SnowflakeID(userID),
		Username: p.Username,
		Password: p.Password,
	}

	// 3. 入库（dao 层会对密码进行 SHA256 加密后存储）
	return mysql.InsertUser(user)
}

// Login 用户登录业务逻辑
// 步骤: 从数据库查询用户并验证密码 -> 生成 JWT Token -> 返回用户信息
func Login(p *models.ParamLogin) (user *models.User, err error) {
	// 1. 构造用户对象，调用 dao 层验证用户名和密码
	user = &models.User{Username: p.Username, Password: p.Password}
	if err = mysql.Login(user); err != nil {
		return nil, err
	}

	// 2. 密码验证通过，生成 JWT Token（有效期1小时）
	token, err := jwt.CreateJWTToken(int64(user.UserID), user.Username)
	if err != nil {
		return nil, err
	}
	// 3. 将 token 附加到用户对象返回给 controller 层
	user.Token = token
	return user, nil
}

// GetUserProfile 获取用户资料页所需的核心公开信息。
func GetUserProfile(ctx context.Context, currentUserID, targetUserID int64) (profile *models.UserProfile, err error) {
	user, err := mysql.GetUserProfileByID(targetUserID)
	if err != nil {
		return nil, err
	}

	postCount, err := mysql.CountNormalPostsByAuthor(targetUserID)
	if err != nil {
		return nil, err
	}
	postScore, err := getUserPostScore(ctx, targetUserID)
	if err != nil {
		return nil, err
	}
	commentScore, err := GetUserCommentScore(targetUserID)
	if err != nil {
		return nil, err
	}
	followerCount, err := mysql.CountFollowers(targetUserID)
	if err != nil {
		return nil, err
	}
	followingCount, err := mysql.CountFollowing(targetUserID)
	if err != nil {
		return nil, err
	}
	isFollowing := false
	if currentUserID != 0 && currentUserID != targetUserID {
		isFollowing, err = isFollowingUserWithCache(ctx, currentUserID, targetUserID)
		if err != nil {
			return nil, err
		}
	}

	signature := user.Signature
	if signature == "" {
		signature = defaultUserSignature
	}

	return &models.UserProfile{
		UserID:         user.UserID,
		Username:       user.Username,
		Signature:      signature,
		CreateTime:     user.CreateTime,
		PostCount:      postCount,
		PostScore:      postScore,
		CommentScore:   commentScore,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		IsFollowing:    isFollowing,
		IsSelf:         currentUserID == targetUserID,
	}, nil
}

func getUserPostScore(ctx context.Context, userID int64) (int64, error) {
	postIDs, err := mysql.GetNormalPostIDsByAuthor(userID)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, postID := range postIDs {
		total += redis.GetPostVoteValue(ctx, strconv.FormatInt(postID, 10))
	}
	return total, nil
}

// UpdateUserProfile 更新当前登录用户资料，并返回更新后的资料。
func UpdateUserProfile(ctx context.Context, userID int64, p *models.ParamUpdateProfile) (*models.UserProfile, error) {
	signature := strings.TrimSpace(p.Signature)
	if err := mysql.UpdateUserSignature(userID, signature); err != nil {
		return nil, err
	}
	return GetUserProfile(ctx, userID, userID)
}

// GetUserPosts 获取指定用户发布的帖子列表。
func GetUserPosts(ctx context.Context, userID int64, p *models.ParamPostList, currentUserID int64) ([]*models.PostListItem, error) {
	if _, err := mysql.GetUserProfileByID(userID); err != nil {
		return nil, err
	}
	p.AuthorID = userID
	return GetPostListWithViewer(ctx, p, currentUserID)
}

// GetUserComments 获取用户的最新评论列表
// 返回用户最新评论，包含帖子标题，用于个人中心展示
func GetUserComments(ctx context.Context, userID int64, limit int, currentUserID int64) (*models.UserCommentListResponse, error) {
	// 1. 验证用户存在
	if _, err := mysql.GetUserProfileByID(userID); err != nil {
		return nil, err
	}

	// 2. 获取用户最新评论
	comments, err := mysql.GetLatestCommentsByAuthorID(userID, limit)
	if err != nil {
		return nil, err
	}

	// 3. 统计用户评论总数
	total, err := mysql.CountCommentsByAuthorID(userID)
	if err != nil {
		return nil, err
	}

	return &models.UserCommentListResponse{
		List:  comments,
		Total: total,
	}, nil
}

// FollowUser 关注用户。
func FollowUser(currentUserID, targetUserID int64) error {
	if currentUserID == targetUserID {
		return api.ErrorFollowSelf
	}
	if _, err := mysql.GetUserProfileByID(targetUserID); err != nil {
		return err
	}
	if err := mysql.FollowUser(currentUserID, targetUserID); err != nil {
		return err
	}
	_ = redis.FollowUserCache(context.Background(), currentUserID, targetUserID)
	return nil
}

// UnfollowUser 取消关注用户。
func UnfollowUser(currentUserID, targetUserID int64) error {
	if currentUserID == targetUserID {
		return api.ErrorFollowSelf
	}
	if _, err := mysql.GetUserProfileByID(targetUserID); err != nil {
		return err
	}
	if err := mysql.UnfollowUser(currentUserID, targetUserID); err != nil {
		return err
	}
	_ = redis.UnfollowUserCache(context.Background(), currentUserID, targetUserID)
	return nil
}

// IsFollowingUser 检查当前用户是否关注目标用户。
func IsFollowingUser(currentUserID, targetUserID int64) (bool, error) {
	if currentUserID == targetUserID {
		return false, nil
	}
	if _, err := mysql.GetUserProfileByID(targetUserID); err != nil {
		return false, err
	}
	return isFollowingUserWithCache(context.Background(), currentUserID, targetUserID)
}

// GetUserFollowers 获取用户粉丝列表。
func GetUserFollowers(ctx context.Context, userID int64, page, size int, viewerID int64) (*models.UserFollowList, error) {
	if _, err := mysql.GetUserProfileByID(userID); err != nil {
		return nil, err
	}
	page, size = normalizeUserFollowPagination(page, size)
	list, total, err := mysql.GetFollowers(userID, page, size)
	if err != nil {
		return nil, err
	}
	enrichUserFollowItems(ctx, list, viewerID)
	return &models.UserFollowList{List: list, Total: total}, nil
}

// GetUserFollowing 获取用户关注列表。
func GetUserFollowing(ctx context.Context, userID int64, page, size int, viewerID int64) (*models.UserFollowList, error) {
	if _, err := mysql.GetUserProfileByID(userID); err != nil {
		return nil, err
	}
	page, size = normalizeUserFollowPagination(page, size)
	list, total, err := mysql.GetFollowing(userID, page, size)
	if err != nil {
		return nil, err
	}
	enrichUserFollowItems(ctx, list, viewerID)
	return &models.UserFollowList{List: list, Total: total}, nil
}

func enrichUserFollowItems(ctx context.Context, list []*models.UserFollowItem, viewerID int64) {
	if viewerID == 0 {
		return
	}
	for _, item := range list {
		itemID := int64(item.UserID)
		if itemID == viewerID {
			continue
		}
		item.IsFollowing, _ = isFollowingUserWithCache(ctx, viewerID, itemID)
		item.IsFollowedBy, _ = isFollowingUserWithCache(ctx, itemID, viewerID)
		item.IsMutual = item.IsFollowing && item.IsFollowedBy
	}
}

func isFollowingUserWithCache(ctx context.Context, followerID, followingID int64) (bool, error) {
	isFollowing, err := redis.IsFollowingUserCache(ctx, followerID, followingID)
	if err == nil && isFollowing {
		return true, nil
	}

	isFollowing, err = mysql.IsFollowingUser(followerID, followingID)
	if err != nil {
		return false, err
	}
	if isFollowing {
		_ = redis.FollowUserCache(ctx, followerID, followingID)
	}
	return isFollowing, nil
}

func normalizeUserFollowPagination(page, size int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 || size > maxUserFollowListSize {
		size = maxUserFollowListSize
	}
	return page, size
}
