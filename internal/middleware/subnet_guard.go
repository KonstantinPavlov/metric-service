package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/labstack/echo/v4"
)

type SubnetGuard struct {
	trustedNet *net.IPNet
}

func NewSubnetGuard(subNetMask string) (*SubnetGuard, error) {
	if subNetMask == "" {
		return &SubnetGuard{trustedNet: nil}, nil
	}
	_, trustedNet, err := net.ParseCIDR(subNetMask)
	if err != nil {
		return nil, fmt.Errorf("invalid trusted subnet CIDR %q: %w", subNetMask, err)
	}

	return &SubnetGuard{
		trustedNet: trustedNet,
	}, nil
}

func (g *SubnetGuard) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if g.trustedNet == nil {
				return next(c)
			}

			clientIPStr := c.Request().Header.Get("X-Real-IP")
			clientIP := net.ParseIP(clientIPStr)

			if clientIP == nil || !g.trustedNet.Contains(clientIP) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}
