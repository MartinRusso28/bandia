package config

import (
	"strings"
	"testing"
)

func TestConfigFailsClosed(t *testing.T) {
	vars := map[string]string{}
	get := func(k string) string { return vars[k] }
	if _, err := Load(get, true); err == nil {
		t.Fatal("missing database accepted")
	}
	vars["BANDIA_DATABASE_URL"] = "postgres://private:secret@localhost/bandia"
	if _, err := Load(get, true); err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("missing token must fail without exposing DSN")
	}
	if _, err := Load(get, false); err != nil {
		t.Fatal("migrations need no manager token", err)
	}
	vars["BANDIA_MANAGER_TOKEN"] = strings.Repeat("a", 32)
	c, err := Load(get, true)
	if err != nil || c.Address != "127.0.0.1:8080" {
		t.Fatal(c.Address, err)
	}
	vars["BANDIA_MANAGER_TOKEN"] += "\n"
	if _, err := Load(get, true); err == nil {
		t.Fatal("whitespace in token accepted")
	}
}
