package config

import (
	"reflect"
	"testing"
	"time"
)

func TestGetRedisPoolSize(t *testing.T) {
	curRedisPoolSize := redisPoolSize
	defer func() {
		redisPoolSize = curRedisPoolSize
	}()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "Default",
			input:    "",
			expected: defaultRedisPoolSize,
		},
		{
			name:     "Value",
			input:    "20",
			expected: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisPoolSize = tt.input
			if v := GetRedisPoolSize(); v != tt.expected {
				t.Errorf("GetRedisPoolSize() = %v, want %v", v, tt.expected)
			}
		})
	}
}

func TestGetRedisHost(t *testing.T) {
	curRedisHost := redisHost
	defer func() {
		redisHost = curRedisHost
	}()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Default",
			input:    "",
			expected: defaultRedisHost,
		},
		{
			name:     "Value",
			input:    "localhost:6379",
			expected: "localhost:6379",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisHost = tt.input
			if v := GetRedisHost(); v != tt.expected {
				t.Errorf("GetRedisHost() = %v, want %v", v, tt.expected)
			}
		})
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

	tests := []struct {
		name     string
		input    string
		expected time.Duration
	}{
		{
			name:     "Default",
			input:    "",
			expected: defaultTime,
		},
		{
			name:     "Value",
			input:    "10",
			expected: 10 * time.Second,
		},
		{
			name:     "Invalid integer",
			input:    "invalid",
			expected: defaultTime,
		},
		{
			name:     "Negative integer",
			input:    "-10",
			expected: defaultTime,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionTimeout = tt.input
			if v := GetSessionTimeout(); v != tt.expected {
				t.Errorf("GetSessionTimeout() = %v, want %v", v, tt.expected)
			}
		})
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

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Default",
			input:    "",
			expected: defaultAppPort,
		},
		{
			name:     "Value",
			input:    "8080",
			expected: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appPort = tt.input
			if v := GetAppPort(); v != tt.expected {
				t.Errorf("GetAppPort() = %v, want %v", v, tt.expected)
			}
		})
	}
}

func TestGetRPDisplayName(t *testing.T) {
	curRPDisplayName := rpDisplayName
	defer func() {
		rpDisplayName = curRPDisplayName
	}()

	// Test case
	rpDisplayName = "RPDisplayName"
	if v := GetRPDisplayName(); v != "RPDisplayName" {
		t.Errorf("GetRPDisplayName() = %v, want %v", v, "RPDisplayName")
	}
}

func TestGetRPID(t *testing.T) {
	curRPID := rpID
	defer func() {
		rpID = curRPID
	}()

	// Test case
	rpID = "RPID"
	if v := GetRPID(); v != "RPID" {
		t.Errorf("GetRPID() = %v, want %v", v, "RPID")
	}
}

func TestGetRPOrigins(t *testing.T) {
	curRPOrigins := rpOrigins
	defer func() {
		rpOrigins = curRPOrigins
	}()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Default",
			input:    "",
			expected: []string{},
		},
		{
			name:     "Single value",
			input:    "http://localhost:8080",
			expected: []string{"http://localhost:8080"},
		},
		{
			name:     "Multiple values",
			input:    "http://localhost:8080,http://localhost:8081",
			expected: []string{"http://localhost:8080", "http://localhost:8081"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rpOrigins = tt.input
			if v := GetRPOrigins(); !reflect.DeepEqual(v, tt.expected) {
				t.Errorf("GetRPOrigins() = %v, want %v", v, tt.expected)
			}
		})
	}
}
