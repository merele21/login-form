package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" env-default:"local" env-required:"true"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`

	GRPC     GRPCConfig     `yaml:"grpc"`
	HTTP     HTTPConfig     `yaml:"http"`
	TLS      TLSConfig      `yaml:"tls"`
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type HTTPConfig struct {
	Port               int      `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	CORSAllowedOrigins []string `yaml:"cors_allowed_origins"` // опционально, для фронта
}

type PostgresConfig struct {
	DSN          string `yaml:"dsn" env:"POSTGRES_DSN"` // напр. postgres://sso:sso@postgres:5432/sso?sslmode=disable
	MaxOpenConns int    `yaml:"max_open_conns" env-default:"20"`
	MaxIdleConns int    `yaml:"max_idle_conns" env-default:"5"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR" env-default:"redis:6379"`
	DB       int    `yaml:"db" env-default:"0"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"` // опционально
}

type TLSConfig struct {
	CertFile string `yaml:"cert_file" env:"TLS_CERT_FILE"`
	KeyFile  string `yaml:"key_file" env:"TLS_KEY_FILE"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable
// priority: flag > env > default
// default value is empty string
func fetchConfigPath() string {
	var res string

	// --config="path/to/config.yaml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
