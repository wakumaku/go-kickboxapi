// Package kickboxapi client based on
// https://docs.kickbox.com/v2.0/reference
package kickboxapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// BaseURL of kickbox services endpoint
const BaseURL = "https://api.kickbox.com"

// ServiceType .
type ServiceType int

// Service types supported
const (
	_ ServiceType = iota
	Verify
	VerifyMultiple
	CheckJobStatus
	CreditBalance
	DisposableEmailCheck
)

// Endpoint definition
type Endpoint struct {
	Method string
	Path   string
}

// Endpoints is an endpoint list
var Endpoints = map[ServiceType]Endpoint{
	Verify:               Endpoint{"GET", "/{API_VERSION}/verify"},
	VerifyMultiple:       Endpoint{"PUT", "/{API_VERSION}/verify-batch"},
	CheckJobStatus:       Endpoint{"GET", "/{API_VERSION}/verify-batch/{JOB_ID}"},
	CreditBalance:        Endpoint{"GET", "/{API_VERSION}/balance"},
	DisposableEmailCheck: Endpoint{"GET", "/{API_VERSION}/disposable/{EMAIL_ADDRESS}"},
}

// Errors
var (
	ErrEmptyResponse              = errors.New("Empty body response received from service")
	ErrResponseHadEmptyStructure  = errors.New("Could not unmarshall correctly the response")
	ErrUnknownErrorVerifyingEmail = errors.New("Unknown error verifying email")
)

// VerifyResponse when validating emails
type VerifyResponse struct {
	Result     string  `json:"result,omitempty"`       //:"undeliverable",
	Reason     string  `json:"reason,omitempty"`       //:"rejected_email",
	Role       bool    `json:"role,omitempty"`         //:false,
	Free       bool    `json:"free,omitempty"`         //:false,
	Disposable bool    `json:"disposable,omitempty"`   //:false,
	AcceptAll  bool    `json:"accept_all,omitempty"`   //:false,
	DidYouMean string  `json:"did_you_mean,omitempty"` //:"bill.lumbergh@gmail.com",
	Sendex     float32 `json:"sendex,omitempty"`       //:0.23,
	Email      string  `json:"email,omitempty"`        //:"bill.lumbergh@gamil.com",
	User       string  `json:"user,omitempty"`         //:"bill.lumbergh",
	Domain     string  `json:"domain,omitempty"`       //:"gamil.com",
	Success    bool    `json:"success,omitempty"`      //:true,
	Message    string  `json:"message,omitempty"`      //:null
}

// IsValid .
func (v *VerifyResponse) IsValid() bool {
	return v.Result == "deliverable"
}

func (v *VerifyResponse) Error() error {
	if !v.Success {
		if v.Message != "" {
			return errors.New(v.Message)
		}
		return ErrUnknownErrorVerifyingEmail
	}

	return nil
}

// VerifyMultipleResponse .
type VerifyMultipleResponse struct {
	ID      int    `json:"id"`      //:123,
	Success bool   `json:"success"` //:true,
	Message string `json:"message"` //:null
}

func (v *VerifyMultipleResponse) Error() error {
	if !v.Success {
		if v.Message != "" {
			return errors.New(v.Message)
		}
		return ErrUnknownErrorVerifyingEmail
	}

	return nil
}

// CheckJobStatusResponse .
type CheckJobStatusResponse struct {
	ID          int    `json:"id"`           //: 465,
	Name        string `json:"name"`         //: "Batch API Process - 05-12-2015-01-58-08",
	DownloadURL string `json:"download_url"` //: "https://{{DOWNLOAD_ADDRESS_HERE}},
	CreatedAt   string `json:"created_at"`   //: "2015-05-12T18:58:08.000Z",
	Status      string `json:"status"`       //:"processing"
	Progress    struct {
		Deliverable   int `json:"deliverable"`   //: 0,
		Undeliverable int `json:"undeliverable"` //: 0,
		Risky         int `json:"risky"`         //: 0,
		Unknown       int `json:"unknown"`       //: 0,
		Total         int `json:"total"`         //: 0,
		Unprocessed   int `json:"unprocessed"`   //: 2
	} `json:"progress"`
	Stats struct {
		Deliverable   int     `json:"deliverable"`   //: 1,
		Undeliverable int     `json:"undeliverable"` //: 1,
		Risky         int     `json:"risky"`         //: 0,
		Unknown       int     `json:"unknown"`       //: 0,
		Sendex        float32 `json:"sendex"`        //: 0.35,
		Addresses     int     `json:"addresses"`     //: 2
	} `json:"stats"`
	Success  bool   `json:"success"`  //: true,
	Message  string `json:"message"`  //: null
	Error    string `json:"error"`    //: null,
	Duration int    `json:"duration"` //: 0,
}

// IsStarting indicates that the process is starting
func (c *CheckJobStatusResponse) IsStarting() bool {
	return c.Status == "starting"
}

// IsProcessing indicates that the validation is being processed
func (c *CheckJobStatusResponse) IsProcessing() bool {
	return c.Status == "processing"
}

// IsCompleted indicates that the validation is completed
func (c *CheckJobStatusResponse) IsCompleted() bool {
	return c.Status == "completed"
}

