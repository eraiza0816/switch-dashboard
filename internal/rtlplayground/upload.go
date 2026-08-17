package rtlplayground

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

func (c *Client) UploadFirmware(firmwareData []byte) error {
	_, err := c.post("/upload", func() (*http.Request, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("uploadedfile", "firmware.bin")
		if err != nil {
			return nil, fmt.Errorf("create form file: %w", err)
		}
		if _, err := io.Copy(fw, bytes.NewReader(firmwareData)); err != nil {
			return nil, fmt.Errorf("copy firmware: %w", err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("close multipart: %w", err)
		}
		req, err := http.NewRequest("POST", c.baseURL+"/upload", &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		return req, nil
	})
	return err
}

func (c *Client) UploadConfig(configText string) error {
	_, err := c.post("/config", func() (*http.Request, error) {
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)
		fw, err := w.CreateFormFile("uploadedfile", "config.txt")
		if err != nil {
			return nil, fmt.Errorf("create form file: %w", err)
		}
		if _, err := io.WriteString(fw, configText); err != nil {
			return nil, fmt.Errorf("write config: %w", err)
		}
		if err := w.Close(); err != nil {
			return nil, fmt.Errorf("close multipart: %w", err)
		}
		req, err := http.NewRequest("POST", c.baseURL+"/config", &buf)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", w.FormDataContentType())
		return req, nil
	})
	return err
}
