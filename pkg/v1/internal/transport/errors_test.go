package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestError_Reporting(t *testing.T) {
	t.Run("KeepsDiagnosticDetails", func(t *testing.T) {
		err := &Error{
			Kind:       KindForbidden,
			StatusCode: http.StatusForbidden,
			RequestID:  "req-0a1b2c3d",
			Message:    "policy does not allow this operation",
		}

		message := err.Error()

		require.Contains(t, message, string(KindForbidden))
		require.Contains(t, message, "403")
		require.Contains(t, message, "req-0a1b2c3d")
		require.Contains(t, message, "policy does not allow this operation")
	})

	t.Run("UnwrapsCause", func(t *testing.T) {
		cause := errors.New("connection reset by peer")

		require.ErrorIs(t, &Error{Kind: KindTransport, Err: cause}, cause)
	})

	t.Run("IsKind", func(t *testing.T) {
		err := &Error{Kind: KindNotFound, StatusCode: http.StatusNotFound}

		require.True(t, IsKind(err, KindNotFound))
		require.False(t, IsKind(err, KindForbidden))
		require.False(t, IsKind(errors.New("plain error"), KindNotFound))
	})

	t.Run("IsKindWalksTheChain", func(t *testing.T) {
		err := &Error{Kind: KindIncompleteList, Err: &Error{Kind: KindForbidden}}

		require.True(t, IsKind(err, KindIncompleteList))
		require.True(t, IsKind(err, KindForbidden))
		require.False(t, IsKind(err, KindNotFound))
	})
}

func TestError_ClassesOfAPIResponses(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
		expected   Kind
	}{
		{
			name:       "InvalidRequest",
			statusCode: http.StatusBadRequest,
			body:       `{"badRequest": {"code": 400, "message": "Invalid share size."}}`,
			expected:   KindInvalidRequest,
		},
		{
			name:       "UnprocessableRequest",
			statusCode: http.StatusUnprocessableEntity,
			body:       `{"badRequest": {"code": 422, "message": "Unprocessable entity."}}`,
			expected:   KindInvalidRequest,
		},
		{
			name:       "RejectedToken",
			statusCode: http.StatusUnauthorized,
			body:       `{"error": {"message": "The request you have made requires authentication."}}`,
			expected:   KindForbidden,
		},
		{
			name:       "ForbiddenOperation",
			statusCode: http.StatusForbidden,
			body:       apiForbiddenJSON,
			expected:   KindForbidden,
		},
		{
			name:       "MissingShare",
			statusCode: http.StatusNotFound,
			body:       apiNotFoundJSON,
			expected:   KindNotFound,
		},
		{
			name:       "RejectedMicroversion",
			statusCode: http.StatusNotAcceptable,
			body: `{"computeFault": {"code": 406, "message": "Version 2.75 is not supported by the API. ` +
				`Minimum is 2.0 and maximum is 2.65."}}`,
			expected: KindMicroversion,
		},
		{
			name:       "StateConflict",
			statusCode: http.StatusConflict,
			body:       `{"conflictingRequest": {"code": 409, "message": "Share status must be available."}}`,
			expected:   KindConflict,
		},
		{
			name:       "ExhaustedQuota",
			statusCode: http.StatusRequestEntityTooLarge,
			body: `{"overLimit": {"code": 413, "message": "Requested share exceeds allowed project/user ` +
				`or share type gigabytes quota.", "retryAfter": "0"}}`,
			expected: KindOverQuota,
		},
		{
			name:       "OversizedRequestRejectedByAProxy",
			statusCode: http.StatusRequestEntityTooLarge,
			body:       `<html><body>413 Request Entity Too Large</body></html>`,
			expected:   KindInvalidRequest,
		},
		{
			name:       "ThrottledByAProxy",
			statusCode: http.StatusTooManyRequests,
			body:       `<html><body>429 Too Many Requests</body></html>`,
			expected:   KindRateLimited,
		},
		{
			name:       "ServerError",
			statusCode: http.StatusInternalServerError,
			body:       apiFaultJSON,
			expected:   KindServerError,
		},
		{
			name:       "UnavailableService",
			statusCode: http.StatusServiceUnavailable,
			body:       ``,
			expected:   KindServerError,
		},
		{
			name:       "UnexpectedResponse",
			statusCode: http.StatusTeapot,
			body:       `{"teapot": {"message": "I am a teapot."}}`,
			expected:   KindUnexpected,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := &testHTTPClient{answers: []testAnswer{{
				status: testCase.statusCode,
				body:   testCase.body,
				header: http.Header{headerRequestID: []string{"req-" + testCase.name}},
			}}}

			meta, err := DoRequest(
				t.Context(), testClient(httpClient), http.MethodGet, "shares", http.StatusOK, nil, nil,
			)

			require.Error(t, err)

			var sdkErr *Error
			require.ErrorAs(t, err, &sdkErr)

			require.Equal(t, testCase.expected, sdkErr.Kind)
			require.Equal(t, testCase.statusCode, sdkErr.StatusCode)
			require.Equal(t, "req-"+testCase.name, sdkErr.RequestID)
			require.Equal(t, testCase.statusCode, meta.StatusCode)
		})
	}
}

