package app

import (
	endpoint "rvneural/rss/internal/endpoint/app"
	"rvneural/rss/internal/transport/rest"
)

type App struct {
	app *endpoint.App
}

func New() *App {
	app := endpoint.New()
	api := rest.New()
	app.GetHandler("/feed", api.GetFeed)
	app.PostHandler("/feed", api.AddFeed)
	app.DeleteHandler("/feed/:id", api.DeleteFeed)
	app.GetHandler("/list", api.GetFeedList)
	return &App{
		app: app,
	}
}

func (a *App) Run() error {
	return a.app.Run()
}
