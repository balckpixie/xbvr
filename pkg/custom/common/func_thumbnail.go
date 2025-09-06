package common

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/common"
)


// 共通で呼べるサムネイル削除処理
// - File にサムネイルがあれば削除
// - File の情報を更新して保存
// - 関連 Scene があれば更新
// - 成功/失敗を error で返す
func DeleteThumbnail(file *models.File) error {
	if !file.HasThumbnail {
		return nil
	}

	thumbFile := filepath.Join(
		common.VideoThumbnailDir,
		strconv.FormatUint(uint64(file.ID), 10)+".jpg",
	)

	// サムネイルファイル削除
	if err := os.Remove(thumbFile); err != nil {
		log.Errorf("error deleting thumbnail file: %v", err)
		return err
	}

	// File 情報を更新
	file.HasThumbnail = false
	file.ThumbnailParameters = ""
	if err := file.Save(); err != nil {
		log.Warnf("failed to save file %v: %v", file.ID, err)
		return err
	}
	log.Infof("Thumbnails deleted File_ID %v - Saved", file.ID)

	// Scene がある場合は更新
	if file.SceneID != 0 {
		var scene models.Scene
		scene.GetIfExistByPK(file.SceneID)
		scene.UpdateStatus()
	}

	return nil
}