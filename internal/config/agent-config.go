package config

import "os"

type AgentConfig struct {
	DbConfig
}

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		DbConfig: DbConfig{
			Host:     getEnv("DB_HOST", "database"),
			Port:     3306,
			User:     getEnv("DB_USER", "rc-user"),
			Password: getEnv("DB_PASSWORD", "rc-password"),
			Database: getEnv("DB_DATABASE", "executions"),
		},
	}
}

func getEnv(key string, defaultVaule string) string {
	value:= os.Getenv(key)
	if value==""{
		return defaultVaule
	}
	return value
}