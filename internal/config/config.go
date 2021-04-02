package config

import (
	"os"
	"ratelimit/pkg/logger"
	"ratelimit/pkg/ratelimit/supported"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Logger    LoggerConfig    `mapstructure:"logger"`
	HTTP      HTTPConfig      `mapstructure:"http"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
}

type LoggerConfig struct {
	Level  string           `mapstructure:"level"`
	Format logger.LogFormat `mapstructure:"format"`
}

type HTTPConfig struct {
	Port uint16 `mapstructure:"port"`
}

type RateLimitConfig struct {
	Driver      supported.SupportedDriver `mapstructure:"driver"`
	Limit       int64                     `mapstructure:"limit"`
	Frequency   time.Duration             `mapstructure:"frequency"`
	RedisOption RedisOptionConfig         `mapstructure:"redis_option"`
}

type RedisOptionConfig struct {
	Host         string        `mapstructure:"host"`
	Port         uint          `mapstructure:"port"`
	DB           int           `mapstructure:"db"`
	Password     string        `mapstructure:"password"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxPoolSize  int           `mapstructure:"max_pool_size"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	MaxRetry     int           `mapstructure:"max_retry"`
}

func NewConfig(configPath string) (Config, error) {
	var file *os.File
	file, _ = os.Open(configPath)

	v := viper.New()
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	/* default */
	v.SetDefault("logger.level", "INFO")
	v.SetDefault("logger.format", logger.ConsoleFormat)
	v.SetDefault("http.port", "8080")
	v.SetDefault("ratelimit.driver", supported.MEMORY)
	v.SetDefault("ratelimit.limit", 60)
	v.SetDefault("ratelimit.frequency", time.Minute)
	v.SetDefault("ratelimit.redis_option.host", "")
	v.SetDefault("ratelimit.redis_option.port", 6379)
	v.SetDefault("ratelimit.redis_option.db", 0)
	v.SetDefault("ratelimit.redis_option.password", "")
	v.SetDefault("ratelimit.redis_option.min_idle_conns", 10)
	v.SetDefault("ratelimit.redis_option.max_pool_size", 20)
	v.SetDefault("ratelimit.redis_option.dial_timeout", 15*time.Second)
	v.SetDefault("ratelimit.redis_option.max_retry", 1000)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.ReadConfig(file)

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return config, err
	}

	return config, nil
}
