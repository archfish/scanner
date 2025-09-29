package web

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// Routes 总路由
func Routes(custom func(r *gin.RouterGroup)) *gin.Engine {
	r := gin.Default()
	r.NoRoute(func(ctx *gin.Context) {
		RenderError(ctx, fmt.Errorf("Hello dEar :)"), http.StatusMethodNotAllowed, nil)
	})

	if custom != nil {
		custom(&r.RouterGroup)
	}

	return r
}

// ListenAndServe 启动一个API服务，stopCh关闭时代表服务关闭
func ListenAndServe(port string, custom func(r *gin.RouterGroup)) *http.Server {
	httpServe := &http.Server{
		Addr:    port,
		Handler: Routes(custom),
	}

	go func() {
		slog.Info(fmt.Sprintf("rest api listen port: %s", port))
		err := httpServe.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintln("服务终止", err))
		}
	}()

	return httpServe
}

const (
	errorCode   = "1"
	successCode = "0"
)

// JSONResponse 默认响应结构
type JSONResponse struct {
	Msg  string
	Code string
	Data interface{} `json:",omitempty"`
	Help string      `json:",omitempty"`
}

func (j *JSONResponse) toJSON() []byte {
	res, err := json.Marshal(j)

	if err != nil {
		rsp := JSONResponse{Msg: err.Error(), Code: errorCode}
		return rsp.toJSON()
	}

	return res
}

func (j *JSONResponse) RenderJSON(ctx *gin.Context, status int) {
	res := j.toJSON()
	ctx.Data(status, "application/json", res)
	slog.Info(string(res))
}

func RenderError(ctx *gin.Context, err error, status int, data interface{}) {
	resp := JSONResponse{Msg: err.Error(), Data: data, Code: errorCode}

	resp.RenderJSON(ctx, status)
}

func RenderSuccess(ctx *gin.Context, data interface{}) {
	resp := JSONResponse{Msg: "成功", Data: data, Code: successCode}
	resp.RenderJSON(ctx, http.StatusOK)
}

func SendData(ctx *gin.Context, filename string, reader io.Reader) {
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", "attachment; filename="+url.QueryEscape(filename)) // 用来指定下载下来的文件名
	ctx.Header("Content-Transfer-Encoding", "binary")
	io.Copy(ctx.Writer, reader)
}
