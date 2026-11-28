package api

import (
	"encoding/json"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/emicklei/go-restful/v3"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/models"
)

var enrichAllRunning int32

type dmmActressSearchResp struct {
	Result struct {
		Actress []struct {
			Name     string `json:"name"`
			Ruby     string `json:"ruby"`
			Bust     string `json:"bust"`
			Cup      string `json:"cup"`
			Waist    string `json:"waist"`
			Hip      string `json:"hip"`
			Height   string `json:"height"`
			Birthday string `json:"birthday"`
			ImageURL struct {
				Small string `json:"small"`
				Large string `json:"large"`
			} `json:"imageURL"`
		} `json:"actress"`
	} `json:"result"`
}

func (i ActorResource) enrichActor(req *restful.Request, resp *restful.Response) {
	actorID, err := strconv.Atoi(req.PathParameter("actor-id"))
	if err != nil {
		resp.WriteErrorString(http.StatusBadRequest, "invalid actor id")
		return
	}

	var actor models.Actor
	if err := actor.GetIfExistByPKWithSceneAvg(uint(actorID)); err != nil || actor.ID == 0 {
		resp.WriteErrorString(http.StatusNotFound, "actor not found")
		return
	}

	changed := enrichActorProfile(&actor)
	if changed {
		if err := actor.Save(); err != nil {
			log.Errorf("enrichActor save error: %v", err)
		}
	}

	resp.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
		"actor":   actor,
		"changed": changed,
	})
}

func (i ActorResource) enrichAllActors(req *restful.Request, resp *restful.Response) {
	if !atomic.CompareAndSwapInt32(&enrichAllRunning, 0, 1) {
		resp.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
			"started": false,
			"message": "Actor enrichment is already running in background",
		})
		return
	}

	commonDb, _ := models.GetCommonDB()
	var actors []models.Actor
	commonDb.Where(
		"(cup_size IS NULL OR cup_size = '' OR band_size = 0 OR waist_size = 0 OR hip_size = 0 OR height = 0)",
	).Find(&actors)

	total := len(actors)
	go func(list []models.Actor) {
		defer atomic.StoreInt32(&enrichAllRunning, 0)
		updated := 0
		for idx := range list {
			a := list[idx]
			if enrichActorProfile(&a) {
				if err := a.Save(); err == nil {
					updated++
				}
			}
			time.Sleep(350 * time.Millisecond)
		}
		log.Infof("Batch actor enrichment completed: checked=%d updated=%d", len(list), updated)
	}(actors)

	resp.WriteHeaderAndEntity(http.StatusOK, map[string]interface{}{
		"started": true,
		"total":   total,
	})
}

func needsBodyEnrichment(a *models.Actor) bool {
	return a.CupSize == "" || a.BandSize == 0 || a.WaistSize == 0 || a.HipSize == 0 || a.Height == 0
}

func enrichActorProfile(actor *models.Actor) bool {
	if actor == nil || actor.Name == "" {
		return false
	}
	changed := false

	// 1. DMM Affiliate API ActressSearch
	if enrichFromDMM(actor) {
		changed = true
	}

	// 2. Fallback: javdatabase.com
	if needsBodyEnrichment(actor) {
		if enrichFromJavDatabase(actor) {
			changed = true
		}
	}

	return changed
}

