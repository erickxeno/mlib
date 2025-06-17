package mac

import (
	"encoding/base64"
	"net/http"
)

type AuthType = string

const (
	Base  AuthType = "Base"
	Admin AuthType = "Admin"
)

type Credentials struct {
	SecretKey string `json:"secret_key"`
	AccessKey string `json:"access_key"`
	Type      string `json:"type"` // such as: Base, Admin, etc.
}

// ---------------------------------------------------------------------------------------

type AuthStrategyI interface {
	Authorize(sk []byte, req *http.Request, suInfo string) ([]byte, string, error)
}

type Mac struct {
	AccessKey string
	SecretKey []byte
	Strategy  AuthStrategyI
}

func BuildMac(cfg Credentials) (Mac, error) {
	if cfg.Type == "" {
		return Mac{}, ErrMissAuthType
	}
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return Mac{}, ErrMissAkSk
	}

	var strategy AuthStrategyI
	switch cfg.Type {
	case Base:
		strategy = AuthStrategy{}
	case Admin:
		strategy = AdminAuthStrategy{}
	default:
		return Mac{}, ErrUnknownAuthType
	}

	return Mac{
		AccessKey: cfg.AccessKey,
		SecretKey: []byte(cfg.SecretKey),
		Strategy:  strategy,
	}, nil
}

func NewMac(ak, sk string, strategy AuthStrategyI) *Mac {
	return &Mac{
		AccessKey: ak,
		SecretKey: []byte(sk),
		Strategy:  strategy,
	}
}

func (mac *Mac) Auth(req *http.Request) error {
	sign, authType, err := mac.Strategy.Authorize(mac.SecretKey, req, "")
	if err != nil {
		return err
	}

	auth := authType + " " + mac.AccessKey + ":" + base64.URLEncoding.EncodeToString(sign)
	req.Header.Set("Authorization", auth)
	return nil
}

func (mac *Mac) AdminAuth(req *http.Request, suInfo string) error {
	sign, authType, err := mac.Strategy.Authorize(mac.SecretKey, req, suInfo)
	if err != nil {
		return err
	}

	auth := authType + ":" + mac.AccessKey + ":" + base64.URLEncoding.EncodeToString(sign)
	req.Header.Set("Authorization", auth)
	return nil
}

// ---------------- Mac Transport ----------------
type Transport struct {
	mac       Mac
	Transport http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	err = t.mac.Auth(req)
	if err != nil {
		return
	}
	return t.Transport.RoundTrip(req)
}

func (t *Transport) NestedObject() interface{} {
	return t.Transport
}

func NewTransport(mac Mac, transport http.RoundTripper) *Transport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	t := &Transport{Transport: transport, mac: mac}
	return t
}

func NewClient(mac Mac, transport http.RoundTripper) *http.Client {
	t := NewTransport(mac, transport)
	return &http.Client{Transport: t}
}

// --------------------- Admin Transport ---------------------

type AdminTransport struct {
	Transport
	suInfo string
}

func (t *AdminTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	err = t.mac.AdminAuth(req, t.suInfo)
	if err != nil {
		return
	}
	return t.Transport.RoundTrip(req)
}

func NewAdminTransport(mac Mac, suInfo string, transport http.RoundTripper) *AdminTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	base := Transport{Transport: transport, mac: mac}
	return &AdminTransport{Transport: base, suInfo: suInfo}
}

// ---------------------------------------------------------------------------------------
