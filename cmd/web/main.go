package main

import (
	"errors"

	"github.com/asim/go-micro/v3"

	"github.com/Mikhalevich/filesharing-web-service/internal/handler"
	"github.com/Mikhalevich/filesharing-web-service/internal/router"
	"github.com/Mikhalevich/filesharing-web-service/internal/wrapper"
	"github.com/Mikhalevich/filesharing/pkg/service"
)

type config struct {
	service.Config           `yaml:"service"`
	GatewayHost              string `yaml:"gateway_host"`
	SessionExpirePeriodInSec int    `yaml:"session_expire_period"`
}

func (c *config) Service() service.Config {
	return c.Config
}

func (c *config) Validate() error {
	if c.GatewayHost == "" {
		return errors.New("gateway_host is required")
	}

	if c.SessionExpirePeriodInSec <= 0 {
		return errors.New("invalid session_expire_period")
	}

	return nil
}

func main() {
	var cfg config
	service.Run("web", &cfg, func(srv micro.Service, s service.Servicer) error {
		cookieSession := wrapper.NewCookieSession(int64(cfg.SessionExpirePeriodInSec))
		h := handler.New(cfg.GatewayHost, cookieSession, s.Logger())

		router.MakeRoutes(s.Router(), true, h, s.Logger())
		return nil
	})
}