// CreditBalanceResponse .
type CreditBalanceResponse struct {
	Balance int    `json:"balance"` //: 1337,
	Success bool   `json:"success"` //: true,
	Message string `json:"message"` //: null
}

func (v *CreditBalanceResponse) Error() error {
	if !v.Success {
		if v.Message != "" {
			return errors.New(v.Message)
		}
		return ErrUnknownErrorVerifyingEmail
	}

	return nil
}

// DisposableEmailCheckResponse .
type DisposableEmailCheckResponse struct {
	Disposable bool `json:"disposable"`
}

// Client holding credentials and connection
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New Client
func New(apiKey string, httpClient *http.Client) *Client {
	return NewWith(BaseURL, apiKey, httpClient)
}

// NewWith a baseURL for test stuff
func NewWith(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 2 * time.Second,
		}
	}

	return &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// Verify .
func (c *Client) Verify(email string) (*VerifyResponse, error) {
	segments := map[string]string{"API_VERSION": "v2"}
	params := map[string]string{"email": email}
	return c.callVerifyService(segments, params)
}

// VerifyMultiple .
func (c *Client) VerifyMultiple(urlCallback, filename string, data []byte) (*VerifyMultipleResponse, error) {
	segments := map[string]string{"API_VERSION": "v2"}
	params := map[string]string{}
	headers := map[string]string{}
	if urlCallback != "" {
		headers["X-Kickbox-Callback"] = urlCallback
	}
	if filename != "" {
		headers["X-Kickbox-Filename"] = filename
	}
	return c.callVerifyMultipleService(headers, segments, params, data)
}

// CheckJobStatus .
func (c *Client) CheckJobStatus(jobID int) (*CheckJobStatusResponse, error) {
	segments := map[string]string{
		"API_VERSION": "v2",
		"JOB_ID":      strconv.Itoa(jobID),
	}
	return c.callCheckJobStatusService(segments)
}

// CreditBalance .
func (c *Client) CreditBalance() (*CreditBalanceResponse, error) {
	segments := map[string]string{
		"API_VERSION": "v2",
	}
	return c.callCreditBalanceService(segments)
}

// Disposable .
func (c *Client) Disposable(email string) (*DisposableEmailCheckResponse, error) {
	segments := map[string]string{
		"API_VERSION":   "v1",
		"EMAIL_ADDRESS": email,
	}
	return c.callDisposableService(segments)
}

func (c *Client) callVerifyService(segments, params map[string]string) (*VerifyResponse, error) {
	endpoint := Endpoints[Verify]

	request, err := c.buildRequest(endpoint.Method, endpoint.Path, segments, params, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.callService(request)

	var response VerifyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, err
}

func (c *Client) callVerifyMultipleService(headers, segments, params map[string]string, data []byte) (*VerifyMultipleResponse, error) {
	endpoint := Endpoints[VerifyMultiple]

	request, err := c.buildRequest(endpoint.Method, endpoint.Path, segments, params, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	body, err := c.callService(request)

	var response VerifyMultipleResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, err
}

func (c *Client) callCheckJobStatusService(segments map[string]string) (*CheckJobStatusResponse, error) {
	endpoint := Endpoints[CheckJobStatus]

	request, err := c.buildRequest(endpoint.Method, endpoint.Path, segments, map[string]string{}, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.callService(request)

	var response CheckJobStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, err
}

func (c *Client) callCreditBalanceService(segments map[string]string) (*CreditBalanceResponse, error) {
	endpoint := Endpoints[CreditBalance]

	request, err := c.buildRequest(endpoint.Method, endpoint.Path, segments, map[string]string{}, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.callService(request)

	var response CreditBalanceResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, err
}

func (c *Client) callDisposableService(segments map[string]string) (*DisposableEmailCheckResponse, error) {
	endpoint := Endpoints[DisposableEmailCheck]

	request, err := c.buildRequest(endpoint.Method, endpoint.Path, segments, map[string]string{}, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.callService(request)

	var response DisposableEmailCheckResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	return &response, err
}

func (c *Client) callService(request *http.Request) ([]byte, error) {
	_, body, err := c.doRequest(request)
	if err != nil {
		return nil, fmt.Errorf("Error doing request: %s", err.Error())
	}

	if len(body) == 0 {
		return nil, ErrEmptyResponse
	}

	return body, nil
}

func (c *Client) buildRequest(method, path string, segments, params map[string]string, body io.Reader) (*http.Request, error) {
	// Default required values
	params["apikey"] = c.apiKey

	for search, replace := range segments {
		path = strings.Replace(path, "{"+search+"}", replace, -1)
	}

	URL, err := c.buildURL(path, params)

	if err != nil {
		return nil, err
	}

	return http.NewRequest(method, URL, body)
}

func (c *Client) buildURL(path string, params map[string]string) (string, error) {

	u, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	u.Path += path

	queryString := u.Query()
	for k, v := range params {
		queryString.Set(k, v)
	}

	u.RawQuery = queryString.Encode()

	return u.String(), nil
}

func (c *Client) doRequest(request *http.Request) (int, []byte, error) {
	resp, err := c.httpClient.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}

	return resp.StatusCode, body, nil
}
