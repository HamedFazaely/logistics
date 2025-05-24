package logging

import (
	"sync"

	"gitlab.com/Hamed1984/logistics/pkg/conf"
	"go.uber.org/zap"
)

var logger *zap.Logger
var once sync.Once

func GetLogger(c *conf.Configuration) *zap.Logger {
	once.Do(func() {

		logger = zap.Must(zap.NewProduction())
		if c.Development {
			logger = zap.Must(zap.NewDevelopment())
		}
	})
	return logger
}
