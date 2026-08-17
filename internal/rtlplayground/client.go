package rtlplayground

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	psk        []byte

	password string
	reAuthMu sync.Mutex
}

func New(ip, password string) (*Client, error) {
	c, err := newBaseClient(ip)
	if err != nil {
		return nil, err
	}
	if err := c.login(password); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	return c, nil
}

// NewWithPSK creates a client and authenticates with the encrypted PSK
// login challenge.  Use this when the switch runs in PSK mode, where the
// firmware rejects password logins (see doc/authentication.md in
// RTLPlayground).  With an empty pskHex it behaves like New.
func NewWithPSK(ip, password, pskHex string) (*Client, error) {
	c, err := newBaseClient(ip)
	if err != nil {
		return nil, err
	}
	if pskHex != "" {
		if err := c.SetPSK(pskHex); err != nil {
			return nil, err
		}
	}
	if err := c.login(password); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	return c, nil
}

func newBaseClient(ip string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}
	c := &Client{
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 5 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		baseURL: fmt.Sprintf("http://%s", ip),
	}
	return c, nil
}

func NewWithClient(ip string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("http://%s", ip),
	}
}

// login authenticates against the firmware's POST /login endpoint.  With a
// pre-shared key configured the password form is replaced by the encrypted
// login challenge (enc=), since the firmware rejects password logins while a
// PSK is set.  The firmware answers every login attempt with a 302 redirect:
// a valid password is redirected to index.html with a Set-Cookie, an invalid
// one to login.html without a cookie.  Go's default redirect following would
// hide the difference, so the first response is inspected directly.
func (c *Client) login(password string) error {
	c.password = password
	var form string
	if len(c.psk) == aeadKeyLen {
		challenge, err := c.pskLoginChallenge()
		if err != nil {
			return fmt.Errorf("psk login challenge: %w", err)
		}
		form = "enc=" + challenge
	} else {
		form = "pwd=" + url.QueryEscape(password)
	}
	client := c.httpClient
	if client.Jar != nil {
		// Share the cookie jar so the session cookie lands where all
		// subsequent requests look for it.
		client = &http.Client{
			Jar:     client.Jar,
			Timeout: client.Timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}
	req, err := http.NewRequest("POST", c.baseURL+"/login", strings.NewReader(form))
	if err != nil {
		return fmt.Errorf("login: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post /login: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("login: HTTP 401")
	}
	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login: HTTP %d", resp.StatusCode)
	}
	if resp.StatusCode == http.StatusFound && resp.Header.Get("Location") != "index.html" {
		return fmt.Errorf("login: invalid password (redirected to %s)", resp.Header.Get("Location"))
	}
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" && cookie.Value != "" {
			return nil
		}
	}
	return fmt.Errorf("login: no session cookie in response")
}

// pskLoginChallenge encrypts the fixed login challenge with the pre-shared
// key and returns hex(nonce[12] || ct || tag), the format the firmware's
// /login expects in the enc= field.
func (c *Client) pskLoginChallenge() (string, error) {
	nonce := make([]byte, aeadNonceLen)
	if _, err := rand.Read(nonce); err != nil {
		// crypto/rand should never fail; fall back to a counter-based nonce
		for i := range nonce {
			nonce[i] = byte(i + 1)
		}
	}
	enc, err := aeadEncrypt(c.psk, nonce, []byte("RTLP-LOGIN-1"))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(enc), nil
}

// reAuthenticate refreshes the firmware session after a 401. JSON API
// requests never refresh the session on the device (only web-page accesses
// do, see RTLPlayground httpd.c), so a long-lived client must re-login once
// the session expires (SESSION_TIMEOUT = 200s).
func (c *Client) reAuthenticate() error {
	c.reAuthMu.Lock()
	defer c.reAuthMu.Unlock()
	if c.password == "" {
		return fmt.Errorf("no password configured")
	}
	return c.login(c.password)
}

