package rtlplayground

import (
	"fmt"
	"net/http"
	"strings"
)

// ExecuteCommand sends a CLI command to the switch.  The command text is
// validated locally first because the firmware has minimal input validation.
// When a pre-shared key is configured (SetPSK), the command is sent encrypted
// through the /enc endpoint instead of the plaintext /cmd endpoint.
//
// The firmware expects the raw command text as the POST /cmd body (see
// httpd.c execute_commands and html/main.js fetchAPI) — not a form-encoded
// "cmd=..." value, which would be tokenized as a bogus command.
func (c *Client) ExecuteCommand(cmd string) error {
	if err := ValidateCommand(cmd); err != nil {
		return fmt.Errorf("validation: %w", err)
	}
	if c.HasPSK() {
		_, err := c.PostEnc(cmd)
		return err
	}
	_, err := c.post("/cmd", func() (*http.Request, error) {
		req, err := http.NewRequest("POST", c.baseURL+"/cmd", strings.NewReader(cmd))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "text/plain")
		return req, nil
	})
	return err
}

// Reboot resets the switch.  The firmware has no /reset HTTP endpoint (the
// WebUI's GET /reset is not implemented in httpd.c); rebooting is the
// "reset" CLI command (cmd_parser.c), which ExecuteCommand sends via /cmd or
// /enc.
func (c *Client) Reboot() error {
	return c.ExecuteCommand("reset")
}
