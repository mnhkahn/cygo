package http

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"time"

	cyurl "github.com/mnhkahn/cygo/net/url"
)

var ErrLog *log.Logger

var HTTP_METHOD = map[string]string{
	"GET":  "GET",
	"POST": "POST",
	//	"HEAD":    "HEAD",
	//	"PUT":     "PUT",
	//	"TRACE":   "TRACE",
	"OPTIONS": "OPTIONS",
	//	"DELETE":  "DELETE",
}

var SUPPORT_BODY_HTTP_METHOD = map[string]bool{
	"POST":   true,
	"PUT":    true,
	"DELETE": true,
}

type Server struct {
	Addr             *cyurl.Host
	Routes           *Route
	AllowHttpMethods []string
}

var DEFAULT_SERVER *Server
var ViewsTemplFiles map[string]string
var AppPath string
var ViewPath string

func Serve(addr string) {
	DEFAULT_SERVER.Addr = cyurl.ParseHost(addr)
	ln, err := net.Listen("tcp", DEFAULT_SERVER.Addr.String())
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	log.Printf("<<<Server Accepting on Port %s>>>\n", DEFAULT_SERVER.Addr.Port())
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

	views, _ := ioutil.ReadDir(ViewPath)
	for _, view := range views {
		ViewsTemplFiles[view.Name()] = ViewPath + "/" + view.Name()
	}

	DEFAULT_SERVER.Routes = NewRoute()

	Router("/", "OPTIONS", &Controller{}, "Option")
	Router("/favicon.ico", "GET", &Controller{}, "Favicon")
}

func handleConnection(conn net.Conn) {
	serve_time := time.Now()

	defer conn.Close()

	ctx := NewContext()
	ctx.ReqAddr = cyurl.ParseHost(conn.RemoteAddr().String())
	ctx.Req = NewRequest()
	ctx.Resp = NewResponse()

	var err error

	reader := bufio.NewReader(conn)

	// Read startline
	line, _, err := reader.ReadLine()
	if len(line) != 0 && err == nil {
		ctx.Req.ParseStartLine(line)
	}

	// Read header
	for {
		line, _, err := reader.ReadLine()
		if len(line) == 0 || err != nil {
			break
		}
		ctx.Req.ParseHeader(line)
	}

	// Read body
	if _, exists := SUPPORT_BODY_HTTP_METHOD[ctx.Req.Method]; exists {
		body := make([]byte, ctx.Req.ContentLength)
		_, err := reader.Read(body)
		if err == nil {
			ctx.Req.Body = string(body)
		} else {
			ctx.Req.Body = err.Error()
		}
	}

	if ctx.Req.Headers.Get(HTTP_HEAD_X_FORWARDED_FOR) != "" {
		ctx.ReqAddr = cyurl.ParseHost(ctx.Req.Headers.Get(HTTP_HEAD_X_FORWARDED_FOR))
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

	writer := bufio.NewWriter(conn)
	_, err = writer.WriteString(fmt.Sprintf("%s %d %s\r\n", ctx.Resp.Proto, ctx.Resp.StatusCode, StatusText(ctx.Resp.StatusCode)))
	if err != nil {
		ErrLog.Println(err)
	} else {
		writer.Flush()
	}
	for k, v := range ctx.Resp.Headers {
		for _, vv := range v {
			_, err = writer.WriteString(fmt.Sprintf("%s: %s\r\n", k, vv))
			if err != nil {
				ErrLog.Println(err)
			} else {
				writer.Flush()
			}
		}
	}
	_, err = writer.WriteString(CRLF)
	if err != nil {
		ErrLog.Println(err)
	} else {
		writer.Flush()
	}
	_, err = writer.WriteString(ctx.Resp.Body)
	if err != nil {
		ErrLog.Println(err)
	} else {
		writer.Flush()
	}
	//	ctx.elapse = time.Now().Sub(serve_time)
	log.Println(fmt.Sprintf(LOG_CONTEXT, ctx.ReqAddr.Address(), "-", serve_time.Format(LOG_TIME_FORMAT), ctx.Req.Method, ctx.Req.Url.RawPath, ctx.Req.Proto, ctx.Resp.StatusCode, ctx.Req.ContentLength, "-", ctx.Req.UserAgent, 0))
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
