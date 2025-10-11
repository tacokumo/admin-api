package pg

import (
	"fmt"
	"net/url"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c Config) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   fmt.Sprintf("%s:%d", c.Host, c.Port),
		Path:   c.DBName,
	}
	q := u.Query()
	if c.SSLMode != "" {
		q.Set("sslmode", c.SSLMode)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
