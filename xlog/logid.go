package xlog

import (
	"encoding/base64"
	"encoding/binary"
	"os"
	"strconv"
	"time"

	"github.com/erickxeno/logs/logid"
)

var pid = uint32(os.Getpid())

var genReqId = defaultGenReqId

func defaultGenReqId() string {
	var b [12]byte
	binary.LittleEndian.PutUint32(b[:], pid)
	binary.LittleEndian.PutUint64(b[4:], uint64(time.Now().UnixNano()))
	return base64.URLEncoding.EncodeToString(b[:])
}

func GenReqId() string {

	return genReqId()
}

func SetGenReqId(f func() string) {
	if f == nil {
		f = defaultGenReqId
	}
	genReqId = f
}

func GenReqIdV2() string {
	return strconv.FormatUint(logid.GetID(), 10)
}
