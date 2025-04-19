package models

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// SnowflakeID 将 int64 雪花 ID 以字符串形式进行 JSON 序列化，避免 JS 精度丢失
type SnowflakeID int64

func (id SnowflakeID) String() string {
	return strconv.FormatInt(int64(id), 10)
}

func (id SnowflakeID) Int64() int64 {
	return int64(id)
}

func (id SnowflakeID) MarshalJSON() ([]byte, error) {
	return json.Marshal(id.String())
}

func (id *SnowflakeID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid snowflake id: %s", s)
	}
	*id = SnowflakeID(v)
	return nil
}
