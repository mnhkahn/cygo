package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var ErrLog *log.Logger

var HTTP_METHOD = map[string]string{
	"GET": "GET",
	//	"POST":    "POST",
	//	"HEAD":    "HEAD",
	//	"PUT":     "PUT",
	//	"TRACE":   "TRACE",
	"OPTIONS": "OPTIONS",
	//	"DELETE":  "DELETE",
}

type Address struct {
	Host string
	Port string
}

func NewAddress(addr_str string) *Address {
	addr := new(Address)
	addr_strs := strings.Split(addr_str, ":")
	if len(addr_strs) == 2 {
		addr.Host, addr.Port = addr_strs[0], addr_strs[1]
	} else {
		addr.Host = addr_str
	}
	return addr
}

func (this *Address) String() string {
	return this.Host + ":" + this.Port
}

type Server struct {
	Addr             *Address
	Routes           *Route
	AllowHttpMethods []string
}

var DEFAULT_SERVER *Server
var ViewsTemplFiles map[string]string
var AppPath string
var ViewPath string

func Serve(addr string) {
	DEFAULT_SERVER.Addr = NewAddress(addr)
	ln, err := net.Listen("tcp", DEFAULT_SERVER.Addr.String())
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	log.Printf("<<<Server Accepting on Port %s>>>\n", DEFAULT_SERVER.Addr.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			ErrLog.Println(err)
		}
		go handleConnection(conn)
	}
}

func init() {
	log.SetFlags(0)
	log.Println(CYEAM_LOG)
	errlogFile, logErr := os.OpenFile("error.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)

	if logErr != nil {
		fmt.Println("Fail to find", "error.log", " start Failed")
	}

	ErrLog = log.New(errlogFile, "", log.LstdFlags|log.Llongfile)

	DEFAULT_SERVER = new(Server)

	AppPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	ViewPath = AppPath + "/views"
	ViewsTemplFiles = make(map[string]string)

	var err error
	views, _ := ioutil.ReadDir(ViewPath)
	for _, view := range views {
		ViewsTemplFiles[view.Name()], err = ioutil.ReadFile(ViewPath + "/" + view.Name())
		if err != nil {
			ErrLog.Println(err)
		}
	}

	DEFAULT_SERVER.Routes = NewRoute()

	Router("/", "OPTIONS", &Controller{}, "Option")
	Router("/favicon.ico", "GET", &Controller{}, "Favicon")
}

func handleConnection(conn net.Conn) {
	serve_time := time.Now()

	defer conn.Close()

	ctx := NewContext()
	ctx.ReqAddr = NewAddress(conn.RemoteAddr().String())
	ctx.Req = NewRequest()
	ctx.Resp = NewResponse()

	for {
		buf := make([]byte, 1024)
		reqLen, err := conn.Read(buf)
		if reqLen > 0 && err == nil {
			ctx.Req.Raw.Write(buf)
			if reqLen < len(buf) && err == nil {
				break
			}
		} else if err == io.EOF {
			return
		} else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			break
		} else {
			ErrLog.Println("Error to read message because of ", err, reqLen)
			ctx.Resp.StatusCode = StatusInternalServerError
			break
		}
	}

	ctx.Req.Init()
	if ctx.Req.Headers.Get(HTTP_HEAD_X_FORWARDED_FOR) != "" {
		ctx.ReqAddr = NewAddress(ctx.Req.Headers.Get(HTTP_HEAD_X_FORWARDED_FOR))
	}
	ctx.Resp.Proto = ctx.Req.Proto

	// black url list
	if _, exists := BLACK_URL[ctx.Req.Url.Path]; exists {
		ctx.Resp.StatusCode = StatusOK
		ctx.Resp.Body = "Fuck You!"
	} else if DEFAULT_SERVER.Routes.routes[ctx.Req.Method][ctx.Req.Url.Path] != nil {
		DEFAULT_SERVER.Routes.routes[ctx.Req.Method][ctx.Req.Url.Path].ServeHTTP(ctx)
	} else {
		if _, exists := HTTP_METHOD[ctx.Req.Method]; !exists {
			ctx.Resp.StatusCode = StatusMethodNotAllowed
		} else {
			ctx.Resp.StatusCode = StatusNotFound
		}
	}

	if ctx.Resp.StatusCode == StatusNotFound {
		ctx.Resp.Body = DEFAULT_ERROR_PAGE
	}
	ctx.Resp.Headers.Add(HTTP_HEAD_DATE, serve_time.Format(time.RFC1123))
	ctx.Resp.Headers.Add(HTTP_HEAD_CONTENTLENGTH, fmt.Sprintf("%d", len(ctx.Resp.Body)))

	buffers := bytes.Buffer{}
	buffers.WriteString(fmt.Sprintf("%s %d %s\r\n", ctx.Resp.Proto, ctx.Resp.StatusCode, StatusText(ctx.Resp.StatusCode)))
	for k, v := range ctx.Resp.Headers {
		for _, vv := range v {
			buffers.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
		}
	}
	buffers.WriteString(CRLF)
	buffers.WriteString(ctx.Resp.Body)
	_, err := conn.Write(buffers.Bytes())
	if err != nil {
		ErrLog.Println(err)
	}
	//	ctx.elapse = time.Now().Sub(serve_time)
	log.Println(fmt.Sprintf(LOG_CONTEXT, ctx.ReqAddr.Host, "-", serve_time.Format(LOG_TIME_FORMAT), ctx.Req.Method, ctx.Req.Url.RawPath, ctx.Req.Proto, ctx.Resp.StatusCode, len(ctx.Req.Body), "-", ctx.Req.UserAgent, 0))
}

func Router(path string, method string, ctrl ControllerIfac, methodName string) {
	if _, exists := HTTP_METHOD[method]; !exists {
		ErrLog.Println("Method not allowed", method, path, methodName)
		return
	}
	handler := new(Handle)
	handler.ctrl = ctrl
	handler.methodName = methodName
	handler.fn = reflect.ValueOf(handler.ctrl).MethodByName(handler.methodName)
	DEFAULT_SERVER.Routes.routes[method][path] = handler
}

var DEFAULT_ERROR_PAGE = "<iframe scrolling='no' frameborder='0' src='http://yibo.iyiyun.com/js/yibo404/key/2354' width='640' height='464' style='display:block;'></iframe>"
