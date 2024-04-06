package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/markphelps/optional"
	"github.com/xbapps/xbvr/pkg/common"
	"github.com/xbapps/xbvr/pkg/models"
)

type RequestMatchFile struct {
	SceneID string `json:"scene_id"`
	FileID  uint   `json:"file_id"`
}

type RequestUnmatchFile struct {
	FileID uint `json:"file_id"`
}

type RequestRenameFile struct {
	FileID      uint   `json:"file_id"`
	NewFilename string `json:"filename"`
}

type RequestFileList struct {
	State       optional.String   `json:"state"`
	CreatedDate []optional.String `json:"createdDate"`
	Sort        optional.String   `json:"sort"`
	Resolutions []optional.String `json:"resolutions"`
	Framerates  []optional.String `json:"framerates"`
	Bitrates    []optional.String `json:"bitrates"`
	Filename    optional.String   `json:"filename"`
}

type FilesResource struct{}

func (i FilesResource) WebService() *restful.WebService {
	tags := []string{"Files"}

	ws := new(restful.WebService)

	ws.Path("/api/files").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.POST("/list").To(i.listFiles).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/match").To(i.matchFile).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/unmatch").To(i.unmatchFile).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.DELETE("/file/{file-id}").To(i.removeFile).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	ws.Route(ws.POST("/rename").To(i.renameFile).
		Metadata(restfulspec.KeyOpenAPITags, tags))

	return ws
}

func (i FilesResource) listFiles(req *restful.Request, resp *restful.Response) {
	db, _ := models.GetDB()
	defer db.Close()

	var r RequestFileList
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	var files []models.File
	tx := db.Model(&files)

	// State
	switch r.State.OrElse("") {
	case "matched":
		tx = tx.Where("files.scene_id != 0")
	case "unmatched":
		tx = tx.Where("files.scene_id = 0")
	}

	// Resolution
	resolutionClauses := []string{}
	if len(r.Resolutions) > 0 {
		for _, resolution := range r.Resolutions {
			if resolution.OrElse("") == "below4k" {
				resolutionClauses = append(resolutionClauses, "video_height between 0 and 1899")
			}
			if resolution.OrElse("") == "4k" {
				resolutionClauses = append(resolutionClauses, "video_height between 1900 and 2449")
			}
			if resolution.OrElse("") == "5k" {
				resolutionClauses = append(resolutionClauses, "video_height between 2450 and 2899")
			}
			if resolution.OrElse("") == "6k" {
				resolutionClauses = append(resolutionClauses, "video_height between 2900 and 3299")
			}
			if resolution.OrElse("") == "above6k" {
				resolutionClauses = append(resolutionClauses, "video_height between 3300 and 9999")
			}
		}
		tx = tx.Where("(" + strings.Join(resolutionClauses, " OR ") + ") AND video_height != 0")
	}

	// Bitrate
	bitrateClauses := []string{}
	if len(r.Bitrates) > 0 {
		for _, bitrate := range r.Bitrates {
			if bitrate.OrElse("") == "low" {
				bitrateClauses = append(bitrateClauses, "video_bit_rate between 0 and 14999999")
			}
			if bitrate.OrElse("") == "medium" {
				bitrateClauses = append(bitrateClauses, "video_bit_rate between 15000000 and 24999999")
			}
			if bitrate.OrElse("") == "high" {
				bitrateClauses = append(bitrateClauses, "video_bit_rate between 25000000 and 35000000")
			}
			if bitrate.OrElse("") == "ultra" {
				bitrateClauses = append(bitrateClauses, "video_bit_rate between 35000001 and 999999999")
			}
		}
		tx = tx.Where("(" + strings.Join(bitrateClauses, " OR ") + ") AND video_bit_rate != 0")
	}

	// Framerate
	framerateClauses := []string{}
	if len(r.Framerates) > 0 {
		for _, framerate := range r.Framerates {
			if framerate.OrElse("") == "30fps" {
				framerateClauses = append(framerateClauses, "video_avg_frame_rate_val = 30.0")
			}
			if framerate.OrElse("") == "60fps" {
				framerateClauses = append(framerateClauses, "video_avg_frame_rate_val = 60.0")
			}
			if framerate.OrElse("") == "other" {
				framerateClauses = append(framerateClauses, "(video_avg_frame_rate_val != 30.0 AND video_avg_frame_rate_val != 60.0)")
			}
		}
		tx = tx.Where("(" + strings.Join(framerateClauses, " OR ") + ") AND video_avg_frame_rate_val != 0")
	}

	// Filename
	if len(r.Filename.OrElse("")) > 0 {
		tx = tx.Where("filename like ?", "%"+r.Filename.OrElse("")+"%")
	}

	// Creation date
	if len(r.CreatedDate) == 2 {
		t0, _ := time.Parse(time.RFC3339, r.CreatedDate[0].OrElse(""))
		t1, _ := time.Parse(time.RFC3339, r.CreatedDate[1].OrElse(""))
		tx = tx.Where("files.created_time > ? AND files.created_time < ?", t0, t1)
	}

	// Sorting
	switch r.Sort.OrElse("") {
	case "filename_asc":
		tx = tx.Order("filename asc")
	case "filename_desc":
		tx = tx.Order("filename desc")
	case "created_time_asc":
		tx = tx.Order("created_time asc")
	case "created_time_desc":
		tx = tx.Order("created_time desc")
	case "duration_asc":
		tx = tx.Order("video_duration asc")
	case "duration_desc":
		tx = tx.Order("video_duration desc")
	case "size_asc":
		tx = tx.Order("size asc")
	case "size_desc":
		tx = tx.Order("size desc")
	case "video_height_asc":
		tx = tx.Order("video_height asc")
	case "video_height_desc":
		tx = tx.Order("video_height desc")
	case "video_width_asc":
		tx = tx.Order("video_width asc")
	case "video_width_desc":
		tx = tx.Order("video_width desc")
	case "video_bitrate_asc":
		tx = tx.Order("video_bit_rate asc")
	case "video_bitrate_desc":
		tx = tx.Order("video_bit_rate desc")
	case "video_avgfps_val_asc":
		tx = tx.Order("video_avg_frame_rate_val asc")
	case "video_avgfps_val_desc":
		tx = tx.Order("video_avg_frame_rate_val desc")
	}

	tx.Find(&files)

	resp.WriteHeaderAndEntity(http.StatusOK, files)
}