func (c *Client) get(path string) ([]byte, error) {
	body, status, err := c.rawGet(path)
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized {
		if err := c.reAuthenticate(); err != nil {
			return nil, fmt.Errorf("get %s: re-login: %w", path, err)
		}
		if body, status, err = c.rawGet(path); err != nil {
			return nil, err
		}
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("get %s: HTTP %d", path, status)
	}
	return body, nil
}

func (c *Client) rawGet(path string) ([]byte, int, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, 0, fmt.Errorf("get %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read %s: %w", path, err)
	}
	return body, resp.StatusCode, nil
}

// post issues a POST request built by makeReq.  The builder is called again
// when the session has expired (401), so request bodies can be regenerated.
func (c *Client) post(path string, makeReq func() (*http.Request, error)) ([]byte, error) {
	do := func() ([]byte, int, error) {
		req, err := makeReq()
		if err != nil {
			return nil, 0, fmt.Errorf("post %s: build request: %w", path, err)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			// The firmware signals a completed configuration upload by
			// closing the connection (httpd.c: "ugly hack to signal
			// connection finished after config upload"), so an EOF here
			// means the config was accepted, not that the transfer failed.
			if path == "/config" && (errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)) {
				return nil, http.StatusOK, nil
			}
			return nil, 0, fmt.Errorf("post %s: %w", path, err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, resp.StatusCode, fmt.Errorf("read %s: %w", path, err)
		}
		return body, resp.StatusCode, nil
	}
	body, status, err := do()
	if err != nil {
		return nil, err
	}
	if status == http.StatusUnauthorized {
		if err := c.reAuthenticate(); err != nil {
			return nil, fmt.Errorf("post %s: re-login: %w", path, err)
		}
		if body, status, err = do(); err != nil {
			return nil, err
		}
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("post %s: HTTP %d", path, status)
	}
	return body, nil
}

func (c *Client) postForm(path string, data url.Values) ([]byte, error) {
	return c.post(path, func() (*http.Request, error) {
		req, err := http.NewRequest("POST", c.baseURL+path, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
}

// SetPSK configures the pre-shared key used for encrypted commands via the
// /enc endpoint.  The key must be exactly 64 hex chars (32 bytes).
func (c *Client) SetPSK(hexPSK string) error {
	if err := validatePSK(hexPSK); err != nil {
		return err
	}
	key, err := hex.DecodeString(hexPSK)
	if err != nil {
		return fmt.Errorf("pre-shared key must be hex: %w", err)
	}
	c.psk = key
	return nil
}

// HasPSK reports whether a pre-shared key is configured.
func (c *Client) HasPSK() bool {
	return len(c.psk) == aeadKeyLen
}

// PostEnc sends a command encrypted with the pre-shared key to the /enc
// endpoint and returns the decrypted response text.
// Body format: nonce[12] || ciphertext || tag[16].
func (c *Client) PostEnc(cmd string) (string, error) {
	if len(c.psk) != aeadKeyLen {
		return "", fmt.Errorf("pre-shared key not configured (need %d bytes)", aeadKeyLen)
	}
	nonce := make([]byte, aeadNonceLen)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	pkt, err := aeadEncrypt(c.psk, nonce, []byte(cmd))
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Post(c.baseURL+"/enc", "application/octet-stream", bytes.NewReader(pkt))
	if err != nil {
		return "", fmt.Errorf("post /enc: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("enc request failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read /enc: %w", err)
	}
	if len(body) < aeadNonceLen+aeadTagLen {
		return "", fmt.Errorf("short encrypted response (%d bytes)", len(body))
	}
	pt, err := aeadDecrypt(c.psk, body[:aeadNonceLen],
		body[aeadNonceLen:len(body)-aeadTagLen], body[len(body)-aeadTagLen:])
	if err != nil {
		return "", fmt.Errorf("response decrypt failed: %w", err)
	}
	return string(pt), nil
}

// EncAPI fetches a JSON API path through the encrypted /enc endpoint
// ("api <path>") and returns the decrypted response body.
func (c *Client) EncAPI(path string) ([]byte, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	respText, err := c.PostEnc("api " + path)
	if err != nil {
		return nil, err
	}
	return []byte(respText), nil
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}
