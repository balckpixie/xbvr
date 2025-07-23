package tasks

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/ffprobe"
	"github.com/xbapps/xbvr/pkg/models"

	"github.com/xbapps/xbvr/pkg/tasks"
)

func GenerateThumnbnails(endTime *time.Time) {
	if !models.CheckLock("thumnbnails") {
		models.CreateLock("thumnbnails")
		defer models.RemoveLock("thumnbnails")
		log.Infof("Generating thumnbnails")
		db, _ := models.GetDB()
		defer db.Close()

		var files []models.File
		db.Model(&models.File{}).Where("type = ?", "video").Where("scene_id != ?", 0).Where("has_thumbnail = ?", false).Order("created_time desc").Find(&files)
		for _, file := range files {

			if endTime != nil && time.Now().After(*endTime) {
				return
			}
			log.Infof("Thumbnail Checking %v", file.ID)
			if file.Exists() && file.Type == "video" {
				// log.Infof("Thumbnail Rendering File_ID %v", strconv.FormatUint(uint64(file.ID), 10))
				log.Infof("Thumbnail Rendering File_ID %v - Start", file.ID)
				destFile := filepath.Join(common.VideoThumbnailDir, strconv.FormatUint(uint64(file.ID), 10)+".jpg")
				err := RenderThumnbnails(
					file.GetPath(),
					destFile,
					file.VideoProjection,
					config.Config.Library.Preview.StartTime,
					config.Config.Library.Preview.SnippetLength,
					config.Config.Library.Preview.SnippetAmount,
					config.Config.Library.Preview.Resolution,
					config.Config.Library.Preview.ExtraSnippet,
				)
				if err == nil {
					log.Infof("Thumbnail Rendering File_ID %v - Finish", file.ID)
					file.HasThumbnail = true
					file.Save()
					// break
				} else {
					log.Warn(err)
				}
			}
		}
	}
	log.Infof("Thumnbnails generated")
}

func RenderThumnbnails(inputFile string, destFile string, videoProjection string, startTime int, snippetLength float64, snippetAmount int, resolution int, extraSnippet bool) error {

	os.MkdirAll(common.VideoThumbnailDir, os.ModePerm)

	// Get video duration
	ffdata, err := ffprobe.GetProbeData(inputFile, time.Second*10)
	if err != nil {
		return err
	}
	vs := ffdata.GetFirstVideoStream()
	dur := ffdata.Format.DurationSeconds

	row := (int)((dur-5)/600) + 1

	crop := "iw/2:ih:iw/2:ih" // LR videos
	if vs.Height == vs.Width {
		crop = "iw/2:ih/2:iw/4:ih/2" // TB videos
	}
	if videoProjection == "flat" {
		crop = "iw:ih:iw:ih" // LR videos
	}
	// Mono 360 crop args: (no way of accurately determining)
	// "iw/2:ih:iw/4:ih"
	vfArgs := fmt.Sprintf("crop=%v,scale=%v:-1:flags=lanczos,fps=fps=1/%v:round=down,tile=20x%v", crop, 200, 30, row)

	args := []string{}
	if isCUDAEnabled() {
		args = []string{
			"-y",
			"-ss", "5",
			"-hwaccel", "cuda",
			"-skip_frame",
			"nokey",
			"-i", inputFile,
			// "-t", "60",
			"-vf", vfArgs,
			// "-frame_pts", "true",
			"-q:v", "3",
			// "-pix_fmt", "rgb24",
			// "-c:v", "mjpeg",
			//"-f", "image2pipe",
			//"-",
			destFile,
		}
		log.Infof("Use Internal hwaccel decoders CUDA")
	} else {
		args = []string{
			"-y",
			"-ss", "5",
			// "-hwaccel", "cuda",
			"-skip_frame",
			"nokey",
			"-i", inputFile,
			// "-t", "60",
			"-vf", vfArgs,
			// "-frame_pts", "true",
			"-q:v", "3",
			// "-pix_fmt", "rgb24",
			// "-c:v", "mjpeg",
			//"-f", "image2pipe",
			//"-",
			destFile,
		}
	}

	cmd := tasks.BuildCmdEx(tasks.GetBinPath("ffmpeg"), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Stderr: %s\n", stderr.String()) // Stderrを出力
		return err
	}
	return nil
}

func isCUDAEnabled() bool {
	args := []string{
		"-hide_banner",
		"-hwaccels",
	}
	cmd := tasks.BuildCmdEx(tasks.GetBinPath("ffmpeg"), args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false // エラーが発生した場合はCUDAが利用できないとみなす
	}

	// コマンドの出力を確認し、CUDAが利用可能かどうかを判定する
	output := out.String()
	return strings.Contains(output, "cuda")
}