func (i FilesResource) matchFile(req *restful.Request, resp *restful.Response) {
	db, _ := models.GetDB()
	defer db.Close()

	var r RequestMatchFile
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	// Assign Scene to File
	var scene models.Scene
	err = scene.GetIfExist(r.SceneID)
	if err != nil {
		log.Error(err)
		return
	}

	var f models.File
	err = db.Preload("Volume").Where(&models.File{ID: r.FileID}).First(&f).Error
	if err == nil {
		f.SceneID = scene.ID
		f.Save()
	}

	// Add File to the list of Scene filenames so it will be discovered when file is moved
	var pfTxt []string
	err = json.Unmarshal([]byte(scene.FilenamesArr), &pfTxt)
	if err != nil {
		log.Error(err)
		return
	}

	pfTxt = append(pfTxt, f.Filename)
	tmp, err := json.Marshal(pfTxt)
	if err == nil {
		scene.FilenamesArr = string(tmp)
	}

	models.AddAction(scene.SceneID, "match", "filenames_arr", scene.FilenamesArr)

	// Finally, update scene available/accessible status
	scene.UpdateStatus()

	//ここまで標準処理、以降追加処理（ファイル名を自動付与する）
	RenameFile(scene, f)
	//ここまで追加処理

	resp.WriteHeaderAndEntity(http.StatusOK, nil)
}

