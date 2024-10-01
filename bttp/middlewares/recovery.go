package middlewares

import (
	"fmt"
	"github.com/bswaterb/goX/bttp"
	"log"
	"runtime"
	"strings"
)

func NewRecoveryMiddleware() bttp.Handler {
	return func(c *bttp.Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.InternalError("Internal Server Error")
			}
		}()
		c.Next()
	}
}

func trace(msg string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // skip first 3 caller

	var str strings.Builder
	str.WriteString(msg + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}
