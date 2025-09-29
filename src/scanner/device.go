package scanner

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/google/gousb"
)

type (
	ScanMode    string
	Compression string
)

// ScanMode represents the scanning mode for the scanner.
// Black & White|Gray[Error Diffusion]|True Gray|24bit Color|24bit Color[Fast] [24bit Color[Fast]]
var (
	ScanModeCGRAY ScanMode = "CGRAY"

	CompressionJPEG    Compression = "JPEG"
	CompressionRLENGTH Compression = "RLENGTH"
)

var DefaultDeviceOptions = DeviceOptions{
	ConfigNum:      1,
	InterfaceNum:   1,
	InterfaceAlt:   0,
	OutEndpointNum: 4,
	InEndpointNum:  5,
}

type DeviceOptions struct {
	ConfigNum      int
	InterfaceNum   int
	InterfaceAlt   int
	OutEndpointNum int
	InEndpointNum  int
}

// DeviceInfo 设备基本信息
// lsusb 输出信息里的 VendorID:ProductID
// 例如：Bus 001 Device 002: ID 17ef:5629 Lenovo M7206
type DeviceInfo struct {
	Name      string
	VendorID  string
	ProductID string
}

// Transfer 统一规则取核心数据
func (info *DeviceInfo) Transfer(desc *gousb.DeviceDesc) {
	info.VendorID = "0x" + desc.Vendor.String()
	info.ProductID = "0x" + desc.Product.String()
}

// ParseVendorID 将设备ID解析为整型
func (info *DeviceInfo) ParseVendorID() uint16 {
	return ParseID(info.VendorID)
}

// ParseProductID 将产品ID解析为整型
func (info *DeviceInfo) ParseProductID() uint16 {
	return ParseID(info.ProductID)
}

func ParseID(id string) uint16 {
	number, err := strconv.ParseUint(id, 0, 16)
	if err != nil {
		slog.Error("invalid ID format, example: 0x17ef", "got:", id)
		return 0
	}
	return uint16(number)
}

// ListUSBDevice 获取机器上的所有USB设备
func ListUSBDevice() []DeviceInfo {
	ctx := gousb.NewContext()
	defer ctx.Close()

	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return true // 打开所有设备以获取信息
	})

	if err != nil {
		slog.Error("Failed to list USB devices", "error", err)
		return nil
	}

	defer func() {
		for _, d := range devices {
			d.Close()
		}
	}()

	var deviceList []DeviceInfo
	for _, device := range devices {
		if device.Desc == nil {
			continue
		}

		// 尝试获取产品名称
		name, err := device.Manufacturer()
		if err != nil {
			slog.Error("Failed to get USB device name", "error", err)
		}
		// 白名单，现在只支持
		if strings.Contains(strings.ToLower(name), "xhci") {
			continue
		}

		deviceInfo := DeviceInfo{
			Name: name,
		}
		deviceInfo.Transfer(device.Desc)

		deviceList = append(deviceList, deviceInfo)
	}

	return deviceList
}
