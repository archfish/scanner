package scanner

import (
	"fmt"
	"log/slog"

	"github.com/google/gousb"
)

// brotherDeviceState Brother扫描仪设备状态（隐藏在内部实现中）
type brotherDeviceState struct {
	ctx    *gousb.Context
	device *gousb.Device

	cfg *gousb.Config
	cif *gousb.Interface
	out *gousb.OutEndpoint
	in  *gousb.InEndpoint
}

// Close 关闭设备连接
func (ds *brotherDeviceState) Close() (err error) {
	if ds.cif != nil {
		ds.cif.Close()
	}
	if ds.cfg != nil {
		if cerr := ds.cfg.Close(); cerr != nil {
			err = fmt.Errorf("close usb config: %w", cerr)
		}
	}
	if ds.device != nil {
		if cerr := ds.device.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close usb device: %w", cerr)
		}
	}
	if ds.ctx != nil {
		if cerr := ds.ctx.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close usb context: %w", cerr)
		}
	}
	return err
}

// openBrotherDevice 打开Brother扫描仪设备
func openBrotherDevice(vendorID, productID uint16, opts DeviceOptions) *brotherDeviceState {
	state := &brotherDeviceState{
		ctx: gousb.NewContext(),
	}
	dev, err := state.ctx.OpenDeviceWithVIDPID(gousb.ID(vendorID), gousb.ID(productID))
	if err != nil {
		slog.Error("open device with VID/PID", "error", err)
		return state
	}
	state.device = dev

	state.cfg, err = state.device.Config(opts.ConfigNum)
	if err != nil {
		slog.Error("select usb device config", "error", err)
		return state
	}

	state.cif, err = state.cfg.Interface(opts.InterfaceNum, opts.InterfaceAlt)
	if err != nil {
		slog.Error("claim usb device interface", "error", err)
		return state
	}

	state.out, err = state.cif.OutEndpoint(opts.OutEndpointNum)
	if err != nil {
		slog.Error("open out endpoint", "error", err)
		return state
	}

	state.in, err = state.cif.InEndpoint(opts.InEndpointNum)
	if err != nil {
		slog.Error("open in endpoint", "error", err)
		return state
	}

	return state
}

// brotherScanRequest Brother扫描请求参数
type brotherScanRequest struct {
	horizontalDPI, verticalDPI uint16
	mode                       ScanMode
	compression                Compression
	brightness                 uint16 // 50
	contrast                   uint16 // 50
	left, top, width, height   uint16
}

func (req brotherScanRequest) Bytes() []byte {
	executeScan := "\x1bX\x0aR=%d,%d\x0aM=%s\x0aC=%s\x0aJ=MID\x0aB=%d\x0aN=%d\x0aA=%d,%d,%d,%d\x0aS=NORMAL_SCAN\x0aP=0\x0aG=0\x0aL=0\x0a\x80"
	return []byte(fmt.Sprintf(executeScan, req.horizontalDPI, req.verticalDPI, req.mode, req.compression, req.brightness, req.contrast, req.left, req.top, req.width, req.height))
}

// brotherNegotiateResponse Brother协商响应
type brotherNegotiateResponse struct {
	unknown                    [3]byte
	horizontalDPI, verticalDPI uint16 // or the other way around, always the same?
	unknown2                   uint16
	scanWidth, outWidth        uint16
	scanHeight, outHeight      uint16
}

func (nr *brotherNegotiateResponse) parse(d []byte) error {
	if len(d) < 10 {
		return fmt.Errorf("too short")
	}

	nr.unknown = [3]byte{d[0], d[1], d[2]}
	if _, err := fmt.Sscanf(string(d[3:len(d)-1]), "%d,%d,%d,%d,%d,%d,%d", &nr.horizontalDPI, &nr.verticalDPI, &nr.unknown2, &nr.scanWidth, &nr.outWidth, &nr.scanHeight, &nr.outHeight); err != nil {
		return err
	}
	return nil
}

var brotherNegotiateRequest = func(resolution uint16, mode ScanMode) []byte {
	return []byte(fmt.Sprintf("\x1bI\x0aR=%d,%d\x0aM=%s\x0a\x80", resolution, resolution, mode))
}
