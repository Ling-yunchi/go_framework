package main

import (
	"log"
	"net/http"
	"time"
	"wego"
)

func onlyForV2() wego.HandlerFunc {
	return func(c *wego.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Next()
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func main() {
	r := wego.New()

	r.Use(wego.Logger())
	r.GET("/", func(c *wego.Context) {
		c.HTML(http.StatusOK, "<h1>Hello WeGo!</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2())
	{
		v2.GET("/hello/:name", func(c *wego.Context) {
			c.String(http.StatusOK, "Hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.Run(":8080")
}
