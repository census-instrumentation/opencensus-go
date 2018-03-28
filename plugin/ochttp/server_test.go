package ochttp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/net/http2"

	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

func httpHandler(statusCode, respSize int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		body := make([]byte, respSize)
		w.Write(body)
	})
}

func updateMean(mean float64, sample, count int) float64 {
	if count == 1 {
		return float64(sample)
	}
	return mean + (float64(sample)-mean)/float64(count)
}

func TestHandlerStatsCollection(t *testing.T) {
	for _, v := range DefaultServerViews {
		v.Subscribe()
	}

	views := []string{
		"opencensus.io/http/server/request_count",
		"opencensus.io/http/server/latency",
		"opencensus.io/http/server/request_bytes",
		"opencensus.io/http/server/response_bytes",
	}

	// TODO: test latency measurements?
	tests := []struct {
		name, method, target                 string
		count, statusCode, reqSize, respSize int
	}{
		{"get 200", "GET", "http://opencensus.io/request/one", 10, 200, 512, 512},
		{"post 503", "POST", "http://opencensus.io/request/two", 5, 503, 1024, 16384},
		{"no body 302", "GET", "http://opencensus.io/request/three", 2, 302, 0, 0},
	}
	totalCount, meanReqSize, meanRespSize := 0, 0.0, 0.0

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := bytes.NewBuffer(make([]byte, test.reqSize))
			r := httptest.NewRequest(test.method, test.target, body)
			w := httptest.NewRecorder()
			h := &Handler{
				Handler: httpHandler(test.statusCode, test.respSize),
			}
			h.StartOptions.Sampler = trace.NeverSample()

			for i := 0; i < test.count; i++ {
				h.ServeHTTP(w, r)
				totalCount++
				// Distributions do not track sum directly, we must
				// mimic their behaviour to avoid rounding failures.
				meanReqSize = updateMean(meanReqSize, test.reqSize, totalCount)
				meanRespSize = updateMean(meanRespSize, test.respSize, totalCount)
			}
		})
	}

	for _, viewName := range views {
		v := view.Find(viewName)
		if v == nil {
			t.Errorf("view not found %q", viewName)
			continue
		}
		rows, err := view.RetrieveData(viewName)
		if err != nil {
			t.Error(err)
			continue
		}
		if got, want := len(rows), 1; got != want {
			t.Errorf("len(%q) = %d; want %d", viewName, got, want)
			continue
		}
		data := rows[0].Data

		var count int
		var sum float64
		switch data := data.(type) {
		case *view.CountData:
			count = int(*data)
		case *view.DistributionData:
			count = int(data.Count)
			sum = data.Sum()
		default:
			t.Errorf("Unkown data type: %v", data)
			continue
		}

		if got, want := count, totalCount; got != want {
			t.Fatalf("%s = %d; want %d", viewName, got, want)
		}

		// We can only check sum for distribution views.
		switch viewName {
		case "opencensus.io/http/server/request_bytes":
			if got, want := sum, meanReqSize*float64(totalCount); got != want {
				t.Fatalf("%s = %g; want %g", viewName, got, want)
			}
		case "opencensus.io/http/server/response_bytes":
			if got, want := sum, meanRespSize*float64(totalCount); got != want {
				t.Fatalf("%s = %g; want %g", viewName, got, want)
			}
		}
	}
}

// Test to ensure that our Handler proxies to its response the
// call to (http.Hijack).Hijacker() and that that successfully
// passes with HTTP/1.1 connections. See Issue #642
func TestHandlerProxiesHijack_HTTP1(t *testing.T) {
	cst := httptest.NewServer(&Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var writeMsg func(string)
			defer func() {
				err := recover()
				writeMsg(fmt.Sprintf("Proto=%s\npanic=%v", r.Proto, err != nil))
			}()
			conn, _, _ := w.(http.Hijacker).Hijack()
			writeMsg = func(msg string) {
				fmt.Fprintf(conn, "%s 200\nContentLength: %d", r.Proto, len(msg))
				fmt.Fprintf(conn, "\r\n\r\n%s", msg)
				conn.Close()
			}
		}),
	})
	defer cst.Close()

	testCases := []struct {
		name string
		tr   *http.Transport
		want string
	}{
		{
			name: "http1-transport",
			tr:   new(http.Transport),
			want: "Proto=HTTP/1.1\npanic=false",
		},
		{
			name: "http2-transport",
			tr: func() *http.Transport {
				tr := new(http.Transport)
				http2.ConfigureTransport(tr)
				return tr
			}(),
			want: "Proto=HTTP/1.1\npanic=false",
		},
	}

	for _, tc := range testCases {
		c := &http.Client{Transport: &Transport{Base: tc.tr}}
		res, err := c.Get(cst.URL)
		if err != nil {
			t.Errorf("(%s) unexpected error %v", tc.name, err)
			continue
		}
		blob, _ := ioutil.ReadAll(res.Body)
		res.Body.Close()
		if g, w := string(blob), tc.want; g != w {
			t.Errorf("(%s) got = %q; want = %q", tc.name, g, w)
		}
	}
}

// Test to ensure that our Handler proxies to its response the
// call to (http.Hijack).Hijacker() and that that crashes since http.Hijacker
// and HTTP/2.0 connections are incompatible, but the detection is only at runtime.
// and ensure that we can stream and flush to the connection. See Issue #642
func TestHandlerProxiesHijack_HTTP2(t *testing.T) {
	cst := httptest.NewUnstartedServer(&Handler{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				switch {
				case err == nil:
				default:
					// Unhandled error
				case strings.Contains(err.(error).Error(), "Hijack"):
					// Confirmed HTTP/2.0, let's stream to it
					for i := 0; i < 5; i++ {
						fmt.Fprintf(w, "%d\n", i)
						w.(http.Flusher).Flush()
					}
				}
			}()
			conn, _, _ := w.(http.Hijacker).Hijack()
			if conn != nil {
				data := fmt.Sprintf("Surprisingly got the Hijacker() Proto: %s", r.Proto)
				fmt.Fprintf(conn, "%s 200\nContent-Length:%d\r\n\r\n%s", r.Proto, len(data), data)
				conn.Close()
				return
			}
		}),
	})
	cst.TLS = &tls.Config{NextProtos: []string{"h2"}}
	cst.StartTLS()
	defer cst.Close()

	if wantPrefix := "https://"; !strings.HasPrefix(cst.URL, wantPrefix) {
		t.Fatalf("URL got = %q wantPrefix = %q", cst.URL, wantPrefix)
	}

	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	http2.ConfigureTransport(tr)
	c := &http.Client{Transport: tr}
	res, err := c.Get(cst.URL)
	if err != nil {
		t.Fatalf("Unexpected error %v", err)
	}
	blob, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if g, w := string(blob), "0\n1\n2\n3\n4\n"; g != w {
		t.Errorf("got = %q; want = %q", g, w)
	}
}
