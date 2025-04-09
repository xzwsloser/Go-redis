package config

import (
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	InitConfig()
	log.Println(GetRedisServerConfig().Address)
	log.Println(GetRedisServerConfig().Port)
}
