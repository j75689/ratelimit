package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	ratelimiterErrs "ratelimit/pkg/ratelimit/errors"
)

func (httpServer *HttpServer) HandleRatelimitAPI(c *gin.Context) {
	ip := c.ClientIP()
	count := c.GetInt64("count")
	c.String(http.StatusOK, "IP: %s, Request: %d", ip, count)
}

func (httpServer *HttpServer) HandleRatelimitMiddleware(c *gin.Context) {
	ip := c.ClientIP()
	token, err := httpServer.ratelimter.Acquire(ip)
	if err != nil {
		if errors.Is(err, ratelimiterErrs.ErrNotEnoughToken) {
			c.String(http.StatusTooManyRequests, "Error: Too Many Request, IP: %s", ip)
		} else {
			c.String(http.StatusInternalServerError, "Error: %v", err)
		}
		c.Abort()
		return
	}
	c.Set("count", token.Number())
	c.Next()
}
