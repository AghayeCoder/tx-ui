package controller

import (
	"x-ui/web/service"

	"github.com/gin-gonic/gin"
)

type APIController struct {
	BaseController
	inboundController *InboundController
	serverController  *ServerController
	Tgbot             service.Tgbot
}

func NewAPIController(g *gin.RouterGroup) *APIController {
	a := &APIController{}
	a.initRouter(g)
	a.serverRouter(g)
	return a
}

func (a *APIController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/panel/api/inbounds")
	g.Use(a.checkLogin)

	a.inboundController = NewInboundController(g)

	inboundRoutes := []struct {
		Method  string
		Path    string
		Handler gin.HandlerFunc
	}{
		{"GET", "/createbackup", a.createBackup},
		{"GET", "/list", a.inboundController.getInbounds},
		{"GET", "/get/:id", a.inboundController.getInbound},
		{"GET", "/getClientTraffics/:email", a.inboundController.getClientTraffics},
		{"GET", "/getClientTrafficsById/:id", a.inboundController.getClientTrafficsById},
		{"POST", "/add", a.inboundController.addInbound},
		{"POST", "/del/:id", a.inboundController.delInbound},
		{"POST", "/update/:id", a.inboundController.updateInbound},
		{"POST", "/clientIps/:email", a.inboundController.getClientIps},
		{"POST", "/clearClientIps/:email", a.inboundController.clearClientIps},
		{"POST", "/addClient", a.inboundController.addInboundClient},
		{"POST", "/:id/delClient/:clientId", a.inboundController.delInboundClient},
		{"POST", "/updateClient/:clientId", a.inboundController.updateInboundClient},
		{"POST", "/:id/resetClientTraffic/:email", a.inboundController.resetClientTraffic},
		{"POST", "/resetAllTraffics", a.inboundController.resetAllTraffics},
		{"POST", "/resetAllClientTraffics/:id", a.inboundController.resetAllClientTraffics},
		{"POST", "/delDepletedClients/:id", a.inboundController.delDepletedClients},
		{"POST", "/onlines", a.inboundController.onlines},
		{"POST", "/depleted", a.inboundController.depleted},
		{"POST", "/disabled", a.inboundController.disabled},
	}

	for _, route := range inboundRoutes {
		g.Handle(route.Method, route.Path, route.Handler)
	}
}

func (a *APIController) serverRouter(g *gin.RouterGroup) {
	g = g.Group("/panel/api/server")
	g.Use(a.checkLogin)
	a.serverController = NewServerController(g)
	serverRoutes := []struct {
		Method  string
		Path    string
		Handler gin.HandlerFunc
	}{
		{"GET", "/restartCore", a.serverController.restartXrayService},
		{"GET", "/status", a.serverController.status},
	}

	for _, route := range serverRoutes {
		g.Handle(route.Method, route.Path, route.Handler)
	}
}

func (a *APIController) createBackup(c *gin.Context) {
	a.Tgbot.SendBackupToAdmins()
}
