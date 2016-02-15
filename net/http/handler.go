package http

import "reflect"

type Handler interface {
	ServeHTTP(ctx *Context)
}

type Handle struct {
	ctrl       ControllerIfac
	methodName string
	fn         reflect.Value
}

func (this *Handle) ServeHTTP(ctx *Context) {
	this.ctrl.Init(ctx)
	this.fn.Call(nil)
}
