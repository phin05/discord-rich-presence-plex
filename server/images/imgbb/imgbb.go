package imgbb

import (
	"context"
	"drpp/server/images"
	"encoding/json/v2"
	"fmt"
	"strconv"
	"time"
)

// https://api.imgbb.com/

type Service struct {
	apiKey        string
	expirySeconds int
}

func NewService(apiKey string, expiryMinutes int) *Service {
	return &Service{
		apiKey:        apiKey,
		expirySeconds: expiryMinutes * 60,
	}
}

type response struct {
	Success bool `json:"success"`
	Status  int  `json:"status"`
	Data    struct {
		Url string `json:"url"`
	} `json:"data"`
}

func (s *Service) Upload(ctx context.Context, pngBytes []byte) (string, error) {
	formFields := map[string]string{
		"key": s.apiKey,
	}
	if s.expirySeconds > 0 {
		formFields["expiration"] = strconv.Itoa(s.expirySeconds)
	}
	bodyBytes, err := images.Upload(
		ctx,
		pngBytes,
		"https://api.imgbb.com/1/upload",
		"image",
		formFields,
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}
	var response response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	if !response.Success {
		return "", fmt.Errorf("status %d: %s", response.Status, string(bodyBytes))
	}
	return response.Data.Url, nil
}

func (s *Service) Lifespan() time.Duration {
	return time.Duration(s.expirySeconds) * time.Second
}