// A caller treats only a structured Manila fault as a definite rejection of a create request.
func TestError_StructuredManilaFault(t *testing.T) {
	cases := []struct {
		name       string
		statusCode int
		body       string
		structured bool
	}{
		{
			name:       "BadRequestFromManila",
			statusCode: http.StatusBadRequest,
			body:       `{"badRequest": {"code": 400, "message": "IP address: 10.0.0.5 is already in use."}}`,
			structured: true,
		},
		{
			name:       "ForbiddenFromManila",
			statusCode: http.StatusForbidden,
			body:       `{"forbidden": {"code": 403, "message": "Maximum number of share-networks allowed (10) exceeded."}}`,
			structured: true,
		},
		{
			name:       "BadRequestFromAProxy",
			statusCode: http.StatusBadRequest,
			body:       `<html><body>400 Bad Request</body></html>`,
			structured: false,
		},
		{
			name:       "ThrottledByAProxy",
			statusCode: http.StatusTooManyRequests,
			body:       `<html><body>429 Too Many Requests</body></html>`,
			structured: false,
		},
		{
			name:       "EmptyClientError",
			statusCode: http.StatusBadRequest,
			body:       ``,
			structured: false,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			httpClient := &testHTTPClient{answers: []testAnswer{{status: testCase.statusCode, body: testCase.body}}}

			_, err := DoRequest(
				t.Context(), testClient(httpClient), http.MethodPost, "shares", http.StatusOK, nil, nil,
			)

			var sdkErr *Error
			require.ErrorAs(t, err, &sdkErr)
			require.Equal(t, testCase.structured, sdkErr.StructuredFault)
		})
	}
}

func TestError_QuotaIsNotRateLimit(t *testing.T) {
	require.NotEqual(t, KindRateLimited, KindOverQuota)
	require.Equal(t, KindOverQuota, kindOfFault(http.StatusRequestEntityTooLarge, &fault{RetryAfter: "0"}))
	require.Equal(t, KindRateLimited, kindOfFault(http.StatusTooManyRequests, nil))

	require.Equal(t, KindInvalidRequest, kindOfFault(http.StatusRequestEntityTooLarge, &fault{Code: 413}))
	require.Equal(t, KindInvalidRequest, kindOfFault(http.StatusRequestEntityTooLarge, nil))
}

func TestError_LongQuotaFaultKeepsClassification(t *testing.T) {
	body, err := json.Marshal(map[string]any{
		"overLimit": map[string]any{
			"code":       http.StatusRequestEntityTooLarge,
			"message":    strings.Repeat("ShareSizeExceedsAvailableQuota ", 50),
			"retryAfter": "0",
		},
	})
	require.NoError(t, err)
	require.Greater(t, len(body), maxDiagnosticMessageLength)

	httpClient := &testHTTPClient{answers: []testAnswer{{
		status: http.StatusRequestEntityTooLarge,
		body:   string(body),
	}}}

	_, err = DoRequest(
		t.Context(), testClient(httpClient), http.MethodPost, "shares", http.StatusAccepted, nil, nil,
	)

	var sdkErr *Error
	require.ErrorAs(t, err, &sdkErr)
	require.Equal(t, KindOverQuota, sdkErr.Kind)
	require.True(t, sdkErr.StructuredFault)
	require.Len(t, sdkErr.Message, maxDiagnosticMessageLength)
	require.True(t, strings.HasSuffix(sdkErr.Message, diagnosticTruncationMarker))
}

func TestError_DiagnosticMessage(t *testing.T) {
	t.Run("FromAPIBody", func(t *testing.T) {
		httpClient := &testHTTPClient{answers: []testAnswer{{
			status: http.StatusNotFound,
			body:   `{"itemNotFound": {"message": "Share 0e1c2b3a could not be found.", "code": 404}}`,
			header: http.Header{headerComputeRequestID: []string{"req-legacy-header"}},
		}}}

		_, err := DoRequest(
			t.Context(), testClient(httpClient), http.MethodGet, "shares/0e1c2b3a", http.StatusOK, nil, nil,
		)

		var sdkErr *Error
		require.ErrorAs(t, err, &sdkErr)

		require.Equal(t, "Share 0e1c2b3a could not be found.", sdkErr.Message)
		require.Equal(t, "req-legacy-header", sdkErr.RequestID)
		require.True(t, sdkErr.StructuredFault)
	})

	t.Run("FromPlainBody", func(t *testing.T) {
		httpClient := &testHTTPClient{answers: []testAnswer{
			{status: http.StatusBadGateway, body: "<html>\n  <body>502 Bad Gateway</body>\n</html>"},
		}}

		_, err := DoRequest(
			t.Context(), testClient(httpClient), http.MethodGet, "shares", http.StatusOK, nil, nil,
		)

		var sdkErr *Error
		require.ErrorAs(t, err, &sdkErr)

		require.Equal(t, "<html> <body>502 Bad Gateway</body> </html>", sdkErr.Message)
		require.Empty(t, sdkErr.RequestID)
		require.False(t, sdkErr.StructuredFault)
	})

	t.Run("IsTruncated", func(t *testing.T) {
		httpClient := &testHTTPClient{answers: []testAnswer{
			{status: http.StatusInternalServerError, body: strings.Repeat("a", 4096)},
		}}

		_, err := DoRequest(
			t.Context(), testClient(httpClient), http.MethodGet, "shares", http.StatusOK, nil, nil,
		)

		var sdkErr *Error
		require.ErrorAs(t, err, &sdkErr)

		require.Len(t, sdkErr.Message, maxDiagnosticMessageLength)
	})

	t.Run("OfEmptyBody", func(t *testing.T) {
		httpClient := &testHTTPClient{answers: []testAnswer{{status: http.StatusForbidden, body: ""}}}

		_, err := DoRequest(
			t.Context(), testClient(httpClient), http.MethodGet, "shares", http.StatusOK, nil, nil,
		)

		var sdkErr *Error
		require.ErrorAs(t, err, &sdkErr)

		require.True(t, IsKind(sdkErr, KindForbidden))
		require.Empty(t, sdkErr.Message)
		require.False(t, sdkErr.StructuredFault)
	})
}
