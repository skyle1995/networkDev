package cmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"networkDev/database"
	"networkDev/middleware"
	"networkDev/server"
	"networkDev/utils"
	"networkDev/utils/logger"
	"networkDev/web"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd 代表服务器命令
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "启动HTTP服务器",
	Long:  `启动一个简单的HTTP服务器，监听配置文件中指定的端口。`,
	Run:   runServer,
}

func init() {
	// 将服务器命令添加到根命令
	rootCmd.AddCommand(serverCmd)

	// 添加服务器特定的标志
	serverCmd.Flags().StringP("host", "H", "", "服务器监听地址 (覆盖配置文件)")
	serverCmd.Flags().IntP("port", "p", 0, "服务器监听端口 (覆盖配置文件)")
}

// runServer 运行HTTP服务器
func runServer(cmd *cobra.Command, args []string) {
	// 获取配置
	host := getServerHost(cmd)
	port := getServerPort(cmd)
	addr := fmt.Sprintf("%s:%d", host, port)

	// 获取全局日志实例
	logger := logger.GetLogger()
	logger.LogServerStart(host, port)

	// 初始化Redis（如果配置存在，失败不致命）
	utils.InitRedis()

	// 初始化数据库（根据 viper 配置选择 SQLite 或 MySQL）
	// 如果初始化失败则回退并退出
	if _, err := database.Init(); err != nil {
		logrus.WithError(err).Fatal("数据库初始化失败")
	}
	// 执行自动迁移（确保表结构存在）
	if err := database.AutoMigrate(); err != nil {
		logrus.WithError(err).Fatal("数据库自动迁移失败")
	}
	// 初始化默认系统设置（包含管理员账号）
	if err := database.SeedDefaultSettings(); err != nil {
		logrus.WithError(err).Fatal("默认系统设置初始化失败")
	}

	// 创建HTTP服务器
	server := createHTTPServer(addr)

	// 启动服务器
	startServer(server)
}

// getServerHost 获取服务器监听地址
func getServerHost(cmd *cobra.Command) string {
	if host, _ := cmd.Flags().GetString("host"); host != "" {
		return host
	}
	return viper.GetString("server.host")
}

// getServerPort 获取服务器监听端口
func getServerPort(cmd *cobra.Command) int {
	if port, _ := cmd.Flags().GetInt("port"); port != 0 {
		return port
	}
	return viper.GetInt("server.port")
}

// createHTTPServer 创建HTTP服务器
func createHTTPServer(addr string) *http.Server {
	// 配置Gin模式和日志
	configureGin()

	// 创建Gin引擎
	router := gin.New()
	
	// 添加恢复中间件
	router.Use(gin.Recovery())

	// 添加日志中间件
	router.Use(middleware.WrapHandler())

	// 添加开发模式中间件（统一管理开发模式功能）
	router.Use(middleware.DevModeMiddleware(router))

	// 加载模板
	if err := loadTemplates(router); err != nil {
		logrus.WithError(err).Fatal("模板加载失败")
	}

	// 注册路由
	registerRoutes(router)

	return &http.Server{
		Addr:    addr,
		Handler: router,
	}
}

// loadTemplates 加载模板到Gin引擎
func loadTemplates(router *gin.Engine) error {
	tmpl, err := web.ParseTemplates()
	if err != nil {
		return err
	}
	router.SetHTMLTemplate(tmpl)
	return nil
}

// registerRoutes 注册HTTP路由
func registerRoutes(router *gin.Engine) {
	// 使用server包中的路由注册函数
	server.RegisterRoutes(router)
}

// startServer 启动服务器并处理优雅关闭
func startServer(server *http.Server) {
	// 获取全局日志实例
	logger := logger.GetLogger()

	// 创建一个通道来接收操作系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中启动服务器
	go func() {
		logger.WithField("addr", server.Addr).Info("HTTP服务器已启动")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.LogError(err, "服务器启动失败")
			os.Exit(1)
		}
	}()

	// 等待中断信号
	<-sigChan
	logger.Info("收到关闭信号，正在优雅关闭服务器...")

	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.LogError(err, "服务器关闭时出错")
	} else {
		logger.LogServerStop()
	}
}

// configureGin 配置Gin的全局设置
func configureGin() {
	// 禁用Gin的颜色输出，提高控制台兼容性
	gin.DisableConsoleColor()
	
	// 设置Gin的输出为丢弃，因为我们使用自定义日志中间件
	gin.DefaultWriter = io.Discard
	gin.DefaultErrorWriter = io.Discard
	
	// 根据配置设置Gin模式
	if viper.GetString("app.mode") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
}
