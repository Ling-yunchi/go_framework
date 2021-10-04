package main

import (
	"net/http"
	"wego"
)

func main() {
	e := wego.Default()

	e.GET("/hello", func(c *wego.Context) {
		c.HTML(http.StatusOK, "<h1>Hello WeGo!</h1>")
	})
	e.POST("/hello/:name", func(c *wego.Context) {
		c.String(http.StatusOK, "hello %s", c.Params["name"])
	})

	_ = e.Run(":8080")
}
