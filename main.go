package main

import (
	"net/http"

	"gee"
)


func main() {
	//r := gee.New()
	//r.Use( gee.Logger() )
	//r.Use( gee.Recovery() )
	r := gee.Default()
	r.GET("/", func(c *gee.Context) {
		c.String(http.StatusOK, "Hello Geektutu\n")
	})
	// index out of range for testing Recovery()
	r.GET("/panic", func(c *gee.Context) {
		names := []string{"geektutu"}
		c.String(http.StatusOK, names[100])
	})

	r.Run(":9999")
}