package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/Mikhalevich/filesharing-web-service/handler"
	"github.com/Mikhalevich/filesharing-web-service/router"
	"github.com/Mikhalevich/filesharing-web-service/wrapper"
	"github.com/sirupsen/logrus"
)

type params struct {
	Host                     string
	GatewayHost              string
	SessionExpirePeriodInSec int
}

func loadParams() (*params, error) {
	var p params

	p.Host = os.Getenv("FS_HOST")
	if p.Host == "" {
		return nil, errors.New("host name is empty, please specify FS_HOST variable")
	}

	p.GatewayHost = os.Getenv("FS_GATEWAY_HOST")
	if p.GatewayHost == "" {
		return nil, errors.New("Gateway host is empty, please specify FS_GATEWAY_HOST")
	}

	p.SessionExpirePeriodInSec = 60 * 60 * 24
	expirePeriodEnv := os.Getenv("FS_SESSION_EXPIRE_PERIOD_SEC")
	if expirePeriodEnv != "" {
		period, err := strconv.Atoi(expirePeriodEnv)
		if err != nil {
			return nil, fmt.Errorf("unable to convert expire session period to integer value expirePeriod: %s, error: %w", expirePeriodEnv, err)
		}
		p.SessionExpirePeriodInSec = period
	}

	return &p, nil
}

func main() {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	params, err := loadParams()
	if err != nil {
		logger.Errorln(fmt.Errorf("load params error: %w", err))
		return
	}

	cookieSession := wrapper.NewCookieSession(int64(params.SessionExpirePeriodInSec))
	h := handler.NewHandler(params.GatewayHost, cookieSession, logger)
	r := router.NewRouter(true, h, logger)

	logger.Infof("Running params = %v", params)

	err = http.ListenAndServe(params.Host, r.Handler())
	if err != nil {
		logger.Errorln(err)
	}
}
