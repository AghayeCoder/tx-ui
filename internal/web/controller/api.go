package controller

import (
	"net/http"
	"x-ui/internal/web/service"
	"x-ui/internal/web/session"

	"github.com/gin-gonic/gin"
)

type APIController struct {
	BaseController
	inboundController *InboundController
	serverController  *ServerController
	Tgbot             service.Tgbot
}

type I18nRequest struct {
	Keys []string `json:"keys"`
}

func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	return a
}

func (a *APIController) checkAPIAuth(c *gin.Context) {
	if !session.IsLogin(c) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.Next()
}

func (a *APIController) initRouter(g *gin.RouterGroup) {
	// Main API group
	api := g.Group("/panel/api")
	api.Use(a.checkAPIAuth)

	// Inbounds API
	inbounds := api.Group("/inbounds")
	a.inboundController = NewInboundController(inbounds)

	// Server API
	server := api.Group("/server")
	a.serverController = NewServerController(server)

	// Extra routes
	api.GET("/backuptotgbot", a.BackuptoTgbot)
	api.POST("/i18n", a.I18n)
}

func (a *APIController) BackuptoTgbot(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
	jsonMsg(c, "telegram backup requested", nil)
}

func (a *APIController) I18n(c *gin.Context) {
	var req I18nRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Keys = []string{}
	}

	out := map[string]string{}
	seen := map[string]struct{}{}
	for _, key := range req.Keys {
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out[key] = I18nWeb(c, key)
	}

	jsonObj(c, out, nil)
}
