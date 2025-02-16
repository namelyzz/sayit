package config

import (
    "fmt"
    "github.com/fsnotify/fsnotify"
    "github.com/spf13/viper"
)

var Conf = new(AppConfig)

type AppConfig struct {
    Name      string `mapstructure:"name"`
    Mode      string `mapstructure:"mode"`
    Version   string `mapstructure:"version"`
    StartTime string `mapstructure:"start_time"`
    MachineID int64  `mapstructure:"machine_id"`
    Port      int    `mapstructure:"port"`

    *LogConfig   `mapstructure:"log"`
    *MySQLConfig `mapstructure:"mysql"`
    *RedisConfig `mapstructure:"redis"`
}

type MySQLConfig struct {
    Host         string `mapstructure:"host"`
    User         string `mapstructure:"user"`
    Password     string `mapstructure:"password"`
    DB           string `mapstructure:"dbname"`
    Port         int    `mapstructure:"port"`
    MaxOpenConns int    `mapstructure:"max_open_conns"`
    MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
    Host         string `mapstructure:"host"`
    Password     string `mapstructure:"password"`
    Port         int    `mapstructure:"port"`
    DB           int    `mapstructure:"db"`
    PoolSize     int    `mapstructure:"pool_size"`
    MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type LogConfig struct {
    Level      string `mapstructure:"level"`
    Filename   string `mapstructure:"filename"`
    MaxSize    int    `mapstructure:"max_size"`
    MaxAge     int    `mapstructure:"max_age"`
    MaxBackups int    `mapstructure:"max_backups"`
}

func Init(filepath string) (err error) {
    // 告诉 viper 要读取的配置文件的具体路径
    viper.SetConfigFile(filepath)

    // 实际读取并解析配置文件内容
    err = viper.ReadInConfig()
    if err != nil {
        fmt.Printf("viper.ReadInConfig failed, err:%v\n", err)
        return
    }

    // 反序列化，将配置数据绑定到程序的结构体变量中
    if err = viper.Unmarshal(Conf); err != nil {
        fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
    }

    // 配置热更新
    viper.WatchConfig() // 启动文件监控，viper 会在后台监控配置文件的变化
    // 注册回调函数，当文件发生变化时自动调用
    // fsnotify.Event 包含文件变化的事件信息
    /*
       热更新流程
       用户修改配置文件
       fsnotify 检测到文件变化
       viper 自动重新读取配置文件
       回调函数被触发，重新解析配置到 Conf 变量
       程序运行时配置立即更新
    */
    viper.OnConfigChange(func(in fsnotify.Event) {
        fmt.Println("配置文件修改了")
        if err = viper.Unmarshal(Conf); err != nil {
            fmt.Printf("viper.Unmarshal failed, err:%v\n", err)
        }
    })

    return nil
}
