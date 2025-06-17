package mac

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	sk = []byte("secret_key")
	su = "su_info"
)

func Test_Sign(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")

	act, err := SignRequestWithHeader(sk, req)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET /path/to/api?param=value\nHost: example.com\nContent-Type: application/json\n\n"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_SignWithHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xeno-Meta-App", "value")

	act, err := SignRequestWithHeader(sk, req)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET /path/to/api?param=value\nHost: example.com\nContent-Type: application/json" +
		"\nX-Xeno-Meta-App: value" +
		"\n\n"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_SignAdmin(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")

	act, err := SignAdminRequestWithHeader(sk, req, su)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET /path/to/api?param=value\nHost: example.com\nContent-Type: application/json" +
		"\nAuthorization: Admin " + su +
		"\n\n"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_SignAdminWithHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://example.com/path/to/api?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Xeno-Meta-App", "value")

	act, err := SignAdminRequestWithHeader(sk, req, su)
	assert.NoError(t, err)

	h := hmac.New(sha1.New, sk)
	h.Write([]byte("GET /path/to/api?param=value\nHost: example.com\nContent-Type: application/json" +
		"\nAuthorization: Admin " + su +
		"\nX-Xeno-Meta-App: value" +
		"\n\n"))
	exp := h.Sum(nil)

	assert.Equal(t, exp, act)
}

func Test_signHeaderValues(t *testing.T) {
	w := bytes.NewBuffer(nil)

	header := make(http.Header)
	header.Set("X-Base-Meta", "value")

	signHeaderValues(header, w)
	assert.Empty(t, w.String())

	header.Set("X-Xeno-Cxxxx", "valuec")
	header.Set("X-Xeno-Bxxxx", "valueb")
	header.Set("X-Xeno-axxxx", "valuea")
	header.Set("X-Xeno-e", "value")
	header.Set("X-Xeno-", "value")
	header.Set("X-Xeno", "value")
	header.Set("", "value")
	signHeaderValues(header, w)

	assert.Equal(t, `
X-Xeno-Axxxx: valuea
X-Xeno-Bxxxx: valueb
X-Xeno-Cxxxx: valuec
X-Xeno-E: value`, w.String())
}
