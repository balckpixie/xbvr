package migrations

import (
	"github.com/jinzhu/gorm"
	"gopkg.in/gormigrate.v1"
	"github.com/xbapps/xbvr/pkg/models"
)

func CustomMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		// ① Actor テーブルに FaceImageUrl カラムを追加
		{
			ID: "bp_0001-actor-add-faceimageurl-column",
			Migrate: func(tx *gorm.DB) error {
				type Actor struct {
					FaceImageUrl string `json:"face_image_url" xbvrbackup:"face_image_url"`
				}
				return tx.AutoMigrate(Actor{}).Error
			},
		},

		// ② FaceImageUrl が null の Actor に ImageUrl をコピーして初期化
		{
			ID: "bp_0002-actor-set-faceimageurl-default",
			Migrate: func(tx *gorm.DB) error {
				var actors []models.Actor
				err := tx.Where("face_image_url IS NULL").Find(&actors).Error
				if err != nil {
					return err
				}

				for _, actor := range actors {
					changed := false
					if actor.FaceImageUrl == "" {
						actor.FaceImageUrl = actor.ImageUrl
						changed = true
					}
					if changed {
						err = tx.Save(&actor).Error
						if err != nil {
							return err
						}
					}
				}
				return nil
			},
		},

		// ③ File テーブルに HasThumbnail フィールドを追加
		{
			ID: "bp_0003-file-add-thumbnail-flag",
			Migrate: func(tx *gorm.DB) error {
				type File struct {
					HasThumbnail bool `json:"has_thumbnail" gorm:"default:false"`
				}
				return tx.AutoMigrate(File{}).Error
			},
		},
	}
}
