package scanner

import (
	"fmt"
	"io"
	"time"
)

var _ Scanner = (*BrotherScanner)(nil)

var (
	ResponseTimeoutIt   = 100
	WaitBetweenRequests = 30 * time.Millisecond
)

// BrotherScanner Brother扫描仪实现
type BrotherScanner struct {
	usb  DeviceInfo
	opts DeviceOptions

	state *brotherDeviceState
}

// NewBrotherScanner 创建Brother扫描仪实例
func NewBrotherScanner(usb DeviceInfo, opts DeviceOptions) *BrotherScanner {
	return &BrotherScanner{
		usb:  usb,
		opts: opts,
	}
}

func (scanner *BrotherScanner) Connect() error {
	scanner.state = openBrotherDevice(scanner.usb.ParseVendorID(), scanner.usb.ParseProductID(), scanner.opts)
	if scanner.state.device == nil {
		return fmt.Errorf("打开设备失败！")
	}

	return nil
}

func (scanner *BrotherScanner) Scan(out io.Writer, opts ScanOptions) error {
	if err := scanner.control(1); err != nil {
		return fmt.Errorf("1st pre-init control transfer: %w", err)
	}
	if _, err := scanner.queryCapabilities(); err != nil {
		return fmt.Errorf("query capabilities: %w", err)
	}
	if err := scanner.control(2); err != nil {
		return fmt.Errorf("1st post-query control transfer: %w", err)
	}
	if err := scanner.control(1); err != nil {
		return fmt.Errorf("2nd post-query control transfer: %w", err)
	}

	neg, err := scanner.negotiateScannerSettings(opts)
	if err != nil {
		return fmt.Errorf("negotiate scanner settings: %w", err)
	}

	if err := scanner.postNegotiate(); err != nil {
		return fmt.Errorf("post negotiate: %w", err)
	}

	top := mmToPixels(opts.Top, neg.verticalDPI)
	left := mmToPixels(opts.Left, neg.horizontalDPI)

	if err := scanner.startScan(brotherScanRequest{
		horizontalDPI: neg.horizontalDPI,
		verticalDPI:   neg.verticalDPI,
		mode:          opts.Mode,
		compression:   CompressionJPEG,
		brightness:    50,
		contrast:      50,
		top:           top,
		left:          left,
		width:         min(mmToPixels(opts.Width, neg.horizontalDPI), mmToPixels(float64(neg.scanWidth), neg.horizontalDPI)),
		height:        min(mmToPixels(opts.Height, neg.verticalDPI), mmToPixels(float64(neg.scanHeight), neg.verticalDPI)),
	}); err != nil {
		return fmt.Errorf("start scan: %w", err)
	}

	if err := scanner.readScanData(out); err != nil {
		return fmt.Errorf("read scan data: %w", err)
	}

	if err := scanner.control(2); err != nil {
		return fmt.Errorf("post-scan control: %w", err)
	}
	return nil
}

func (scanner *BrotherScanner) Disconnect() error {
	return scanner.state.Close()
}

// We only see 0x0c0 control transfers, they always have value 0x0002 and index 0, the data is always 5 bytes long.
func (scanner *BrotherScanner) control(request uint8) error {
	data := make([]byte, 5)
	_, err := scanner.state.device.Control(0xc0, request, 0x0002, 0, data)
	if err != nil {
		return fmt.Errorf("control transfer: %w", err)
	}
	return nil
}

/*
Request:
0040   1b 51 0a 80                                       .Q..

Response:
0040   c1 00 1c 09 ff 3f 00 00 00 00 00 00 00 01 04 01   .....?..........
0050   01 01 01 01 00 00 00 00 00 00 00 00 00 01         ..............
*/
func (scanner *BrotherScanner) queryCapabilities() ([]byte, error) {
	cmd := []byte{0x1b, 0x51, 0x0a, 0x80} // 0x51 = 'Q'
	if _, err := scanner.state.out.Write(cmd); err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	// NOTE: this is probably returning a list of capabilities the scanner supports
	//       it needs to be understood further
	if resp, err := scanner.waitForResponse(281); err != nil {
		return nil, err
	} else {
		return resp, nil
	}
}

/*
Request:
0040   1b 49 0a 52 3d 33 30 30 2c 33 30 30 0a 4d 3d 43   .I.R=300,300.M=C
0050   47 52 41 59 0a 80                                 GRAY..

Response:
0040   00 1d 00 33 30 30 2c 33 30 30 2c 32 2c 32 30 39   ...300,300,2,209
0050   2c 32 34 38 30 2c 32 39 31 2c 33 34 33 37 2c 00   ,2480,291,3437,.
*/
func (scanner *BrotherScanner) negotiateScannerSettings(opts ScanOptions) (*brotherNegotiateResponse, error) {
	if _, err := scanner.state.out.Write(brotherNegotiateRequest(opts.DPI, opts.Mode)); err != nil {
		return nil, fmt.Errorf("sending command: %w", err)
	}
	rawData, err := scanner.waitForResponse(281)
	if err != nil {
		return nil, err
	}

	resp := &brotherNegotiateResponse{}
	if err := resp.parse(rawData); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	return resp, nil
}

