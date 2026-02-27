package config

import (
	"context"
	"drpp/server/api"
	"net/http"
)

type Controller struct {
	service *Service
}

func NewController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(server *api.Server) {
	api.RegisterRoute(server, c.Get, http.MethodGet, "config", http.StatusOK)
	api.RegisterRoute(server, c.Put, http.MethodPut, "config", http.StatusOK)
	api.RegisterRoute(server, c.GetAutostart, http.MethodGet, "config/autostart", http.StatusOK)
	api.RegisterRoute(server, c.SetAutostart, http.MethodPut, "config/autostart", http.StatusOK)
}

func (c *Controller) Get(ctx context.Context, input any) (Config, error) {
	return c.service.Config(), nil
}

func (c *Controller) Put(ctx context.Context, input Config) (Config, error) {
	if err := c.service.SetConfig(input); err != nil {
		return Config{}, err
	}
	return input, nil
}

func (c *Controller) GetAutostart(ctx context.Context, input any) (bool, error) {
	return isAutostartEnabled()
}

func (c *Controller) SetAutostart(ctx context.Context, input bool) (bool, error) {
	if err := setAutostartEnabled(input); err != nil {
		return false, err
	}
	return input, nil
}
