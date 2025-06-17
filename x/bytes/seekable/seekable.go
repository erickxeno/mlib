// Package seekable 提供了一个用于读取和替换 http.Request 请求体的工具。
// 它允许对 HTTP 请求体进行多次读取，这在某些场景下非常有用，比如：
// 1. 需要多次读取请求体内容
// 2. 需要在读取请求体后仍然保持原始请求体内容
// 3. 需要限制请求体大小
package seekable

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/erickxeno/mlib/x/bytes"
)

// ---------------------------------------------------

// Seekabler 接口定义了可重复读取的基本操作
type Seekabler interface {
	// Bytes 返回当前可读取的所有字节
	Bytes() []byte
	// Read 实现了 io.Reader 接口，用于读取数据
	Read(val []byte) (n int, err error)
	// SeekToBegin 将读取位置重置到开始位置
	SeekToBegin() error
}

// SeekableCloser 接口组合了 Seekabler 和 io.Closer 接口
// 提供了可重复读取和关闭资源的能力
type SeekableCloser interface {
	Seekabler
	io.Closer
}

// readCloser 结构体组合了 Seekabler 和 io.Closer
// 用于包装 HTTP 请求体，使其支持重复读取
type readCloser struct {
	Seekabler
	io.Closer
}

// 预定义的错误类型
var (
	ErrNoBody       = errors.New("no body")        // 请求体为空
	ErrTooLargeBody = errors.New("too large body") // 请求体过大
)

// MaxBodyLength 定义了请求体的最大长度限制（16MB）
var MaxBodyLength int64 = 16 * 1024 * 1024

// New 创建一个新的 SeekableCloser 实例
// 参数:
//   - req: HTTP 请求对象
//
// 返回:
//   - SeekableCloser: 支持重复读取的请求体包装器
//   - error: 可能的错误
func New(req *http.Request) (r SeekableCloser, err error) {
	if req.Body == nil {
		return nil, ErrNoBody
	}
	var ok bool
	if r, ok = req.Body.(SeekableCloser); ok {
		// 请求体已经支持可重复读取，不需要做额外的包装处理
		return
	}
	b, err2 := ReadAll(req)
	if err2 != nil {
		return nil, err2
	}
	r = bytes.NewReader(b)
	req.Body = readCloser{r, req.Body}
	return
}

// readCloser2 结构体用于处理大文件的情况
// 当请求体超过最大限制时，会使用这个结构体来包装原始请求体
type readCloser2 struct {
	io.Reader
	io.Closer
}

// ReadAll 读取 HTTP 请求体的所有内容
// 参数:
//   - req: HTTP 请求对象
//
// 返回:
//   - []byte: 请求体内容
//   - error: 可能的错误
//
// 功能:
//  1. 检查请求体大小是否超过限制
//  2. 根据 ContentLength 采用不同的读取策略
//  3. 对于未知大小的请求体，使用 LimitReader 限制读取大小
func ReadAll(req *http.Request) (b []byte, err error) {
	if req.ContentLength > MaxBodyLength {
		return nil, ErrTooLargeBody
	} else if req.ContentLength > 0 {
		b = make([]byte, int(req.ContentLength))
		_, err = io.ReadFull(req.Body, b)
		return
	} else if req.ContentLength == 0 {
		return nil, ErrNoBody
	}
	b, err = ioutil.ReadAll(io.LimitReader(req.Body, MaxBodyLength+1))
	if int64(len(b)) > MaxBodyLength {
		r := io.MultiReader(bytes.NewReader(b), req.Body)
		req.Body = readCloser2{r, req.Body}
		return nil, ErrTooLargeBody
	}
	return
}

// ---------------------------------------------------
