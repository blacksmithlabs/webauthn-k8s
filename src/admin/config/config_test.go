package config

import (
	"testing"
	"time"
)

func TestGetRedisPoolSize(t *testing.T) {
	curRedisPoolSize := redisPoolSize
	defer func() {
		redisPoolSize = curRedisPoolSize
	}()

	// Test case 1 default
	redisPoolSize = ""
	if v := GetRedisPoolSize(); v != defaultRedisPoolSize {
		t.Errorf("GetRedisPoolSize() = %v, want %v", v, defaultRedisPoolSize)
	}

	// Test case 2 value
	redisPoolSize = "20"
	if v := GetRedisPoolSize(); v != 20 {
		t.Errorf("GetRedisPoolSize() = %v, want %v", v, 20)
	}
}

func TestGetRedisHost(t *testing.T) {
	curRedisHost := redisHost
	defer func() {
		redisHost = curRedisHost
	}()

	// Test case 1 default
	redisHost = ""
	if v := GetRedisHost(); v != defaultRedisHost {
		t.Errorf("GetRedisHost() = %v, want %v", v, defaultRedisHost)
	}

	// Test case 2 value
	redisHost = "localhost:6379"
	if v := GetRedisHost(); v != "localhost:6379" {
		t.Errorf("GetRedisHost() = %v, want %v", v, "localhost:6379")
	}
}

func TestGetRedisPassword(t *testing.T) {
	curRedisPassword := redisPassword
	defer func() {
		redisPassword = curRedisPassword
	}()

	// Test case 1
	redisPassword = "password"
	if v := GetRedisPassword(); v != "password" {
		t.Errorf("GetRedisPassword() = %v, want %v", v, "password")
	}
}

func TestGetSessionSecret(t *testing.T) {
	curSessionSecret := sessionSecret
	defer func() {
		sessionSecret = curSessionSecret
	}()

	// Test case
	sessionSecret = "secret"
	if v := GetSessionSecret(); string(v) != "secret" {
		t.Errorf("GetSessionSecret() = %v, want %v", v, "secret")
	}
}

func TestGetSessionTimeout(t *testing.T) {
	curSessionTimeout := sessionTimeout
	defer func() {
		sessionTimeout = curSessionTimeout
	}()

	defaultTime := defaultSessionTimeout * time.Second

	// Test case 1 default
	sessionTimeout = ""
	if v := GetSessionTimeout(); v != defaultTime {
		t.Errorf("GetSessionTimeout() = %v, want %v", v, defaultTime)
	}

	// Test case 2 value
	sessionTimeout = "10"
	if v := GetSessionTimeout(); v != 10*time.Second {
		t.Errorf("GetSessionTimeout() = %v, want %v", v, 10*time.Second)
	}

	// Test case 3 invalid integer
	sessionTimeout = "invalid"
	if v := GetSessionTimeout(); v != defaultTime {
		t.Errorf("GetSessionTimeout() = %v, want %v", v, defaultTime)
	}

	// Test case 4 negative integer
	sessionTimeout = "-10"
	if v := GetSessionTimeout(); v != defaultTime {
		t.Errorf("GetSessionTimeout() = %v, want %v", v, defaultTime)
	}
}

func TestGetPostgresUrl(t *testing.T) {
	curPostgresUrl := postgresUrl
	defer func() {
		postgresUrl = curPostgresUrl
	}()

	// Test case 1
	postgresUrl = "postgres://username:password@localhost:5432/database_name"
	if v := GetPostgresUrl(); v != "postgres://username:password@localhost:5432/database_name" {
		t.Errorf("GetPostgresUrl() = %v, want %v", v, "postgres://username:password@localhost:5432/database_name")
	}
}

func TestGetAppPort(t *testing.T) {
	curAppPort := appPort
	defer func() {
		appPort = curAppPort
	}()

	// Test case 1 default
	appPort = ""
	if v := GetAppPort(); v != defaultAppPort {
		t.Errorf("GetAppPort() = %v, want %v", v, defaultAppPort)
	}

	// Test case 2 value
	appPort = "8080"
	if v := GetAppPort(); v != "8080" {
		t.Errorf("GetAppPort() = %v, want %v", v, "8080")
	}
}
