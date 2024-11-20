package app

import (
	"github.com/gin-gonic/gin"
)

type App struct {
	engine *gin.Engine
}

func New() *App {
	app := gin.Default()
	return &App{
		engine: app,
	}
}

func (a *App) GetHandler(pattern string, f gin.HandlerFunc) {
	a.engine.GET(pattern, f)
}

func (a *App) PostHandler(pattern string, f gin.HandlerFunc) {
	a.engine.POST(pattern, f)
}

func (a *App) DeleteHandler(pattern string, f gin.HandlerFunc) {
	a.engine.DELETE(pattern, f)
}

func (a *App) Run() error {
	return a.engine.Run(":7001")
}
