//+build wireinject

//The build tag makes sure the stub is not built in the final build.

package http

import (
	"ratelimit/internal/config"
	"ratelimit/internal/delivery/http"
	"ratelimit/internal/wireset"

	"github.com/google/wire"
)

func Initialize(configPath string) (Application, error) {
	wire.Build(
		newApplication,
		config.NewConfig,
		wireset.InitLogger,
		wireset.InitRateLimiter,
		http.NewHttpServer,
	)
	return Application{}, nil
}
