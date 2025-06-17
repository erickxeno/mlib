package errors

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("测试错误")
	if err.Error() != "测试错误" {
		t.Errorf("New() = %v, want %v", err.Error(), "测试错误")
	}
}

func TestErrorf(t *testing.T) {
	err := Errorf("测试错误: %s", "详情")
	if err.Error() != "测试错误: 详情" {
		t.Errorf("Errorf() = %v, want %v", err.Error(), "测试错误: 详情")
	}
}

func TestWrapM(t *testing.T) {
	originalErr := New("原始错误")
	err := WrapWithMsg(originalErr, "包装错误")

	t.Log("raw:", err)
	t.Logf("%%s: %s", err)
	t.Logf("%%v: %v", err)
	t.Logf("%%+v: %+v", err)
	t.Logf("%%#v: %#v", err)
}

func TestWrapC(t *testing.T) {

	t.Run("WrapWithCode", func(t *testing.T) {
		err := WrapWithCode(1001, fmt.Errorf("包装信息"))
		RegisterErrorCode(1001, 401, "错误1001")

		// 这里为了使用的方便，修改了 format 函数，修改了 format 函数，
		t.Log("raw:", err)
		t.Logf("%%s: %s", err)
		t.Logf("%%v: %v", err)
		t.Logf("%%+v: %+v", err)
		t.Logf("%%#v: %#v", err)

		err2 := WrapWithMsg(err, "包装信息2")
		t.Log("raw:", err2)
		t.Logf("%%s: %s", err2)
		t.Logf("%%v: %v", err2)
		t.Logf("%%+v: %+v", err2)
		t.Logf("%%#v: %#v", err2)

		// 测试错误消息
		if err.Error() != "1001:错误1001,包装信息" {
			t.Errorf("WithCode().Error() = %v, want %v", err.Error(), "1001:错误1001,包装信息")
		}

		// 测试错误码
		if w, ok := err.(*withCode); !ok {
			t.Error("WithCode() should return *withCode")
		} else if w.code != 1001 {
			t.Errorf("WithCode() code = %v, want %v", w.code, 1001)
		}
	})

	t.Run("WrapC", func(t *testing.T) {
		originalErr := New("原始错误")
		err := WrapC(1001, originalErr)
		t.Log("raw:", err)
		t.Logf("%%s: %s", err)
		t.Logf("%%v: %v", err)
		t.Logf("%%+v: %+v", err)
		t.Logf("%%#v: %#v", err)

		// 测试错误消息
		if err.Error() != "1001:错误1001,原始错误" {
			t.Errorf("WrapC().Error() = %v, want %v", err.Error(), "1001:错误1001,原始错误")
		}

		// 测试错误码和错误链
		if w, ok := err.(*withCode); !ok {
			t.Error("WrapC() should return *withCode")
		} else {
			if w.code != 1001 {
				t.Errorf("WrapC() code = %v, want %v", w.code, 1001)
			}
			if w.err != originalErr {
				t.Error("WrapC() cause should be original error")
			}
		}
	})
}

func TestErrorFormat(t *testing.T) {
	originalErr := New("原始错误")
	RegisterErrorCode(1001, 401, "错误1001")

	err := WrapC(1001, originalErr)

	// 测试基础格式化
	basic := fmt.Sprintf("%v", err)
	if basic != "1001:错误1001" {
		t.Errorf("basic format = %v, want %v", basic, "包装错误: 详情")
	}

	// 测试详细格式化
	detailed := fmt.Sprintf("%+v", err)
	if detailed == "" {
		t.Error("detailed format should not be empty")
	}

	// 测试 JSON 格式化
	jsonFormat := fmt.Sprintf("%#v", err)
	t.Log("jsonFormat:", jsonFormat)
	var jsonData []map[string]interface{}
	if err := json.Unmarshal([]byte(jsonFormat), &jsonData); err != nil {
		t.Errorf("JSON format should be valid: %v", err)
	}
}

func TestErrorChain(t *testing.T) {
	RegisterErrorCode(1001, 401, "错误1001")
	RegisterErrorCode(1002, 402, "错误1002")

	t.Run("WrapMultiWithMsg", func(t *testing.T) {
		originalErr := New("原始错误")
		err1 := WrapWithMsg(originalErr, "第一层包装")
		err2 := WrapWithMsg(err1, "第二层包装")
		err3 := WrapWithMsg(err2, "第三层包装")
		fmt.Println(err3)
		t.Log("raw:", err3)
		t.Logf("%%s: %s", err3)
		t.Logf("%%v: %v", err3)
		t.Logf("%%+v: %+v", err3)
		t.Logf("%%#v: %#v", err3)
	})
	t.Run("WrapMultiWithMsgOneWithCode", func(t *testing.T) {
		originalErr := New("原始错误")
		err1 := WrapWithMsg(originalErr, "第一层包装")
		err2 := WrapWithMsg(err1, "第二层包装")
		err3 := WrapWithMsg(err2, "第三层包装")
		err4 := WrapWithCode(1001, err3)

		fmt.Println(err4)
		t.Log("raw:", err4)
		t.Logf("%%s: %s", err4)
		t.Logf("%%v: %v", err4)
		t.Logf("%%+v: %+v", err4)
		t.Logf("%%#v: %#v", err4)
	})
	t.Run("WrapMultiType", func(t *testing.T) {
		originalErr := New("原始错误")
		err1 := WrapWithMsg(originalErr, "第一层包装")
		err2 := WrapWithCode(1001, err1)
		err3 := WrapWithMsg(err2, "第三层包装")
		err4 := WrapWithCode(1002, err3)
		fmt.Println(err4)
		t.Log("raw:", err4)
		t.Logf("%%s: %s", err4)
		t.Logf("%%v: %v", err4)
		t.Logf("%%+v: %+v", err4)
		t.Logf("%%#v: %#v", err4)

		// 测试错误链
		if w4, ok := err4.(*withCode); !ok {
			t.Error("err4 should be *withCode")
		} else {
			if w4.err != err3 {
				t.Error("err4 cause should be err3")
			}
			w3, ok := w4.err.(*withMessage)
			if !ok {
				t.Error("err1 should be *withMessage")
			} else if w3.err != err2 {
				t.Error("err3 cause should be err2")
			}
			w2, ok := w3.err.(*withCode)
			if !ok {
				t.Error("err3 should be *withCode")
			} else if w2.err != err1 {
				t.Error("err2 cause should be err1")
			}
			w1, ok := w2.err.(*withMessage)
			if !ok {
				t.Error("err1 should be *withMessage")
			} else if w1.err != originalErr {
				t.Error("err1 cause should be originalErr")
			}
		}
	})
}

func TestNilError(t *testing.T) {
	RegisterErrorCode(1001, 401, "错误1001")
	// 测试 nil 错误包装
	if err := WrapWithMsg(nil, "包装错误"); err != nil {
		t.Error("WrapM(nil) should return nil")
	}

	err := WrapWithCode(1001, nil)
	// if  err != nil {
	// 	t.Error("WrapC(nil) should return nil")
	// }
	t.Logf("%%s: %s", err)
	t.Logf("%%v: %v", err)
	t.Logf("%%+v: %+v", err)
	t.Logf("%%#v: %#v", err)
}
