# 错误处理模块

本模块提供了一个完整的错误处理解决方案，支持错误包装、错误码管理和格式化输出。

## 主要特性

- 支持错误包装和链式传递
- 内置错误码系统
- 支持 JSON 格式输出
- 支持错误堆栈跟踪

## 错误类型

### 1. 基础错误 (fundamental)
最基本的错误类型，包含错误消息。

```go
err := errors.New("基础错误")
err2 := errors.Errorf("发生错误：%s"，"错误描述")
```

### 2. 带消息的错误 (withMessage)
包装现有错误并添加额外消息。

```go
err := errors.WrapM(originalErr, "额外信息")
// 或使用格式化字符串
err := errors.WrapMF(originalErr, "错误：%s", "详细信息")
```

### 3. 带错误码的错误 (withCode)
包装错误并关联错误码。

注意：需要提前将错误码使用 `Register` 或 `MustRegister` 注册

```go
RegisterErrorCode(100, 200, "发生 100 错误")

err := errors.WrapC(originalErr, 100)
// 或使用格式化字符串
err := errors.WrapCF(1001, "错误：%s", "详细信息")
```

## 错误码系统

错误码系统通过 `code.go` 实现，支持：

- 错误码定义
- 错误码说明
- 错误码国际化
- 错误码分类

### 错误码定义

```go
const (
    ErrCodeSuccess = 0
    ErrCodeUnknown = 1000
    // ... 其他错误码
)
```

### 错误码说明

```go
type ErrCode struct {
	// ErrCode refers to the integer code of the ErrCode.
	ErrCode int

	// HTTPCode status that should be used for the associated error code.
	HTTPCode int

	// External (user) facing error text.
	Msg string
}
```
错误码需要系统初始化时使用 `Register` 或 `MustRegister` 进行注册。可参考`code_test.go` 中定义

#### 函数 `ParseCoder` 用于从错误中解析出错误码信息。

```go
// 获取错误码信息
if coder := errors.ParseCoder(err); coder != nil {
    code := coder.Code()           // 获取错误码
    httpStatus := coder.HTTPStatus() // 获取对应的 HTTP 状态码
    message := coder.String()      // 获取错误消息
}
```

#### 函数 `IsCode` 用于检查错误是否包含特定的错误码。

```go
// 检查错误码
if errors.IsCode(err, 1001) {
    // 处理特定错误码的情况
}
```

## 格式化输出

支持多种格式化输出方式，并对多层 Wrap 的消息，支持拆分按序输出（特别是使用`%#v` JSON 格式输出），实现错误堆栈跟踪：
- `%v`：简要信息输出（不包含内部错误原因）
- `%s`：详细错误信息输出
- `%+v`：详细错误信息输出
- `%#v`：以 JSON 格式输出

### 1. 标准输出
```go
fmt.Print(err)         // 简要输出
fmt.Printf("%v", err)  // 简要输出
fmt.Printf("%s", err)  // 详细输出（包含堆栈），与 %+v 效果相同
fmt.Printf("%+v", err) // 详细输出（包含堆栈）
```

### 2. JSON 输出
```go
fmt.Printf("%#v", err) // JSON 格式输出
```

## 使用示例

### 1. 创建基础错误
```go
err := New("发生错误")
// 添加第一层包装信息
err1 := WrapWithMsg(err, "第一层包装")
// 添加第二层包装信息
err2 := WrapWithMsg(err1, "第二层包装")
// 添加第三层包装信息
err3 := WrapWithMsg(err2, "第三层包装")
// 添加第四层 ErrorCode 包装信息
err4 := WrapWithCode(1001, err3)

fmt.Println(err4)           // 简要信息
t.Log("raw:", err4)         // 简要信息
t.Logf("%%s: %s", err4)     // 详细输出（包含堆栈），与 %+v 效果相同
t.Logf("%%v: %v", err4)     // 简要信息
t.Logf("%%+v: %+v", err4)   // 详细输出（包含堆栈）
t.Logf("%%#v: %#v", err4)   // 以 JSON 字符串数组输出（包含堆栈）

// 获取错误原因
cause := errors.Cause(err)
```

## 注意事项

1. 错误码系统支持国际化，可以通过修改 `code.go` 中的 `message` 字段实现
2. 格式化输出支持自定义分隔符，默认为逗号
3. 错误堆栈跟踪仅在详细输出模式下可用
4. JSON 输出模式下会包含完整的错误信息 