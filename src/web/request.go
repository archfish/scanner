package web

import "scanner/src/scanner"

type DeviceListReq struct {
	DeviceType string
}

// ScanReq 扫描参数
type ScanReq struct {
	Device scanner.DeviceInfo   `json:"device"`
	Option *scanner.ScanOptions `json:"option"`
}

// ScanResp 扫描结果
type ScanResp struct {
	URL      string
	FileType string

	Req *ScanReq
}