func RenameFile(scene models.Scene, file models.File) models.Scene {

	// 元のファイル名から sceneNo を取得
	sceneNo := GetSceneNo(file)
	// 拡張子を取得
	extension := filepath.Ext(file.Filename)
	// Actor.Name をカンマで結合した文字列を生成
	var actorNames []string
	for _, actor := range scene.Cast {
		actorNames = append(actorNames, actor.Name)
	}
	var castString string
	if len(actorNames) > 0 {
		castString = strings.Join(actorNames, ",")
	} else {
		castString = "Unknown"
	}
	if len(sceneNo) > 0 {
		sceneNo = "-" + sceneNo
	}
	title := scene.Title
	if trimPrefix(scene.SceneID, title) == "" {
		title = scene.Synopsis
	}
	newFileName := fmt.Sprintf("%s｜%s%s｜%s%s", castString, scene.SceneID, sceneNo, trimPrefix(scene.SceneID, title), extension)

	db, _ := models.GetDB()
	defer db.Close()

	vol := models.Volume{}
	err := db.First(&vol, file.VolumeID).Error

	if err == nil {
		if len(actorNames) > 1 {
			castString = "_Group"
		}
		newPath := filepath.Join(vol.Path, sanitizeFilename(castString))
		newFileName = sanitizeFilename(newFileName)
		if filepath.Join(file.Path, file.Filename) != filepath.Join(newPath, newFileName) {
			scene = renameFileByFileId(uint(file.ID), newPath, newFileName)
		}
	}

	return scene
}

func sanitizeFilename(filename string) string {
	// 無効な文字を取り除く正規表現
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)

	// 無効な文字を削除し、指定した文字列で置き換える
	sanitizedFilename := invalidChars.ReplaceAllString(filename, "-")

	// 空白をアンダースコアに置き換える
	// sanitizedFilename = strings.ReplaceAll(sanitizedFilename, " ", "_")

	return sanitizedFilename
}

func trimPrefix(a, b string) string {
	// 文字列Bが文字列Aで始まる場合、文字列Bから文字列Aを取り除く
	if strings.HasPrefix(b, a) {
		trimmed := strings.TrimPrefix(b, a)
		return strings.TrimSpace(trimmed)
	}

	// 文字列Bが文字列Aを含まない場合は、文字列Bをそのまま返す
	return b
}

// func GetSceneNo(file models.File) string {
// 	// ダミーの実装（実際のファイル名から sceneNo を適切に取得する実装が必要）
// 	return "1"
// }

func GetSceneNo(file models.File) string {

	// pattern := `[a-zA-Z0-9]{2,6}-\d{2,6}`
	// base := filepath.Base(file.Filename)
	// input := base[:len(base)-len(filepath.Ext(base))]
	// re := regexp.MustCompile(pattern)
	// matches := re.FindStringSubmatch(input)

	// if len(matches) > 0 {
	// 	// ② 一致した部分があれば、それを返す
	// 	secondPattern := `-(R\d{1,2})`
	// 	re2 := regexp.MustCompile(secondPattern)
	// 	matches2 := re2.FindStringSubmatch(input)
	// 	if len(matches2) > 0 {
	// 		return matches2[1] // ③ R1 にマッチする部分を返す
	// 	}
	// }

	pattern := `([a-zA-Z0-9]{2,6}-\d{2,6})(.*)`
	base := filepath.Base(file.Filename)
	input := base[:len(base)-len(filepath.Ext(base))]

	// 正規表現パターンに一致する部分を検索
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)

	if len(matches) > 0 {
		// matchedPart := matches[1]  // 最初のサブマッチンググループ: ([a-zA-Z0-9]{2,6}-\d{2,6})
		nextPart := matches[2]
		secondPattern := `-([a-zA-Z0-9]{1})`
		re2 := regexp.MustCompile(secondPattern)
		matches2 := re2.FindStringSubmatch(nextPart)
		if len(matches2) > 0 {
			return matches2[1]
		}
	}

	// ④ 一致しない場合は、文字列の最後の１文字を返す（ただし数字の場合のみ）
	if len(input) > 0 {
		lastChar := input[len(input)-1:]
		if _, err := strconv.Atoi(lastChar); err == nil {
			if len(matches) > 0 {
				if areEqualBackwards(input, matches[1]) {
					return ""
				}
			}
			return lastChar
		}
	}

	return ""
}

func areEqualBackwards(strA, strB string) bool {
    lenA, lenB := len(strA), len(strB)
    i, j := lenA - 1, lenB - 1
    
    for i >= 0 && j >= 0 {
        if strA[i] != strB[j] {
            return false
        }
        i--
        j--
    }
    
    return true
}

