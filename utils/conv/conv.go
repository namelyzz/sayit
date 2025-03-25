package conv

import (
	"go.uber.org/zap"
	"strconv"
)

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
