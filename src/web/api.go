package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed view/*
var webFiles embed.FS

// 静态文件处理函数
func staticHandler(c *gin.Context) {
	path := c.Request.URL.Path
	if path == "/" {
		path = "/index.html"
	}

	// 尝试从本地文件系统读取
	localPath := "./view" + path
	if _, err := os.Stat(localPath); err == nil {
		c.File(localPath)
		return
	}

	// 从嵌入的文件系统读取（view目录现在位于web目录下）
	embedPath := "view" + path
	file, err := webFiles.ReadFile(embedPath)
	if err != nil {
		c.String(http.StatusNotFound, "File not found: %s", path)
		return
	}

	// 根据文件扩展名设置正确的Content-Type
	contentType := "application/octet-stream"
	switch {
	case strings.HasSuffix(path, ".html"):
		contentType = "text/html; charset=utf-8"
	case strings.HasSuffix(path, ".js"):
		contentType = "application/javascript"
	case strings.HasSuffix(path, ".css"):
		contentType = "text/css"
	case strings.HasSuffix(path, ".json"):
		contentType = "application/json"
		// 可以根据需要添加更多的文件类型
	}

	c.Data(http.StatusOK, contentType, file)
}

// Routes 总路由
func Routes(custom func(r *gin.RouterGroup)) *gin.Engine {
	r := gin.Default()
	r.NoRoute(staticHandler)

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
