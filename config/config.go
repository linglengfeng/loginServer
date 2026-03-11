package config

import (
	"log"

	"github.com/spf13/viper"
)

var (
	// Config 主配置文件实例
	Config *viper.Viper
)

// init 初始化配置系统
func init() {
	if err := initConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}
}

// initConfig 初始化主配置文件
func initConfig() error {
	Config = viper.New()
	Config.AddConfigPath("./")
	Config.SetConfigName("config")
	Config.SetConfigType("json")

	if err := Config.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

// GetConfigPath 获取配置文件路径（用于调试）
func GetConfigPath() string {
	if Config == nil {
		return ""
	}
	return Config.ConfigFileUsed()
}

// ValidateConfig 验证主配置是否已正确加载
func ValidateConfig() bool {
	return Config != nil && Config.ConfigFileUsed() != ""
}
