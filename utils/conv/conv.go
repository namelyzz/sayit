package conv

import (
	"go.uber.org/zap"
	"strconv"
)

// Strings2Int64s 将字符串切片转换为 int64 切片
// 使用场景: Redis 返回的帖子 ID 是字符串类型，需要转换为 int64 用于 MySQL 查询
// 容错: 转换失败的字符串会被跳过并记录警告日志
func Strings2Int64s(strs []string) (res []int64) {
	res = make([]int64, 0, len(strs))
	for _, s := range strs {
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			zap.L().Warn("Strings2Ints 数据转化失败", zap.Error(err), zap.String("id", s))
			continue
		}
		res = append(res, i)
	}
	return res
}
