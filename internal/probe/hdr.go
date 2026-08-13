package probe

import (
	"strings"
)

func (p *ProbeResult) IsHDR() bool {
	v := p.PrimaryVideoStream()
	if v == nil {
		return false
	}

	transfer := strings.ToLower(v.ColorTransfer)
	primaries := strings.ToLower(v.ColorPrimaries)
	pixFmt := strings.ToLower(v.PixFmt)

	if strings.Contains(transfer, "smpte2084") || strings.Contains(transfer, "arib-std-b67") || strings.Contains(transfer, "hlg") {
		return true
	}
	if strings.Contains(primaries, "bt2020") && (strings.Contains(pixFmt, "10") || strings.Contains(pixFmt, "12")) {
		return true
	}
	if v.BitsPerRawSample == "10" || v.BitsPerRawSample == "12" || strings.Contains(pixFmt, "p10") || strings.Contains(pixFmt, "p12") {
		if strings.Contains(primaries, "bt2020") || strings.Contains(v.ColorSpace, "bt2020") {
			return true
		}
	}
	return false
}

func (p *ProbeResult) HDRType() string {
	if !p.IsHDR() {
		return "SDR"
	}
	v := p.PrimaryVideoStream()
	if v == nil {
		return "SDR"
	}

	transfer := strings.ToLower(v.ColorTransfer)
	if strings.Contains(transfer, "smpte2084") {
		return "HDR10 (PQ)"
	}
	if strings.Contains(transfer, "arib-std-b67") || strings.Contains(transfer, "hlg") {
		return "HLG"
	}
	return "HDR (BT.2020)"
}

func ToneMapFilter() string {
	return "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease,scale=trunc(iw/2)*2:trunc(ih/2)*2,format=gbrpf32le,tonemap=hable:desat=0.5,format=yuv420p"
}
