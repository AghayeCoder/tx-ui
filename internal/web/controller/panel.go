package controller

import "github.com/gin-gonic/gin"

type PanelController struct {
	BaseController

	settingController     *SettingController
	xraySettingController *XraySettingController
}

func NewPanelController(g *gin.RouterGroup) *PanelController {
	a := &PanelController{}
	a.initRouter(g)
	return a
}

func (a *PanelController) initRouter(g *gin.RouterGroup) {
	g = g.Group("/panel")
	g.Use(a.checkLogin)

	g.GET("/", a.index)
	g.GET("/inbounds", a.inbounds)
	g.GET("/settings", a.settings)
	g.GET("/xray", a.xraySettings)

	a.settingController = NewSettingController(g)
	a.xraySettingController = NewXraySettingController(g)
}

func (a *PanelController) index(c *gin.Context) {
	html(c, "panel.html", "pages.index.title", nil)
}

func (a *PanelController) inbounds(c *gin.Context) {
	html(c, "panel.html", "pages.inbounds.title", nil)
}

func (a *PanelController) settings(c *gin.Context) {
	html(c, "panel.html", "pages.settings.title", nil)
}

func (a *PanelController) xraySettings(c *gin.Context) {
	html(c, "panel.html", "pages.xray.title", nil)
}
