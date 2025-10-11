package config

import "github.com/tacokumo/admin-api/pkg/envconfig"

type Config struct {
	Addr          string `env:"ADDR"`
	Port          string `env:"PORT"`
	AdminDBConfig AdminDBConfig
	Auth          AuthConfig
	CORS          CORSConfig
	TLS           TLSConfig
}

type AuthConfig struct {
	Auth0Domain   string `env:"AUTH0_DOMAIN"`
	Auth0Audience string `env:"AUTH0_AUDIENCE"`
	Auth0ClientID string `env:"AUTH0_CLIENT_ID"`
	// Auth0ClientSecret is required only when you use Management API.
	Auth0ClientSecret string `env:"AUTH0_CLIENT_SECRET"`
}

type AdminDBConfig struct {
	Host             string `env:"ADMIN_DB_HOST"`
	Port             int    `env:"ADMIN_DB_PORT"`
	User             string `env:"ADMIN_DB_USER"`
	Password         string `env:"ADMIN_DB_PASSWORD"`
	DBName           string `env:"ADMIN_DB_NAME"`
	InitialConnRetry int    `env:"ADMIN_DB_INITIAL_CONN_RETRY"`
}

type CORSConfig struct {
	AllowOrigins     string `env:"CORS_ALLOW_ORIGINS"`
	AllowMethods     string `env:"CORS_ALLOW_METHODS"`
	AllowHeaders     string `env:"CORS_ALLOW_HEADERS"`
	ExposeHeaders    string `env:"CORS_EXPOSE_HEADERS"`
	AllowCredentials bool   `env:"CORS_ALLOW_CREDENTIALS"`
	MaxAge           int    `env:"CORS_MAX_AGE"`
}

type TLSConfig struct {
	Enabled  bool   `env:"TLS_ENABLED"`
	CertFile string `env:"TLS_CERT_FILE"`
	KeyFile  string `env:"TLS_KEY_FILE"`
}

func LoadFromEnv() (Config, error) {
	return envconfig.LoadFromEnv[Config]()
}
