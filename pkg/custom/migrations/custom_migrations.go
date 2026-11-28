package migrations

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/xbapps/xbvr/pkg/models"
	"gopkg.in/gormigrate.v1"

	shared "github.com/xbapps/xbvr/pkg/custom/shared"
	customconfig "github.com/xbapps/xbvr/pkg/custom/config"
)

func CustomMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		// No.1 Actor テーブルに FaceImageUrl カラムを追加
		{
			ID: "bp_0001-actor-add-faceimageurl-column",
			Migrate: func(tx *gorm.DB) error {
				type Actor struct {
					FaceImageUrl string `json:"face_image_url" xbvrbackup:"face_image_url"`
				}
				return tx.AutoMigrate(Actor{}).Error
			},
		},

		// No.2 FaceImageUrl が null の Actor に ImageUrl をコピーして初期化
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

		// No.3 File テーブルに HasThumbnail フィールドを追加
		{
			ID: "bp_0003-file-add-thumbnail-flag",
			Migrate: func(tx *gorm.DB) error {
				type File struct {
					HasThumbnail bool `json:"has_thumbnail" gorm:"default:false"`
				}
				return tx.AutoMigrate(File{}).Error
			},
		},

		// No.4 Actor テーブルに Furigana カラムを追加
		{
			ID: "bp_0004-actor-add-furigana-column",
			Migrate: func(tx *gorm.DB) error {
				type Actor struct {
					Furigana string `json:"furigana" xbvrbackup:"furigana"`
				}
				return tx.AutoMigrate(Actor{}).Error
			},
		},

		// No.5 Furigana が null の Actor に Aliasesからひらがな文字を探して初期化
		{
			ID: "bp_0005-actor-set-furigana-default",
			Migrate: func(tx *gorm.DB) error {
				var actors []models.Actor
				err := tx.Where("furigana IS NULL").Find(&actors).Error
				if err != nil {
					return err
				}

				for _, actor := range actors {
					changed := false
					if actor.Furigana == "" {
						if actor.Aliases != "" {
							// ひらがな文字のみを抽出
							furiganaList := shared.FilterHiraganaOnly(actor.Aliases)
							if len(furiganaList) > 0 {
								actor.Furigana = furiganaList[0]
								changed = true
							}
						}
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

		// No.6 File テーブルに ThumbnailParameters フィールドを追加
		{
			ID: "bp_0006-file-add-thumbnail-params",
			Migrate: func(tx *gorm.DB) error {
				type File struct {
					ThumbnailParameters string `json:"thumbnail_parameters"`
				}
				return tx.AutoMigrate(File{}).Error
			},
		},

		// No.7 HasThumbnail = true and ThumbnailParameters = null の Files に 規定の設定値を追加
		{
			ID: "bp_0007-file-set-thumbnailparameters-default",
			Migrate: func(tx *gorm.DB) error {
				var files []models.File
				err := tx.Where("has_thumbnail=1 AND thumbnail_parameters IS NULL").Find(&files).Error
				if err != nil {
					return err
				}

				for _, file := range files {
					if len(file.ThumbnailParameters) == 0 {
						params := customconfig.ThumbnailParams{
							Start:         15,
							Interval:      15,
							Resolution:    200,
							UseCUDAEncode: true,
						}

						jsonBytes, err := json.Marshal(params)
						if err != nil {
							return err
						}

						jsonString := string(jsonBytes)
						file.ThumbnailParameters = jsonString

						if err := tx.Save(&file).Error; err != nil {
							return err
						}
					}
				}
				return nil
			},
		},

		// No.8 Actor テーブルに詳細評価・ジャンルタグカラムを追加
		{
			ID: "bp_0008-actor-add-detailed-ratings-and-genre-tags",
			Migrate: func(tx *gorm.DB) error {
				type Actor struct {
					RatingFace  float64 `json:"rating_face"`
					RatingStyle float64 `json:"rating_style"`
					RatingBust  float64 `json:"rating_bust"`
					RatingVibe  float64 `json:"rating_vibe"`
					GenreTags   string  `json:"genre_tags" sql:"type:text;"`
				}
				return tx.AutoMigrate(Actor{}).Error
			},
		},

		// No.9 Scene テーブルに画質・演出評価カラムを追加
		{
			ID: "bp_0009-scene-add-detailed-ratings",
			Migrate: func(tx *gorm.DB) error {
				type Scene struct {
					RatingVideoQuality float64 `json:"rating_video_quality"`
					RatingDirection    float64 `json:"rating_direction"`
				}
				return tx.AutoMigrate(Scene{}).Error
			},
		},

		// No.10 Scene テーブルにプレイ・抜け感評価カラムを追加
		{
			ID: "bp_0010-scene-add-play-and-appeal-ratings",
			Migrate: func(tx *gorm.DB) error {
				type Scene struct {
					RatingPlay   float64 `json:"rating_play"`
					RatingAppeal float64 `json:"rating_appeal"`
				}
				return tx.AutoMigrate(Scene{}).Error
			},
		},

		// No.11 Scene テーブルにジャンルタグカラムを追加
		{
			ID: "bp_0011-scene-add-genre-tags",
			Migrate: func(tx *gorm.DB) error {
				type Scene struct {
					GenreTags string `json:"genre_tags" sql:"type:text;"`
				}
				return tx.AutoMigrate(Scene{}).Error
			},
		},
	}
}
