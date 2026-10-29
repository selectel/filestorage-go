// Package testhttp provides a recording HTTP client for the resource package tests.
package testhttp

import (
	"io"
	"net/http"
	"strings"
	"testing"

	v1 "github.com/selectel/filestorage-go/pkg/v1"
	"github.com/stretchr/testify/require"
)

const (
	TokenID   = "fake-project-scoped-token" //nolint:gosec // G101: this is a fake value used by the tests.
	ProjectID = "test-project-id"
	Endpoint  = "https://api.example.com/v2/" + ProjectID
	BasePath  = "/v2/" + ProjectID
)

type Answer struct {
	Status int
	Body   string
}

// Client records every request and replays the answers in order; the last answer repeats.
type Client struct {
	Requests []*http.Request
	Bodies   []string
	Answers  []Answer
	Err      error
}

func (client *Client) Do(request *http.Request) (*http.Response, error) {
	body := ""
	if request.Body != nil {
		raw, _ := io.ReadAll(request.Body)
		body = string(raw)
	}

	client.Requests = append(client.Requests, request)
	client.Bodies = append(client.Bodies, body)

	if client.Err != nil {
		return nil, client.Err
	}

	answer := client.Answers[0]
	if len(client.Answers) > 1 {
		client.Answers = client.Answers[1:]
	}

	return &http.Response{
		StatusCode: answer.Status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(answer.Body)),
	}, nil
}

func (client *Client) LastRequest(t *testing.T) *http.Request {
	t.Helper()

	require.NotEmpty(t, client.Requests)

	return client.Requests[len(client.Requests)-1]
}

func (client *Client) LastBody(t *testing.T) string {
	t.Helper()

	require.NotEmpty(t, client.Bodies)

	return client.Bodies[len(client.Bodies)-1]
}

func NewSDKClient(httpClient v1.HTTPClient) *v1.Client {
	client, err := v1.NewClient(v1.Config{Endpoint: Endpoint, Token: TokenID, HTTPClient: httpClient})
	if err != nil {
		panic(err)
	}

	return client
}
