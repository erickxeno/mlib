package ext

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"

	errs "github.com/erickxeno/mlib/errors"
)

// 常量定义
const (
	chunkBits = 16                   // 64K 块大小（以位为单位）
	chunkLen  = (1 << chunkBits) - 4 // 每个块的数据长度（64K - 4字节）
)

const (
	BufSize = chunkLen + 4 // 缓冲区大小（包含4字节CRC32校验码）
)

// 预定义错误
var (
	ErrUnmatchedChecksum = errors.New("unmatched checksum")      // CRC32校验不匹配
	ErrClosed            = errors.New("has already been closed") // 已关闭
)

// -----------------------------------------

// EncodeSize 计算编码后的大小
// 参数:
//   - fsize: 原始文件大小
//
// 返回:
//   - int64: 编码后的大小（包含CRC32校验码）
func EncodeSize(fsize int64) int64 {
	chunkCount := (fsize + (chunkLen - 1)) / chunkLen
	return fsize + 4*chunkCount
}

// DecodeSize 计算解码后的大小
// 参数:
//   - totalSize: 编码后的总大小
//
// 返回:
//   - int64: 解码后的原始大小
func DecodeSize(totalSize int64) int64 {
	chunkCount := (totalSize + (BufSize - 1)) / BufSize
	return totalSize - 4*chunkCount
}

// -----------------------------------------

// ReaderError 包装读取错误
type ReaderError struct {
	error
}

// WriterError 包装写入错误
type WriterError struct {
	error
}

// Encode 对数据进行CRC32编码
// 参数:
//   - w: 输出写入器
//   - in: 输入读取器
//   - fsize: 文件大小
//   - chunk: 缓冲区（可选）
//
// 返回:
//   - error: 可能的错误
func Encode(w io.Writer, in io.Reader, fsize int64, chunk []byte) (err error) {
	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Encode failed: invalid len(chunk)")
	}

	i := 0
	for fsize >= chunkLen {
		_, err = io.ReadFull(in, chunk[4:])
		if err != nil {
			return ReaderError{err}
		}
		crc := crc32.ChecksumIEEE(chunk[4:])
		binary.LittleEndian.PutUint32(chunk, crc)
		_, err = w.Write(chunk)
		if err != nil {
			return WriterError{err}
		}
		fsize -= chunkLen
		i++
	}

	if fsize > 0 {
		n := fsize + 4
		_, err = io.ReadFull(in, chunk[4:n])
		if err != nil {
			return ReaderError{err}
		}
		crc := crc32.ChecksumIEEE(chunk[4:n])
		binary.LittleEndian.PutUint32(chunk, crc)
		_, err = w.Write(chunk[:n])
		if err != nil {
			err = WriterError{err}
		}
	}
	return
}

// ---------------------------------------------------------------------------

// ReaderWriterAt 接口组合了 io.ReaderAt 和 io.WriterAt
type ReaderWriterAt interface {
	io.ReaderAt
	io.WriterAt
}

// AppendEncode 向已有的CRC32编码文件追加数据
// 参数:
//   - rw: 支持随机读写的接口
//   - base: 为所有批次写入前的 Writer 基础偏移量（CRC32编码文件在底层存储中的起始偏移量）
//   - fsize: 已写入的原始文件数据大小（不包含CRC32校验码，注意：不是待写入的文件大小）
//   - in: 输入读取器
//   - size: 要追加的数据大小
//   - chunk: 缓冲区（可选）
//
// 返回:
//   - error: 可能的错误
func AppendEncode(rw ReaderWriterAt, base int64, fsize int64, in io.Reader, size int64, chunk []byte) (err error) {
	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Encode failed: invalid len(chunk)")
	}

	offset := base + EncodeSize(fsize)
	if oldSize := fsize % chunkLen; oldSize > 0 {
		// 旧文件的最后一个 chunk 需要特殊处理。
		// 处理流程为：读取旧内容、写入新内容、写入总 crc32。
		r := RangeDecoder(rw, base, chunk, fsize-oldSize, fsize, fsize)
		_, err = io.ReadFull(r, chunk[4:4+oldSize])
		if err != nil {
			// 从 rw 读失败，认为是 writer 错误。
			return WriterError{err}
		}
		addSize := chunkLen - oldSize
		if addSize > size {
			addSize = size
		}
		add := chunk[4+oldSize : 4+oldSize+addSize]
		_, err = io.ReadFull(in, add)
		if err != nil {
			return ReaderError{err}
		}
		// 如果 header 写成功但 data 写失败，这个 chunk 就无法正常读写了。
		// 因此下面的操作是先写 data 再写 header。
		_, err = rw.WriteAt(add, offset)
		if err != nil {
			return WriterError{err}
		}
		crc := crc32.ChecksumIEEE(chunk[4 : 4+oldSize+addSize])
		pos := base + (fsize/chunkLen)<<chunkBits
		defer func() {
			if err != nil {
				return
			}
			binary.LittleEndian.PutUint32(chunk[:4], crc)
			_, err = rw.WriteAt(chunk[:4], pos)
			if err != nil {
				err = WriterError{err}
			}
		}()
		size -= addSize
		offset += addSize
	}
	if size == 0 {
		return nil
	}
	w := &Writer{
		WriterAt: rw,
		Offset:   offset,
	}
	return Encode(w, in, size, chunk)
}

