package controllers

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"blacksmithlabs.dev/webauthn-k8s/auth/cache"
	"blacksmithlabs.dev/webauthn-k8s/auth/database"
)

func HealthCheck(c *gin.Context) {
	var (
		pgstatus    string = "OK"
		cachestatus string = "OK"
	)

	pgconn, err := database.ConnectDb(c)
	if err != nil {
		pgstatus = fmt.Errorf("ERROR connecting to postgres: %v", err).Error()
	} else {
		err := pgconn.Ping(c)
		if err != nil {
			filterre := regexp.MustCompile(`(user|database)=\w+`)
			errmsg := fmt.Errorf("ERROR connecting to postgres: %v", err).Error()
			pgstatus = filterre.ReplaceAllString(errmsg, "$1=*****")
		}
	}

	cacheconn := cache.ConnectCache()
	resp, err := cacheconn.Ping(c).Result()
	if err != nil {
		cachestatus = fmt.Errorf("ERROR connecting to cache: %v", err).Error()
	} else if resp != "PONG" {
		cachestatus = fmt.Errorf("ERROR connecting to cache: unexpected response: %v", resp).Error()
	}

	status := "OK"
	statusCode := http.StatusOK
	if pgstatus != "OK" || cachestatus != "OK" {
		status = "DEGRADED"
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, gin.H{
		"status":   status,
		"postgres": pgstatus,
		"cache":    cachestatus,
	})
}
