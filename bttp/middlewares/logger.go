package middlewares

import (
	"fmt"
	"github.com/bswaterb/goX/bttp"
	"time"
)

func NewLoggerMiddleware() bttp.Handler {
	return func(c *bttp.Context) {
		start := time.Now()
		c.Next()
		fmt.Printf("[%s] %s 当前请求执行耗时: %d ms\n", time.Now().Format(time.DateTime), c.Path, time.Since(start).Milliseconds())
	}
}
