package common

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xbapps/xbvr/pkg/custom/shared"
	"github.com/xbapps/xbvr/pkg/models"
)

func ResetProjection(fileId uint, projection string) models.Scene {

	var scene models.Scene
	var file models.File
	db, _ := models.GetDB()
	defer db.Close()

	err := db.Preload("Volume").Where(&models.File{ID: fileId}).First(&file).Error
	if err == nil {
		if (projection == "") {
			if vrType, err := shared.DetectVRType(file.GetPath(),time.Second*5); err == nil {
					projection = vrType
			}
		}
		db.Model(&file).Where("id = ?", fileId).Update("video_projection", projection)
		
		_ = DeleteThumbnail(&file)
		if file.SceneID != 0 {
			scene.GetIfExistByPK(file.SceneID)
			scene.UpdateStatus()
		}
	} else {
		log.Errorf("error renaming file %v", err)
	}
	return scene
}

func RenameFileBySceneID(scene models.Scene, file models.File) models.Scene {

	extension := filepath.Ext(file.Filename)
	sceneNo := GetSceneNo(file)
	if len(sceneNo) > 0 {
		sceneNo = "-" + sceneNo
	}
	// 許可キーワード（すべて小文字）
	allowed := map[string]bool{
		"sbs": true, "lr": true, "ou": true, "tb": true, "rl": true, "bt": true,
		"lrf": true, "tbf": true, "rlf": true, "btf": true,
		"180": true, "360": true, "180f": true, "360eac": true,
		"fisheye": true, "fisheye190": true, "rf52": true,
		"alpha": true,
		"4k":    true, "5k": true, "6k": true, "7k": true, "8k": true,
		"60fps": true, "30fps": true,
	}
	baseName := strings.TrimSuffix(file.Filename, extension)
	suffix := ""
	if idx := strings.Index(baseName, "_"); idx != -1 {
		// アンダースコア以降を取り出して、アンダースコアで分割
		suffixRaw := baseName[idx+1:]
		parts := strings.Split(suffixRaw, "_")
		var validParts []string
		for _, part := range parts {
			if allowed[strings.ToLower(part)] {
				validParts = append(validParts, part)
			}
		}
		if len(validParts) > 0 {
			suffix = "_" + strings.Join(validParts, "_")
		}
	}

	newFileName := sanitizeFilename(fmt.Sprintf("%s%s%s%s", scene.SceneID, sceneNo, suffix, extension))

	db, _ := models.GetDB()
	defer db.Close()
	vol := models.Volume{}
	err := db.First(&vol, file.VolumeID).Error
	if err == nil {
		newPath := vol.Path
		newFileName = sanitizeFilename(newFileName)
		if filepath.Join(file.Path, file.Filename) != filepath.Join(newPath, newFileName) {
			scene = RenameFileByFileId(uint(file.ID), newPath, newFileName)
		}
	}

	return scene
}

func RenameFileBySceneInfo(scene models.Scene, file models.File) models.Scene {

	sceneNo := GetSceneNo(file)
	extension := filepath.Ext(file.Filename)
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
	newFileName := fmt.Sprintf("%s|%s%s|%s%s", castString, scene.SceneID, sceneNo, trimPrefix(scene.SceneID, title), extension)

	db, _ := models.GetDB()
	defer db.Close()

	vol := models.Volume{}
	err := db.First(&vol, file.VolumeID).Error

	if err == nil {
		newPath := vol.Path
		newFileName = sanitizeFilename(newFileName)
		if filepath.Join(file.Path, file.Filename) != filepath.Join(newPath, newFileName) {
			scene = RenameFileByFileId(uint(file.ID), newPath, newFileName)
		}
	}

	return scene
}

func RenameFileByFileId(fileId uint, newPath string, newfilename string) models.Scene {

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
		// UTF-8での文字数制限に合わせて処理
		if utf8.RuneCountInString(nameOnly) > maxFilenameLength-4 {
			// スライスの範囲を文字数で調整
			runes := []rune(nameOnly)
			nameOnly = string(runes[:maxFilenameLength-4])
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
			log.Errorf("error renaming file %v", err)
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
	ext := filepath.Ext(newFileName)
	base := strings.TrimSuffix(newFileName, ext)

	i := 1
	for {
		newPath := newFileName
		_, err := os.Stat(newPath)
		if err != nil {
			renamedFileName = newFileName
			break
		}
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

func GetSceneNo(file models.File) string {

	base := filepath.Base(file.Filename)
	input := base[:len(base)-len(filepath.Ext(base))]

	// 正規表現パターンに一致する部分を検索
	//IPVR-001-1.mp4
	pattern := `([a-zA-Z]{2,6}-(\d{3,6})-([a-zA-Z0-9]{1}))`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 0 {
		if len(matches[3]) > 0 {
			return matches[3]
		}
	}

	//4k2.com@juvr00208_3_8k.mp4
	pattern = `([a-zA-Z0-9]{2,9}(-|_)(\d{1,6}))`
	re = regexp.MustCompile(pattern)
	matches = re.FindStringSubmatch(input)

	if len(matches) > 0 {
		// matchedPart := matches[1]  // 最初のサブマッチンググループ: ([a-zA-Z0-9]{2,6}-\d{2,6})
		nextPart := matches[3]
		secondPattern := `([a-zA-Z0-9]{1})`
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

func sanitizeFilename(filename string) string {
	// 無効な文字を取り除く正規表現
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|]`)

	// 無効な文字を削除し、指定した文字列で置き換える
	sanitizedFilename := invalidChars.ReplaceAllString(filename, "-")

	return sanitizedFilename
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

//////////////////////////

// スライス内の古い要素を新しい要素で置き換える関数
func replaceElement(slice []string, oldElement string, newElement string) {
	for i, element := range slice {
		if element == oldElement {
			slice[i] = newElement
			return
		}
	}
}

// 文字列Bが文字列Aで始まる場合、文字列Bから文字列Aを取り除く
// 文字列Bが文字列Aを含まない場合は、文字列Bをそのまま返す
func trimPrefix(a, b string) string {
	if strings.HasPrefix(b, a) {
		trimmed := strings.TrimPrefix(b, a)
		return strings.TrimSpace(trimmed)
	}
	return b
}

// 2つの文字列 strA と strB の末尾から順に比較し、一致しているかどうかを判定
// 途中で一致しない文字があれば false を返し、片方の文字列が短くて最後まで一致していれば true を返します
func areEqualBackwards(strA, strB string) bool {
	lenA, lenB := len(strA), len(strB)
	i, j := lenA-1, lenB-1

	for i >= 0 && j >= 0 {
		if strA[i] != strB[j] {
			return false
		}
		i--
		j--
	}
	return true
}
