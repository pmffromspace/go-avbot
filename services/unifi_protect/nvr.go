package unifi_protect

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

func init() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

func (e *Service) Authenticate() error {
	e.connected = false

	r := strings.NewReader(fmt.Sprintf(`{"password": "%s", "username": "%s"}`, e.Password, e.User))

	if err := e.Call(http.MethodPost, "/api/auth/login", r, nil); err != nil {
		return err
	}

	e.connected = true

	return nil
}

func (e *Service) GetSocketEvents() (*WebsocketEvent, error) {
	return NewWebsocketEvent(e)
}

func (e *Service) Call(method string, url string, body io.Reader, out interface{}) error {
	var fullBody []byte
	if body != nil {
		fullBody, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		body = bytes.NewReader(fullBody)
	}

	if err := e.httpDo(method, url, body, out); err != nil {
		if err != ErrUnauthorized {
			return err
		}

		// Reconnect and retry
		if err := e.Authenticate(); err != nil {
			return err
		}
		// Re-create a body reader from the full body
		if body != nil {
			body = bytes.NewReader(fullBody)
		}
		return e.httpDo(method, url, body, out)
	}
	return nil
}

func (e *Service) httpDo(method string, url string, body io.Reader, out interface{}) error {
	request, err := http.NewRequest(method, fmt.Sprintf("https://%s:%d%s", e.Host, e.Port, url), body)
	if err != nil {
		return err
	}

	request.Header.Set("Cookie", e.cookies)
	request.Header.Add("X-CSRF-Token", e.csrfToken)

	if body != nil {
		request.Header.Add("Content-Type", "application/json")
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(request)

	if err != nil {
		return err
	}

	if resp.StatusCode == 401 {
		return ErrUnauthorized
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("invalid return code %d", resp.StatusCode)
	}

	e.csrfToken = resp.Header.Get("X-CSRF-Token")
	e.cookies = resp.Header.Get("Set-Cookie")

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

func (e *Service) CallIO(method string, url string) (io.ReadCloser, int64, error) {
	out, length, err := e.httpDoIO(method, url)
	if err != nil {
		if err != ErrUnauthorized {
			return nil, 0, err
		}

		// Reconnect and retry
		if err := e.Authenticate(); err != nil {
			return nil, 0, err
		}

		return e.httpDoIO(method, url)
	}
	return out, length, nil
}

func (e *Service) httpDoIO(method string, url string) (io.ReadCloser, int64, error) {
	request, err := http.NewRequest(method, fmt.Sprintf("https://%s:%d%s", e.Host, e.Port, url), nil)
	if err != nil {
		return nil, 0, err
	}

	request.Header.Set("Cookie", e.cookies)
	request.Header.Add("X-CSRF-Token", e.csrfToken)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Do(request)

	if err != nil {
		return nil, 0, err
	}

	if resp.StatusCode == 401 {
		return nil, 0, ErrUnauthorized
	}

	if resp.StatusCode/100 != 2 {
		return nil, 0, fmt.Errorf("invalid return code %d", resp.StatusCode)
	}

	e.csrfToken = resp.Header.Get("X-CSRF-Token")
	e.cookies = resp.Header.Get("Set-Cookie")

	return resp.Body, resp.ContentLength, nil
}

func (e *Service) GetBootstrap() (*Bootstrap, error) {
	bootstrap := &Bootstrap{}
	return bootstrap, e.Call(http.MethodGet, "/proxy/protect/api/bootstrap", nil, bootstrap)
}
