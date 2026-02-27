package litterbox

import (
	"context"
	"drpp/server/images"
	"fmt"
	"net/url"
	"time"
)

// https://litterbox.catbox.moe/tools.php

type Service struct {
	expiryHours int
}

func NewService(expiryHours int) *Service {
	return &Service{
		expiryHours: expiryHours,
	}
}

func (s *Service) Upload(ctx context.Context, pngBytes []byte) (string, error) {
	bodyBytes, err := images.Upload(
		ctx,
		pngBytes,
		"https://litterbox.catbox.moe/resources/internals/api.php",
		"fileToUpload",
		map[string]string{
			"reqtype": "fileupload",
			"time":    fmt.Sprintf("%dh", s.expiryHours),
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("upload: %w", err)
	}
	result := string(bodyBytes)
	if parsed, err := url.Parse(result); err != nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("invalid response: %s", result)
	}
	return result, nil
}

func (s *Service) Lifespan() time.Duration {
	return time.Duration(s.expiryHours) * time.Hour
}
