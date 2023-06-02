package file

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"go.uber.org/zap"
)

type SeaweedFS struct {
	log.Log
	ctx *config.Context
}

func NewSeaweedFS(ctx *config.Context) *SeaweedFS {
	return &SeaweedFS{
		Log: log.NewTLog("SeaweedFS"),
		ctx: ctx,
	}
}

// UploadFile 上传文件
func (s *SeaweedFS) UploadFile(filePath string, contentType string, copyFileWriter func(io.Writer) error) (map[string]interface{}, error) {
	fileDir, fileName := filepath.Split(filePath)
	s.Debug("filePath->", zap.String("filePath", filePath), zap.String("fileDir", fileDir), zap.String("fileName", fileName))
	newFileDir := fileDir
	if !filepath.IsAbs(fileDir) {
		newFileDir = fmt.Sprintf("/%s", newFileDir)
	}
	resultMap, err := uploadFile(fmt.Sprintf("%s%s", s.ctx.GetConfig().UploadURL, newFileDir), fileName, copyFileWriter)
	return resultMap, err
}

func (s *SeaweedFS) DownloadURL(path string, filename string) (string, error) {

	return fmt.Sprintf("%s%s", s.ctx.GetConfig().FileDownloadURL, path), nil
}
