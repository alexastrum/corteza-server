package rbac

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var _ = tls.Config{}

type (
	Client struct {
		Transport *http.Transport
		Client    *http.Client

		debugLevel string
		config     configuration
	}

	ClientInterface interface {
		Users() *Users
		Roles() *Roles
		Resources() *Resources
		Sessions() *Sessions
	}
)

func (c *Client) Users() *Users         { return &Users{c} }
func (c *Client) Roles() *Roles         { return &Roles{c} }
func (c *Client) Resources() *Resources { return &Resources{c} }
func (c *Client) Sessions() *Sessions   { return &Sessions{c} }

var _ ClientInterface = &Client{}

func New() (*Client, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	timeout := time.Duration(config.timeout) * time.Second

	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: timeout,
		}).Dial,
		TLSHandshakeTimeout: timeout,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{
		Transport: transport,
		Client:    client,
		config:    config,
	}, nil
}

func (c *Client) Debug(debugLevel string) *Client {
	c.debugLevel = debugLevel
	return c
}

func (c *Client) Get(url string) (*http.Response, error) {
	return c.Request("GET", url, nil)
}

func (c *Client) Post(url string, body interface{}) (*http.Response, error) {
	return c.Request("POST", url, body)
}

func (c *Client) Patch(url string, body interface{}) (*http.Response, error) {
	return c.Request("PATCH", url, body)
}

func (c *Client) Delete(url string) (*http.Response, error) {
	return c.Request("DELETE", url, nil)
}

func (c *Client) Request(method, url string, body interface{}) (*http.Response, error) {
	link := strings.TrimRight(c.config.baseURL, "/") + "/" + strings.TrimLeft(url, "/")

	if c.debugLevel == "info" {
		fmt.Println("RBAC >>>", method, link)
	}

	request := func() (*http.Request, error) {
		if body != nil {
			b, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			return http.NewRequest(method, link, bytes.NewBuffer(b))
		}
		return http.NewRequest(method, link, nil)
	}

	req, err := request()
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.config.auth)))
	// req.Header.Add("X-TENANT-ID", c.config.tenant)
	req.Header["X-TENANT-ID"] = []string{c.config.tenant}

	if c.debugLevel == "debug" {
		fmt.Println("RBAC >>> (request)")
		b, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			fmt.Println("RBAC >>> Error:", err)
		} else {
			if b != nil {
				fmt.Println(strings.TrimSpace(string(b)))
			}
		}
		fmt.Println("---")
	}

	resp, err := c.Client.Do(req)
	if c.debugLevel == "debug" {
		fmt.Println("RBAC <<< (response)")
		if err != nil {
			fmt.Println("RBAC <<< Error:", err)
		} else {

			b, err := httputil.DumpResponse(resp, true)
			if err != nil {
				fmt.Println("RBAC <<< Error:", err)
			} else {
				if b != nil {
					fmt.Println(string(b))
				}
			}
		}
		fmt.Println("-----------------")
	}
	if err != nil {
		if c.debugLevel == "info" {
			fmt.Println("RBAC <<< Response error", err)
		}
		return nil, err
	}
	if c.debugLevel == "info" {
		fmt.Println("RBAC <<< Response", resp.StatusCode)
	}
	return resp, nil
}