package copyparty

import (
	"context"
	"drpp/server/images"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// https://github.com/9001/copyparty/blob/hovudstraum/docs/devnotes.md#http-api

type Service struct {
	url           string
	password      string
	expirySeconds int
}

func NewService(url string, password string, expiryMinutes int) *Service {
	return &Service{
		url:           strings.TrimSuffix(url, "/") + "/",
		password:      password,
		expirySeconds: expiryMinutes * 60,
	}
}

func (s *Service) Upload(ctx context.Context, pngBytes []byte) (string, error) {
	headers := map[string]string{
		"Accept": "url",
		"Rand":   "16",
		"CK":     "no",
	}
	if s.password != "" {
		headers["PW"] = s.password
	}
	if s.expirySeconds > 0 {
		headers["Life"] = strconv.Itoa(s.expirySeconds)
	}
	bodyBytes, err := images.Upload(
		ctx,
		pngBytes,
		s.url,
		"f",
		nil,
		headers,
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
	return time.Duration(s.expirySeconds) * time.Second
}
