package scanner

import (
	"io"
)

var _ Scanner = (*LenovoScanner)(nil)

// LenovoScanner 联想扫描仪实现
// 通过嵌入BrotherScanner来复用其功能
type LenovoScanner struct {
	*BrotherScanner
}

// NewLenovoScanner 创建联想扫描仪实例
func NewLenovoScanner(usb DeviceInfo, opts DeviceOptions) *LenovoScanner {
	// 创建一个BrotherScanner实例
	brotherScanner := NewBrotherScanner(usb, opts)

	return &LenovoScanner{
		BrotherScanner: brotherScanner,
	}
}

// Connect 连接联想扫描仪设备
func (scanner *LenovoScanner) Connect() error {
	// 使用BrotherScanner的连接逻辑
	return scanner.BrotherScanner.Connect()
}

// Scan 使用BrotherScanner的扫描功能
// 联想M7206扫描仪使用与Brother相同的协议
func (scanner *LenovoScanner) Scan(out io.Writer, opts ScanOptions) error {
	// 直接使用BrotherScanner的扫描功能
	return scanner.BrotherScanner.Scan(out, opts)
}

// Disconnect 断开联想扫描仪设备连接
func (scanner *LenovoScanner) Disconnect() error {
	// 使用BrotherScanner的断开连接逻辑
	return scanner.BrotherScanner.Disconnect()
}

// LenovoScannerFactory 联想扫描仪工厂
type LenovoScannerFactory struct{}

// 确保LenovoScannerFactory实现了ScannerFactory接口
var _ ScannerFactory = (*LenovoScannerFactory)(nil)

// Match 检查是否匹配联想设备
func (lsf *LenovoScannerFactory) Match(usb DeviceInfo) bool {
	// 联想M7206扫描仪
	vendorID := usb.ParseVendorID()

	// 联想的VendorID通常是0x17ef
	if vendorID == 0x17ef {
		return true
	}

	return false
}

// CreateScanner 创建联想扫描仪实例
func (lsf *LenovoScannerFactory) CreateScanner(usb DeviceInfo, opts DeviceOptions) Scanner {
	return NewLenovoScanner(usb, opts)
}

// 注册联想扫描仪工厂
func init() {
	GlobalDeviceRegistry.Register(&LenovoScannerFactory{})
}
