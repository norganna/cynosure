// Package http is a HTTP(S) checker.
//
// Accepts no config.
//
// The "wait" parameter will be either a plain URL, or JSON of the form:
//    {
//      "url": URL,
//      "method": METHOD,
//      "body": POST,
//      "headers": {
//        HEADER: VALUE,
//        ...
//      },
//      "expect": {
//        "code": CODE,
//        "body": FIND,
//        "headers": {
//          HEADER: FIND,
//        }
//      }
//    }
//
// If only a URL is supplied, it will expand to:
//
//    {"url": URL}
//
// Default values (if none specified):
//    method = "GET"
//    expect = {"code": 200}
//
// CODE can be any http status code number (not a string), e.g. 200, 404, 302 etc.
// FIND can be any substring to find within the corresponding content.
//
// If multiple expect values are specified, then all must match.
package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/norganna/cynosure/common"
	"github.com/norganna/cynosure/deps"
)

// Kind contains the kind string of this provider.
const Kind = "http"

func init() {
	deps.RegisterProvider(Kind, create)
}

// create returns the Broker.
func create(_ common.StringMap) (deps.Broker, error) {
	b := &broker{}
	return b, nil
}

type broker struct{}

type expectation struct {
	Code    int              `json:"code,omitempty"`
	Body    string           `json:"body,omitempty"`
	Headers common.StringMap `json:"headers,omitempty"`
}

type request struct {
	URL     string           `json:"url,omitempty"`
	Method  string           `json:"method,omitempty"`
	Body    string           `json:"body,omitempty"`
	Headers common.StringMap `json:"headers,omitempty"`
	Expect  expectation      `json:"expect,omitempty"`
}

func (b *broker) Dep(wait string) (deps.Depender, error) {
	r := &request{}
	if wait[0:1] != "{" {
		r.URL = wait
	} else {
		err := json.Unmarshal([]byte(wait), r)
		if err != nil {
			return nil, common.Error(err, "failed to parse dependency condition JSON")
		}
	}

	u, err := url.Parse(r.URL)
	if err != nil {
		return nil, common.Error(err, "failed to parse URL %s", r.URL)
	}

	return &dep{
		r: r,
		u: u,
	}, nil
}

type dep struct {
	r *request
	u *url.URL
}

func (d *dep) Check() (msg string, ok bool) {
	msg = fmt.Sprintf("%s %s", Kind, d.u.Host)

	var reader *bytes.Buffer
	if d.r.Body != "" {
		reader = bytes.NewBuffer([]byte(d.r.Body))
	}

	req, err := http.NewRequest(d.r.Method, d.r.URL, reader)
	if err != nil {
		return msg + " " + err.Error(), false
	}

	if d.r.Headers != nil {
		for header, value := range d.r.Headers {
			req.Header.Add(header, value)
		}
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return msg + " " + err.Error(), false
	}

	var waiting []string

	expectCode := d.r.Expect.Code
	if expectCode > 0 {
		if expectCode != res.StatusCode {
			waiting = append(waiting, fmt.Sprintf("code %d", expectCode))
		}
	}

	expectBody := []byte(d.r.Expect.Body)
	if len(expectBody) > 0 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return msg + " " + err.Error(), false
		}

		if bytes.Index(body, expectBody) == -1 {
			waiting = append(waiting, "body text")
		}
	}

	h := d.r.Expect.Headers
	if h != nil {
		for header, match := range h {
			value := res.Header.Get(header)
			if value == "" || match != "" && strings.Index(value, match) == -1 {
				waiting = append(waiting, "header "+header)
			}
		}
	}

	if len(waiting) > 0 {
		return msg + " waiting for " + strings.Join(waiting, ", "), false
	}

	return msg + " good", ok
}
