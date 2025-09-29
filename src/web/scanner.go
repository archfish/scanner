package web

import (
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"scanner/src/scanner"
	"time"

	"github.com/gin-gonic/gin"
)

//go:embed index.html
var webFiles embed.FS

// 扫描件存储位置
var DefaultAttachmentPath = "./attachment"

func AddWebRoutes(r *gin.RouterGroup) {
	// 提供静态文件服务（如果src/web目录存在）
	if _, err := os.Stat("./web"); err == nil {
		// 提供index.html文件
		r.GET("/", func(c *gin.Context) {
			c.File("./web/index.html")
		})
	} else {
		// 否则使用嵌入的文件
		r.GET("/", func(c *gin.Context) {
			file, err := webFiles.ReadFile("index.html")
			if err != nil {
				c.String(http.StatusNotFound, "File not found")
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", file)
		})
	}

	// 确保附件目录存在
	if _, err := os.Stat(DefaultAttachmentPath); os.IsNotExist(err) {
		os.MkdirAll(DefaultAttachmentPath, 0755)
	}

	r.Group("/api").
		POST("/scan", Scan).
		GET("/devices", ListUSBDevice).
		GET("/download/:attachID", Download) // 使用附件ID下载
}

// ListUSBDevice 查看本机所有USB设备
func ListUSBDevice(ctx *gin.Context) {
	devices := scanner.ListUSBDevice()
	RenderSuccess(ctx, devices)
}

// Scan 执行扫描
func Scan(ctx *gin.Context) {
	var req ScanReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		RenderError(ctx, err, http.StatusBadRequest, nil)
		return
	}
	if req.Option == nil {
		req.Option = &scanner.DefaultScanOptions
	}

	// 如果没有传入设备信息，尝试使用第一个可用设备
	if req.Device.VendorID == "" || req.Device.ProductID == "" {
		devices := scanner.ListUSBDevice()
		if len(devices) > 0 {
			req.Device = devices[0]
		} else {
			RenderError(ctx, fmt.Errorf("no USB device found"), http.StatusNotFound, nil)
			return
		}
	}

	// 创建M7206扫描器实例
	scan := scanner.NewCommonScanner(req.Device, scanner.DefaultDeviceOptions)

	// 初始化USB上下文
	if err := scan.Connect(); err != nil {
		RenderError(ctx, err, http.StatusInternalServerError, nil)
		return
	}
	defer scan.Disconnect()

	slog.Info("Successfully opened scanner device", "vendorID", req.Device.VendorID, "productID", req.Device.ProductID)

	// 使用getAttachment()创建可重复访问的路径
	filepath := getAttachment()
	file, err := os.Create(filepath)
	if err != nil {
		RenderError(ctx, err, http.StatusInternalServerError, nil)
		return
	}
	defer file.Close()

	// 执行扫描
	if err := scan.Scan(file, *req.Option); err != nil {
		RenderError(ctx, err, http.StatusInternalServerError, nil)
		return
	}

	// 使用文件名作为attachID
	attachID := filepath[len(DefaultAttachmentPath)+1:] // 移除前缀路径

	result := ScanResp{
		URL:      fmt.Sprintf("/api/download/%s", attachID),
		FileType: "jpeg",
		Req:      &req,
	}

	RenderSuccess(ctx, result)
}

// Download 下载扫描件
func Download(ctx *gin.Context) {
	attachID := ctx.Param("attachID")
	// 构建完整文件路径
	filepath := fmt.Sprintf("%s/%s", DefaultAttachmentPath, attachID)

	f, err := os.Open(filepath)
	if err != nil {
		RenderError(ctx, err, http.StatusBadRequest, nil)
		return
	}
	defer f.Close()

	// 使用响应头中的文件名
	SendData(ctx, attachID, f)
}

func getAttachment() string {
	return fmt.Sprintf("%s/%s.jpg", DefaultAttachmentPath, time.Now().Local().Format("20060102T150405"))
}
