package snowflake

import (
    "github.com/bwmarrin/snowflake"
    "time"
)

// 声明一个全局变量，用于存储 snowflake 节点实例
// 这个节点将在整个应用程序中用于生成唯一的ID
// 雪花 ID 是 64 位整数，由以下部分组成：
// - 时间戳（自自定义纪元以来的毫秒数）
// - 机器 / 工作节点 ID（用于区分分布式系统中的不同节点）
// - 序列号（用于处理同一机器上同一毫秒内生成的多个 ID）
// 这些 ID 具有唯一性、可按时间排序，且生成时无需节点间协调。
var node *snowflake.Node

func Init(startTime string, machineID int64) (err error) {
    var st time.Time
    st, err = time.Parse("2006-01-02", startTime)
    if err != nil {
        return
    }

    // 设置 snowflake 算法的纪元（起始时间点）
    // snowflake 算法使用一个基准时间来计算ID的时间戳部分
    // UnixNano() 返回从1970年1月1日到指定时间的纳秒数
    // 除以 1000000 将纳秒转换为毫秒（因为snowflake使用毫秒时间戳）
    snowflake.Epoch = st.UnixNano() / 1000000

    // 使用提供的 machineID 创建一个新的 snowflake 节点
    // machineID 在分布式系统中必须是唯一的，范围通常是 0-1023
    // 这个节点将用于生成全局唯一的ID
    node, err = snowflake.NewNode(machineID)
    return
}

func GenID() int64 {
    return node.Generate().Int64()
}
