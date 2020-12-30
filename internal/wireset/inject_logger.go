package wireset

import (
	"ratelimit/internal/config"
	"ratelimit/pkg/logger"

	"github.com/rs/zerolog"
)

func InitLogger(config config.Config) (zerolog.Logger, error) {
	return logger.NewLogger(config.Logger.Level, config.Logger.Format)
}
