package config

import (
	"bytes"
	_ "embed"
	"errors"
	"io/fs"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

//go:embed config.json
var DefaultConfig string

// Init 初始化配置文件
func Init(cfgFilePath string) {
	viper.SetConfigFile(cfgFilePath)
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			log.Warn("未找到配置文件，使用默认配置")
			err = os.WriteFile(cfgFilePath, []byte(DefaultConfig), 0o644)
			if err != nil {
				log.WithFields(
					log.Fields{
						"err": err,
					},
				).Error("写入默认配置文件失败")
			} else {
				log.WithFields(
					log.Fields{
						"file": cfgFilePath,
					},
				).Info("写入默认配置文件成功")
			}
			// 写完默认配置后再读一次
			err = viper.ReadConfig(bytes.NewBuffer([]byte(DefaultConfig)))
			if err != nil {
				log.WithFields(
					log.Fields{
						"err": err,
					},
				).Error("读取默认配置文件失败")
			} else {
				log.Info("已成功读取默认配置")
			}
		} else {
			log.WithFields(
				log.Fields{
					"err": err,
				},
			).Fatal("配置文件解析错误")
		}
	}
	log.WithFields(
		log.Fields{
			"file": viper.ConfigFileUsed(),
		},
	).Info("使用配置文件")

	// 验证配置并设置默认值
	if _, err := ValidateAndSetDefaults(); err != nil {
		log.WithFields(
			log.Fields{
				"err": err,
			},
		).Fatal("配置验证失败")
	}
}

// CreateDefaultConfig 创建默认配置文件
func CreateDefaultConfig(filePath string) error {
	return os.WriteFile(filePath, []byte(DefaultConfig), 0o644)
}
