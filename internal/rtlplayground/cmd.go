package rtlplayground

import "net/url"

func (c *Client) ExecuteCommand(cmd string) error {
	_, err := c.postForm("/cmd", url.Values{"cmd": {cmd}})
	return err
}

func (c *Client) Reboot() error {
	_, err := c.get("/reset")
	return err
}
