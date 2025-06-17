package bytes

import (
	"io"
	"syscall"
)

// ---------------------------------------------------

// Reader 是一个字节读取器，实现了 io.Reader 接口
// 它提供了对字节切片的读取功能，支持：
// 1. 基本的读取操作
// 2. 随机访问（Seek）
// 3. 获取剩余可读字节
// 4. 重置读取位置
type Reader struct {
	b   []byte // 存储要读取的数据
	off int    // 当前读取位置
}

// NewReader 创建一个新的 Reader 实例
// 参数:
//   - val: 要读取的字节切片
//
// 返回:
//   - *Reader: 新创建的 Reader 实例
func NewReader(val []byte) *Reader {
	return &Reader{val, 0}
}

// Len 返回剩余可读的字节数
func (r *Reader) Len() int {
	if r.off >= len(r.b) {
		return 0
	}
	return len(r.b) - r.off
}

// Bytes 返回从当前读取位置到末尾的所有字节
func (r *Reader) Bytes() []byte {
	return r.b[r.off:]
}

// SeekToBegin 将读取位置重置到开始位置
func (r *Reader) SeekToBegin() (err error) {
	r.off = 0
	return
}

// Seek 实现了 io.Seeker 接口，支持随机访问
// 参数:
//   - offset: 偏移量
//   - whence: 基准位置（0: 开始, 1: 当前位置, 2: 末尾）
func (r *Reader) Seek(offset int64, whence int) (ret int64, err error) {
	switch whence {
	case 0:
	case 1:
		offset += int64(r.off)
	case 2:
		offset += int64(len(r.b))
	default:
		err = syscall.EINVAL
		return
	}
	if offset < 0 {
		err = syscall.EINVAL
		return
	}
	if offset >= int64(len(r.b)) {
		r.off = len(r.b)
	} else {
		r.off = int(offset)
	}
	ret = int64(r.off)
	return
}

// Read 实现了 io.Reader 接口，用于读取数据
func (r *Reader) Read(val []byte) (n int, err error) {
	n = copy(val, r.b[r.off:])
	if n == 0 && len(val) != 0 {
		err = io.EOF
		return
	}
	r.off += n
	return
}

// Close 实现了 io.Closer 接口
func (r *Reader) Close() (err error) {
	return
}

// ---------------------------------------------------

// Writer 是一个字节写入器，实现了 io.Writer 接口
// 它提供了对字节切片的写入功能，支持：
// 1. 基本的写入操作
// 2. 获取已写入的字节数
// 3. 获取已写入的内容
// 4. 重置写入位置
type Writer struct {
	b []byte // 存储要写入的缓冲区
	n int    // 当前写入位置
}

// NewWriter 创建一个新的 Writer 实例
// 参数:
//   - buff: 用于写入的缓冲区
//
// 返回:
//   - *Writer: 新创建的 Writer 实例
func NewWriter(buff []byte) *Writer {
	return &Writer{buff, 0}
}

// Write 实现了 io.Writer 接口，用于写入数据
// 如果写入的字节数小于要写入的数据长度，返回 io.ErrShortWrite 错误
func (p *Writer) Write(val []byte) (n int, err error) {
	n = copy(p.b[p.n:], val)
	if n == 0 && len(val) > 0 {
		err = io.EOF
		return
	}
	if n < len(val) {
		return n, io.ErrShortWrite
	}
	p.n += n
	return
}

// Len 返回已写入的字节数
func (p *Writer) Len() int {
	return p.n
}

// Bytes 返回已写入的所有字节
func (p *Writer) Bytes() []byte {
	return p.b[:p.n]
}

// Reset 重置写入位置到开始位置
func (p *Writer) Reset() {
	p.n = 0
}

// ---------------------------------------------------

// Buffer 是一个字节缓冲区，实现了 io.ReaderAt 和 io.WriterAt 接口
// 它提供了对字节切片的随机读写功能，支持：
// 1. 随机读取（ReadAt）
// 2. 随机写入（WriteAt）
// 3. 字符串写入（WriteStringAt）
// 4. 调整缓冲区大小（Truncate）
type Buffer struct {
	b []byte // 存储数据的缓冲区
}

// NewBuffer 创建一个新的 Buffer 实例
func NewBuffer() *Buffer {
	return new(Buffer)
}

// ReadAt 实现了 io.ReaderAt 接口，支持从指定位置读取数据
func (p *Buffer) ReadAt(buf []byte, off int64) (n int, err error) {
	ioff := int(off)
	if len(p.b) <= ioff {
		return 0, io.EOF
	}
	n = copy(buf, p.b[ioff:])
	if n != len(buf) {
		err = io.EOF
	}
	return
}

// WriteAt 实现了 io.WriterAt 接口，支持向指定位置写入数据
// 如果写入位置超出缓冲区大小，会自动扩展缓冲区
func (p *Buffer) WriteAt(buf []byte, off int64) (n int, err error) {
	ioff := int(off)
	iend := ioff + len(buf)
	if len(p.b) < iend {
		if len(p.b) == ioff {
			p.b = append(p.b, buf...)
			return len(buf), nil
		}
		zero := make([]byte, iend-len(p.b))
		p.b = append(p.b, zero...)
	}
	copy(p.b[ioff:], buf)
	return len(buf), nil
}

// WriteStringAt 支持向指定位置写入字符串
// 如果写入位置超出缓冲区大小，会自动扩展缓冲区
func (p *Buffer) WriteStringAt(buf string, off int64) (n int, err error) {
	ioff := int(off)
	iend := ioff + len(buf)
	if len(p.b) < iend {
		if len(p.b) == ioff {
			p.b = append(p.b, buf...)
			return len(buf), nil
		}
		zero := make([]byte, iend-len(p.b))
		p.b = append(p.b, zero...)
	}
	copy(p.b[ioff:], buf)
	return len(buf), nil
}

// Truncate 调整缓冲区大小
// 如果新大小大于当前大小，会用零字节填充
// 如果新大小小于当前大小，会截断缓冲区
func (p *Buffer) Truncate(fsize int64) (err error) {
	size := int(fsize)
	if len(p.b) < size {
		zero := make([]byte, size-len(p.b))
		p.b = append(p.b, zero...)
	} else {
		p.b = p.b[:size]
	}
	return nil
}

// Buffer 返回底层的字节切片
func (p *Buffer) Buffer() []byte {
	return p.b
}

// Len 返回缓冲区的当前大小
func (p *Buffer) Len() int {
	return len(p.b)
}
