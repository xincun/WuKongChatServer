package file

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/WuKongIM/WuKongChatServer/internal/config"
	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	limlog "github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/util"
	"github.com/nfnt/resize"
	"go.uber.org/zap"
)

type IUploadService interface {
	UploadFile(filePath string, contentType string, copyFileWriter func(io.Writer) error) (map[string]interface{}, error)
	// 获取下载地址
	DownloadURL(path string, filename string) (string, error)
}

// IService IService
type IService interface {
	IUploadService
	DownloadAndMakeCompose(uploadPath string, downloadURLs []string) (map[string]interface{}, error)
	DownloadImage(url string) (*os.File, error)
}

// NewService NewService
func NewService(ctx *config.Context) IService {
	var uploadService IUploadService
	if ctx.GetConfig().UploadService == config.UploadServiceMinio {
		uploadService = NewServiceMinio(ctx)
	} else {
		uploadService = NewSeaweedFS(ctx)
	}
	return &Service{
		Log: log.NewTLog("Service"),
		ctx: ctx,
		downloadClient: &http.Client{
			Timeout: time.Second * 30,
		},
		uploadService: uploadService,
	}
	// return NewServiceMinio(ctx)
}

// Service Service
type Service struct {
	downloadClient *http.Client
	log.Log
	ctx           *config.Context
	uploadClient  *http.Client
	uploadService IUploadService
}

func (s *Service) UploadFile(filePath string, contentType string, copyFileWriter func(io.Writer) error) (map[string]interface{}, error) {
	return s.uploadService.UploadFile(filePath, contentType, copyFileWriter)
}

func (s *Service) DownloadURL(path string, filename string) (string, error) {

	return s.uploadService.DownloadURL(path, filename)
}

func (s *Service) DownloadImage(url string) (*os.File, error) {
	var w = sync.WaitGroup{}
	w.Add(1)
	localPaths := make([]string, 0, 1)
	timestamp := time.Now().UnixNano()
	name := fmt.Sprintf("%s_%d_1.png", "tmp", timestamp)
	localPaths = append(localPaths, name)
	go s.downloadImage(url, name, &w)
	w.Wait()
	defer s.removeFile(localPaths)
	file, err := os.Open(localPaths[0])
	if err != nil {
		s.Error("读取下载的文件失败", zap.String("localPath", localPaths[0]), zap.Error(err))
		return nil, err
	}
	return file, nil
}

// DownloadAndMakeCompose 下载并组合图片
func (s *Service) DownloadAndMakeCompose(uploadPath string, downloadURLs []string) (map[string]interface{}, error) {
	var w = sync.WaitGroup{}
	w.Add(len(downloadURLs))
	localPaths := make([]string, 0, len(downloadURLs))
	timestamp := time.Now().UnixNano()
	for i, downloadPath := range downloadURLs {
		name := fmt.Sprintf("%s_%d_%d.png", "tmp", timestamp, i)
		localPaths = append(localPaths, name)
		go s.downloadImage(downloadPath, name, &w)
	}
	w.Wait()
	defer s.removeFile(localPaths)
	files := make([]*os.File, 0, len(localPaths))
	for _, localPath := range localPaths {
		file, err := os.Open(localPath)
		if err != nil {
			s.Warn("读取下载的图片失败", zap.String("localPath", localPath), zap.Error(err))
		}
		if file == nil {
			file, err = os.Open(s.ctx.GetConfig().DefaultAvatar)
			if err != nil {
				s.Warn("读取默认头像失败！", zap.String("avatar", s.ctx.GetConfig().DefaultAvatar), zap.Error(err))
			}
		}
		files = append(files, file)
	}
	img, err := s.MakeCompose(files)
	if err != nil {
		s.Error("组合图片失败！", zap.Error(err))
		return nil, err
	}
	// uploadURL := fmt.Sprintf("%s/public%s", s.ctx.GetConfig().UploadURL, uploadPath)
	// 上传文件
	resultMap, err := s.UploadFile(uploadPath, "image/png", func(w io.Writer) error {
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
	})
	if err != nil {
		s.Error("上传文件失败！", zap.Error(err))
		return nil, err
	}
	return resultMap, nil
}

