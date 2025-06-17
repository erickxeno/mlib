package seekable

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSeekable_EOFIfReqAlreadyParsed(t *testing.T) {
	ast := require.New(t)
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	ast.NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	req.ParseForm()
	_, err = New(req)
	ast.Equal(err.Error(), "EOF")
}

func TestSeekable_WorkaroundForEOF(t *testing.T) {
	ast := require.New(t)
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	ast.NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, _ = New(req)
	req.ParseForm()
	ast.Equal(req.FormValue("a"), "1")
	_, err = New(req)
	ast.NoError(err)
}

func testSeekabler(t *testing.T, r io.ReadCloser, data string) {
	ast := require.New(t)
	b, _ := ioutil.ReadAll(r)
	ast.Equal(data, string(b))
	sr, ok := r.(Seekabler)
	ast.True(ok)
	err := sr.SeekToBegin()
	ast.NoError(err)
	b, _ = ioutil.ReadAll(sr)
	ast.Equal(data, string(b))
}

func TestSeekable(t *testing.T) {
	ast := require.New(t)
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	ast.NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, err = New(req)
	ast.NoError(err)
	testSeekabler(t, req.Body, body)

	req, err = http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	ast.NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	_, err = New(req)
	ast.NoError(err)
	testSeekabler(t, req.Body, body)
}

func TestSeekableLength(t *testing.T) {
	ast := require.New(t)
	old := MaxBodyLength
	defer func() {
		MaxBodyLength = old
	}()
	MaxBodyLength = 2
	body := "a=1"
	req, err := http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	ast.NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", "3")
	_, err = New(req)
	ast.Equal(ErrTooLargeBody, err)
	b, _ := ioutil.ReadAll(req.Body)
	ast.Equal(body, string(b))

	body = "a=11111111"
	req, err = http.NewRequest("POST", "/a", bytes.NewBufferString(body))
	ast.NoError(err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = -1
	_, err = New(req)
	ast.Equal(ErrTooLargeBody, err)
	b, _ = ioutil.ReadAll(req.Body)
	ast.Equal(body, string(b))
}
