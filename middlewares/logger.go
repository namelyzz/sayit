package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/namelyzz/sayit/config"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"
)

// 全局日志编码器，可通过 zap.L() 全局获取
var lg *zap.Logger

/*
Init 初始化全局日志组件
1. 读取配置
2. 使用 lumberjack 返回一个 WriteSyncer
3. 构建 zap 的 encoder
4. 根据 mode 决定是否将日志输出到终端
5. 组合成为 zapcore.Core
6. 生成 logger，设置为全局 logger

参数：

	cfg  -> 日志配置，包括文件名、大小、保留份数、日志级别等
	mode -> 模式："dev" 表示开发环境，会额外输出到控制台
*/
func Init(cfg *config.LogConfig, mode string) (err error) {
	// 1. 创建日志写入器，作用是写入文件和日志切割
	writeSyncer := getLogWriter(cfg.Filename, cfg.MaxSize, cfg.MaxBackups, cfg.MaxAge)

	// 2. 创建日志编码器，用于定义日志格式，比如时间、字段名等等
	encoder := getEncoder()

	// 3. 设置日志级别，info / debug / error 等
	level := new(zapcore.Level)
	// 从配置中读取配置等级并解析为 zapcore.Level 类型
	err = level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return err
	}

	// 4. 创建日志核心（Core），通过 encoder + writeSyncer + level 组合而成
	var core zapcore.Core
	if mode == "dev" {
		// 开发模式，在终端打印日志

		// 创建人类刻度的控制台编码器，可以设置颜色、缩进等
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

		// NewTee 允许把日志同时写入多个目标（文件 + 控制台）
		core = zapcore.NewTee(
			// 写入文件的 core（结构化 JSON 格式）
			zapcore.NewCore(encoder, writeSyncer, level),
			// 写入控制台的 core（开发格式）
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel),
		)
	} else {
		// 生产模式下，只写入文件，不打印到控制台
		core = zapcore.NewCore(encoder, writeSyncer, level)
	}

	// 5. 构建 logger 对象
	// zap.AddCaller()：在日志中显示调用位置（文件名:行号）
	lg = zap.New(core, zap.AddCaller())

	// 6. 设置为全局 logger。这样在其他包可以直接用 zap.L() 调用
	zap.ReplaceGlobals(lg)
	zap.L().Info("init logger success")
	return nil
}

// getLogWriter 使用 lumberjack 实现日志文件切割
//   - filename: 日志文件名
//   - maxSize: 单个文件最大尺寸（MB）
//   - maxBackup: 最多保留多少个旧日志文件
//   - maxAge: 日志保留天数
func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	// lumberjack.Logger 是一个 io.Writer
	// 用于日志文件滚动（按文件大小、保留份数、保留天数）
	lumberJackLogger := &lumberjack.Logger{
		Filename:   filename,  // 日志文件路径
		MaxSize:    maxSize,   // 单个日志文件最大体积 (MB)
		MaxBackups: maxBackup, // 旧日志文件最大保留数
		MaxAge:     maxAge,    // 日志保留天数
	}

	// zap 要求写入器是 zapcore.WriteSyncer 类型
	// AddSync() 用于将普通 io.Writer 封装成 zapcore.WriteSyncer
	return zapcore.AddSync(lumberJackLogger)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()             // 生产环境推荐使用 ProductionEncoderConfig（JSON 格式）
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder         // 设置时间格式为 ISO8601（例如 "2025-02-07T12:34:56.789Z"）
	encoderConfig.TimeKey = "time"                                // 时间字段名（默认是 "ts"）
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder       // 日志级别使用大写（INFO、ERROR）
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder // 时长字段使用秒（例如 duration=1.234）
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder       // 调用方信息使用短路径（如 "main.go:42"）
	return zapcore.NewJSONEncoder(encoderConfig)                  // 使用 JSON 格式编码（结构化日志）
}

/*
GinLogger 用于 Gin 框架的日志中间件，使用 zap 日志系统来记录 HTTP 请求的访问日志
流程大致如下：
1. 记录请求开始时间
2. 保存请求路径和查询参数
3. 执行下一个中间件或业务逻辑
4. 记录请求处理完成后的日志
参照了 gin.Logger() 的源码，详见 LoggerWithConfig 函数的返回部分
这里做的是增加了耗时，并把一些关键日志信息用 zap 记录
*/
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 记录请求开始时间
		start := time.Now()

		// 保存请求的 URL 和查询参数
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 让请求继续执行
		c.Next()

		// 请求执行后，计算耗时
		cost := time.Since(start)

		// 使用 zap 记录日志
		lg.Info(
			path,
			zap.Int("status", c.Writer.Status()),            // HTTP 状态码（例如 200, 404, 500）
			zap.String("method", c.Request.Method),          // 请求方法（GET、POST 等）
			zap.String("path", path),                        // 请求路径
			zap.String("query", query),                      // 查询字符串
			zap.String("ip", c.ClientIP()),                  // 客户端 IP
			zap.String("user-agent", c.Request.UserAgent()), // 客户端 UA 信息
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()), // Gin 中的内部错误
			zap.Duration("cost", cost), // 请求耗时
		)
	}
}

/*
GinRecovery 自定义的 zap 版本的 gin.Recovery()
用于捕获代码运行中的 panic（程序崩溃），记录错误日志，防止整个服务挂掉
相较于 GinLogger, GinRecovery 的改动并不多，只是把 gin.Recovery 中本身用 logger 的地方改为 zap 的 logger
1. 捕获 panic，防止程序崩溃
2. 记录详细错误日志（包括堆栈）
3. 返回 500 状态码给客户端
参数：

	stack -> 是否在日志中打印堆栈信息
*/
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 使用 defer 语句会在函数退出前执行，用于捕获 panic
		defer func() {
			if err := recover(); err != nil {
				// ----------- 1. 检查是否是连接断开的错误 -----------
				// 有时候 panic 是因为客户端主动断开连接（例如“broken pipe”），
				// 这种错误不需要打印堆栈。

				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok { // 网络操作错误
					if se, ok := ne.Err.(*os.SyscallError); ok { // 系统调用错误
						// 转成小写后判断是否包含关键字
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") ||
							strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// ----------- 2. 打印请求信息 -----------
				// DumpRequest 可以把 HTTP 请求的头部、URL、方法打印出来
				// 第二个参数 false 表示不打印 body（避免太大）
				httpRequest, _ := httputil.DumpRequest(c.Request, false)

				if brokenPipe {
					// 如果是 broken pipe 错误，记录后直接返回
					lg.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)

					// 将错误附加到 Gin 的上下文中（方便上层处理）
					c.Error(err.(error)) // nolint: errcheck
					// 终止请求，不再继续执行后续中间件
					c.Abort()
					return
				}

				// ----------- 3. 记录 panic 错误日志 -----------
				if stack {
					// 如果要求打印堆栈信息
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),                      // panic 的错误信息
						zap.String("request", string(httpRequest)), // 请求详情
						zap.String("stack", string(debug.Stack())), // 程序堆栈
					)
				} else {
					// 只打印错误，不打印堆栈
					lg.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				// ----------- 4. 返回 500 内部服务器错误 -----------
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		// ----------- 5. 继续执行后续中间件或 handler -----------
		// 如果没有 panic，程序会正常执行到这里
		c.Next()
	}
}
