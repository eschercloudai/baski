package util

import (
	"github.com/eschercloudai/baski/pkg/server/generated"
	"io"
	"net/http"
	"os"
)

//go:generate mockgen -source=interfaces.go -destination=../mock/interfaces.go -package=mock

type HandlerInterface interface {
	Healthz(w http.ResponseWriter, r *http.Request)
	ApiV1GetScan(w http.ResponseWriter, r *http.Request, imageId generated.ImageID)
}
type S3Interface interface {
	FetchFromS3(string) ([]byte, error)
	PutToS3(string, string, string, io.ReadSeeker) error
}

type VaultInterface interface {
	Fetch(mountPath, secretPath, data string) ([]byte, error)
}

type SSHInterface interface {
	CopyFromRemoteServer(srcPath, dstPath, filename string) (*os.File, error)
	SSHClose() error
	SFTPClose() error
}
