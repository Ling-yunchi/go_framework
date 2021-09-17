package main

import (
	"net/http"
	"wego"
)

func main() {
	r := wego.Default()

	r.GET("/", func(c *wego.Context) {
		c.HTML(http.StatusOK, "<h1>Hello WeGo!</h1>")
	})
	r.GET("/panic", func(c *wego.Context) {
		panic("test panic recovery")
		c.String(200, "you shouldn't see this sentence")
	})

	r.Run(":8080")
}
