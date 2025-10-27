package timeutil

import (
	"fmt"
	"time"
)

// serverStartTime 记录进程启动时间（近似服务器启动时间）
var serverStartTime = time.Now()

// GetServerStartTime 获取服务器启动时间
// 返回: 服务器启动的时间戳
func GetServerStartTime() time.Time {
	return serverStartTime
}

// GetServerUptime 获取服务器运行时长
// 返回: 从服务器启动到现在的时间间隔
func GetServerUptime() time.Duration {
	return time.Since(serverStartTime)
}

// GetServerUptimeString 获取服务器运行时长的字符串表示
// 返回: 格式化的运行时长字符串（中文单位）
func GetServerUptimeString() string {
	duration := time.Since(serverStartTime)

	// 获取总秒数并转换为整数
	totalSeconds := int(duration.Seconds())

	// 计算天、小时、分钟、秒
	days := totalSeconds / 86400
	hours := (totalSeconds % 86400) / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	// 根据时长长度选择合适的格式
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟%d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}
