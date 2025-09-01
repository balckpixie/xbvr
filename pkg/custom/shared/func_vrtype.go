package shared

import (
	"time"
	
	"github.com/xbapps/xbvr/pkg/ffprobe"
)

// ProbeData の構造体は既存定義に合わせてください
// ここでは Streams[i].Width, Height, Tags を使用する想定です
type ProbeData struct {
	Streams []struct {
		Width  int               `json:"width"`
		Height int               `json:"height"`
		Tags   map[string]string `json:"tags"`
	} `json:"streams"`
}

// DetectVRType は GetProbeData を利用してVR種別を判定
func DetectVRType(filePath string, timeout time.Duration) (string, error) {
	data, err := ffprobe.GetProbeData(filePath, timeout)
	if err != nil {
		return "UNKNOWN", err
	}
	if len(data.Streams) == 0 {
		return "UNKNOWN", nil
	}

	stream := data.Streams[0]
	w, h := stream.Width, stream.Height
	if h == 0 {
		return "UNKNOWN", nil
	}
	ratio := float64(w) / float64(h)

	// --- 0. 特殊解像度ヒューリスティック ---
	specialCases := map[[2]int]string{
		{8192, 4096}: "180_sbs",
		{7680, 3840}: "180_sbs",
		{5760, 2880}: "180_sbs",
		{3840, 1920}: "180_sbs",
		{1920, 960}:  "180_mono",
	}
	if val, ok := specialCases[[2]int{w, h}]; ok {
		return val, nil
	}

	proj := stream.Tags.Projection
	mode := stream.Tags.StereoMode
	hasProj := proj != ""
	hasMode := mode != ""

	if hasProj {
		if proj == "equirectangular" { // 360°
			if hasMode {
				if mode == "top_bottom" {
					return "360_tb", nil
				}
				if mode == "left_right" {
					return "360_sbs", nil
				}
			}
			return "360_mono", nil
		} else if proj == "rectilinear" { // 180°
			if hasMode {
				if mode == "left_right" {
					return "180_sbs", nil
				}
				if mode == "top_bottom" {
					return "180_tb", nil
				}
			}
			return "180_mono", nil
		}
	}

	// --- 2. 解像度から判定 ---
	// 16:9 → flat に固定
	if ratio > 1.75 && ratio < 1.79 {
		return "flat", nil
	}

	switch {
	case ratio > 1.95 && ratio < 2.05:
		return "360_mono", nil
	case ratio > 0.95 && ratio < 1.05:
		return "360_tb", nil
	case ratio > 3.5:
		return "180_sbs", nil
	case ratio > 0.8 && ratio < 1.2 && h > w: // 縦長
		return "180_tb", nil
	case ratio > 1.6 && ratio < 2.0: // 180 mono (16:9は除外済)
		return "180_mono", nil
	default:
		return "flat", nil
	}
}
