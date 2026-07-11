package logger

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ZapMiddleware(log *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			req := c.Request()
			res := c.Response()
			fields := []zap.Field{
				zap.String("uri", req.RequestURI),
				zap.String("method", req.Method),
				zap.Duration("latency", time.Since(start)),
				zap.Int("status_code", res.Status),
				zap.Int64("bytes_out", res.Size),
			}
			if err != nil {
				fields = append(fields, zap.Error(err))
				log.Error("HTTP request", fields...)
			} else {
				log.Info("HTTP request", fields...)
			}
			return nil
		}
	}
}
