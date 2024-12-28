package models

import (
<<<<<<< HEAD
	"fmt"
=======
>>>>>>> f9a5af58215e2f45b39000e9b63ae1ef22d12ac1
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
<<<<<<< HEAD
	"github.com/xbapps/xbvr/pkg/common"
=======
>>>>>>> f9a5af58215e2f45b39000e9b63ae1ef22d12ac1
)

type File struct {
	ID        uint      `gorm:"primary_key" json:"id" xbvrbackup:"-"`
	CreatedAt time.Time `json:"created_at" xbvrbackup:"-"`
	UpdatedAt time.Time `json:"updated_at" xbvrbackup:"-"`

	VolumeID    uint      `json:"volume_id" xbvrbackup:"-"`
	Volume      Volume    `json:"-" xbvrbackup:"-"`
	Path        string    `json:"path" xbvrbackup:"path"`
	Filename    string    `json:"filename" xbvrbackup:"filename"`
	Size        int64     `json:"size" xbvrbackup:"size"`
	OsHash      string    `json:"oshash" xbvrbackup:"oshash"`
	CreatedTime time.Time `json:"created_time" xbvrbackup:"created_time"`
	UpdatedTime time.Time `json:"updated_time" xbvrbackup:"updated_time"`

	Type    string `json:"type" xbvrbackup:"type"`
	SceneID uint   `gorm:"index" json:"scene_id" xbvrbackup:"-"`
	Scene   Scene  `json:"-" xbvrbackup:"-"`

	VideoWidth           int     `json:"video_width" xbvrbackup:"video_width"`
	VideoHeight          int     `json:"video_height" xbvrbackup:"video_height"`
	VideoBitRate         int     `json:"video_bitrate" xbvrbackup:"video_bitrate"`
	VideoAvgFrameRate    string  `json:"-" xbvrbackup:"video_avgfps"`
	VideoAvgFrameRateVal float64 `json:"video_avgfps_val" xbvrbackup:"video_avgfps_val"`
	VideoCodecName       string  `json:"video_codec_name" xbvrbackup:"video_codec_name"`
	VideoDuration        float64 `json:"duration" xbvrbackup:"duration"`
	VideoProjection      string  `json:"projection" xbvrbackup:"projection"`

	HasHeatmap          bool `json:"has_heatmap" xbvrbackup:"-"`
	IsSelectedScript    bool `json:"is_selected_script" xbvrbackup:"is_selected_script"`
	IsExported          bool `json:"is_exported" xbvrbackup:"-"`
	RefreshHeatmapCache bool `json:"refresh_heatmap_cache" xbvrbackup:"-"`
<<<<<<< HEAD

	HasThumbnail bool `json:"has_thumbnail" gorm:"default:false" xbvrbackup:"-"`
=======
>>>>>>> f9a5af58215e2f45b39000e9b63ae1ef22d12ac1
}

func (f *File) GetPath() string {
	return filepath.Join(f.Path, f.Filename)
}

func (f *File) Save() error {
	db, _ := GetDB()
	defer db.Close()

	var err error = retry.Do(
		func() error {
			err := db.Save(&f).Error
			if err != nil {
				return err
			}
			return nil
		},
	)

	if err != nil {
		log.Fatal("Failed to save ", err)
	}

	return nil
}

func (f *File) GetIfExistByPK(id uint) error {
	db, _ := GetDB()
	defer db.Close()

	return db.Where(&File{ID: id}).First(f).Error
}

func (f *File) Exists() bool {
	switch f.Volume.Type {
	case "local":
		if _, err := os.Stat(f.GetPath()); os.IsNotExist(err) {
			return false
		}
		return true
	case "putio":
		// NOTE: we're assuming files weren't removed via Put.io web UI, so there's no need to check
		return true
	default:
<<<<<<< HEAD
		// return false
		if _, err := os.Stat(f.GetPath()); os.IsNotExist(err) {
			return false
		}
		return true
=======
		return false
>>>>>>> f9a5af58215e2f45b39000e9b63ae1ef22d12ac1
	}
}

func (f *File) CalculateFramerate() error {
	v1, err := strconv.ParseFloat(strings.Split(f.VideoAvgFrameRate, "/")[0], 64)
	if err != nil {
		return err
	}

	v2, err := strconv.ParseFloat(strings.Split(f.VideoAvgFrameRate, "/")[1], 64)
	if err != nil {
		return err
	}

	f.VideoAvgFrameRateVal = math.Ceil(v1 / v2)
	return nil
}
<<<<<<< HEAD

func (o *File) ThumbnailExists() bool {

	if _, err := os.Stat(filepath.Join(common.VideoThumbnailDir, fmt.Sprintf("%v.jpg", strconv.FormatUint(uint64(o.ID), 10)))); os.IsNotExist(err) {
		return false
	}
	return true
}
=======
>>>>>>> f9a5af58215e2f45b39000e9b63ae1ef22d12ac1
