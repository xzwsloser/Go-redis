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

type DBConfig struct {
	Number int `yaml:"Number"`
}

type AofConfig struct {
	Load           string `yaml:"Load"`
	AppendOnly     string `yaml:"AppendOnly"`
	AppendFileName string `yaml:"AppendFileName"`
	AppendFileSync string `yaml:"AppendFileSync"`
}

func init() {
	InitConfig()
}

var (
	redisServerConfig *RedisServerConfig = new(RedisServerConfig)
	logConfig         *LogConfig         = new(LogConfig)
	dbConfig          *DBConfig          = new(DBConfig)
	aofConfig         *AofConfig         = new(AofConfig)
)

func GetRedisServerConfig() *RedisServerConfig {
	return redisServerConfig
}

func GetLogConfig() *LogConfig {
	return logConfig
}

func GetDBConfig() *DBConfig {
	return dbConfig
}

func GetAofConfig() *AofConfig {
	return aofConfig
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

	err = viper.UnmarshalKey("DB", dbConfig)
	if err != nil {
		panic(err)
	}

	err = viper.UnmarshalKey("Aof", aofConfig)
	if err != nil {
		panic(err)
	}
}
