package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

var DEFAULT_CONTROLLER *Controller = new(Controller)

type ControllerIfac interface {
	Init(ctx *Context)
	Finish()
}

type Controller struct {
	Ctx       *Context
	TemplPath string
}

func (this *Controller) Init(ctx *Context) {
	this.Ctx = ctx
}

func (this *Controller) BasicAuth(realm string) {
	this.Ctx.Resp.StatusCode = StatusUnauthorized
	this.Ctx.Resp.Headers.Add(HTTP_HEAD_WWW_AUTHENTICATE, fmt.Sprintf("Basic realm=%s", realm))
}

func (this *Controller) Option() {
	allowMethods := []string{}
	for _, method := range HTTP_METHOD {
		allowMethods = append(allowMethods, method)
	}
	this.Ctx.Resp.Headers.Add(HTTP_HEAD_ALLOW, strings.Join(allowMethods, ", "))
}

func (this *Controller) getString(param string) string {
	if len(this.Ctx.Req.Url.Query().Get(param)) > 0 {
		return this.Ctx.Req.Url.Query().Get(param)[0]
	}
	return ""
}

func (this *Controller) GetString(param string) string {
	return this.getString(param)
}

func (this *Controller) GetInt(param string) int {
	if this.getString(param) != "" {
		i, _ := strconv.Atoi(this.getString(param))
		return i
	}
	return 0
}

func (this *Controller) ServeRaw(v []byte) {
	this.Ctx.Resp.Body = string(v)
}

func (this *Controller) ServeJson(j interface{}) {
	this.Ctx.Resp.Headers.Add(HTTP_HEAD_CONTENTTYPE, this.ContentType(filepath.Ext(params[0].(string))))
	v, _ := json.Marshal(j)
	this.Ctx.Resp.Body = string(v)
}

func (this *Controller) ServeView(params ...interface{}) {
	if len(params) <= 0 {

	} else {
		if templ, exists := ViewsTemplFiles[params[0].(string)]; exists {
			if len(params) == 1 {
				this.Ctx.Resp.Headers.Add(HTTP_HEAD_CONTENTTYPE, this.ContentType(filepath.Ext(params[0].(string))))
				v, _ := ioutil.ReadFile(templ)
				this.Ctx.Resp.Body = string(v)
			} else if len(params) == 2 {
				body := new(bytes.Buffer)

				t, err := template.ParseFiles(templ)
				if err != nil {
					this.Ctx.Resp.Body = err.Error()
					return
				} else {
					this.Ctx.Resp.Body = string(body.Bytes())
				}

				err = t.Execute(body, params[1])
				if err != nil {
					this.Ctx.Resp.Body = err.Error()
				} else {
					this.Ctx.Resp.Body = string(body.Bytes())
				}
			}
		} else {
			this.debugLog(fmt.Sprintf("Can't find the template file %v", params))
			ErrLog.Println("Can't find the template file", params)
		}
	}
}

func (this *Controller) ServeFile(params ...interface{}) {
	if len(params) <= 0 {

	} else if len(params) == 1 {
		if templ, exists := ViewsTemplFiles[params[0].(string)]; exists {
			this.Ctx.Resp.Body = string(templ)
		} else {
			ErrLog.Println("Can't find the template file", params)
		}
	} else {

	}
}

func (this *Controller) debugLog(log string) {
	this.Ctx.Resp.Body = log
}

func (this *Controller) Favicon() {
	this.Ctx.Resp.StatusCode = StatusFound
	this.Ctx.Resp.Headers.Add(HTTP_HEAD_LOCATION, "http://7b1h1l.com1.z0.glb.clouddn.com/c32.ico")
}

func (this *Controller) Finish() {
}

// TypeByExtension returns the MIME type associated with the file extension ext. The extension ext should begin with a leading dot, as in ".html". When ext has no associated type, TypeByExtension returns "".
func (this *Controller) ContentType(ext string) string {
	if !strings.HasPrefix(ext, ".") {
		return ""
	}
	return mime.TypeByExtension(ext)
}
