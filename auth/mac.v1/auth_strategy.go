package mac

import (
	"net/http"
)

func checkSk(sk []byte) error {
	if len(sk) == 0 {
		return ErrMissSK
	}
	return nil
}

func checkSuInfo(suInfo string) error {
	if suInfo == "" {
		return ErrMissSuInfo
	}
	return nil
}

type AuthStrategy struct{}

func (s AuthStrategy) Authorize(sk []byte, req *http.Request, _ string) ([]byte, string, error) {
	if err := checkSk(sk); err != nil {
		return nil, "", err
	}

	bs, err := SignRequestWithHeader(sk, req)
	return bs, "Base", err
}

type AdminAuthStrategy struct{}

func (s AdminAuthStrategy) Authorize(sk []byte, req *http.Request, suInfo string) ([]byte, string, error) {
	if err := checkSk(sk); err != nil {
		return nil, "", err
	}
	if err := checkSuInfo(suInfo); err != nil {
		return nil, "", err
	}

	bs, err := SignAdminRequestWithHeader(sk, req, suInfo)
	return bs, "Admin " + suInfo, err
}
