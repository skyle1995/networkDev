package logger

import (
	log "github.com/sirupsen/logrus"
)

// LogServerStart 记录服务器启动日志
// host: 服务器监听地址
// port: 服务器监听端口
func (l *Logger) LogServerStart(host string, port int) {
	l.WithFields(log.Fields{
		"host": host,
		"port": port,
	}).Info("HTTP服务器启动")
}

// LogServerStop 记录服务器停止日志
func (l *Logger) LogServerStop() {
	l.Info("HTTP服务器停止")
}

// LogConfigLoad 记录配置加载日志
// configFile: 配置文件路径
func (l *Logger) LogConfigLoad(configFile string) {
	l.WithField("config_file", configFile).Info("配置文件加载")
}