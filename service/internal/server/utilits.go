package server

import (
	"github.com/goccy/go-json"
	"github.com/valyala/fasthttp"
)

const (
	JSON string = "application/json"
	XML  string = "application/xml"
)

func MakeRequest(client *fasthttp.HostClient, t string) (res []byte, err error) {
	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.Header.Set(fasthttp.HeaderContentType, t)
	req.SetRequestURI(client.Addr)
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)

	err = client.Do(req, resp)
	if err != nil {
		return
	}

	return resp.Body(), nil
}

func makeResponse(ctx *fasthttp.RequestCtx, c any) {
	request, err := json.Marshal(c)
	if err != nil {
		ctx.Error(err.Error(), fasthttp.StatusInternalServerError)
		return
	}

	ctx.Response.Header.Set(fasthttp.HeaderContentType, JSON)
	ctx.Response.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(request)
}