// ---------------------------------------------------------------------------

// simpleEncoder 是一个简单的CRC32编码器
// 实现了 io.Reader 接口，用于读取并编码数据
type simpleEncoder struct {
	chunk []byte    // 缓冲区，用于存储当前块的数据
	in    io.Reader // 输入读取器
	off   int       // 当前读取位置
}

// SimpleEncoder 创建一个新的简单编码器
// 参数:
//   - in: 输入读取器
//   - chunk: 缓冲区（可选）
//
// 返回:
//   - *simpleEncoder: 新创建的编码器
func SimpleEncoder(in io.Reader, chunk []byte) (enc *simpleEncoder) {
	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Decoder failed: invalid len(chunk)")
	}

	enc = &simpleEncoder{chunk, in, BufSize}
	return
}

// Read 实现了 io.Reader 接口
// 读取并编码数据，如果遇到错误的块会抛弃之前的数据
func (r *simpleEncoder) Read(b []byte) (n int, err error) {
	if r.off == len(r.chunk) {
		err = r.fetch()
		if err != nil {
			return
		}
	}

	n = copy(b, r.chunk[r.off:])
	r.off += n
	return
}

// fetch 获取下一个数据块并进行CRC32编码
func (r *simpleEncoder) fetch() (err error) {
	var n int
	n, err = ReadSize(r.in, r.chunk[4:])
	if err != nil {
		return
	}

	crc := crc32.ChecksumIEEE(r.chunk[4 : n+4])
	binary.LittleEndian.PutUint32(r.chunk, crc)
	r.off = 0
	r.chunk = r.chunk[:n+4]
	return
}

// ReadSize 读取指定大小的数据
// 参数:
//   - r: 输入读取器
//   - buf: 缓冲区
//
// 返回:
//   - int: 实际读取的字节数
//   - error: 可能的错误
func ReadSize(r io.Reader, buf []byte) (n int, err error) {
	size := len(buf)
	for n < size && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])
		n += nn
	}
	if err == io.EOF && n != 0 {
		err = nil
	}
	return
}

// simpleDecoder 是一个简单的CRC32解码器
// 实现了 io.Reader 接口，用于读取并解码数据
type simpleDecoder struct {
	chunk []byte    // 缓冲区，用于存储当前块的数据
	in    io.Reader // 输入读取器
	off   int       // 当前读取位置
}

// SimpleDecoder 创建一个新的简单解码器
// 参数:
//   - in: 输入读取器
//   - chunk: 缓冲区（可选）
//
// 返回:
//   - *simpleDecoder: 新创建的解码器
func SimpleDecoder(in io.Reader, chunk []byte) (dec *simpleDecoder) {
	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Decoder failed: invalid len(chunk)")
	}

	dec = &simpleDecoder{chunk, in, BufSize}
	return
}

// Read 实现了 io.Reader 接口
// 读取并解码数据
func (r *simpleDecoder) Read(b []byte) (n int, err error) {
	if r.off == len(r.chunk) {
		err = r.fetch()
		if err != nil {
			return
		}
	}

	n = copy(b, r.chunk[r.off:])
	r.off += n
	return
}

