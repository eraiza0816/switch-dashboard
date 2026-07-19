package rtlplayground

func (c *Client) DownloadConfig() (string, error) {
	data, err := c.get("/config")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
