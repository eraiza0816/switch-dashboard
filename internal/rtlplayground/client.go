package rtlplayground

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func New(ip, password string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("cookiejar: %w", err)
	}
	c := &Client{
		httpClient: &http.Client{Jar: jar},
		baseURL:    fmt.Sprintf("http://%s", ip),
	}
	if err := c.login(password); err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	return c, nil
}

func NewWithClient(ip string, httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("http://%s", ip),
	}
}

func (c *Client) login(password string) error {
	resp, err := c.httpClient.PostForm(c.baseURL+"/login", url.Values{"pwd": {password}})
	if err != nil {
		return fmt.Errorf("post /login: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.httpClient.Get(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", path, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get %s: HTTP %d", path, resp.StatusCode)
	}
	return body, nil
}

func (c *Client) post(path, contentType string, body io.Reader) ([]byte, error) {
	resp, err := c.httpClient.Post(c.baseURL+path, contentType, body)
	if err != nil {
		return nil, fmt.Errorf("post %s: %w", path, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return respBody, nil
}

func (c *Client) postForm(path string, data url.Values) ([]byte, error) {
	return c.post(path, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}
