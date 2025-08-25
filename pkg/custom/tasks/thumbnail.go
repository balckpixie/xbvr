package tasks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/ffprobe"
	"github.com/xbapps/xbvr/pkg/models"

	"github.com/xbapps/xbvr/pkg/tasks"

	shared "github.com/xbapps/xbvr/pkg/custom/shared"
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

				// ファイル削除（存在しなければ無視）
				rmerr := os.Remove(destFile)
				if rmerr != nil {
					if !os.IsNotExist(rmerr) {
						log.Error("削除エラー:", rmerr)
					}
				}

				projection := file.VideoProjection
				if vrType, err := shared.DetectVRType(file.GetPath(),time.Second*5); err == nil {
					projection = vrType
				}

				err := RenderThumnbnails(
					file.GetPath(),
					destFile,
					projection,
					config.Config.Custom.ThumbnailParams.Start,
					config.Config.Custom.ThumbnailParams.Interval,
					config.Config.Custom.ThumbnailParams.Resolution,
					config.Config.Custom.ThumbnailParams.UseCUDAEncode,
				)
				if err == nil {
					log.Infof("Thumbnail Rendering File_ID %v - Finish", file.ID)
					jsonBytes, err := json.Marshal(config.Config.Custom.ThumbnailParams)
					if err != nil {
						log.Warn(err)
					} else {
						jsonString := string(jsonBytes)
						file.ThumbnailParameters = jsonString
						file.HasThumbnail = true

						if err := file.Save(); err != nil {
							log.Warnf("failed to save file %v: %v", file.ID, err)
						} else {
							log.Infof("Thumbnails generated File_ID %v - Saved", file.ID)
						}
					}
				} else {
					log.Warn(err)
				}
			}
		}
	}
	log.Infof("Thumnbnails generate task finishd")
}

