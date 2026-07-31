package rtlplayground

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
)

func (c *Client) UploadFirmware(firmwareData []byte) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("uploadedfile", "firmware.bin")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(fw, bytes.NewReader(firmwareData)); err != nil {
		return fmt.Errorf("copy firmware: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart: %w", err)
	}
	_, err = c.post("/upload", w.FormDataContentType(), &buf)
	return err
}

func (c *Client) UploadConfig(configText string) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("uploadedfile", "config.txt")
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.WriteString(fw, configText); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("close multipart: %w", err)
	}
	_, err = c.post("/config", w.FormDataContentType(), &buf)
	return err
}
