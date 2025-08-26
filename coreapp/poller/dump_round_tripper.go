package poller

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"go.uber.org/zap"
)

type dumpRoundTripper struct {
	transport http.RoundTripper
}

func newDumpRoundTripper(transport http.RoundTripper) *dumpRoundTripper {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &dumpRoundTripper{transport: transport}
}

func (d *dumpRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqDump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		zap.L().Error("failed to dump request", zap.Error(err))
	} else {
		zap.L().Debug(fmt.Sprintf("---REQUEST---\n%s\n", string(reqDump)))
	}

	resp, err := d.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	// We need to read the response body to dump it, but we also need to allow the original caller to read it.
	// So we read it into a buffer, and then replace the response body with a new reader from the buffer.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error("failed to read response body", zap.Error(err))
		return resp, nil
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		zap.L().Error("failed to dump response", zap.Error(err))
	} else {
		zap.L().Debug(fmt.Sprintf("---RESPONSE---\n%s\n", string(respDump)))
	}

	// Restore the body so it can be read again.
	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return resp, nil
}
