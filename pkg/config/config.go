package config

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/tacokumo/admin-api/pkg/envconfig"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Addr          string        `env:"ADDR" yaml:"addr"`
	Port          string        `env:"PORT" yaml:"port"`
	AdminDBConfig AdminDBConfig `yaml:"admin_db"`
	Auth          AuthConfig    `yaml:"auth"`
	CORS          CORSConfig    `yaml:"cors"`
	TLS           TLSConfig     `yaml:"tls"`
}

type AuthConfig struct {
	Auth0Domain   string `env:"AUTH0_DOMAIN" yaml:"auth0_domain"`
	Auth0Audience string `env:"AUTH0_AUDIENCE" yaml:"auth0_audience"`
	Auth0ClientID string `env:"AUTH0_CLIENT_ID" yaml:"auth0_client_id"`
	// Auth0ClientSecret is required only when you use Management API.
	Auth0ClientSecret string `env:"AUTH0_CLIENT_SECRET" yaml:"auth0_client_secret"`
}

type AdminDBConfig struct {
	Host             string `env:"ADMIN_DB_HOST" yaml:"host"`
	Port             int    `env:"ADMIN_DB_PORT" yaml:"port"`
	User             string `env:"ADMIN_DB_USER" yaml:"user"`
	Password         string `env:"ADMIN_DB_PASSWORD" yaml:"password"`
	DBName           string `env:"ADMIN_DB_NAME" yaml:"db_name"`
	InitialConnRetry int    `env:"ADMIN_DB_INITIAL_CONN_RETRY" yaml:"initial_conn_retry"`
}

type CORSConfig struct {
	AllowOrigins     string `env:"CORS_ALLOW_ORIGINS" yaml:"allow_origins"`
	AllowMethods     string `env:"CORS_ALLOW_METHODS" yaml:"allow_methods"`
	AllowHeaders     string `env:"CORS_ALLOW_HEADERS" yaml:"allow_headers"`
	ExposeHeaders    string `env:"CORS_EXPOSE_HEADERS" yaml:"expose_headers"`
	AllowCredentials bool   `env:"CORS_ALLOW_CREDENTIALS" yaml:"allow_credentials"`
	MaxAge           int    `env:"CORS_MAX_AGE" yaml:"max_age"`
}

type TLSConfig struct {
	Enabled  bool   `env:"TLS_ENABLED" yaml:"enabled"`
	CertFile string `env:"TLS_CERT_FILE" yaml:"cert_file"`
	KeyFile  string `env:"TLS_KEY_FILE" yaml:"key_file"`
}

func LoadFromEnv() (Config, error) {
	return envconfig.LoadFromEnv[Config]()
}

// LoadFromYAML loads configuration from a YAML file.
func LoadFromYAML(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, errors.Wrapf(err, "failed to read config file: %s", path)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, errors.Wrapf(err, "failed to unmarshal yaml config")
	}

	return cfg, nil
}

// LoadFromYAMLWithEnvOverride loads configuration from a YAML file and overrides
// with environment variables if they are set.
func LoadFromYAMLWithEnvOverride(path string) (Config, error) {
	cfg, err := LoadFromYAML(path)
	if err != nil {
		return cfg, err
	}

	if err := envconfig.OverrideFromEnv(&cfg); err != nil {
		return cfg, errors.Wrap(err, "failed to override config from env")
	}

	return cfg, nil
}
