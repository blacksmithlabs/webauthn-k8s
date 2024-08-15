package controllers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"blacksmithlabs.dev/webauthn-k8s/app/config"
)

var sessionTimeout = config.GetSessionTimeout()

func getSession(c *gin.Context) sessions.Session {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: int(sessionTimeout.Seconds()),
	})
	return session
}