/*
Request:
0040   1b 44 0a 41 44 46 0a 80                           .D.ADF..

Response:
0040   d0                                                .
*/
func (scanner *BrotherScanner) postNegotiate() error {
	cmd := []byte{0x1b, 0x44, 0x0a, 0x41, 0x44, 0x46, 0x0a, 0x80}
	if _, err := scanner.state.out.Write(cmd); err != nil {
		return fmt.Errorf("send command: %w", err)
	}

	_, err := scanner.waitForResponse(64)
	if err != nil {
		return err
	}
	return nil
}

/*
Response for JPEG compression:

This is a preamble that sometimes starts a frame, it probably contains the length of the image data in the current "batch", maybe page?
0000   00 60 e4 e8 46 9f ff ...

I'm currently ignoring that and just writing the whole file to output.
*/
func (scanner *BrotherScanner) readScanData(out io.Writer) error {
	buf := make([]byte, 4096*4)
	for {
		n, err := scanner.state.in.Read(buf)
		if err != nil {
			return err
		}
		if n == 0 {
			continue
		}
		if n == 1 && buf[0] == 0x80 {
			break
		}
		startOffset := 0
		// Sometimes we get some headers
		// NOTE: understand what they mean, we're currently skipping them, but it will probably break
		if len(buf) > 12 && buf[0] == 0x64 && buf[1] == 0x07 && buf[2] == 0x00 {
			startOffset = 12
		}
		if _, err := out.Write(buf[startOffset:n]); err != nil {
			return fmt.Errorf("send to writer: %w", err)
		}
	}
	return nil
}

/*
Request:
0040   1b 58 0a 52 3d 31 30 30 2c 31 30 30 0a 4d 3d 47   .X.R=100,100.M=G
0050   52 41 59 36 34 0a 43 3d 52 4c 45 4e 47 54 48 0a   RAY64.C=RLENGTH.
0060   42 3d 35 30 0a 4e 3d 35 30 0a 41 3d 33 32 2c 34   B=50.N=50.A=32,4
0070   32 2c 38 31 36 2c 31 31 34 35 0a 53 3d 4e 4f 52   2,816,1145.S=NOR
0080   4d 41 4c 5f 53 43 41 4e 0a 50 3d 30 0a 47 3d 30   MAL_SCAN.P=0.G=0
0090   0a 4c 3d 30 0a 80                                 .L=0..

Response:
scan data

or
0040   1b 52                                             .R

in case of error?
*/
func (scanner *BrotherScanner) startScan(request brotherScanRequest) error {
	if _, err := scanner.state.out.Write(request.Bytes()); err != nil {
		return fmt.Errorf("send command: %w", err)
	}
	return nil
}

func (scanner *BrotherScanner) waitForResponse(len int) ([]byte, error) {
	buf := make([]byte, len)
	for i := 0; ; i++ {
		if i > ResponseTimeoutIt {
			return nil, fmt.Errorf("timeout waiting for response")
		}
		packetLen, err := scanner.state.in.Read(buf)
		if err != nil {
			return nil, fmt.Errorf("read response: %w", err)
		}
		if packetLen == 0 {
			time.Sleep(WaitBetweenRequests)
			continue
		}
		return buf[:packetLen], nil
	}
}

func mmToPixels(mm float64, dpi uint16) uint16 {
	return uint16(mm * float64(dpi) / 25.4)
}

// BrotherScannerFactory Brother扫描仪工厂
type BrotherScannerFactory struct{}

// 确保BrotherScannerFactory实现了ScannerFactory接口
var _ ScannerFactory = (*BrotherScannerFactory)(nil)

// Match 检查是否匹配Brother设备
func (bsf *BrotherScannerFactory) Match(usb DeviceInfo) bool {
	// 可以根据VendorID来判断是否为Brother设备
	// 这里简化处理，实际应该有完整的VendorID列表
	vendorID := usb.ParseVendorID()
	// Brother的VendorID通常是0x04f9
	return vendorID == 0x04f9
}

// CreateScanner 创建Brother扫描仪实例
func (bsf *BrotherScannerFactory) CreateScanner(usb DeviceInfo, opts DeviceOptions) Scanner {
	return NewBrotherScanner(usb, opts)
}

// 注册Brother扫描仪工厂
func init() {
	GlobalDeviceRegistry.Register(&BrotherScannerFactory{})
}
