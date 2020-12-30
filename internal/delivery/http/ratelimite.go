package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		httpServer.ipMap.Store(ip, 0)
		c.String(http.StatusTooManyRequests, "Error: Too Many Request, IP: %s", ip)
		c.Abort()
		return
	}
	c.Set("count", token.Number())
	c.Next()
}
