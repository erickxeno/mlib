package ext

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

// mockReaderWriterAt 实现 ReaderWriterAt 接口用于测试
type mockReaderWriterAt struct {
	data []byte
}

func newMockReaderWriterAt() *mockReaderWriterAt {
	return &mockReaderWriterAt{
		data: make([]byte, 0),
	}
}

func (m *mockReaderWriterAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[off:])
	if n == 0 {
		return 0, io.EOF
	}
	return n, nil
}

func (m *mockReaderWriterAt) WriteAt(p []byte, off int64) (n int, err error) {
	// 扩展数据切片以容纳写入的数据
	if off+int64(len(p)) > int64(len(m.data)) {
		newData := make([]byte, off+int64(len(p)))
		copy(newData, m.data)
		m.data = newData
	}
	copy(m.data[off:], p)
	return len(p), nil
}

// Write 实现 io.Writer 接口，用于 Encode 函数
func (m *mockReaderWriterAt) Write(p []byte) (n int, err error) {
	// 追加到数据末尾
	m.data = append(m.data, p...)
	return len(p), nil
}

func (m *mockReaderWriterAt) GetData() []byte {
	return m.data
}

func TestAppendEncode(t *testing.T) {
	t.Run("test append to empty file", func(t *testing.T) {
		ast := require.New(t)

		// 创建空的模拟存储
		rw := newMockReaderWriterAt()
		base := int64(0)
		fsize := int64(0) // 空文件

		// 准备要追加的数据
		originalData := []byte("Hello, World! This is test data for CRC32 encoding.")
		in := bytes.NewReader(originalData)
		size := int64(len(originalData))

		// 执行追加编码
		err := AppendEncode(rw, base, fsize, in, size, nil)
		ast.NoError(err)

		// 验证编码后的数据大小
		expectedEncodedSize := EncodeSize(size)
		ast.Equal(expectedEncodedSize, int64(len(rw.GetData())))

		// 解码验证
		decodedData, err := decodeAll(rw, base, expectedEncodedSize)
		ast.NoError(err)
		ast.Equal(originalData, decodedData)
	})

	t.Run("test append to file with complete chunks", func(t *testing.T) {
		ast := require.New(t)

		// 创建模拟存储并写入完整块的数据
		rw := newMockReaderWriterAt()
		base := int64(0)

		// 写入一个完整的块（65532字节）
		completeChunkData := make([]byte, chunkLen)
		for i := range completeChunkData {
			completeChunkData[i] = byte(i % 256)
		}

		// 先编码写入完整块
		err := Encode(rw, bytes.NewReader(completeChunkData), int64(len(completeChunkData)), nil)
		ast.NoError(err)

		fsize := int64(len(completeChunkData))

		// 准备要追加的数据
		appendData := []byte("This is appended data.")
		in := bytes.NewReader(appendData)
		size := int64(len(appendData))

		// 执行追加编码，fsize 是已写入的原始文件数据大小（不包含CRC32校验码）
		err = AppendEncode(rw, base, fsize, in, size, nil)
		ast.NoError(err)

		// 验证总大小
		totalOriginalSize := fsize + size
		expectedEncodedSize := EncodeSize(totalOriginalSize)
		ast.Equal(expectedEncodedSize, int64(len(rw.GetData())))

		// 解码验证完整数据
		decodedData, err := decodeAll(rw, base, expectedEncodedSize)
		ast.NoError(err)

		// 验证原始数据 + 追加数据
		expectedData := append(completeChunkData, appendData...)
		ast.Equal(expectedData, decodedData)
	})

	t.Run("test append to file with incomplete chunk (oldSize > 0)", func(t *testing.T) {
		ast := require.New(t)

		// 创建模拟存储并写入不完整块的数据
		rw := newMockReaderWriterAt()
		base := int64(0)

		// 写入不完整的块（比如 1000 字节，小于 chunkLen=65532）
		incompleteChunkData := make([]byte, 1000)
		for i := range incompleteChunkData {
			incompleteChunkData[i] = byte(i % 256)
		}

		// 先编码写入不完整块
		err := Encode(rw, bytes.NewReader(incompleteChunkData), int64(len(incompleteChunkData)), nil)
		ast.NoError(err)

		fsize := int64(len(incompleteChunkData))

		// 验证 oldSize > 0 的条件
		oldSize := fsize % chunkLen
		ast.True(oldSize > 0, "oldSize should be greater than 0 to test this logic")
		ast.Equal(int64(1000), oldSize)

		// 准备要追加的数据（足够填满当前块并开始新块）
		appendData := make([]byte, chunkLen+5000) // 超过一个块的大小
		for i := range appendData {
			appendData[i] = byte((i + 1000) % 256) // 不同的数据模式
		}

		in := bytes.NewReader(appendData)
		size := int64(len(appendData))

		// 执行追加编码
		err = AppendEncode(rw, base, fsize, in, size, nil)
		ast.NoError(err)

		// 验证总大小
		totalOriginalSize := fsize + size
		expectedEncodedSize := EncodeSize(totalOriginalSize)
		ast.Equal(expectedEncodedSize, int64(len(rw.GetData())))

		// 解码验证完整数据
		decodedData, err := decodeAll(rw, base, expectedEncodedSize)
		ast.NoError(err)

		// 验证原始数据 + 追加数据
		expectedData := append(incompleteChunkData, appendData...)
		ast.Equal(expectedData, decodedData)
	})

	t.Run("test append small data to incomplete chunk", func(t *testing.T) {
		ast := require.New(t)

		// 创建模拟存储并写入不完整块的数据
		rw := newMockReaderWriterAt()
		base := int64(0)

		// 写入不完整的块（比如 50000 字节，小于 chunkLen=65532）
		incompleteChunkData := make([]byte, 50000)
		for i := range incompleteChunkData {
			incompleteChunkData[i] = byte(i % 256)
		}

		// 先编码写入不完整块
		err := Encode(rw, bytes.NewReader(incompleteChunkData), int64(len(incompleteChunkData)), nil)
		ast.NoError(err)

		fsize := int64(len(incompleteChunkData))
		oldSize := fsize % chunkLen
		ast.True(oldSize > 0)

		// 准备要追加的小数据（不足以填满当前块）
		appendData := make([]byte, 1000) // 小于剩余空间
		for i := range appendData {
			appendData[i] = byte((i + 50000) % 256)
		}

		in := bytes.NewReader(appendData)
		size := int64(len(appendData))

		// 执行追加编码
		err = AppendEncode(rw, base, fsize, in, size, nil)
		ast.NoError(err)

		// 验证总大小
		totalOriginalSize := fsize + size
		expectedEncodedSize := EncodeSize(totalOriginalSize)
		ast.Equal(expectedEncodedSize, int64(len(rw.GetData())))

		// 解码验证完整数据
		decodedData, err := decodeAll(rw, base, expectedEncodedSize)
		ast.NoError(err)

		// 验证原始数据 + 追加数据
		expectedData := append(incompleteChunkData, appendData...)
		ast.Equal(expectedData, decodedData)
	})

	t.Run("test append with non-zero base offset", func(t *testing.T) {
		ast := require.New(t)

		// 创建模拟存储，从非零位置开始
		rw := newMockReaderWriterAt()
		base := int64(1024) // 从第1024字节开始
		fsize := int64(0)   // 空文件

		// 准备要追加的数据
		originalData := []byte("Test data with non-zero base offset.")
		in := bytes.NewReader(originalData)
		size := int64(len(originalData))

		// 执行追加编码
		err := AppendEncode(rw, base, fsize, in, size, nil)
		ast.NoError(err)

		// 验证编码后的数据位置和大小
		expectedEncodedSize := EncodeSize(size)
		ast.Equal(base+expectedEncodedSize, int64(len(rw.GetData())))

		// 解码验证
		decodedData, err := decodeAll(rw, base, expectedEncodedSize)
		ast.NoError(err)
		ast.Equal(originalData, decodedData)
	})
}

