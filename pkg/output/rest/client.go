package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ernesto2108/anvil/pkg/errors"
)

type Client interface {
	Request(ctx context.Context, method, url string, opts ...Options) error
}

type Options func(r *Option)

type Option struct {
	BasicAuth       *basicAuth
	Headers, Params map[string]string
	Body, Response  interface{}
	Type            string
	Timeout         int64
}

type basicAuth struct {
	username string
	password string
}

func WithBasicAuth(username, password string) Options {
	return func(r *Option) {
		r.BasicAuth = &basicAuth{
			username: username,
			password: password,
		}
	}
}

func WithHeaders(headers map[string]string) Options {
	return func(r *Option) {
		r.Headers = headers
	}
}

func WithParams(params map[string]string) Options {
	return func(r *Option) {
		r.Params = params
	}
}

func WithBody(body interface{}) Options {
	return func(r *Option) {
		r.Body = body
	}
}

func WithResponse(response interface{}) Options {
	return func(r *Option) {
		r.Response = response
	}
}

func WithTimeout(timeout int64) Options {
	return func(r *Option) {
		r.Timeout = timeout
	}
}

func WithType(value string) Options {
	return func(r *Option) {
		r.Type = value
	}
}

type Error struct {
	Response *http.Response
	Body     []byte
}

func (e *Error) Error() string {
	if e.Response != nil {
		return fmt.Sprintf("HTTP %d: %s", e.Response.StatusCode, e.Response.Status)
	}

	return "unknown HTTP errors"
}

type client struct {
	defaultTimeout time.Duration
}

func NewClient(timeout int) Client {
	return &client{
		defaultTimeout: time.Duration(timeout) * time.Second,
	}
}

func (c client) Request(ctx context.Context, method, url string, opts ...Options) error {
	var (
		bodyRequest bytes.Buffer
		opt         Option
	)

	for _, v := range opts {
		v(&opt)
	}

	if err := encodeRequestBody(&bodyRequest, opt.Body); err != nil {
		return err
	}

	timeout := c.defaultTimeout
	if opt.Timeout > 0 {
		timeout = time.Duration(opt.Timeout) * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := createHTTPRequest(reqCtx, method, url, &bodyRequest, opt)
	if err != nil {
		return err
	}

	httpClient := &http.Client{}

	return c.executeRequest(httpClient, request, opt.Response)
}

func encodeRequestBody(buffer *bytes.Buffer, body interface{}) error {
	if body == nil {
		return nil
	}

	if err := json.NewEncoder(buffer).Encode(body); err != nil {
		return errors.New(
			errors.BadRequestErr,
			errors.WithMessage("failed to encode request body"),
		)
	}

	return nil
}

func createHTTPRequest(ctx context.Context,
	method, url string,
	bodyBuffer *bytes.Buffer,
	opt Option) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, bodyBuffer)
	if err != nil {
		return nil, errors.New(
			errors.BadRequestErr,
			errors.WithMessage("failed to create request"),
		)
	}

	if len(opt.Params) > 0 {
		q := request.URL.Query()
		for k, v := range opt.Params {
			q.Add(k, v)
		}

		request.URL.RawQuery = q.Encode()
	}

	if len(opt.Headers) > 0 {
		for k, v := range opt.Headers {
			request.Header.Set(k, v)
		}
	}

	if opt.BasicAuth != nil {
		request.SetBasicAuth(
			opt.BasicAuth.username,
			opt.BasicAuth.password,
		)
	}

	return request, nil
}

func (c client) executeRequest(client *http.Client, request *http.Request, responseObj interface{}) error {
	response, err := client.Do(request)
	if err != nil {
		return errors.New(
			errors.InternalErr,
			errors.WithMessage("failed to execute request"),
		)
	}

	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close response body: %w", closeErr)
		}
	}()

	bodyResponse, err := io.ReadAll(response.Body)
	if err != nil {
		return errors.New(
			errors.InternalErr,
			errors.WithMessage("failed to read response body"),
		)
	}

	return handleResponse(response.StatusCode, bodyResponse, responseObj)
}

func handleResponse(statusCode int, bodyResponse []byte, response interface{}) error {
	if statusCode == http.StatusNoContent {
		return nil
	}

	if statusCode == http.StatusOK || statusCode == http.StatusCreated {
		if response != nil {
			if err := json.Unmarshal(bodyResponse, response); err != nil {
				return errors.New(
					errors.InternalErr,
					errors.WithMessage(fmt.Sprintf("failed to decode response body: %s", string(bodyResponse))),
					errors.WithError(err),
				)
			}
		}

		return nil
	}

	errorMappings := map[int]error{
		http.StatusBadRequest: errors.New(
			errors.BadRequestErr,
			errors.WithMessage(fmt.Sprintf("bad request (status code: %d)", statusCode)),
		),
		http.StatusUnauthorized: errors.New(
			errors.UnauthorizedErr,
			errors.WithMessage(fmt.Sprintf("unauthorized (status code: %d)", statusCode)),
		),
		http.StatusNotFound: errors.New(
			errors.NotFoundErr,
			errors.WithMessage(fmt.Sprintf("not found (status code: %d)", statusCode)),
		),
		http.StatusInternalServerError: errors.New(
			errors.InternalErr,
			errors.WithMessage(fmt.Sprintf("server error (status code: %d)", statusCode)),
		),
	}

	if err, exists := errorMappings[statusCode]; exists {
		return err
	}

	return errors.New(
		errors.InternalErr,
		errors.WithMessage("unexpected status code"),
	)
}
