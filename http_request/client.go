package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type httpClient struct {
	*http.Client
}

func newHttpClient() *httpClient {
	return &httpClient{
		&http.Client{Timeout: 1 * time.Minute},
	}
}

func Do(ctx context.Context, method, endpoint, userAgent string, headers map[string]string, params url.Values, response interface{}) error {
	var body io.Reader
	switch method {
	case http.MethodPost:
		body = bytes.NewBufferString(params.Encode())
	case http.MethodGet:
		if params != nil {
			u, _ := url.Parse(endpoint)
			u.RawQuery = params.Encode()
			endpoint = u.String()
		}
	default:
		return fmt.Errorf(http.StatusText(http.StatusBadRequest))
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := newHttpClient().Do(req)
	if err != nil {
		return err
	}

	var respBody []byte
	respBody, err = io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusBadRequest:
			var response struct {
				Error            string `json:"error"`
				ErrorDescription string `json:"error_description"`
			}
			e := json.Unmarshal(respBody, &response)
			if e == nil && response.ErrorDescription == "Token expired or revoked" {
				return fmt.Errorf("Error Token Revoked.")
			}
			return fmt.Errorf(http.StatusText(http.StatusBadRequest))
		default:
			return fmt.Errorf(http.StatusText(resp.StatusCode))
		}
	}
	if response != nil {
		err := json.Unmarshal(respBody, &response)
		if err != nil {
			return err
		}
	}
	return nil
}
