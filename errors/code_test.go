package errors

import "testing"

var (
	exampleCodes = []Coder{
		//ErrCode{0, 200, "成功"},
		// 2xx 系列
		ErrCode{200, 200, "OK"},
		ErrCode{201, 201, "Created"},
		ErrCode{202, 202, "Accepted"},
		ErrCode{204, 204, "No Content"},

		// 4xx 系列
		ErrCode{400, 400, "Bad Request"},
		ErrCode{401, 401, "Unauthorized"},
		ErrCode{403, 403, "Forbidden"},
		ErrCode{404, 404, "Not Found"},
		ErrCode{405, 405, "Method Not Allowed"},
		ErrCode{408, 408, "Request Timeout"},
		ErrCode{409, 409, "Conflict"},
		ErrCode{413, 413, "Payload Too Large"},
		ErrCode{415, 415, "Unsupported Media Type"},
		ErrCode{429, 429, "Too Many Requests"},

		// 5xx 系列
		ErrCode{500, 500, "Internal Server Error"},
		ErrCode{501, 501, "Not Implemented"},
		ErrCode{502, 502, "Bad Gateway"},
		ErrCode{503, 503, "Service Unavailable"},
		ErrCode{504, 504, "Gateway Timeout"},

		// 业务错误码
		ErrCode{1000, 400, "参数错误"},
		ErrCode{1001, 400, "参数验证失败"},
		ErrCode{1002, 400, "参数类型错误"},
		ErrCode{1003, 400, "参数缺失"},

		ErrCode{2000, 401, "未授权访问"},
		ErrCode{2001, 401, "token已过期"},
		ErrCode{2002, 401, "token无效"},

		ErrCode{3000, 403, "权限不足"},
		ErrCode{3001, 403, "资源访问受限"},

		ErrCode{4000, 404, "资源不存在"},
		ErrCode{4001, 404, "接口不存在"},

		ErrCode{5000, 500, "服务器内部错误"},
		ErrCode{5001, 500, "数据库操作失败"},
		ErrCode{5002, 500, "缓存操作失败"},
		ErrCode{5003, 500, "第三方服务调用失败"},
		ErrCode{5004, 500, "系统繁忙，请稍后重试"},
	}
)

func init() {
	for _, code := range exampleCodes {
		if code.Code() == 0 {
			continue
		}
		MustRegister(code)
	}
}

func TestErrCode(t *testing.T) {
	err := New("test")
	err = WrapC(1001, err)
	t.Logf("%+v", err)
}
