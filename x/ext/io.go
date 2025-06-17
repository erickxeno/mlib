package ext

import (
	"io"

	"github.com/erickxeno/mlib/x/bytes"
)

// ---------------------------------------------------

// ReadWriterAt 接口组合了 io.ReaderAt 和 io.WriterAt 接口
// 提供了对数据的随机读写能力
type ReadWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

// ---------------------------------------------------

// Writer 是一个带有偏移量的写入器
// 它包装了 io.WriterAt 接口，并自动维护写入位置
// 适用于需要顺序写入但底层支持随机写入的场景
type Writer struct {
	io.WriterAt       // 底层的随机写入器
	Offset      int64 // 当前写入位置
}

// Write 实现了 io.Writer 接口
// 在当前位置写入数据，并自动更新偏移量
func (p *Writer) Write(val []byte) (n int, err error) {
	n, err = p.WriteAt(val, p.Offset)
	p.Offset += int64(n)
	return
}

// ---------------------------------------------------

// Reader 是一个带有偏移量的读取器
// 它包装了 io.ReaderAt 接口，并自动维护读取位置
// 适用于需要顺序读取但底层支持随机读取的场景
type Reader struct {
	io.ReaderAt       // 底层的随机读取器
	Offset      int64 // 当前读取位置
}

// Read 实现了 io.Reader 接口
// 从当前位置读取数据，并自动更新偏移量
func (p *Reader) Read(val []byte) (n int, err error) {
	n, err = p.ReadAt(val, p.Offset)
	p.Offset += int64(n)
	return
}

// ---------------------------------------------------

// NilReader 是一个空读取器
// 它总是返回 EOF 错误，用于表示没有数据可读
type NilReader struct{}

// NilWriter 是一个空写入器
// 它总是成功写入所有数据，但实际上不执行任何操作
type NilWriter struct{}

// Read 实现了 io.Reader 接口
// 总是返回 EOF 错误
func (r NilReader) Read(val []byte) (n int, err error) {
	return 0, io.EOF
}

// Write 实现了 io.Writer 接口
// 总是返回成功，但不实际写入数据
func (r NilWriter) Write(val []byte) (n int, err error) {
	return len(val), nil
}

// ---------------------------------------------------

// NewBytesReader 创建一个新的字节读取器
// 参数:
//   - val: 要读取的字节切片
//
// 返回:
//   - *bytes.Reader: 新创建的读取器
func NewBytesReader(val []byte) *bytes.Reader {
	return bytes.NewReader(val)
}

// ---------------------------------------------------

// NewBytesWriter 创建一个新的字节写入器
// 参数:
//   - buff: 用于写入的缓冲区
//
// 返回:
//   - *bytes.Writer: 新创建的写入器
func NewBytesWriter(buff []byte) *bytes.Writer {
	return bytes.NewWriter(buff)
}

// ---------------------------------------------------

// optimisticMultiWriter 是一个乐观的多写入器
// 它尝试向多个写入器写入数据，即使某些写入器失败也会继续
// 只有当所有写入器都失败时才返回错误
type optimisticMultiWriter struct {
	writers []io.Writer // 目标写入器列表
	errs    []error     // 每个写入器的错误状态
	fail    int         // 失败的写入器数量
}

// OptimisticMultiWriter 创建一个新的乐观多写入器
// 参数:
//   - writers: 要写入的目标写入器列表
//
// 返回:
//   - *optimisticMultiWriter: 新创建的多写入器
func OptimisticMultiWriter(writers ...io.Writer) *optimisticMultiWriter {
	return &optimisticMultiWriter{
		writers: writers,
		errs:    make([]error, len(writers)),
		fail:    0,
	}
}

// Write 实现了 io.Writer 接口
// 尝试向所有写入器写入数据，即使某些写入器失败也会继续
// 只有当所有写入器都失败时才返回错误
func (t *optimisticMultiWriter) Write(p []byte) (n int, err error) {
	for i, w := range t.writers {
		if t.errs[i] != nil {
			continue
		}
		_, err1 := w.Write(p)
		if err1 != nil {
			t.fail++
			t.errs[i] = err1
		}
	}

	if t.fail == len(t.writers) {
		return 0, io.ErrShortWrite
	}

	return len(p), nil
}

// Errors 返回所有写入器的错误状态
func (t *optimisticMultiWriter) Errors() []error {
	return t.errs
}

// ---------------------------------------------------
