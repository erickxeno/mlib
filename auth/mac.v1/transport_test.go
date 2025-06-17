package mac

import (
	"reflect"
	"testing"
)

func TestBuildMac(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Credentials
		want    Mac
		wantErr error
	}{
		{
			name: "有效的Base凭证",
			cfg: Credentials{
				AccessKey: "testAK",
				SecretKey: "testSK",
				Type:      Base,
			},
			want: Mac{
				AccessKey: "testAK",
				SecretKey: []byte("testSK"),
				Strategy:  AuthStrategy{},
			},
			wantErr: nil,
		},
		{
			name: "有效的Admin凭证",
			cfg: Credentials{
				AccessKey: "testAK",
				SecretKey: "testSK",
				Type:      Admin,
			},
			want: Mac{
				AccessKey: "testAK",
				SecretKey: []byte("testSK"),
				Strategy:  AdminAuthStrategy{},
			},
			wantErr: nil,
		},
		{
			name: "缺少认证类型",
			cfg: Credentials{
				AccessKey: "testAK",
				SecretKey: "testSK",
			},
			want:    Mac{},
			wantErr: ErrMissAuthType,
		},
		{
			name: "缺少AccessKey",
			cfg: Credentials{
				SecretKey: "testSK",
				Type:      Base,
			},
			want:    Mac{},
			wantErr: ErrMissAkSk,
		},
		{
			name: "未知的认证类型",
			cfg: Credentials{
				AccessKey: "testAK",
				SecretKey: "testSK",
				Type:      "Unknown",
			},
			want:    Mac{},
			wantErr: ErrUnknownAuthType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BuildMac(tt.cfg)
			if err != tt.wantErr {
				t.Errorf("BuildMac() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BuildMac() = %v, want %v", got, tt.want)
			}
		})
	}
}