func (i FilesResource) unmatchFile(req *restful.Request, resp *restful.Response) {
	db, _ := models.GetDB()
	defer db.Close()

	var r RequestUnmatchFile
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}

	var f models.File
	err = db.Preload("Volume").Where(&models.File{ID: r.FileID}).First(&f).Error
	var sceneID uint = 0
	if err == nil {
		sceneID = f.SceneID
		if sceneID != 0 {
			f.SceneID = 0
			f.Save()
		}

	}

	var scene models.Scene
	if sceneID != 0 {
		err = scene.GetIfExistByPK(sceneID)
		if err != nil {
			log.Error(err)
			return
		}

		// Remove File from the list of Scene filenames so it will be not be auto-matched again
		var pfTxt []string
		err = json.Unmarshal([]byte(scene.FilenamesArr), &pfTxt)
		if err != nil {
			log.Error(err)
			return
		}

		var newFilenamesArr []string

		for _, fn := range pfTxt {
			if fn != f.Filename {
				newFilenamesArr = append(newFilenamesArr, fn)
			}
		}

		tmp, err := json.Marshal(newFilenamesArr)
		if err == nil {
			scene.FilenamesArr = string(tmp)
		}

		models.AddAction(scene.SceneID, "unmatch", "filenames_arr", scene.FilenamesArr)

		// Finally, update scene available/accessible status
		scene.UpdateStatus()
	}

	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}

func (i FilesResource) renameFile(req *restful.Request, resp *restful.Response) {
	var r RequestRenameFile
	err := req.ReadEntity(&r)
	if err != nil {
		log.Error(err)
		return
	}
	scene := renameFileByFileId(uint(r.FileID), "", r.NewFilename)
	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}

func renameFileByFileId(fileId uint, newPath string, newfilename string) models.Scene {

	var scene models.Scene
	var file models.File
	db, _ := models.GetDB()
	defer db.Close()

	err := db.Preload("Volume").Where(&models.File{ID: fileId}).First(&file).Error
	var oldFilename = file.Filename
	if err == nil {
		// 拡張子を除いた部分を取得
		nameOnly := newfilename[:len(newfilename)-len(filepath.Ext(newfilename))]
		ext := filepath.Ext(newfilename)
		const maxFilenameLength = 90
		if utf8.RuneCountInString(nameOnly) > maxFilenameLength-4 {
			// nameOnly = nameOnly[:len(nameOnly)-len(filepath.Ext(nameOnly))]
			nameOnly = string([]rune(nameOnly)[:maxFilenameLength])
		}
		//暫定用
		if ext == "" {
			ext = ".mp4"
		}

		newfilename = nameOnly + ext

		targetFilename := filepath.Join(newPath, newfilename)
		if newPath == "" {
			targetFilename = filepath.Join(file.Path, newfilename)
		}
		log.Infof("Renaming file %s", filepath.Join(file.Path, file.Filename))
		renamed := false
		switch file.Volume.Type {
		case "local":
			newfilename, err = RenameFileNoDuplicate(filepath.Join(file.Path, file.Filename), targetFilename)
			// err := os.Rename(filepath.Join(file.Path, file.Filename), filepath.Join(file.Path, newfilename))
			if err == nil {
				renamed = true
			} else {
				log.Errorf("error renaming file: %v", err)
			}
		case "putio":
			// id, err := strconv.ParseInt(file.Path, 10, 64)
			// if err != nil {
			// 	return scene
			// }
			// client := file.Volume.GetPutIOClient()
			// err = client.Files.Delete(context.Background(), id)
			// if err == nil {
			// 	renamed = true
			// } else {
			log.Errorf("error renaming file %v", err)
			// }
		}

		if renamed {
			dir := filepath.Dir(newfilename)
			base := filepath.Base(newfilename)
			db.Model(&file).Where("id = ?", fileId).Update("filename", base).Update("path", dir)
			if file.SceneID != 0 {
				scene.GetIfExistByPK(file.SceneID)
				updateFilenameAtrr(scene, oldFilename, base)
				scene.UpdateStatus()
			}

		}
	} else {
		log.Errorf("error renaming file %v", err)
	}
	return scene
}

