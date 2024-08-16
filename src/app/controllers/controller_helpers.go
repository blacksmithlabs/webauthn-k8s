package controllers

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"blacksmithlabs.dev/webauthn-k8s/app/config"
	"blacksmithlabs.dev/webauthn-k8s/app/utils"
)

var sessionTimeout = config.GetSessionTimeout()

var logger = utils.GetLogger()

func getSession(c *gin.Context) sessions.Session {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		Path:   "/",
		MaxAge: int(sessionTimeout.Seconds()),
	})
	return session
}
