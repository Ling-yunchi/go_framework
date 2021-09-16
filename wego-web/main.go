package main

import (
	"net/http"
	"wego"
)

func main() {
	r := wego.New()

	r.GET("/", func(c *wego.Context) {
		c.HTML(http.StatusOK, "<h1> Hello WeGo! </h1>")
	})
	r.GET("/hello", func(c *wego.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})
	r.POST("/login", func(c *wego.Context) {
		c.JSON(http.StatusOK, wego.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	r.Run(":8080")
}
