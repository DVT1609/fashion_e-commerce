package config

import (
	"os"
)

type ConfigMysql struct {
	MysqlRootPassword string
	MysqlDatabase string
	MysqlUser string
	MysqlPassword string
	MysqlHost string
}

// Cách viết chuyên nghiệp: Trả về con trỏ và không dùng "this"
func LoadConfigMysql() *ConfigMysql {
    return &ConfigMysql{
        MysqlRootPassword: os.Getenv("MYSQL_ROOT_PASSWORD"),
        MysqlDatabase:     os.Getenv("MYSQL_DATABASE"),
        MysqlUser:         os.Getenv("MYSQL_USER"),
        MysqlPassword:     os.Getenv("MYSQL_PASSWORD"),
        MysqlHost:         os.Getenv("DB_HOST"),
    }
}