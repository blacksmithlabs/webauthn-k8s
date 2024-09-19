package controllers

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"

	"blacksmithlabs.dev/k8s-webauthn/admin/config"
	"blacksmithlabs.dev/k8s-webauthn/admin/utils"
)

var sessionTimeout = config.GetSessionTimeout()

var logger = utils.GetLogger()

func getSession(c *gin.Context) sessions.Session {
	session := sessions.Default(c)
	session.Options(sessions.Options{
		Path:     "/",
		MaxAge:   int(sessionTimeout.Seconds()),
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
	})
	return session
}
