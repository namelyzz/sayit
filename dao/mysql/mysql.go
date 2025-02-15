package mysql

import (
    "fmt"
    "github.com/namelyzz/sayit/config"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "time"
)

var db *gorm.DB

func Init(cfg *config.MySQLConfig) (err error) {
    // 构建 DSN（Data Source Name）
    // 格式: user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

    gormCfg := &gorm.Config{
        // 开启详细日志（可根据环境配置）
        Logger: logger.Default.LogMode(logger.Info),
        // 禁止外键约束（推荐在大部分服务中关闭）
        DisableForeignKeyConstraintWhenMigrating: true,
    }

    // 连接数据库
    db, err = gorm.Open(mysql.Open(dsn), gormCfg)
    if err != nil {
        return fmt.Errorf("failed to connect database: %w", err)
    }

    // 获取底层 *sql.DB 对象（用于设置连接池）
    sqlDB, err := db.DB()
    if err != nil {
        return fmt.Errorf("failed to get underlying sql.DB: %w", err)
    }

    // 配置连接池
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)    // 最大连接数
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)    // 最大空闲连接数
    sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最长存活时间
    sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最长保持时间

    return nil
}

// Close 关闭数据库连接
func Close() {
    sqlDB, err := db.DB()
    if err == nil {
        _ = sqlDB.Close()
    }
}

// DB 获取全局 GORM 数据库对象
// 用于在其他包中访问数据库
func DB() *gorm.DB {
    return db
}
