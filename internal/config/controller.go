package config

import (
	"drpp/internal/exc"
	"encoding/json"
	"fmt"
	"net/http"
)

type controller struct {
	service *service
}

func NewController(service *service) *controller {
	return &controller{
		service: service,
	}
}

func (c *controller) Get(w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(c.service.config)
}

func (c *controller) Put(w http.ResponseWriter, r *http.Request) error {
	return c.patch(w, r, new(Config))
}

func (c *controller) Patch(w http.ResponseWriter, r *http.Request) error {
	return c.patch(w, r, c.service.config.deepCopy())
}

func (c *controller) patch(w http.ResponseWriter, r *http.Request, config *Config) error {
	err := json.NewDecoder(r.Body).Decode(config)
	if err != nil {
		return exc.Malformed("Invalid JSON body", err.Error())
	}
	if errs := config.validate(); len(errs) != 0 {
		return exc.Invalid("Invalid fields", errs...)
	}
	c.service.config = config
	if err := c.service.Save(); err != nil {
		return fmt.Errorf("save: %w", err)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
