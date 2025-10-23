package cmd

import (
	"io"
	"networkDev/config"
	"networkDev/utils/logger"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd 代表没有调用子命令时的基础命令
var rootCmd = &cobra.Command{
	Use:   "networkDev",
	Short: "一个基于Cobra的网络验证服务器应用",
	Long: `networkDev是一个使用Cobra CLI框架构建的网络验证服务器应用，
集成了Viper配置管理、Logrus日志记录和embed静态资源嵌入功能。`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 在加载配置前配置logrus用于非HTTP日志

		setupLogrusForNonHTTP()

	},
}

// Execute 添加所有子命令到根命令并设置适当的标志
// 这由main.main()调用。只需要对rootCmd执行一次。
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// 在这里定义标志和配置设置
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "配置文件路径 (默认为 config.json)")
}

// setupLogrusForNonHTTP 配置logrus用于非HTTP日志
// 在加载配置文件之前进行基本的logrus设置
func setupLogrusForNonHTTP() {
	// 设置日志格式
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     false,
		DisableColors:   true,
	})

	// 设置默认日志级别
	logrus.SetLevel(logrus.InfoLevel)

	// 设置输出目标（稍后会根据配置文件调整）
	logrus.SetOutput(os.Stdout)
	if cfgFile != "" {
		// 使用命令行指定的配置文件
		config.Init(cfgFile)
	} else {
		// 使用默认配置文件路径
		config.Init("./config.json")
	}

	// 根据配置文件进一步配置logrus
	setupLogrusFromConfig()

	// 初始化HTTP日志处理器
	logger.InitLogger()

	// 记录配置加载完成
	logrus.WithField("config_file", viper.ConfigFileUsed()).Info("配置文件加载完成")
}

// initConfig 读取配置文件和环境变量
func initConfig() {

}

// setupLogrusFromConfig 根据配置文件进一步配置logrus
// 设置日志级别和输出目标
func setupLogrusFromConfig() {
	// 设置日志级别
	if level := viper.GetString("log.level"); level != "" {
		if logLevel, err := logrus.ParseLevel(level); err == nil {
			logrus.SetLevel(logLevel)
		}
	}

	// 设置日志输出目标
	logFile := viper.GetString("log.file")
	if logFile != "" {
		// 确保日志目录存在
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err == nil {
			// 打开日志文件
			if file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
				// 同时输出到控制台和文件
				multiWriter := io.MultiWriter(os.Stdout, file)
				logrus.SetOutput(multiWriter)
			}
		}
	}
	// 当日志文件路径为空时，保持默认输出到控制台，不创建任何目录
}