func enrichFromDMM(actor *models.Actor) bool {
	apiID := strings.TrimSpace(config.Config.Custom.DmmAPIKey.DmmApiId)
	affID := strings.TrimSpace(config.Config.Custom.DmmAPIKey.DmmAffiliateId)
	if apiID == "" || affID == "" {
		return false
	}

	queries := []string{actor.Name}
	if actor.Furigana != "" {
		if strings.HasPrefix(strings.TrimSpace(actor.Furigana), "[") {
			var arr []string
			if err := json.Unmarshal([]byte(actor.Furigana), &arr); err == nil && len(arr) > 0 && arr[0] != "" {
				queries = append(queries, arr[0])
			}
		} else {
			queries = append(queries, actor.Furigana)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	for _, kw := range queries {
		params := url.Values{}
		params.Set("api_id", apiID)
		params.Set("affiliate_id", affID)
		params.Set("keyword", kw)
		params.Set("output", "json")

		res, err := client.Get("https://api.dmm.com/affiliate/v3/ActressSearch?" + params.Encode())
		if err != nil {
			continue
		}
		body, err := io.ReadAll(res.Body)
		res.Body.Close()
		if err != nil || res.StatusCode != http.StatusOK {
			continue
		}

		var dmmResp dmmActressSearchResp
		if err := json.Unmarshal(body, &dmmResp); err != nil || len(dmmResp.Result.Actress) == 0 {
			continue
		}

		var matchedIdx = -1
		for idx, item := range dmmResp.Result.Actress {
			if strings.EqualFold(strings.TrimSpace(item.Name), strings.TrimSpace(actor.Name)) {
				matchedIdx = idx
				break
			}
		}
		if matchedIdx == -1 && len(dmmResp.Result.Actress) == 1 {
			matchedIdx = 0
		}
		if matchedIdx == -1 {
			continue
		}

		item := dmmResp.Result.Actress[matchedIdx]
		changed := false

		if actor.CupSize == "" {
			if cup := models.NormalizeCupSize(item.Cup); cup != "" {
				actor.CupSize = cup
				changed = true
			}
		}
		if actor.BandSize == 0 {
			if v, err := strconv.Atoi(strings.TrimSpace(item.Bust)); err == nil && v > 0 {
				actor.BandSize = v
				changed = true
			}
		}
		if actor.WaistSize == 0 {
			if v, err := strconv.Atoi(strings.TrimSpace(item.Waist)); err == nil && v > 0 {
				actor.WaistSize = v
				changed = true
			}
		}
		if actor.HipSize == 0 {
			if v, err := strconv.Atoi(strings.TrimSpace(item.Hip)); err == nil && v > 0 {
				actor.HipSize = v
				changed = true
			}
		}
		if actor.Height == 0 {
			if v, err := strconv.Atoi(strings.TrimSpace(item.Height)); err == nil && v > 0 {
				actor.Height = v
				changed = true
			}
		}
		if actor.BirthDate.Year() <= 1 && strings.TrimSpace(item.Birthday) != "" {
			if t, err := time.Parse("2006-01-02", strings.TrimSpace(item.Birthday)); err == nil {
				actor.BirthDate = t
				changed = true
			}
		}
		if actor.ImageUrl == "" && strings.TrimSpace(item.ImageURL.Large) != "" {
			actor.ImageUrl = strings.TrimSpace(item.ImageURL.Large)
			changed = true
		}
		return changed
	}

	return false
}

func enrichFromJavDatabase(actor *models.Actor) bool {
	client := &http.Client{Timeout: 12 * time.Second}
	searchURL := "https://www.javdatabase.com/idols/?_idol_search=" + url.QueryEscape(actor.Name)
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return false
	}
	req.Header.Set("User-Agent", userAgent)

	res, err := client.Do(req)
	if err != nil {
		return false
	}
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil || res.StatusCode != http.StatusOK {
		return false
	}

	linkRe := regexp.MustCompile(`https://www\.javdatabase\.com/idols/([a-z0-9\-]+)/`)
	matches := linkRe.FindAllStringSubmatch(string(body), -1)
	var targetURL string
	for _, m := range matches {
		if len(m) >= 2 && m[1] != "page" {
			targetURL = m[0]
			break
		}
	}
	if targetURL == "" {
		return false
	}

	detailReq, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return false
	}
	detailReq.Header.Set("User-Agent", userAgent)

	detailRes, err := client.Do(detailReq)
	if err != nil {
		return false
	}
	detailBodyBytes, err := io.ReadAll(detailRes.Body)
	detailRes.Body.Close()
	if err != nil || detailRes.StatusCode != http.StatusOK {
		return false
	}
	detailHTML := html.UnescapeString(string(detailBodyBytes))

	changed := false

	// Measurements: 89-58-84
	measRe := regexp.MustCompile(`(?i)<b>Measurements:</b>\s*(\d{2,3})\s*-\s*(\d{2,3})\s*-\s*(\d{2,3})`)
	if mm := measRe.FindStringSubmatch(detailHTML); len(mm) == 4 {
		if actor.BandSize == 0 {
			if v, err := strconv.Atoi(mm[1]); err == nil && v > 0 {
				actor.BandSize = v
				changed = true
			}
		}
		if actor.WaistSize == 0 {
			if v, err := strconv.Atoi(mm[2]); err == nil && v > 0 {
				actor.WaistSize = v
				changed = true
			}
		}
		if actor.HipSize == 0 {
			if v, err := strconv.Atoi(mm[3]); err == nil && v > 0 {
				actor.HipSize = v
				changed = true
			}
		}
	}

	// Cup: <a ...>F</a> or DD / DDD
	cupRe := regexp.MustCompile(`(?i)<b>Cup:</b>\s*(?:<[^>]+>)*\s*([A-Za-z]{1,4}(?:\s*(?:カップ|cups?))?)`)
	if cm := cupRe.FindStringSubmatch(detailHTML); len(cm) == 2 {
		if actor.CupSize == "" {
			if cup := models.NormalizeCupSize(cm[1]); cup != "" {
				actor.CupSize = cup
				changed = true
			}
		}
	}

	// Height: <a ...>163 cm</a>
	heightRe := regexp.MustCompile(`(?i)<b>Height:</b>\s*(?:<[^>]+>)*\s*(\d{2,3})\s*cm`)
	if hm := heightRe.FindStringSubmatch(detailHTML); len(hm) == 2 {
		if actor.Height == 0 {
			if v, err := strconv.Atoi(hm[1]); err == nil && v > 0 {
				actor.Height = v
				changed = true
			}
		}
	}

	// DOB: <a ...>1990-08-25</a>
	dobRe := regexp.MustCompile(`(?i)<b>DOB:</b>\s*(?:<[^>]+>)*\s*(\d{4}-\d{2}-\d{2})`)
	if dm := dobRe.FindStringSubmatch(detailHTML); len(dm) == 2 {
		if actor.BirthDate.Year() <= 1 {
			if t, err := time.Parse("2006-01-02", dm[1]); err == nil {
				actor.BirthDate = t
				changed = true
			}
		}
	}

	return changed
}
