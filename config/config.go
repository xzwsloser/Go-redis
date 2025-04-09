package config

import (
	"github.com/spf13/viper"
)

type RedisServerConfig struct {
	Address string `yaml:"Address"`
	Port    int    `yaml:"Port"`
}

type LogConfig struct {
	Stdout   string `yaml:"Stdout"`
	File     string `yaml:"File"`
	Filename string `yaml:"Filename"`
	Color    string `yaml:"Color"`
	Level    string `yaml:"Level"`
}

func init() {
	InitConfig()
}

var (
	redisServerConfig *RedisServerConfig = new(RedisServerConfig)
	logConfig         *LogConfig         = new(LogConfig)
)

func GetRedisServerConfig() *RedisServerConfig {
	return redisServerConfig
}

func GetLogConfig() *LogConfig {
	return logConfig
}

func InitConfig() {
	viper.SetConfigName("redis")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	viper.AddConfigPath("../..")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("Redis", redisServerConfig)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("Log", logConfig)
	if err != nil {
		panic(err)
	}
}