func RenameFileNoDuplicate(filePath, newFileName string) (renamedFileName string, err error) {
	// ファイル名のディレクトリパスと拡張子を取得
	// dir := filepath.Dir(filePath)
	ext := filepath.Ext(newFileName)
	base := strings.TrimSuffix(newFileName, ext)

	// リネーム後のファイル名が重複しないようにする
	i := 1
	for {
		// newPath := filepath.Join(dir, newFileName)
		newPath := newFileName
		_, err := os.Stat(newPath)
		if err != nil {
			// エラーがなければ、そのファイル名を使用
			renamedFileName = newFileName
			break
		}
		// 重複する場合は、ファイル名の末尾に番号を付与してリトライ
		i++
		newFileName = fmt.Sprintf("%s(%d)%s", base, i, ext)
	}

	// フォルダが存在しない場合は作成する
	if err := os.MkdirAll(filepath.Dir(renamedFileName), os.ModePerm); err != nil {
		return "", err
	}

	// リネーム実行
	err = os.Rename(filePath, renamedFileName)
	if err != nil {
		return "", err
	}

	return newFileName, nil
}

func updateFilenameAtrr(scene models.Scene, oldFilename string, newFilename string) {
	var pfTxt []string
	err := json.Unmarshal([]byte(scene.FilenamesArr), &pfTxt)
	if err != nil {
		log.Error(err)
		return
	}

	replaceElement(pfTxt, oldFilename, newFilename)
	tmp, err := json.Marshal(pfTxt)
	if err == nil {
		scene.FilenamesArr = string(tmp)
	}

	models.AddAction(scene.SceneID, "match", "filenames_arr", scene.FilenamesArr)

	// Finally, update scene available/accessible status
	scene.UpdateStatus()
}

// スライス内の古い要素を新しい要素で置き換える関数
func replaceElement(slice []string, oldElement string, newElement string) {
	// スライスをループして古い要素を探す
	for i, element := range slice {
		if element == oldElement {
			// 古い要素が見つかったら、新しい要素で置き換える
			slice[i] = newElement
			return
		}
	}
	// 古い要素が見つからない場合は何もしない
}

func (i FilesResource) removeFile(req *restful.Request, resp *restful.Response) {
	fileId, err := strconv.Atoi(req.PathParameter("file-id"))
	if err != nil {
		return
	}
	scene := removeFileByFileId(uint(fileId))
	resp.WriteHeaderAndEntity(http.StatusOK, scene)
}
func removeFileByFileId(fileId uint) models.Scene {

	var scene models.Scene
	var file models.File
	db, _ := models.GetDB()
	defer db.Close()

	err := db.Preload("Volume").Where(&models.File{ID: fileId}).First(&file).Error
	if err == nil {

		log.Infof("Deleting file %s", filepath.Join(file.Path, file.Filename))
		deleted := false
		switch file.Volume.Type {
		case "local":
			err := os.Remove(filepath.Join(file.Path, file.Filename))
			if err == nil {
				deleted = true
			} else {
				log.Errorf("error deleting file: %v", err)
			}
		case "putio":
			id, err := strconv.ParseInt(file.Path, 10, 64)
			if err != nil {
				return scene
			}
			client := file.Volume.GetPutIOClient()
			err = client.Files.Delete(context.Background(), id)
			if err == nil {
				deleted = true
			} else {
				log.Errorf("error deleting file %v", err)
			}
		}

		if deleted {
			if file.HasThumbnail {
				thumbFile := filepath.Join(common.VideoThumbnailDir, strconv.FormatUint(uint64(file.ID), 10)+".jpg")
				err := os.Remove(thumbFile)
				if err == nil {
					deleted = true
				} else {
					log.Errorf("error deleting thumbnail file: %v", err)
				}
			}
			db.Delete(&file)

			if file.SceneID != 0 {
				scene.GetIfExistByPK(file.SceneID)
				scene.UpdateStatus()
			}
		}
	} else {
		log.Errorf("error deleting file %v", err)
	}
	return scene
}
