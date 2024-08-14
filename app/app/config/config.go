package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultSessionTimeout = 180
const defaultRedisPoolSize = 10
const defaultRedisHost = "localhost:6379"
const defaultAppPort = "8080"

var (
	// Session cache info
	redisPoolSize  = os.Getenv("REDIS_POOL_SIZE")
	redisHost      = os.Getenv("REDIS_HOST")
	redisPassword  = os.Getenv("REDIS_PASSWORD")
	sessionSecret  = os.Getenv("SESSION_SECRET")
	sessionTimeout = os.Getenv("SESSION_TIMEOUT")
	// TODO Postgres info
	// Application Config info
	appPort = os.Getenv("APP_PORT")
	// Webauthn Config info
	rpDisplayName = os.Getenv("RP_DISPLAY_NAME")
	rpID          = os.Getenv("RP_ID")
	rpOrigins     = os.Getenv("RP_ORIGINS")
)

func GetRedisPoolSize() int {
	if redisPoolSize != "" {
		if value, err := strconv.Atoi(redisPoolSize); err != nil {
			fmt.Println("Failed to parse REDIS_POOL_SIZE", err)
		} else if value < 1 {
			fmt.Println("REDIS_POOL_SIZE must be greater than 0")
		} else {
			return int(value)
		}
	}

	return defaultRedisPoolSize
}

func GetRedisHost() string {
	if redisHost == "" {
		return defaultRedisHost
	}

	return redisHost
}

func GetRedisPassword() string {
	return redisPassword
}

func GetSessionSecret() []byte {
	return []byte(sessionSecret)
}

func GetSessionTimeout() time.Duration {
	if sessionTimeout != "" {
		if value, err := strconv.Atoi(sessionTimeout); err != nil {
			fmt.Println("Failed to parse SESSION_TIMEOUT", err)
		} else if value < 1 {
			fmt.Println("SESSION_TIMEOUT must be greater than 0")
		} else {
			return time.Duration(value) * time.Second
		}
	}

	return defaultSessionTimeout * time.Second
}

func GetAppPort() string {
	if appPort == "" {
		return defaultAppPort
	}

	return appPort
}

func GetRPDisplayName() string {
	return rpDisplayName
}

func GetRPID() string {
	return rpID
}

func GetRPOrigins() []string {
	return strings.Split(rpOrigins, ",")
}
