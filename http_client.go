package predictclob

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

// CreateReqClientWithProxy creates a req.Client with the given proxy transport
func CreateReqClientWithProxy(transport *http.Transport, userAgent string, timeout time.Duration) *req.Client {
	browserUserAgent := "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

	if userAgent != "" {
		browserUserAgent = userAgent
	}

	reqClient := req.C().
		SetTLSFingerprintChrome().
		DisableHTTP3().
		SetTimeout(timeout).
		SetCommonRetryCount(0).
		SetUserAgent(browserUserAgent)

	if transport != nil {
		reqTransport := reqClient.GetTransport()

		if transport.Proxy != nil {
			reqTransport.Proxy = transport.Proxy
		}

		if transport.DialContext != nil {
			reqTransport.DialContext = transport.DialContext
		}

		if transport.TLSHandshakeTimeout > 0 {
			reqTransport.TLSHandshakeTimeout = transport.TLSHandshakeTimeout
		}
		if transport.IdleConnTimeout > 0 {
			reqTransport.IdleConnTimeout = transport.IdleConnTimeout
		}
		if transport.MaxIdleConns > 0 {
			reqTransport.MaxIdleConns = transport.MaxIdleConns
		}
		if transport.ExpectContinueTimeout > 0 {
			reqTransport.ExpectContinueTimeout = transport.ExpectContinueTimeout
		}
	}

	return reqClient
}

// validateRequestHeaders validates the request by checking headers and IP
func validateRequestHeaders(reqClient *req.Client, prefix string) {
	log.Printf("%s ========== IP Validation ==========", prefix)

	ipResp, err := reqClient.R().
		SetHeader("User-Agent", "curl/7.68.0").
		SetHeader("Accept", "*/*").
		Get("https://ifconfig.me")

	if err != nil {
		log.Printf("%s Warning: Failed to check IP: %v", prefix, err)
	} else if ipResp.StatusCode == http.StatusOK {
		externalIP := strings.TrimSpace(ipResp.String())
		log.Printf("%s   External IP: %s", prefix, externalIP)
	} else {
		log.Printf("%s Warning: Unexpected status code from ifconfig.me: %d", prefix, ipResp.StatusCode)
	}
}

// DoReqClientRequest is a unified request entry point for req.Client
func DoReqClientRequest(reqClient *req.Client, method, url string, body []byte, customHeaders map[string]string, skipValidation bool) (*req.Response, error) {
	if !skipValidation {
		validateRequestHeaders(reqClient, "[DoReqClientRequest]")
	}

	reqBuilder := reqClient.R().
		SetHeader("Content-Type", "application/json").
		SetBody(body)

	for key, value := range customHeaders {
		reqBuilder.SetHeader(key, value)
	}

	var resp *req.Response
	var err error

	switch method {
	case http.MethodGet:
		resp, err = reqBuilder.Get(url)
	case http.MethodPost:
		resp, err = reqBuilder.Post(url)
	case http.MethodPut:
		resp, err = reqBuilder.Put(url)
	case http.MethodDelete:
		resp, err = reqBuilder.Delete(url)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// DoReqClientRequestFromHTTPRequest converts http.Request and uses unified entry point
func DoReqClientRequestFromHTTPRequest(reqClient *req.Client, request *http.Request, skipValidation bool) (*req.Response, error) {
	if !skipValidation {
		validateRequestHeaders(reqClient, "[DoReqClientRequestFromHTTPRequest]")
	}

	reqBuilder := reqClient.R()

	for key, values := range request.Header {
		for _, value := range values {
			reqBuilder.SetHeader(key, value)
		}
	}

	if request.Body != nil {
		bodyBytes, err := io.ReadAll(request.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %w", err)
		}
		request.Body.Close()
		reqBuilder.SetBody(bodyBytes)
	}

	var resp *req.Response
	var err error

	switch request.Method {
	case http.MethodGet:
		resp, err = reqBuilder.Get(request.URL.String())
	case http.MethodPost:
		resp, err = reqBuilder.Post(request.URL.String())
	case http.MethodPut:
		resp, err = reqBuilder.Put(request.URL.String())
	case http.MethodDelete:
		resp, err = reqBuilder.Delete(request.URL.String())
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", request.Method)
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}
