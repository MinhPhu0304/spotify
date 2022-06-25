package trace

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type tracingTransport struct {
	http.RoundTripper
}

func newTracingTransport(roundTripper http.RoundTripper) *tracingTransport {
	return &tracingTransport{RoundTripper: roundTripper}
}

func (t *tracingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	opName := fmt.Sprintf("HTTP %s %s", req.Method, req.URL.String())
	span := sentry.StartSpan(req.Context(), opName)
	defer span.Finish()

	span.SetTag("url", req.URL.String())
	if span.Data == nil {
		span.Data = make(map[string]interface{})
	}

	// adding sentry header for distributed tracing
	req.Header.Add("sentry-trace", span.TraceID.String())

	response, err := t.RoundTripper.RoundTrip(req)

	span.Data["http_code"] = response.StatusCode

	return response, err
}

func WrapWithTrace(client *http.Client) *http.Client {
	client.Transport = newTracingTransport(client.Transport)
	return client
}
