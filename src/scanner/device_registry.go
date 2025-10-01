package scanner

import (
	"fmt"
)

// ScannerFactory 设备扫描仪工厂接口
type ScannerFactory interface {
	// CreateScanner 创建扫描仪实例
	CreateScanner(usb DeviceInfo, opts DeviceOptions) Scanner
	// Match 匹配设备
	Match(usb DeviceInfo) bool
}

// deviceRegistry 设备注册表
type deviceRegistry struct {
	factories []ScannerFactory
}

// GlobalDeviceRegistry 全局设备注册表
var GlobalDeviceRegistry = &deviceRegistry{
	factories: make([]ScannerFactory, 0),
}

// Register 注册扫描仪工厂
func (dr *deviceRegistry) Register(factory ScannerFactory) {
	dr.factories = append(dr.factories, factory)
}

// CreateScanner 创建适配的扫描仪实例
func (dr *deviceRegistry) CreateScanner(usb DeviceInfo, opts DeviceOptions) (Scanner, error) {
	for _, factory := range dr.factories {
		if factory.Match(usb) {
			return factory.CreateScanner(usb, opts), nil
		}
	}
	return nil, fmt.Errorf("no suitable scanner found for device: VendorID=%s, ProductID=%s", usb.VendorID, usb.ProductID)
}

// 初始化注册表
func init() {
	// 注意：具体的工厂实现在各自的设备文件中注册
}
