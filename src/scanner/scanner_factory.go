package scanner

import (
	"fmt"
	"io"
)

// Scanner 抽象一个扫描仪设备
type Scanner interface {
	// Connect 连接一个设备
	Connect() error
	// Scan 开始扫描
	Scan(out io.Writer, opts ScanOptions) error
	// Close 断开扫描仪
	Disconnect() error
}

// NewScanner 创建扫描仪实例
// 使用工厂模式根据设备信息创建对应的扫描仪实现
func NewScanner(device DeviceInfo, opts DeviceOptions) (Scanner, error) {
	// 使用全局设备注册表来创建扫描仪实例
	scanner, err := GlobalDeviceRegistry.CreateScanner(device, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create scanner: %w", err)
	}
	return scanner, nil
}
