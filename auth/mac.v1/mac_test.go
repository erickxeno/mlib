package mac

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSignRequest(t *testing.T) {
	var tests = []struct {
		method  string
		url     string
		body    io.Reader
		header  map[string]string
		signExp string
	}{
		{
			method:  http.MethodGet,
			url:     "http://foo.com/bar.jpg",
			body:    nil,
			signExp: "lNDZqBwBi92LguarI2DzcLrM7dY=",
		},
		{
			method:  http.MethodGet,
			url:     "http://foo.com/bar.jpg",
			body:    strings.NewReader("a=b&c=d"),
			signExp: "lNDZqBwBi92LguarI2DzcLrM7dY=",
		},
		{
			method:  http.MethodGet,
			url:     "http://foo.com/bar.jpg?a=b&c=d",
			signExp: "y1-OyJF2bryu4scccyEUG0Z_Hw0=",
		},
		{
			method: http.MethodGet,
			url:    "http://foo.com/bar.jpg",
			body:   strings.NewReader("a=b&c=d"),
			header: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
			},
			signExp: "601BWwe7eHhrt8FNawCFa4ISXNY=",
		},
	}

	for _, test := range tests {
		req, _ := http.NewRequest(test.method, test.url, test.body)
		for k, v := range test.header {
			req.Header.Add(k, v)
		}
		sign, err := DefaultRequestSigner.Sign([]byte("12345678901234567890123456789012"), req)
		if err != nil {
			t.Error(err)
		}

		if base64.URLEncoding.EncodeToString(sign) != test.signExp {
			t.Error("unexcept sign str")
		}
	}
}
