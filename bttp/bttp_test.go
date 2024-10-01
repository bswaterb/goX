package bttp

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBttpServer(t *testing.T) {
	e := NewEngine()

	e.GET("/", func(c *Context) {
		c.String(http.StatusOK, "Hello Test")
	})

	e.POST("/path1", func(c *Context) {
		fmt.Println("req info: ", c.Method, c.Path, c.Req.RemoteAddr)
		c.JSON(http.StatusOK, map[string]string{
			"field1": "value1",
			"field2": "value2",
		})
	})

	go func() {
		time.Sleep(2 * time.Second)
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080", nil)
		assert.NoError(t, err)
		resp, err := client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK)

		req, err = http.NewRequest("POST", "http://127.0.0.1:8080/path1", nil)
		assert.NoError(t, err)
		resp, err = client.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusOK)

	}()
	e.Run(":8080")

}
