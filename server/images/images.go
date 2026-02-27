package images

import (
	"bytes"
	"context"
	"drpp/server/logger"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // TODO: https://github.com/golang/go/issues/40173
)

const (
	maxUploadRespBodyBytes   = 1 * 1024 * 1024  // 1 MB
	maxDownloadRespBodyBytes = 16 * 1024 * 1024 // 16 MB
)

var httpClient = http.Client{}

func Upload(ctx context.Context, pngBytes []byte, postEndpoint string, imageFieldName string, formFields map[string]string, headers map[string]string) ([]byte, error) {
	logger.Info("Uploading image")
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	fileField, err := writer.CreateFormFile(imageFieldName, "image.png")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	_, err = fileField.Write(pngBytes)
	if err != nil {
		return nil, fmt.Errorf("write form file: %w", err)
	}
	for key, value := range formFields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("write form field %q: %w", key, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close form writer: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, postEndpoint, &body)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxUploadRespBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	bodyText := string(bodyBytes)
	logger.Debug("POST %s, HTTP %d, %s, %s", postEndpoint, resp.StatusCode, resp.Header, bodyText)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("HTTP status %s", resp.Status)
	}
	return bodyBytes, nil
}

func GetPngBytes(ctx context.Context, sourceUrl string, fitInSquare bool, maxSize int, headers map[string]string) ([]byte, error) {
	logger.Debug("Fetching image")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("fetch image: HTTP status %s", resp.Status)
	}
	img, _, err := image.Decode(io.LimitReader(resp.Body, maxDownloadRespBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	longLength := max(width, height)
	if fitInSquare && width != height {
		newImg := image.NewRGBA(image.Rect(0, 0, longLength, longLength))
		x := (longLength - width) / 2
		y := (longLength - height) / 2
		draw.Draw(newImg, image.Rect(x, y, x+width, y+height), img, bounds.Min, draw.Over)
		img = newImg
		width, height = longLength, longLength
	}
	if longLength > maxSize {
		ratio := float64(maxSize) / float64(longLength)
		width = int(float64(width) * ratio)
		height = int(float64(height) * ratio)
		newImg := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(newImg, newImg.Bounds(), img, img.Bounds(), draw.Over, nil)
		img = newImg
	}
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, img); err != nil {
		return nil, fmt.Errorf("encode to PNG: %w", err)
	}
	return buffer.Bytes(), nil
}