// fetch 获取下一个数据块并进行CRC32校验
func (r *simpleDecoder) fetch() (err error) {
	var n int
	n, err = ReadSize(r.in, r.chunk)
	if err != nil {
		return
	}

	if n <= 4 {
		return errs.WrapWithMsg(ErrUnmatchedChecksum, "crc32util.decode")
	}
	crc := crc32.ChecksumIEEE(r.chunk[4:n])
	if binary.LittleEndian.Uint32(r.chunk) != crc {
		return errs.WrapWithMsg(ErrUnmatchedChecksum, "crc32util.decode")
	}
	r.chunk = r.chunk[:n]
	r.off = 4
	return
}

// encodeWriter 是一个CRC32编码写入器
// 实现了 io.WriteCloser 接口，用于写入并编码数据
type encodeWriter struct {
	w     io.Writer // 底层写入器
	chunk []byte    // 缓冲区，用于保存每次写入后剩余的数据
	off   int       // 当前写入位置
}

// NewEncodeWriteCloser 创建一个新的编码写入器
// 参数:
//   - w: 底层写入器
//
// 返回:
//   - *encodeWriter: 新创建的编码写入器
func NewEncodeWriteCloser(w io.Writer) (ewc *encodeWriter) {
	chunk := make([]byte, BufSize)
	ewc = &encodeWriter{
		w:     w,
		chunk: chunk,
		off:   4,
	}
	return
}

// Write 实现了 io.Writer 接口
// 写入数据并进行CRC32编码
func (w *encodeWriter) Write(p []byte) (n int, err error) {
	offset := w.off
	size := (offset - 4) + len(p)
	var pfrom int = 0
	var pto int
	for size >= chunkLen {
		pto = BufSize - offset + pfrom
		copy(w.chunk[offset:], p[pfrom:pto])
		crc := crc32.ChecksumIEEE(w.chunk[4:])
		binary.LittleEndian.PutUint32(w.chunk, crc)
		_, err = w.w.Write(w.chunk)
		if err != nil {
			return 0, err
		}

		// 处理已经写完的数据
		offset = 4
		pfrom = pto
		size -= chunkLen
	}

	n1 := copy(w.chunk[offset:], p[pfrom:])
	w.off = offset + n1
	return len(p), nil
}

// CloseWithError 带错误关闭写入器
func (w *encodeWriter) CloseWithError(err error) error {
	if err == nil {
		return w.Close()
	}
	return nil
}

// Close 实现了 io.Closer 接口
// 关闭写入器并写出最后的数据
func (w *encodeWriter) Close() (err error) {
	if w.off > 4 {
		// 将缓冲中的数据写出
		crc := crc32.ChecksumIEEE(w.chunk[4:w.off])
		binary.LittleEndian.PutUint32(w.chunk, crc)
		_, err = w.w.Write(w.chunk[:w.off])
		if err != nil {
			return
		}
		w.off = 4
	}
	return nil
}

// decoder 是一个CRC32解码器
// 实现了 io.Reader 接口，用于读取并解码数据
type decoder struct {
	chunk   []byte    // 缓冲区，用于存储当前块的数据
	in      io.Reader // 输入读取器
	lastErr error     // 最后一次错误
	off     int       // 当前读取位置
	left    int64     // 剩余数据大小
}

// Decoder 创建一个新的解码器
// 参数:
//   - in: 输入读取器
//   - n: 原始文件大小
//   - chunk: 缓冲区（可选）
//
// 返回:
//   - *decoder: 新创建的解码器
func Decoder(in io.Reader, n int64, chunk []byte) (dec *decoder) {
	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Decoder failed: invalid len(chunk)")
	}

	dec = &decoder{chunk, in, nil, BufSize, n}
	return
}

// fetch 获取下一个数据块并进行CRC32校验
func (r *decoder) fetch() {
	min := len(r.chunk)
	if r.left+4 < int64(min) {
		min = int(r.left + 4)
	}
	var n2 int
	n2, r.lastErr = io.ReadAtLeast(r.in, r.chunk, min)
	if r.lastErr != nil {
		if r.lastErr == io.EOF {
			r.lastErr = io.ErrUnexpectedEOF
		}
		return
	}
	crc := crc32.ChecksumIEEE(r.chunk[4:n2])
	if binary.LittleEndian.Uint32(r.chunk) != crc {
		r.lastErr = errs.WrapWithMsg(ErrUnmatchedChecksum, "crc32util.decode")
		return
	}
	r.chunk = r.chunk[:n2]
	r.off = 4
}

