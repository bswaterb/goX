package main

import (
	"fmt"
	"net/http"

	"github.com/bswaterb/goX/bttp"
)

func main() {
	RunBttp()
}

func RunBttp() {
	e := bttp.NewEngine()

	e.GET("/", func(c *bttp.Context) {
		c.String(http.StatusOK, "Hello Test")
	})

	e.POST("/path1", func(c *bttp.Context) {
		fmt.Println("req info: ", c.Method, c.Path, c.Req.RemoteAddr)
		c.JSON(http.StatusOK, map[string]string{
			"field1": "value1",
			"field2": "value2",
		})
	})

	e.GET("/path2/:name/search", func(c *bttp.Context) {
		name := c.Params["name"]
		c.JSON(http.StatusOK, map[string]string{
			"name": name,
		})
	})

	e.GET("/path2/a/b/*filepath", func(c *bttp.Context) {
		path := c.Params["filepath"]
		c.JSON(http.StatusOK, map[string]string{
			"filepath": path,
		})
	})

	e.GET("/hello", func(c *bttp.Context) {
		c.String(http.StatusOK, "query params 'name' value is : %s\n", c.Query("name"), c.Path)
	})

	e.Run(":8080")
}