// decodeAll 解码指定范围的所有数据
func decodeAll(rw *mockReaderWriterAt, base int64, encodedSize int64) ([]byte, error) {
	// 使用 RangeDecoder 解码所有数据
	reader := RangeDecoder(rw, base, nil, 0, DecodeSize(encodedSize), DecodeSize(encodedSize))

	// 读取所有解码后的数据
	var result bytes.Buffer
	_, err := io.Copy(&result, reader)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// 测试辅助函数：验证编码和解码的一致性
func TestEncodeDecodeConsistency(t *testing.T) {
	ast := require.New(t)

	// 测试不同大小的数据
	testCases := []struct {
		name string
		size int
	}{
		{"small data", 100},
		{"medium data", 10000},
		{"large data", 100000},
		{"exact chunk size", int(chunkLen)},
		{"multiple chunks", int(chunkLen * 3)},
		{"incomplete chunk", int(chunkLen - 1000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 生成测试数据
			originalData := make([]byte, tc.size)
			for i := range originalData {
				originalData[i] = byte(i % 256)
			}

			// 编码
			var encodedBuffer bytes.Buffer
			err := Encode(&encodedBuffer, bytes.NewReader(originalData), int64(len(originalData)), nil)
			ast.NoError(err)

			// 解码
			decoder := SimpleDecoder(&encodedBuffer, nil)
			var decodedBuffer bytes.Buffer
			_, err = io.Copy(&decodedBuffer, decoder)
			ast.NoError(err)

			// 验证数据一致性
			ast.Equal(originalData, decodedBuffer.Bytes())
		})
	}
}