// Read 实现了 io.Reader 接口
// 读取并解码数据
func (r *decoder) Read(b []byte) (n int, err error) {
	if r.off == len(r.chunk) {
		if r.lastErr != nil {
			err = r.lastErr
			return
		}
		if r.left == 0 {
			err = io.EOF
			return
		}
		r.fetch()
	}
	n = copy(b, r.chunk[r.off:])
	r.off += n
	r.left -= int64(n)
	return
}

// ---------------------------------------------------------------------------

// RangeDecoder 创建一个范围解码器
// 用于读取指定范围的数据
func RangeDecoder(in io.ReaderAt, base int64, chunk []byte, from, to, fsize int64) io.Reader {
	fromBase := (from / chunkLen) << chunkBits
	encodedSize := EncodeSize(fsize) - fromBase
	sect := io.NewSectionReader(in, base+fromBase, encodedSize)
	dec := Decoder(sect, DecodeSize(encodedSize), chunk)
	if (from == 0 || from%chunkLen == 0) && to >= fsize {
		return dec
	}
	return newSectionReader(dec, from%chunkLen, to-from)
}

// ---------------------------------------------------------------------------

// decodeAt 解码指定位置的数据
func decodeAt(w io.Writer, in io.ReaderAt, chunk []byte, idx int64, ifrom, ito int) (err error) {
	n, err := in.ReadAt(chunk, idx<<chunkBits)
	if err != nil {
		if err != io.EOF {
			return
		}
	}
	if n <= 4 {
		if n == 0 {
			return io.EOF
		}
		return errs.WrapWithMsg(ErrUnmatchedChecksum, fmt.Sprintf("crc32util.Decode stop n %d", n))
	}

	crc := crc32.ChecksumIEEE(chunk[4:n])
	if binary.LittleEndian.Uint32(chunk) != crc {
		err = errs.WrapWithMsg(ErrUnmatchedChecksum, "crc32util.Decode")
		return
	}

	ifrom += 4
	ito += 4
	if ito > n {
		ito = n
	}
	if ifrom >= ito {
		return io.EOF
	}
	_, err = w.Write(chunk[ifrom:ito])
	return
}

// DecodeRange 解码指定范围的数据
func DecodeRange(w io.Writer, in io.ReaderAt, chunk []byte, from, to int64) (err error) {
	if from >= to {
		return
	}

	if chunk == nil {
		chunk = make([]byte, BufSize)
	} else if len(chunk) != BufSize {
		panic("crc32util.Decode failed: invalid len(chunk)")
	}

	fromIdx, toIdx := from/chunkLen, to/chunkLen
	fromOff, toOff := int(from%chunkLen), int(to%chunkLen)
	if fromIdx == toIdx { // 只有一行
		return decodeAt(w, in, chunk, fromIdx, fromOff, toOff)
	}
	for fromIdx < toIdx {
		err = decodeAt(w, in, chunk, fromIdx, fromOff, chunkLen)
		if err != nil {
			return
		}
		fromIdx++
		fromOff = 0
	}
	if toOff > 0 {
		err = decodeAt(w, in, chunk, fromIdx, 0, toOff)
	}
	return
}

// ---------------------------------------------------------------------------

// newSectionReader 创建一个新的分段读取器
func newSectionReader(r io.Reader, off int64, n int64) io.Reader {
	return &sectionReader{r: r, base: off, n: n}
}

// sectionReader 是一个分段读取器
// 实现了 io.Reader 接口，用于读取指定范围的数据
type sectionReader struct {
	r       io.Reader // 底层读取器
	base    int64     // 基础偏移量
	n       int64     // 要读取的数据大小
	discard bool      // 是否已经丢弃了基础偏移量的数据
}

// Read 实现了 io.Reader 接口
// 读取指定范围的数据
func (s *sectionReader) Read(p []byte) (int, error) {
	if !s.discard {
		_, err := io.CopyN(io.Discard, s.r, s.base)
		if err != nil {
			return 0, err
		}
		s.discard = true
		s.r = io.LimitReader(s.r, s.n)
	}
	return s.r.Read(p)
}
