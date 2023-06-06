package server

import (
	"net"
	"os"
	"path/filepath"

	"github.com/WuKongIM/WuKongChatServer/pkg/log"
	"github.com/WuKongIM/WuKongChatServer/pkg/wkhttp"
	"github.com/judwhite/go-svc"
	"github.com/unrolled/secure"
	"google.golang.org/grpc"
)

// Server 悟空聊天server
type Server struct {
	r *wkhttp.WKHttp
	log.TLog
	sslAddr    string
	addr       string
	GrpcServer *grpc.Server
	grpcAddr   string
}

// New 创建悟空聊天server
func New(addr string, sslAddr string, grpcAddr string) *Server {
	r := wkhttp.New()
	r.Use(wkhttp.CORSMiddleware())
	s := &Server{
		r:          r,
		addr:       addr,
		sslAddr:    sslAddr,
		grpcAddr:   grpcAddr,
		GrpcServer: grpc.NewServer(),
	}
	return s
}

func (s *Server) Init(env svc.Environment) error {
	if env.IsWindowsService() {
		dir := filepath.Dir(os.Args[0])
		return os.Chdir(dir)
	}
	return nil
}

// Run 运行
func (s *Server) run(sslAddr string, addr ...string) error {

	// s.r.LoadHTMLGlob("configs/webroot/**/*.html")
	s.r.Static("/web", "./configs/web")
	s.r.Any("/v1/ping", func(c *wkhttp.Context) {
		c.ResponseOK()
	})

	if len(addr) != 0 {
		if sslAddr != "" {
			go func() {
				err := s.r.Run(addr...)
				if err != nil {
					panic(err)
				}
			}()
		} else {
			err := s.r.Run(addr...)
			if err != nil {
				return err
			}
		}

	}

	// https 服务
	if sslAddr != "" {
		s.r.Use(TlsHandler(sslAddr))
		currDir, _ := os.Getwd()
		return s.r.RunTLS(sslAddr, currDir+"/configs/ssl/ssl.pem", currDir+"/configs/ssl/ssl.key")
	}

	return nil

}

func (s *Server) Start() error {
	go func() {
		err := s.run(s.sslAddr, s.addr)
		if err != nil {
			panic(err)
		}
	}()

	lis, err := net.Listen("tcp", s.grpcAddr)
	if err != nil {
		return err
	}
	go func() {
		err = s.GrpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

func (s *Server) Stop() error {
	s.GrpcServer.Stop()
	return nil
}

func TlsHandler(sslAddr string) wkhttp.HandlerFunc {
	return func(c *wkhttp.Context) {
		secureMiddleware := secure.New(secure.Options{
			SSLRedirect: true,
			SSLHost:     sslAddr,
		})
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			return
		}

		c.Next()
	}
}

// GetRoute 获取路由
func (s *Server) GetRoute() *wkhttp.WKHttp {
	return s.r
}
