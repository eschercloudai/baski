package util

import (
	"github.com/drewbernetes/baski/pkg/server/generated"
	"io"
	"net/http"
	"os"
)

//go:generate mockgen -source=interfaces.go -destination=../mock/interfaces.go -package=mock

type HandlerInterface interface {
	Healthz(w http.ResponseWriter, r *http.Request)
	ApiV1GetScans(w http.ResponseWriter, r *http.Request)
	ApiV1GetScan(w http.ResponseWriter, r *http.Request, imageId generated.ImageID)
	ApiV1GetTests(w http.ResponseWriter, r *http.Request)
	ApiV1GetTest(w http.ResponseWriter, r *http.Request, imageId generated.ImageID)
}
type S3Interface interface {
	List() ([]string, error)
	Fetch(string) ([]byte, error)
	Put(string, string, io.ReadSeeker) error
}

type VaultInterface interface {
	Fetch(mountPath, secretPath, data string) ([]byte, error)
}

type SSHInterface interface {
	CopyFromRemoteServer(srcPath, dstPath, filename string) (*os.File, error)
	SSHClose() error
	SFTPClose() error
}
