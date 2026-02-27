package imgur

import (
	"context"
	"drpp/server/images"
	"encoding/json/v2"
	"fmt"
	"time"
)

// https://apidocs.imgur.com/

type Service struct {
	clientId string
}

func NewService(clientId string) *Service {
	return &Service{
		clientId: clientId,
	}
}

type response struct {
	Success bool `json:"success"`
	Status  int  `json:"status"`
	Data    struct {
		Error string `json:"error"`
		Link  string `json:"link"`
	} `json:"data"`
}

func (s *Service) Upload(ctx context.Context, pngBytes []byte) (string, error) {
	bodyBytes, err := images.Upload(
		ctx,
		pngBytes,
		"https://api.imgur.com/3/image",
		"image",
		nil,
		map[string]string{
			"Authorization": "Client-ID " + s.clientId,
		},
	)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}
	var response response
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	if !response.Success {
		return "", fmt.Errorf("status %d: %s", response.Status, response.Data.Error)
	}
	return response.Data.Link, nil
}

func (s *Service) Lifespan() time.Duration {
	return 0
}
