package config

import (
	"log"

	"github.com/spf13/viper"
)

const (
	// Mailer 邮件配置标识
	Mailer = "mailer"

	// configDir 子配置文件目录
	configDir = "./config/cfg"
)

var (
	// Config 主配置文件实例
	Config *viper.Viper

	// Cfg 子配置文件映射表
	Cfg = make(map[string]*viper.Viper)
)

// init 初始化配置系统
func init() {
	if err := initConfig(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}
	if err := initCfgFiles(); err != nil {
		log.Fatalf("子配置初始化失败: %v", err)
	}
	log.Println("配置系统初始化成功")
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

// initCfgFiles 初始化所有子配置文件
func initCfgFiles() error {
	// 定义需要加载的子配置文件列表
	cfgFiles := []string{Mailer}

	for _, cfgName := range cfgFiles {
		if err := loadCfgFile(cfgName); err != nil {
			return err
		}
	}

	return nil
}

// loadCfgFile 加载指定的子配置文件
func loadCfgFile(cfgName string) error {
	v := viper.New()
	v.AddConfigPath(configDir)
	v.SetConfigName(cfgName)
	v.SetConfigType("json")

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	Cfg[cfgName] = v
	return nil
}

// GetCfg 获取指定的子配置实例，如果不存在则返回 nil
func GetCfg(cfgName string) *viper.Viper {
	return Cfg[cfgName]
}

// MustGetCfg 获取指定的子配置实例，如果不存在则 panic
func MustGetCfg(cfgName string) *viper.Viper {
	cfg := Cfg[cfgName]
	if cfg == nil {
		log.Panicf("配置 '%s' 不存在", cfgName)
	}
	return cfg
}

// ReloadCfg 重新加载指定的子配置文件
func ReloadCfg(cfgName string) error {
	return loadCfgFile(cfgName)
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
