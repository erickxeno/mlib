package xlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/erickxeno/logs"
)

type Level = logs.Level

func SetCallDepth(depth int) {
	logs.ResetCallDepth()
	logs.AddCallDepth(depth)
}

// Deprecated
func SetOutput(w io.Writer) {
	//logs.SetOutput(w)
}

// Deprecated
func SetFlags(flag int) {
	//logs.SetFlags(flag)
}

func SetOutputLevel(lvl Level) {
	logs.SetLevel(logs.Level(lvl))
}

func init() {
	// Set default call depth offset to 1 for correct line number printing in logs wrapper
	logs.SetDefaultLogCallDepthOffset(1)
	logs.SetDefaultLogPrefixFileDepth(1)
	logs.SetLogPrefixTimePrecision(logs.TimePrecisionMicrosecond)
	logs.SetDefaultLogPrefixWithoutHost(true)
	logs.SetDefaultLogPrefixWithoutPSM(true)
	logs.SetDefaultLogPrefixWithoutCluster(true)
	logs.SetDefaultLogPrefixWithoutStage(true)
	logs.SetDefaultLogPrefixWithoutSpanID(true)
}

// Print calls Output to print to the standard Logger.
// Arguments are handled in the manner of fmt.Print.
func (xlog *Logger) Print(v ...interface{}) {
	if logs.GetLevel() > logs.TraceLevel {
		return
	}
	logs.CtxTrace(xlog.ctx, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard Logger.
// Arguments are handled in the manner of fmt.Printf.
func (xlog *Logger) Printf(format string, v ...interface{}) {
	logs.CtxTrace(xlog.ctx, format, v...)
}

// Println calls Output to print to the standard Logger.
// Arguments are handled in the manner of fmt.Println.
func (xlog *Logger) Println(v ...interface{}) {
	if logs.GetLevel() > logs.TraceLevel {
		return
	}
	logs.CtxTrace(xlog.ctx, fmt.Sprintln(v...))
}

func (xlog *Logger) Tracef(format string, v ...interface{}) {
	logs.CtxTrace(xlog.ctx, format, v...)
}

func (xlog *Logger) Trace(v ...interface{}) {
	if logs.GetLevel() > logs.TraceLevel {
		return
	}
	logs.CtxTrace(xlog.ctx, fmt.Sprint(v...))
}

func (xlog *Logger) TraceStr(content string) {
	if logs.GetLevel() > logs.TraceLevel {
		return
	}
	logs.CtxTrace(xlog.ctx, content)
}

func (xlog *Logger) Debugf(format string, v ...interface{}) {
	logs.CtxDebug(xlog.ctx, format, v...)
}

func (xlog *Logger) Debug(v ...interface{}) {
	if logs.GetLevel() > logs.DebugLevel {
		return
	}
	logs.CtxDebug(xlog.ctx, fmt.Sprint(v...))
}

func (xlog *Logger) DebugStr(content string) {
	if logs.GetLevel() > logs.DebugLevel {
		return
	}
	logs.CtxDebug(xlog.ctx, content)
}

func (xlog *Logger) Infof(format string, v ...interface{}) {
	logs.CtxInfo(xlog.ctx, format, v...)
}

func (xlog *Logger) Info(v ...interface{}) {
	if logs.GetLevel() > logs.InfoLevel {
		return
	}
	logs.CtxInfo(xlog.ctx, fmt.Sprint(v...))
}

func (xlog *Logger) InfoStr(content string) {
	logs.CtxInfo(xlog.ctx, content)
}

func (xlog *Logger) Warnf(format string, v ...interface{}) {
	logs.CtxWarn(xlog.ctx, format, v...)
}

func (xlog *Logger) Warn(v ...interface{}) {
	if logs.GetLevel() > logs.WarnLevel {
		return
	}
	logs.CtxWarn(xlog.ctx, fmt.Sprint(v...))
}

func (xlog *Logger) WarnStr(content string) {
	logs.CtxWarn(xlog.ctx, content)
}

func (xlog *Logger) Errorf(format string, v ...interface{}) {
	logs.CtxError(xlog.ctx, format, v...)
}

func (xlog *Logger) Error(v ...interface{}) {
	if logs.GetLevel() > logs.ErrorLevel {
		return
	}
	logs.CtxError(xlog.ctx, fmt.Sprint(v...))
}

func (xlog *Logger) ErrorStr(content string) {
	logs.CtxError(xlog.ctx, content)
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func (xlog *Logger) Fatal(v ...interface{}) {
	ctx := logs.CtxStackInfo(xlog.ctx, logs.CurrGoroutine)
	logs.CtxFatal(ctx, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func (xlog *Logger) Fatalf(format string, v ...interface{}) {
	ctx := logs.CtxStackInfo(xlog.ctx, logs.CurrGoroutine)
	logs.CtxFatal(ctx, format, v...)
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func (xlog *Logger) Fatalln(v ...interface{}) {
	ctx := logs.CtxStackInfo(xlog.ctx, logs.CurrGoroutine)
	logs.CtxFatal(ctx, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func (xlog *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	ctx := logs.CtxStackInfo(xlog.ctx, logs.CurrGoroutine)
	logs.CtxError(ctx, "panic: "+s)
	logs.Flush()
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func (xlog *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	ctx := logs.CtxStackInfo(xlog.ctx, logs.CurrGoroutine)
	logs.CtxError(ctx, "panic: "+s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func (xlog *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	ctx := logs.CtxStackInfo(xlog.ctx, logs.CurrGoroutine)
	logs.CtxError(ctx, "panic: "+s)
	panic(s)
}

func (xlog *Logger) Stack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, true)
	s += string(buf[:n])
	s += "\n"
	logs.CtxError(xlog.ctx, s)
}

func (xlog *Logger) SingleStack(v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, false)
	s += string(buf[:n])
	s += "\n"
	logs.CtxError(xlog.ctx, s)
}

func Tracef(reqId string, format string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxTrace(ctx, format, v...)
}

func Trace(reqId string, v ...interface{}) {
	if logs.GetLevel() > logs.TraceLevel {
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxTrace(ctx, fmt.Sprintln(v...))
}

func Debugf(reqId string, format string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxDebug(ctx, format, v...)
}

func Debug(reqId string, v ...interface{}) {
	if logs.GetLevel() > logs.DebugLevel {
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxDebug(ctx, fmt.Sprintln(v...))
}

func Infof(reqId string, format string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxInfo(ctx, format, v...)
}

func Info(reqId string, v ...interface{}) {
	if logs.GetLevel() > logs.InfoLevel {
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxInfo(ctx, fmt.Sprintln(v...))
}

func Warnf(reqId string, format string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxWarn(ctx, format, v...)
}

func Warn(reqId string, v ...interface{}) {
	if logs.GetLevel() > logs.WarnLevel {
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxWarn(ctx, fmt.Sprintln(v...))
}

func Errorf(reqId string, format string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxError(ctx, format, v...)
}

func Error(reqId string, v ...interface{}) {
	if logs.GetLevel() > logs.ErrorLevel {
		return
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxError(ctx, fmt.Sprintln(v...))
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(reqId string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxFatal(ctx, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(reqId string, format string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxFatal(ctx, format, v...)
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(reqId string, v ...interface{}) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxFatal(ctx, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(reqId string, v ...interface{}) {
	s := fmt.Sprint(v...)
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	ctx = logs.CtxStackInfo(ctx, logs.CurrGoroutine)
	logs.CtxError(ctx, "panic: "+s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(reqId string, format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	ctx = logs.CtxStackInfo(ctx, logs.CurrGoroutine)
	logs.CtxError(ctx, "panic: "+s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(reqId string, v ...interface{}) {
	s := fmt.Sprintln(v...)
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	ctx = logs.CtxStackInfo(ctx, logs.CurrGoroutine)
	logs.CtxError(ctx, "panic: "+s)
	panic(s)
}

func Stack(reqId string, v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, true)
	s += string(buf[:n])
	s += "\n"
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxError(ctx, s)
}

func SingleStack(reqId string, v ...interface{}) {
	s := fmt.Sprint(v...)
	s += "\n"
	buf := make([]byte, 1024*1024)
	n := runtime.Stack(buf, false)
	s += string(buf[:n])
	s += "\n"
	ctx := context.Background()
	ctx = context.WithValue(ctx, logs.LogIDCtxKey, reqId)
	logs.CtxError(ctx, s)
}
