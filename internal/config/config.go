package config

import (
	"errors"
	"strings"
)

type Config struct{ Address, DatabaseURL, ManagerToken string }

func Load(getenv func(string) string, requireToken bool) (Config, error) {
	c := Config{Address: getenv("BANDIA_HTTP_ADDR"), DatabaseURL: getenv("BANDIA_DATABASE_URL"), ManagerToken: getenv("BANDIA_MANAGER_TOKEN")}
	if c.Address == "" {
		c.Address = "127.0.0.1:8080"
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return c, errors.New("BANDIA_DATABASE_URL is required")
	}
	if requireToken && (len(c.ManagerToken) < 32 || strings.ContainsAny(c.ManagerToken, " \t\r\n")) {
		return c, errors.New("BANDIA_MANAGER_TOKEN must contain at least 32 characters without whitespace")
	}
	return c, nil
}