// MakeCompose 组合图片
func (s *Service) MakeCompose(srcImgFiles []*os.File) (image.Image, error) {

	var minwidth int = 42
	var minheight int = 42
	var maxwidth int = 64
	var maxheight int = 64
	var borderWidth = 4

	groupWith := 128 + borderWidth
	groupHeight := 128 + borderWidth
	bounds := image.Rect(0, 0, groupWith, groupHeight)
	num := len(srcImgFiles)

	m := image.NewRGBA(bounds)
	white := color.RGBA{219, 223, 221, 221}
	draw.Draw(m, bounds, &image.Uniform{white}, image.ZP, draw.Src)

	wx := 0
	hy := -1
	for i := 0; i < num; i++ {

		if num == 3 {
			if i == 0 || i == 1 {
				hy++
				wx = 0
			}
		} else if num == 4 {
			if i%2 == 0 {
				hy++
				wx = 0
			}
		} else if num == 5 {
			if i == 2 || i == 0 {
				hy++
				wx = 0
			}
		} else {
			if i%3 == 0 {
				hy++
				wx = 0
			}
		}
		file := srcImgFiles[i]

		fileExt := filepath.Ext(file.Name())

		var m1 image.Image
		var err error
		if fileExt == "png" {
			m1, err = png.Decode(file)
		} else {
			m1, err = jpeg.Decode(file)
			if err != nil {
				s.Warn("jpeg编码出错！【%s】 将采用png编码", zap.Error(err))
				m1, err = png.Decode(file)
			}
		}

		if err != nil {
			log.Error(fmt.Sprintf("图片编码错误【%s】!", err.Error()))
			continue
		}
		var mbounds image.Rectangle
		//缩略图
		if num >= 5 {
			m1 = resize.Resize(uint(minwidth), uint(minheight), m1, resize.Lanczos3)
		} else {
			m1 = resize.Resize(uint(maxwidth), uint(maxheight), m1, resize.Lanczos3)
		}

		if num == 2 {
			mbounds = image.Rect(borderWidth+int(maxwidth)*wx, borderWidth+int(maxheight)*hy+(maxheight-borderWidth)/2, m1.Bounds().Size().X+int(maxwidth)*wx, m1.Bounds().Size().Y+int(maxheight)*hy+(maxheight-borderWidth)/2)
		} else if num == 3 {
			if i == 0 {
				mbounds = image.Rect((maxwidth-borderWidth)/2, borderWidth, m1.Bounds().Size().X+(maxwidth-borderWidth)/2+int(maxwidth)*wx, m1.Bounds().Size().Y+int(maxheight)*hy)
			} else {
				mbounds = image.Rect(borderWidth+int(maxwidth)*wx, borderWidth+int(maxheight)*hy, m1.Bounds().Size().X+int(maxwidth)*wx, m1.Bounds().Size().Y+int(maxheight)*hy)
			}
		} else if num == 5 {
			h := int(minheight)*hy + (minheight-borderWidth)/2
			if i <= 1 {
				mbounds = image.Rect(borderWidth+int(minwidth)*wx+(minwidth-borderWidth)/2, h+borderWidth, m1.Bounds().Size().X+int(minwidth)*wx+(minwidth-borderWidth)/2, m1.Bounds().Size().Y+h)
			} else {
				mbounds = image.Rect(borderWidth+int(minwidth)*wx, h+borderWidth, m1.Bounds().Size().X+int(minwidth)*wx, m1.Bounds().Size().Y+h)
			}
		} else if num == 6 {
			mbounds = image.Rect(borderWidth+int(minwidth)*wx, borderWidth+int(minheight)*hy+(minheight-borderWidth)/2, m1.Bounds().Size().X+int(minwidth)*wx, m1.Bounds().Size().Y+(minheight-borderWidth)/2+int(minheight)*hy)
		} else {
			if num > 5 {
				mbounds = image.Rect(borderWidth+int(minwidth)*wx, borderWidth+int(minheight)*hy, m1.Bounds().Size().X+int(minwidth)*wx, m1.Bounds().Size().Y+int(minheight)*hy)
			} else {
				mbounds = image.Rect(borderWidth+int(maxwidth)*wx, borderWidth+int(maxheight)*hy, m1.Bounds().Size().X+int(maxwidth)*wx, m1.Bounds().Size().Y+int(maxheight)*hy)
			}
		}

		draw.Draw(m, mbounds, m1, image.ZP, draw.Src)

		wx++
	}

	return m, nil
}

var uploadClient *http.Client
var onceUploadClient sync.Once

// UploadFile 上传文件
func uploadFile(uploadURL, fileName string, copyFileWriter func(io.Writer) error) (map[string]interface{}, error) {
	body := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(body)
	fileWriter, err := bodyWriter.CreateFormFile("file", fileName)
	if err != nil {
		limlog.Error("构建formFile失败！", zap.Error(err))
		return nil, err
	}
	err = copyFileWriter(fileWriter)
	if err != nil {
		limlog.Error("复制文件内容失败！", zap.String("uploadURL", uploadURL), zap.Error(err))
		return nil, err
	}
	bodyWriter.Close()
	fRequest, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		limlog.Error("创建上传请求失败！", zap.String("uploadURL", uploadURL), zap.Error(err))
		return nil, err
	}
	fRequest.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	onceUploadClient.Do(func() {
		uploadClient = &http.Client{
			Timeout: time.Second * 120,
		}
	})
	resp, err := uploadClient.Do(fRequest)
	if err != nil {
		limlog.Error("上传文件失败！", zap.String("uploadURL", uploadURL), zap.Error(err))
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("上传文件返回失败！")
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		limlog.Error("文件上传返回状态有误！", zap.Int("status", resp.StatusCode), zap.String("uploadURL", uploadURL), zap.String("fileName", fileName))
		return nil, errors.New("文件上传返回状态有误！")
	}
	defer resp.Body.Close()
	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		limlog.Error("读取上传返回的数据失败！", zap.Error(err))
		return nil, err
	}
	var resultMap map[string]interface{}
	err = util.ReadJsonByByte(respData, &resultMap)
	if err != nil {
		limlog.Error("上传返回的json格式有误！", zap.String("uploadURL", uploadURL), zap.String("fileName", fileName), zap.String("resp", string(respData)), zap.Error(err))
		return nil, err
	}
	return resultMap, err
}

func (s *Service) downloadImage(url string, imgName string, wg *sync.WaitGroup) {
	defer wg.Done()
	out, err := os.Create(imgName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()
	resp, err := s.downloadClient.Get(url)
	if err != nil {
		s.Error("下载图片错误！", zap.String("url", url), zap.Error(err))
		return
	}
	if resp == nil {
		s.Error("没有返回数据，下载图片失败！", zap.String("url", url))
		return
	}
	defer resp.Body.Close()
	pix, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.Error("读取下载的图片数据失败！", zap.Error(err))
		return
	}
	_, err = io.Copy(out, bytes.NewReader(pix))
}

func (s *Service) removeFile(filePaths []string) {
	for _, url := range filePaths {
		err := os.RemoveAll(url)
		if err != nil {
			s.Warn("移除文件失败！", zap.String("filePath", url), zap.Error(err))
		}
	}
}
