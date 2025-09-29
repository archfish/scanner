package scanner

var DefaultScanOptions = ScanOptions{
	DPI:    400,
	Mode:   ScanModeCGRAY,
	Top:    0,
	Left:   0,
	Width:  211.881,
	Height: 355.567,
}

type ScanOptions struct {
	DPI  uint16
	Mode ScanMode
	// TODO: Compression string, currently JPEG
	// All in [mm]
	Top    float64
	Left   float64
	Width  float64
	Height float64
}
