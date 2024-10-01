package main

import (
	"fmt"
	"github.com/bswaterb/goX/bttp/middlewares"
	"net/http"

	"github.com/bswaterb/goX/bttp"
)

func main() {
	RunBttpGroup()
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

func RunBttpGroup() {
	e := bttp.NewEngine()
	g1 := e.Group("/v1")
	g1.Use(middlewares.NewLoggerMiddleware(), middlewares.NewRecoveryMiddleware())
	{
		g1.GET("/", func(c *bttp.Context) {
			c.String(http.StatusOK, "Hello Test")
		})
		g1.POST("/path1", func(c *bttp.Context) {
			fmt.Println("req info: ", c.Method, c.Path, c.Req.RemoteAddr)
			c.JSON(http.StatusOK, map[string]string{
				"field1": "value1",
				"field2": "value2",
			})
		})
	}

	g2 := e.Group("/v2")
	{
		g2.GET("/", func(c *bttp.Context) {
			c.String(http.StatusOK, "Hello Test")
		})
		g2.GET("/path2/:name/search", func(c *bttp.Context) {
			name := c.Params["name"]
			c.JSON(http.StatusOK, map[string]string{
				"name": name,
			})
		})
		g2.GET("/path2/a/b/*filepath", func(c *bttp.Context) {
			path := c.Params["filepath"]
			c.JSON(http.StatusOK, map[string]string{
				"filepath": path,
			})
		})
	}

	g3 := e.Group("/auth")
	{
		g3.POST("/login", func(c *bttp.Context) {
			c.JSON(http.StatusOK, bttp.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
		g3.POST("/logout", func(c *bttp.Context) {
			fmt.Printf("user %s logout", c.PostForm("username"))
			c.JSON(http.StatusOK, bttp.H{
				"code": 200,
				"msg":  "success",
			})
		})
	}

	e.Run(":8080")
}