func RenderThumnbnails(inputFile string, destFile string, videoProjection string, startTime int, interval int, resolution int, useCUDA bool) error {

	os.MkdirAll(common.VideoThumbnailDir, os.ModePerm)
	// Get video duration
	ffdata, err := ffprobe.GetProbeData(inputFile, time.Second*10)
	if err != nil {
		return err
	}
	dur := ffdata.Format.DurationSeconds
	row := int((dur-float64(startTime))/(20*float64(interval))) + 1
	crop := GetCropFilter(videoProjection)
	vfArgs := fmt.Sprintf("crop=%v,scale=%v:-1:flags=lanczos,fps=1/%v:round=near,tile=20x%v", crop, resolution, interval, row)
	ss := int(math.Max(0, float64(startTime-1)))

	var args []string
	if isCUDAEnabled() && useCUDA{
		args = []string{
			"-y",
			"-hwaccel", "cuda",
			// "-i", inputFile,
			"-ss", strconv.Itoa(ss),
			"-skip_frame",
			"nokey",
			"-i", inputFile,
			"-vf", vfArgs,
			"-vframes", "1",
			"-q:v", "3",
			destFile,
		}
	} else {
		args = []string{
			"-y",
			// "-i", inputFile,
			"-ss", strconv.Itoa(ss),
			"-skip_frame",
			"nokey",
			"-i", inputFile,
			"-vf", vfArgs,
			"-vframes", "1",
			"-q:v", "3",
			destFile,
		}
	}

	log.Infof("Args: %s\n", args) 
	cmd := tasks.BuildCmdEx(tasks.GetBinPath("ffmpeg"), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		log.Error("Error:", err)
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

func GetCropFilter(vrType string) string {
	switch vrType {
	case "flat":
		// 通常動画: そのまま
		return "iw:ih:0:0"
	case "180_sbs":
		// 左右分割 → 右目 (例: 左側を使う場合は x=0)
		return "iw/2:ih:0:ih"
	case "180_tb":
		// 上下分割 → 上 (例: 下を使う場合は y=ih/2)
		return "iw:ih/2:0:0"
	case "180_mono":
		// 180 mono → 全体
		return "iw:ih:0:0"
	case "360_sbs":
		// 360 SBS → 左右分割、左を利用
		return "iw/2:ih:0:0"
	case "360_tb":
		// 360 TB → 上下分割、上を利用
		return "iw:ih/2:0:0"
	case "360_mono":
		// 360 mono → 全体
		return "iw:ih:0:0"
	default:
		// 不明な場合はそのまま
		return "iw:ih:0:0"
	}
}


func RenderThumbnailTile(inputFile string, destFile string, videoProjection string, startTime int, interval int, resolution int, useCUDA bool) error {
	// 出力ディレクトリ作成
	os.MkdirAll(filepath.Dir(destFile), os.ModePerm)

	// 動画情報取得
	ffdata, err := ffprobe.GetProbeData(inputFile, time.Second*10)
	if err != nil {
		return err
	}
	dur := ffdata.Format.DurationSeconds

	// crop フィルタ
	crop := GetCropFilter(videoProjection)

	// CUDA 使用可否を事前に確認
	cudaEnabled := isCUDAEnabled() && useCUDA

	// キーフレーム取得
	keyframes, err := getKeyframes(inputFile)
	if err != nil {
		return err
	}

	// タイル列数・行数
	cols := 20
	totalFrames := int((dur - float64(startTime)) / float64(interval))
	if totalFrames < 1 {
		totalFrames = 1
	}
	rows := (totalFrames + cols - 1) / cols // 切り上げ

	// interval 秒ごとに最も近いキーフレームを選択
	var thumbs []image.Image
	for t := startTime; t < int(dur) && len(thumbs) < cols*rows; t += interval {
		closest := findClosestKeyframe(keyframes, float64(t))
		img, err := getThumbnailImageWithCrop(inputFile, closest, crop, resolution, cudaEnabled)
		if err != nil {
			fmt.Println("サムネイル生成失敗:", err)
			continue
		}
		thumbs = append(thumbs, img)
	}

	if len(thumbs) == 0 {
		return fmt.Errorf("サムネイルが1枚も生成されませんでした")
	}

	// タイル画像作成
	tw := thumbs[0].Bounds().Dx()
	th := thumbs[0].Bounds().Dy()
	tileImg := image.NewRGBA(image.Rect(0, 0, cols*tw, rows*th))
	for i, img := range thumbs {
		x := (i % cols) * tw
		y := (i / cols) * th
		draw.Draw(tileImg, image.Rect(x, y, x+tw, y+th), img, image.Point{0, 0}, draw.Over)
	}

	// JPEG 保存
	outFile, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return jpeg.Encode(outFile, tileImg, &jpeg.Options{Quality: 85})
}


func getThumbnailImageWithCrop(videoPath string, ss float64, crop string, resolution int, useCUDA bool) (image.Image, error) {
	ssStr := fmt.Sprintf("%.3f", ss)

	args := []string{"-y"}
	if useCUDA {
		args = append(args, "-hwaccel", "cuda")
	}
	args = append(args,
		"-ss", ssStr,
		"-i", videoPath,
		"-vframes", "1",
		"-vf", fmt.Sprintf("crop=%s,scale=%d:-1:flags=lanczos", crop, resolution),
		"-f", "image2pipe",
		"-vcodec", "mjpeg", "-",
	)

	// tasks.BuildCmdEx を使用してコマンド作成
	cmd := tasks.BuildCmdEx(tasks.GetBinPath("ffmpeg"), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	// コマンド実行
	if err := cmd.Run(); err != nil {
		fmt.Printf("FFmpeg Error: %s\n", stderr.String())
		return nil, err
	}

	img, err := jpeg.Decode(&stdout)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func getKeyframes(videoPath string) ([]float64, error) {
	cmd := exec.Command(tasks.GetBinPath("ffprobe"),
		"-v", "error",
		"-select_streams", "v:0",
		"-skip_frame", "nokey",
		"-show_frames",
		"-show_entries", "frame=pkt_pts_time",
		"-of", "csv=p=0",
		videoPath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	lines := bytes.Split(out.Bytes(), []byte{'\n'})
	keyframes := make([]float64, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if t, err := strconv.ParseFloat(string(line), 64); err == nil {
			keyframes = append(keyframes, t)
		}
	}
	return keyframes, nil
}

func findClosestKeyframe(keyframes []float64, target float64) float64 {
	if len(keyframes) == 0 {
		return target
	}
	closest := keyframes[0]
	minDiff := absFloat(target - closest)
	for _, k := range keyframes[1:] {
		if diff := absFloat(target - k); diff < minDiff {
			minDiff = diff
			closest = k
		}
	}
	return closest
}

func absFloat(a float64) float64 {
	if a < 0 {
		return -a
	}
	return a
}
