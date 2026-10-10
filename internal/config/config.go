package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Dbname   string `yaml:"dbname"`
	Sslmode  string `yaml:"sslmode"`
}

type JWTConfig struct {
	Secret string        `yaml:"secret"`
	TTL    time.Duration `yaml:"ttl"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Postgres PostgresConfig `yaml:"postgres"`
	JWT      JWTConfig      `yaml:"jwt"`
}

func validate(cfg *Config) error {
	if cfg.Postgres.Password == "" {
		return fmt.Errorf("postgres.password is empty (is DB_PASSWORD set?)")
	}
	if cfg.Postgres.User == "" {
		return fmt.Errorf("postgres.user is empty (is DB_USER set?)")
	}
	if cfg.Postgres.Dbname == "" {
		return fmt.Errorf("postgres.dbname is empty (is DB_NAME set?)")
	}
	if len(cfg.JWT.Secret) < 32 {
		return fmt.Errorf("jwt.secret is too short (is JWT_SECRET at lest 32 bytes?)")
	}
	if cfg.JWT.TTL <= 0 {
		return fmt.Errorf("jwt.ttl is empty (is ttl set?)")
	}
	return nil
}

func Load(path string) (*Config, error) {
	var cfg Config
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	data = []byte(os.ExpandEnv(string(data)))

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, fmt.Errorf("parse config %q: %w", path, err)
	}
	if err := validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (cfg PostgresConfig) DSN() string {
	q := url.Values{}
	q.Set("sslmode", cfg.Sslmode)

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, cfg.Port),
		Path:     cfg.Dbname,
		RawQuery: q.Encode(),
	}

	return u.String()
}
