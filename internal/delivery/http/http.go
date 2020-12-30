package http

import (
	"ratelimit/pkg/ratelimit"
	"sync"

	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	*gin.Engine
	ratelimter ratelimit.Ratelimiter
	ipMap      sync.Map
}

func (server *HttpServer) setRouter() {
	server.Use(server.HandleRatelimitMiddleware)
	server.Any("/ratelimit", server.HandleRatelimitAPI)
}

func NewHttpServer(ratelimiter ratelimit.Ratelimiter) *HttpServer {
	httpServer := &HttpServer{
		Engine:     gin.Default(),
		ratelimter: ratelimiter,
		ipMap:      sync.Map{},
	}
	httpServer.setRouter()

	return httpServer
}
