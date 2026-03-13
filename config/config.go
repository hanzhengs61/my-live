package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port           string `yaml:"port"`
		JwtSecret      string `yaml:"jwt_secret"`
		JwtExpiryHours int    `yaml:"jwt_expiry_hours"`
	} `yaml:"server"`

	Database struct {
		DSN string `yaml:"dsn"`
	} `yaml:"database"`
}

var (
	once     sync.Once
	instance *Config
)

// LoadConfig 加载配置
func LoadConfig() *Config {
	once.Do(func() {
		configPath := "config.yaml"

		data, err := os.ReadFile(configPath)
		if err != nil {
			panic(fmt.Sprintf("读取 config.yaml 失败: %v", err))
		}

		var cfg Config
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			panic(fmt.Sprintf("解析 config.yaml 失败: %v", err))
		}

		// 校验
		if cfg.Server.JwtSecret == "" {
			panic("JWT Secret 未配置")
		}
		if cfg.Database.DSN == "" {
			panic("数据库 DSN 未配置")
		}

		instance = &cfg
	})
	return instance
}

// GetJWTExpiryDuration 返回 JWT 过期时间
func (c *Config) GetJWTExpiryDuration() time.Duration {
	return time.Duration(c.Server.JwtExpiryHours) * time.Hour
}
