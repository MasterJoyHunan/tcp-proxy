package main

import (
	"flag"
	"io"
	"net"

	"tcp_proxy/logger"

	"github.com/kardianos/service"
	"github.com/sirupsen/logrus"
)

var (
	localHost  string
	remoteHost string
	act        string
)

// 安装服务 tcp-proxy -a install -localhost 0.0.0.0:8484 -remotehost 127.0.0.1:8485 // 安装的时候才能使用 -localhost,-remotehost 参数, 否则无法生效
// 卸载服务 tcp-proxy -a uninstall
// 启动服务 tcp-proxy -a start
// 停止服务 tcp-proxy -a stop
// 重启服务 tcp-proxy -a restart

func init() {
	flag.StringVar(&localHost, "localhost", "0.0.0.0:8484", "本机暴露IP+端口")
	flag.StringVar(&remoteHost, "remotehost", "127.0.0.1:8485", "需要代理的IP+端口")
	flag.StringVar(&act, "a", "", "start/stop/restart/install/uninstall")
}

func main() {

	flag.Parse()

	logger.Setup()

	svcConfig := &service.Config{
		Name:        "tcp-proxy",
		DisplayName: "tcp-proxy tcp请求转发",
		Description: "tcp请求转发",
		Arguments: []string{
			"-localhost",
			localHost,
			"-remotehost",
			remoteHost,
		},
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		logrus.Fatal(err)
	}

	if len(act) > 0 {
		err = service.Control(s, act)
		if err != nil {
			logrus.Fatal(err)
		}
		return
	}
	err = s.Run()
	if err != nil {
		logrus.Fatal(err)
	}

}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}
func (p *program) run() {

	logrus.WithField("env", "localhost:"+localHost).Info()
	logrus.WithField("env", "remotehost:"+remoteHost).Info()

	l, err := net.Listen("tcp", localHost)
	if err != nil {
		logrus.Error(err.Error())
		return
	}

	logrus.Info("等待连接....")

	for {
		localConn, err := l.Accept()
		if err != nil {
			logrus.Error(err.Error())
			continue
		}

		logrus.Info("连接成功")

		remoteConn, err := net.Dial("tcp", remoteHost)
		if err != nil {
			logrus.Error(err.Error())
			localConn.Close()
			continue
		}

		go io.Copy(localConn, remoteConn)
		go io.Copy(remoteConn, localConn)
	}
}

func (p *program) Stop(s service.Service) error {
	logrus.Warn("stop tcp-proxy service")
	return nil
}
