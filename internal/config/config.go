package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Admin    AdminConfig
}

type AdminConfig struct {
	APIKey string
}

type ServerConfig struct {
	Port       string
	Mode       string
	PublicURL  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	Secret string
	TTL    int
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + d.Port +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/wwvpn")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:      viper.GetString("server.port"),
			Mode:      viper.GetString("server.mode"),
			PublicURL: viper.GetString("server.public_url"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("database.host"),
			Port:     viper.GetString("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			Name:     viper.GetString("database.name"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		Redis: RedisConfig{
			Host:     viper.GetString("redis.host"),
			Port:     viper.GetString("redis.port"),
			Password: viper.GetString("redis.password"),
			DB:       viper.GetInt("redis.db"),
		},
		JWT: JWTConfig{
			Secret: viper.GetString("jwt.secret"),
			TTL:    viper.GetInt("jwt.ttl"),
		},
		Admin: AdminConfig{
			APIKey: viper.GetString("admin.api_key"),
		},
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("server.public_url", "")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "wwvpn")
	viper.SetDefault("database.password", "wwvpn")
	viper.SetDefault("database.name", "wwvpn")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "change-me-jwt-secret-32-chars!!")
	viper.SetDefault("jwt.ttl", 10080)
	viper.SetDefault("admin.api_key", "change-me-admin-api-key-32chars!")
}
