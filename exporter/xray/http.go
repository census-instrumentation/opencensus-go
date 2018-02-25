package xray

import (
	"go.opencensus.io/plugin/ochttp"
)

// httpRequest – Information about an http request.
type httpRequest struct {
	// Method – The request method. For example, GET.
	Method string `json:"method,omitempty"`

	// URL – The full URL of the request, compiled from the protocol, hostname,
	// and path of the request.
	URL string `json:"url,omitempty"`

	// UserAgent – The user agent string from the requester's client.
	UserAgent string `json:"user_agent,omitempty"`

	// ClientIP – The IP address of the requester. Can be retrieved from the IP
	// packet's Source Address or, for forwarded requests, from an X-Forwarded-For
	// header.
	ClientIP string `json:"client_ip,omitempty"`

	// XForwardedFor – (segments only) boolean indicating that the client_ip was
	// read from an X-Forwarded-For header and is not reliable as it could have
	// been forged.
	XForwardedFor string `json:"x_forwarded_for,omitempty"`

	// Traced – (subsegments only) boolean indicating that the downstream call
	// is to another traced service. If this field is set to true, X-Ray considers
	// the trace to be broken until the downstream service uploads a segment with
	// a parent_id that matches the id of the subsegment that contains this block.
	//
	// TODO - need to understand the impact of this field
	//Traced bool `json:"traced"`
}

// httpResponse - Information about an http response.
type httpResponse struct {
	// Status – number indicating the HTTP status of the response.
	Status int64 `json:"status,omitempty"`

	// ContentLength – number indicating the length of the response body in bytes.
	ContentLength int64 `json:"content_length,omitempty"`
}

type http struct {
	Request  httpRequest  `json:"request"`
	Response httpResponse `json:"response"`
}

func makeHttp(attributes map[string]interface{}) (map[string]interface{}, *http, string) {
	var (
		host     string
		http     http
		filtered = map[string]interface{}{}
	)

	for key, value := range attributes {
		switch key {
		case ochttp.HostAttribute:
			host, _ = value.(string)

		case ochttp.MethodAttribute:
			http.Request.Method, _ = value.(string)

		case ochttp.UserAgentAttribute:
			http.Request.UserAgent, _ = value.(string)

		case ochttp.StatusCodeAttribute:
			http.Response.Status, _ = value.(int64)

		default:
			filtered[key] = value
		}
	}

	if len(filtered) == len(attributes) {
		return attributes, nil, ""
	}

	return filtered, &http, host
}
