package shared

import (
	"path/filepath"
	"strings"
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
// DetectVRType は VR 動画の種類を判定
// 戻り値: 180_mono / 180_sbs / 180_tb / 360_mono / 360_sbs / 360_tb / flat / fisheye / mkx200 ... / UNKNOWN
func DetectVRType(filePath string, timeout time.Duration) (string, error) {
	
	// --- 1. ファイル名から判定 ---
	base := strings.ToLower(filepath.Base(filePath))
	nameparts := strings.FieldsFunc(base, func(r rune) bool {
		return r == '_' || r == '-' || r == '.' || r == ' '
	})

	for i, part := range nameparts {
		if part == "mkx200" || part == "mkx220" || part == "rf52" || part == "fisheye190" || part == "vrca220" || part == "flat" {
			return part, nil
		} else if part == "fisheye" || part == "f180" || part == "180f" {
			return "fisheye", nil
		} else if i < len(nameparts)-1 {
			combined := part + "_" + nameparts[i+1]
			if combined == "mono_360" || combined == "mono_180" {
				return nameparts[i+1] + "_mono", nil
			}
			if combined == "360_mono" || combined == "180_mono" {
				return part + "_mono", nil
			}
		}
	}

	// --- 2. ffprobe のタグから判定 ---
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

	proj := stream.Tags.Projection
	mode := stream.Tags.StereoMode
	if proj != "" || mode != "" {
		switch proj {
		case "equirectangular": // 360°
			switch mode {
			case "left_right":
				return "360_sbs", nil
			case "top_bottom":
				return "360_tb", nil
			default:
				return "360_mono", nil
			}
		case "rectilinear": // 180°
			switch mode {
			case "left_right":
				return "180_sbs", nil
			case "top_bottom":
				return "180_tb", nil
			default:
				return "180_mono", nil
			}
		}
		if mode == "left_right" {
			if ratio > 1.9 && ratio < 2.1 {
				return "180_sbs", nil
			}
			return "360_sbs", nil
		}
		if mode == "top_bottom" {
			if ratio > 0.9 && ratio < 1.1 {
				return "180_tb", nil
			}
			return "360_tb", nil
		}
	}

	// --- 3. よくある解像度から判定 ---
	specialCases := map[[2]int]string{
		{8192, 4096}: "180_sbs",
		{7680, 3840}: "180_sbs",
		{5760, 2880}: "180_sbs",
		{4096, 2048}: "180_sbs",
		{3840, 1920}: "180_sbs",
		{1920, 1080}: "180_mono",
	}
	if val, ok := specialCases[[2]int{w, h}]; ok {
		return val, nil
	}

	// --- 4. 解像度ヒューリスティック ---
	switch {
	case ratio > 3.5 && ratio < 4.1:
		return "360_sbs", nil
	case ratio > 1.9 && ratio < 2.1:
		if h <= 2160 {
			return "180_sbs", nil
		}
		return "360_mono", nil
	case ratio > 0.95 && ratio < 1.05:
		if h <= 2160 {
			return "180_tb", nil
		}
		return "360_tb", nil
	case ratio > 1.6 && ratio < 1.8:
		return "180_mono", nil
	case ratio > 1.75 && ratio < 1.79:
		return "flat", nil
	default:
		return "flat", nil
	}
}