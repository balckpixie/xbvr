<template>
  <div class="modal is-active">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
      @keydown.left="handleLeftArrow"
      @keydown.right="handleRightArrow"
      @keydown.o="prevActor"
      @keydown.p="nextActor"
      @keydown.g="toggleGallery"
    />

    <div class="modal-background" @click="close"></div>

    <div class="modal-card actor-collection-modal">
      <a
        class="docked-nav-btn is-prev"
        :class="{ 'is-disabled': !prevActorObj }"
        @click="prevActor"
        :title="prevActorObj ? `Prev: ${prevActorObj.name} (O)` : 'No previous actor'"
      >
        <span class="nav-arrow">&#10094;</span>
        <img
          v-if="prevActorObj"
          :src="getImageURL(prevActorObj.face_image_url || prevActorObj.image_url, '120,fit')"
          class="nav-actor-face"
          :alt="prevActorObj.name"
        />
      </a>

      <section class="modal-card-body">
        <div class="top-controls">
          <div class="is-flex" style="gap: 0.25rem; flex-wrap: wrap;">
            <span class="tag is-dark is-light" v-if="currentPlayingFile">
              Playing: {{ currentPlayingFile.filename }}
            </span>
          </div>
          <button
            rounded
            class="button is-outlined is-small"
            @click="hideRightPane = !hideRightPane"
          >
            {{ hideRightPane ? 'Show' : 'Hide' }} Detail
          </button>
        </div>

        <splitpanes
          class="default-theme"
          style="flex: 1; min-height: 0; height: 100%;"
          :gutter-size="10"
          :min-pane-size="25"
          :max-pane-size="100"
          ref="splitpanesRef"
        >
          <!-- 左ペイン: ギャラリー兼プレイヤー -->
          <pane :min-size="25" :max-size="100" :size="hideRightPane ? 100 : 60">
            <div class="left-pane-inner">
              <b-tabs v-model="activeMedia" position="is-centered" :animated="false">
                <b-tab-item label="Gallery">
                  <div class="gallery-toolbar" v-if="selectedScene">
                    <b-field grouped>
                      <b-radio-button v-model="galleryMode" native-value="scene" size="is-small">
                        Scene ({{ selectedSceneImages.length }})
                      </b-radio-button>
                      <b-radio-button v-model="galleryMode" native-value="actor" size="is-small">
                        Actor ({{ actorImages.length }})
                      </b-radio-button>
                      <b-radio-button v-model="galleryMode" native-value="all" size="is-small">
                        All ({{ allGalleryImages.length }})
                      </b-radio-button>
                    </b-field>
                  </div>

                  <b-carousel
                    v-if="galleryImages.length > 0"
                    v-model="carouselSlide"
                    @change="scrollToActiveIndicator"
                    :autoplay="false"
                    :indicator-inside="false"
                  >
                    <b-carousel-item v-for="(imgUrl, i) in galleryImages" :key="i">
                      <div
                        class="image is-1by1 is-full"
                        :style="{
                          backgroundImage: `url('${getImageURL(imgUrl, '700,fit')}')`,
                          backgroundSize: 'contain',
                          backgroundPosition: 'center',
                          backgroundRepeat: 'no-repeat'
                        }"
                      ></div>
                    </b-carousel-item>
                    <template slot="indicators" slot-scope="props">
                      <span class="al image" style="width: max-content;">
                        <vue-load-image>
                          <img slot="image" :src="getIndicatorURL(props.i)" style="height: 60px;" />
                          <img slot="preloader" src="/ui/images/blank.png" style="height: 60px;" />
                          <img slot="error" src="/ui/images/blank_female_profile.png" style="height: 60px;" />
                        </vue-load-image>
                      </span>
                    </template>
                  </b-carousel>
                  <div v-else class="has-text-centered has-text-grey py-6">
                    No gallery images available.
                  </div>
                </b-tab-item>

                <b-tab-item label="Player">
                  <div class="player-container">
                    <Vr180Player
                      v-if="activeMedia === 1 && currentPlayingFile"
                      ref="vrPlayer"
                      :file="currentPlayingFile"
                      @ended="onPlayerEnded"
                    />
                    <div v-else class="has-text-centered has-text-grey py-6">
                      Select a video file from the right list to play.
                    </div>
                  </div>
                  <div v-if="currentPlayingScene" class="player-footer-bar mt-2">
                    <div class="player-file-switcher">
                      <button
                        v-for="(f, idx) in getVideoFiles(currentPlayingScene)"
                        :key="f.id"
                        class="button is-small"
                        :class="[
                          f.has_thumbnail ? 'is-success' : 'is-danger',
                          currentPlayingFile && currentPlayingFile.id === f.id ? 'is-current-file' : 'is-outlined'
                        ]"
                        :title="f.filename"
                        @click="playFile(currentPlayingScene, f)"
                      >
                        {{ idx + 1 }}
                      </button>
                    </div>
                    <span v-if="isPlayAllActive" class="tag is-info is-light is-small">
                      すべて再生中
                    </span>
                    <span class="player-scene-title is-size-7 has-text-grey">
                      {{ currentPlayingScene.title }}
                    </span>
                  </div>
                </b-tab-item>
              </b-tabs>
            </div>
          </pane>

          <!-- 右ペイン: Actor 情報 + シーンごとの作品・動画ファイル一覧 -->
          <pane v-if="!hideRightPane" :min-size="25" :max-size="75" :size="40">
            <div class="right-pane-inner">
              <!-- Actor ヘッダ (添付画像レイアウト + ダークテーマ) -->
              <div class="actor-header">
                <div class="actor-header-layout">
                  <div class="actor-avatar-wrapper" v-if="actor.face_image_url || actor.image_url">
                    <img
                      :src="getImageURL(actor.face_image_url || actor.image_url, '300,fit')"
                      class="actor-avatar"
                      alt="Actor"
                      @mouseover="showActorImageTooltip = true"
                      @mouseout="showActorImageTooltip = false"
                    />
                    <div v-if="showActorImageTooltip && actor.image_url" class="actor-hover-tooltip">
                      <img :src="getImageURL(actor.image_url, '700,fit')" alt="Actor Image" />
                    </div>
                  </div>
                  <div class="actor-header-main">
                    <div class="actor-title-row">
                      <div class="actor-name-wrap">
                        <span class="actor-name">{{ actor.name }}</span>
                        <span v-if="actor.furigana" class="actor-furigana">({{ actor.furigana }})</span>
                      </div>
                      <div class="actor-top-right">
                        <div class="overall-rating-pill">
                          <span class="overall-label">総合</span>
                          <star-rating
                            :key="`${actor.id}-overall`"
                            :rating="actor.star_rating || 0"
                            read-only
                            :increment="0.1"
                            :star-size="16"
                            :show-rating="true"
                            active-color="#f43f5e"
                          />
                        </div>
                        <div class="actor-fav-btns">
                          <actor-favourite-button :actor="actor" />
                          <actor-watchlist-button :actor="actor" />
                        </div>
                      </div>
                    </div>

                    <div class="actor-stats-row">
                      <span class="actor-stat-pill">作品数: {{ detailedScenes.length }}本</span>
                      <span class="actor-stat-pill">お気に入り: {{ favouriteSceneCount }}本</span>
                    </div>

                    <ActorRatingAndGenres
                      :actor="actor"
                      compact
                      dark
                      exclude-overall
                      grid2x2
                    />
                  </div>
                </div>
              </div>

              <!-- ツールバー -->
              <div class="scenes-toolbar">
                <div class="is-flex is-align-items-center" style="gap: 0.5rem;">
                  <span class="tag is-info is-light">
                    Scenes: {{ filteredScenes.length }} / {{ detailedScenes.length }}
                  </span>
                  <span class="tag is-success is-light">
                    Videos: {{ totalVideoCount }}
                  </span>
                </div>
                <div class="is-flex is-align-items-center" style="gap: 0.5rem;">
                  <b-switch v-model="onlyWithVideos" size="is-small">
                    Videos only
                  </b-switch>
                  <button class="button is-small is-light" @click="setAllExpanded(true)">Expand all</button>
                  <button class="button is-small is-light" @click="setAllExpanded(false)">Collapse all</button>
                </div>
              </div>

              <!-- シーン・ファイル一覧 -->
              <div class="scenes-scroll">
                <div v-if="loadingScenes" class="has-text-centered py-5">
                  <b-icon pack="mdi" icon="loading" custom-class="mdi-spin" size="is-medium"></b-icon>
                  <p class="is-size-7 has-text-grey mt-2">Loading scenes and files...</p>
                </div>

                <div v-else-if="filteredScenes.length === 0" class="has-text-centered has-text-grey py-5">
                  No matching scenes found.
                </div>

                <div
                  v-else
                  v-for="scene in filteredScenes"
                  :key="scene.id"
                  class="scene-group-card"
                  :class="{ 'is-selected': selectedScene && selectedScene.id === scene.id }"
                >
                  <!-- シーンヘッダ (添付画像レイアウト) -->
                  <div class="scene-group-header" @click="toggleSceneExpand(scene)">
                    <div class="scene-header-upper">
                      <div class="scene-header-thumb-wrap" @click.stop="selectSceneGallery(scene)">
                        <img
                          v-if="scene.cover_url"
                          :src="getImageURL(scene.cover_url, '300,fit')"
                          class="scene-header-cover"
                          alt="Cover"
                        />
                        <div v-else class="scene-header-cover-placeholder">
                          <b-icon pack="mdi" icon="filmstrip" size="is-medium" />
                        </div>
                      </div>

                      <div class="scene-header-info">
                        <div class="scene-header-top-row">
                          <div class="scene-header-title" :title="scene.title">{{ scene.title }}</div>
                          <div class="scene-right-meta-stack">
                            <div class="spec-badges is-right" v-if="getPrimaryVideo(scene)">
                              <span v-if="formatResolutionBadge(getPrimaryVideo(scene))" class="spec-badge is-res">
                                {{ formatResolutionBadge(getPrimaryVideo(scene)) }}
                              </span>
                              <span v-if="getPrimaryVideo(scene).video_codec_name" class="spec-badge is-codec">
                                {{ formatCodecBadge(getPrimaryVideo(scene).video_codec_name) }}
                              </span>
                              <span v-if="getPrimaryVideo(scene).projection" class="spec-badge is-proj">
                                {{ formatProjectionBadge(getPrimaryVideo(scene).projection) }}
                              </span>
                            </div>
                            <div class="scene-overall-pill" @click.stop>
                              <span class="overall-label">総合</span>
                              <star-rating
                                :key="`${scene.id}-scene-overall`"
                                :rating="calcSceneOverall(scene)"
                                read-only
                                :increment="0.1"
                                :star-size="14"
                                :show-rating="true"
                                active-color="#f43f5e"
                              />
                            </div>
                          </div>
                        </div>

                        <div class="mt-1" @click.stop>
                          <SceneRatings :item="scene" dark :star-size="14" />
                        </div>
                      </div>
                    </div>

                    <div class="scene-header-bottom-row">
                      <div class="scene-header-sub">
                        <span>{{ scene.site }}</span>
                        <span v-if="scene.release_date && scene.release_date !== '0001-01-01T00:00:00Z'">
                          • {{ format(parseISO(scene.release_date), 'yyyy-MM-dd') }}
                        </span>
                        <span>• {{ getVideoFiles(scene).length }} video(s)</span>
                      </div>

                      <div class="scene-header-actions-bar" @click.stop>
                        <favourite-button :item="scene" />
                        <edit-button :item="scene" />
                        <delete-button :item="scene" @deleted="onSceneDeleted(scene)" />
                        <button
                          class="scene-icon-btn"
                          title="Show scene gallery"
                          @click="selectSceneGallery(scene)"
                        >
                          <b-icon pack="mdi" icon="image-multiple" size="is-small" />
                        </button>
                        <button
                          class="scene-icon-btn"
                          title="Open Scene Details"
                          @click="openSceneDetails(scene)"
                        >
                          <b-icon pack="mdi" icon="open-in-new" size="is-small" />
                        </button>
                        <button
                          v-if="getVideoFiles(scene).length > 0"
                          class="play-all-btn"
                          :class="{ 'is-active': isPlayAllActive && currentPlayingScene && currentPlayingScene.id === scene.id }"
                          @click="playAllScene(scene)"
                        >
                          <b-icon pack="fas" icon="play" size="is-small" />
                          <span>すべて再生</span>
                        </button>
                        <button
                          class="scene-icon-btn"
                          :title="isSceneExpanded(scene.id) ? 'Collapse' : 'Expand'"
                          @click="toggleSceneExpand(scene)"
                        >
                          <b-icon
                            pack="mdi"
                            :icon="isSceneExpanded(scene.id) ? 'chevron-up' : 'chevron-down'"
                            size="is-small"
                          />
                        </button>
                      </div>
                    </div>
                  </div>

                  <!-- シーン内ファイルリスト (テーブル風レイアウト + 抜粋サムネイル + 操作ボタン) -->
                  <div v-show="isSceneExpanded(scene.id)" class="scene-files-body">
                    <div
                      v-if="!scene.file || scene.file.length === 0"
                      class="is-size-7 has-text-grey-light px-3 py-3"
                    >
                      No files linked to this scene.
                    </div>

                    <template v-else>
                      <div class="file-table-header">
                        <div class="col-play"></div>
                        <div class="col-name">ファイル名</div>
                        <div class="col-duration">再生時間</div>
                        <div class="col-size">容量</div>
                        <div class="col-actions-head">操作</div>
                      </div>

                      <div
                        class="file-table-row"
                        v-for="f in sortedFiles(scene)"
                        :key="f.id"
                        :class="{ 'is-playing': currentPlayingFile && currentPlayingFile.id === f.id }"
                      >
                        <div class="file-row-main">
                          <div class="col-play">
                            <button
                              rounded
                              class="button is-success is-small"
                              @click="playFile(scene, f)"
                              v-show="f.type === 'video'"
                              title="Play video"
                            >
                              <b-icon pack="fas" icon="play" size="is-small"></b-icon>
                            </button>
                            <button
                              rounded
                              class="button is-info is-small is-outlined"
                              disabled
                              v-show="f.type === 'script'"
                            >
                              <b-icon pack="mdi" icon="pulse"></b-icon>
                            </button>
                            <button
                              rounded
                              class="button is-info is-small is-outlined"
                              disabled
                              v-show="f.type === 'hsp'"
                            >
                              <b-icon pack="mdi" icon="safety-goggles"></b-icon>
                            </button>
                            <button
                              rounded
                              class="button is-info is-small is-outlined"
                              disabled
                              v-show="f.type === 'subtitles'"
                            >
                              <b-icon pack="mdi" icon="subtitles"></b-icon>
                            </button>
                          </div>

                          <div class="col-name">
                            <div class="file-name-text" :title="f.filename">{{ f.filename }}</div>
                          </div>

                          <div class="col-duration">
                            <span v-if="f.duration > 1">{{ humanizeSeconds(f.duration) }}</span>
                            <span v-else>-</span>
                          </div>

                          <div class="col-size">
                            {{ prettyBytes(f.size) }}
                          </div>

                          <div class="col-actions">
                            <div class="file-actions-group">
                              <button class="button is-danger is-small is-outlined" title="Rename file" @click="renameFile(f)">
                                <b-icon pack="fas" icon="pen" size="is-small"></b-icon>
                              </button>
                              <button class="button is-danger is-small is-outlined" title="Delete file from disk" @click="removeFile(f)">
                                <b-icon pack="fas" icon="trash" size="is-small"></b-icon>
                              </button>
                              <button class="button is-dark is-small is-outlined" title="Unmatch file from scene" @click="unmatchFile(f)">
                                <b-icon pack="fas" icon="unlink" size="is-small"></b-icon>
                              </button>
                              <b-dropdown aria-role="list" position="is-bottom-left" append-to-body>
                                <template #trigger>
                                  <button class="button is-dark is-small is-outlined" title="More actions">
                                    <b-icon pack="mdi" icon="dots-horizontal" size="is-small"></b-icon>
                                  </button>
                                </template>
                                <b-dropdown-item aria-role="listitem" @click="deleteThumbnail(f)">
                                  Delete thumbnail
                                </b-dropdown-item>
                                <b-dropdown-item aria-role="listitem" @click="resetProjection(f)">
                                  Reset projection
                                </b-dropdown-item>
                                <b-dropdown-item aria-role="listitem" @click="resetFileName(scene, f)">
                                  Reset filename
                                </b-dropdown-item>
                              </b-dropdown>
                            </div>
                          </div>
                        </div>

                        <div v-if="f.type === 'script' && f.has_heatmap" class="file-row-previews heatmapFunscript">
                          <img :src="getHeatmapURL(f.id)" />
                        </div>

                        <!-- サムネイル抜粋表示 (10個・100x100) -->
                        <div v-if="f.type === 'video' && f.has_thumbnail" class="file-row-previews">
                          <FileThumbnailStrip
                            :file="f"
                            :max-count="10"
                            @thumbnail-clicked="sec => playFileAt(scene, f, sec)"
                          />
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </div>
            </div>
          </pane>
        </splitpanes>
      </section>

      <a
        class="docked-nav-btn is-next"
        :class="{ 'is-disabled': !nextActorObj }"
        @click="nextActor"
        :title="nextActorObj ? `Next: ${nextActorObj.name} (P)` : 'No next actor'"
      >
        <img
          v-if="nextActorObj"
          :src="getImageURL(nextActorObj.face_image_url || nextActorObj.image_url, '120,fit')"
          class="nav-actor-face"
          :alt="nextActorObj.name"
        />
        <span class="nav-arrow">&#10095;</span>
      </a>
    </div>

    <button class="modal-close is-large" aria-label="close" @click="close"></button>
  </div>
</template>

<script>
import ky from 'ky'
import prettyBytes from 'pretty-bytes'
import { format, parseISO } from 'date-fns'
import VueLoadImage from 'vue-load-image'
import GlobalEvents from 'vue-global-events'
import StarRating from 'vue-star-rating'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'

import ActorFavouriteButton from '../../components/ActorFavouriteButton'
import ActorWatchlistButton from '../../components/ActorWatchlistButton'
import ActorRatingAndGenres from '../../components/ActorRatingAndGenres'
import FavouriteButton from '../../components/FavouriteButton'
import EditButton from '../../components/EditButton'
import DeleteButton from '../../components/DeleteButton'
import SceneRatings from '../../components/SceneRatings'
import FileThumbnailStrip from '../../components/FileThumbnailStrip'
import Vr180Player from '../../components/Vr180Player'

export default {
  name: 'ActorCollectionDetails',
  components: {
    VueLoadImage,
    GlobalEvents,
    StarRating,
    Splitpanes,
    Pane,
    ActorFavouriteButton,
    ActorWatchlistButton,
    ActorRatingAndGenres,
    FavouriteButton,
    EditButton,
    DeleteButton,
    SceneRatings,
    FileThumbnailStrip,
    Vr180Player
  },
  data () {
    return {
      activeMedia: 0,
      galleryMode: 'all',
      carouselSlide: 0,
      hideRightPane: false,
      onlyWithVideos: false,
      loadingScenes: false,
      detailedScenes: [],
      expandedSceneIds: {},
      selectedScene: null,
      currentPlayingScene: null,
      currentPlayingFile: null,
      playQueue: [],
      isPlayAllActive: false,
      showActorImageTooltip: false
    }
  },
  computed: {
    actor () {
      return this.$store.state.overlay.actorcollection.actor
    },
    prevActorObj () {
      return this.actor ? this.$store.getters['actorList/prevActor'](this.actor) : null
    },
    nextActorObj () {
      return this.actor ? this.$store.getters['actorList/nextActor'](this.actor) : null
    },
    actorImages () {
      if (!this.actor || !this.actor.image_arr) return []
      try {
        return JSON.parse(this.actor.image_arr).filter(im => im && im !== '')
      } catch (e) {
        return []
      }
    },
    selectedSceneImages () {
      if (!this.selectedScene) return []
      return this.extractSceneImages(this.selectedScene)
    },
    allGalleryImages () {
      const seen = new Set()
      const list = []
      const pushUnique = (u) => {
        if (u && !seen.has(u)) {
          seen.add(u)
          list.push(u)
        }
      }
      if (this.selectedScene) {
        this.extractSceneImages(this.selectedScene).forEach(pushUnique)
      }
      this.actorImages.forEach(pushUnique)
      this.detailedScenes.forEach(sc => {
        if (sc.cover_url) pushUnique(sc.cover_url)
      })
      return list
    },
    galleryImages () {
      if (this.galleryMode === 'scene' && this.selectedScene) {
        return this.selectedSceneImages
      }
      if (this.galleryMode === 'actor') {
        return this.actorImages
      }
      return this.allGalleryImages
    },
    filteredScenes () {
      if (!this.onlyWithVideos) return this.detailedScenes
      return this.detailedScenes.filter(sc => this.getVideoFiles(sc).length > 0)
    },
    totalVideoCount () {
      return this.detailedScenes.reduce((sum, sc) => sum + this.getVideoFiles(sc).length, 0)
    },
    favouriteSceneCount () {
      return this.detailedScenes.filter(sc => sc.favourite).length
    }
  },
  watch: {
    'actor.id': {
      immediate: true,
      handler () {
        this.loadActorScenes()
      }
    }
  },
  methods: {
    format,
    parseISO,
    prettyBytes,
    humanizeSeconds (seconds) {
      return new Date(seconds * 1000).toISOString().substr(11, 8)
    },
    getImageURL (u, size) {
      if (!u) return '/ui/images/blank_female_profile.png'
      if (u.startsWith('http://') || u.startsWith('https://')) {
        return `/img/${size || '700,fit'}/${u.replace('://', ':/')}`
      }
      return u
    },
    getIndicatorURL (idx) {
      if (this.galleryImages[idx] !== undefined) {
        return this.getImageURL(this.galleryImages[idx], 'x85')
      }
      return '/ui/images/blank_female_profile.png'
    },
    getHeatmapURL (fileId) {
      return `/api/dms/heatmap/${fileId}`
    },
    extractSceneImages (scene) {
      const imgs = []
      if (scene.cover_url) imgs.push(scene.cover_url)
      if (scene.images) {
        try {
          const parsed = typeof scene.images === 'string' ? JSON.parse(scene.images) : scene.images
          if (Array.isArray(parsed)) {
            parsed.forEach(item => {
              if (item && item.url) imgs.push(item.url)
            })
          }
        } catch (e) {
          // ignore malformed JSON
        }
      }
      return imgs
    },
    async loadActorScenes () {
      if (!this.actor || !this.actor.scenes) {
        this.detailedScenes = []
        return
      }
      this.loadingScenes = true
      this.selectedScene = null
      this.carouselSlide = 0
      this.expandedSceneIds = {}

      try {
        const scenes = await Promise.all(
          this.actor.scenes.map(s =>
            ky.get(`/api/scene/${s.id}`).json().catch(() => s)
          )
        )
        this.detailedScenes = scenes
        const expanded = {}
        scenes.forEach(sc => {
          expanded[sc.id] = true
        })
        this.expandedSceneIds = expanded
        if (scenes.length > 0) {
          this.selectedScene = scenes[0]
        }
      } finally {
        this.loadingScenes = false
      }
    },
    getVideoFiles (scene) {
      if (!scene || !scene.file) return []
      return scene.file.filter(f => f.type === 'video')
    },
    getPrimaryVideo (scene) {
      const videos = this.getVideoFiles(scene)
      return videos.length > 0 ? videos[0] : null
    },
    calcSceneOverall (scene) {
      if (!scene) return 0
      const cast = Array.isArray(scene.cast) ? scene.cast : []
      const castVals = cast.map(a => Number(a.star_rating) || 0).filter(v => v > 0)
      const actorAvg = castVals.length > 0
        ? Math.round((castVals.reduce((a, b) => a + b, 0) / castVals.length) * 10) / 10
        : 0
      const vals = [
        actorAvg,
        Number(scene.rating_video_quality) || 0,
        Number(scene.rating_play) || 0,
        Number(scene.rating_appeal) || 0
      ].filter(v => v > 0)
      if (vals.length === 0) return 0
      return Math.round((vals.reduce((a, b) => a + b, 0) / vals.length) * 10) / 10
    },
    formatResolutionBadge (file) {
      if (!file) return ''
      const w = file.video_width || 0
      const h = file.video_height || 0
      if (w >= 7000 || h >= 3800) return '8K'
      if (w >= 3800 || h >= 2000) return '4K'
      if (w >= 1900 || h >= 1000) return '1080p'
      if (h > 0) return `${h}p`
      return ''
    },
    formatCodecBadge (codec) {
      if (!codec) return ''
      const c = codec.toLowerCase()
      if (c === 'h264') return 'H.264'
      if (c === 'hevc' || c === 'h265') return 'H.265'
      return codec.toUpperCase()
    },
    formatProjectionBadge (proj) {
      if (!proj) return ''
      const map = {
        '180_sbs': '180° SBS',
        '180_tb': '180° TB',
        '180_mono': '180° MONO',
        '360_sbs': '360° SBS',
        '360_tb': '360° TB',
        '360_mono': '360° MONO',
        'flat': 'Flat'
      }
      return map[proj] || proj
    },
    sortedFiles (scene) {
      if (!scene || !scene.file) return []
      const order = { video: 0, script: 1, hsp: 2, subtitles: 3 }
      return [...scene.file].sort((a, b) => (order[a.type] ?? 9) - (order[b.type] ?? 9))
    },
    isSceneExpanded (sceneId) {
      return !!this.expandedSceneIds[sceneId]
    },
    toggleSceneExpand (scene) {
      this.$set(this.expandedSceneIds, scene.id, !this.expandedSceneIds[scene.id])
      this.selectedScene = scene
    },
    setAllExpanded (expanded) {
      const next = {}
      this.detailedScenes.forEach(sc => {
        next[sc.id] = expanded
      })
      this.expandedSceneIds = next
    },
    selectSceneGallery (scene) {
      this.selectedScene = scene
      this.galleryMode = 'scene'
      this.carouselSlide = 0
      this.activeMedia = 0
    },
    openSceneDetails (scene) {
      this.close()
      this.$store.commit('overlay/showDetails', { scene })
    },
    onSceneDeleted (scene) {
      this.detailedScenes = this.detailedScenes.filter(sc => sc.id !== scene.id)
      if (this.selectedScene && this.selectedScene.id === scene.id) {
        this.selectedScene = this.detailedScenes.length > 0 ? this.detailedScenes[0] : null
      }
      if (this.currentPlayingScene && this.currentPlayingScene.id === scene.id) {
        this.currentPlayingScene = null
        this.currentPlayingFile = null
      }
    },
    playAllScene (scene) {
      const videos = this.getVideoFiles(scene)
      if (videos.length === 0) return
      this.playQueue = videos
      this.isPlayAllActive = true
      this.playFileAt(scene, videos[0], null, true)
    },
    onPlayerEnded () {
      if (!this.isPlayAllActive || !this.currentPlayingFile || this.playQueue.length === 0) {
        return
      }
      const currentIdx = this.playQueue.findIndex(f => f.id === this.currentPlayingFile.id)
      if (currentIdx >= 0 && currentIdx + 1 < this.playQueue.length) {
        const nextFile = this.playQueue[currentIdx + 1]
        this.playFileAt(this.currentPlayingScene, nextFile, null, true)
      } else {
        this.isPlayAllActive = false
      }
    },
    playFile (scene, file) {
      this.playQueue = []
      this.isPlayAllActive = false
      this.playFileAt(scene, file, null, false)
    },
    playFileAt (scene, file, startSeconds, keepQueue = false) {
      if (!keepQueue) {
        this.playQueue = []
        this.isPlayAllActive = false
      }
      const isSameFile = this.currentPlayingFile && this.currentPlayingFile.id === file.id
      const wasPlayerActive = this.activeMedia === 1

      this.selectedScene = scene
      this.currentPlayingScene = scene
      this.currentPlayingFile = file
      this.activeMedia = 1

      if (startSeconds !== null && startSeconds !== undefined) {
        this.$nextTick(() => {
          const vrPlayer = this.$refs.vrPlayer
          if (!vrPlayer) return
          if (isSameFile && wasPlayerActive) {
            vrPlayer.seekTo(startSeconds)
          } else {
            const videoEl = vrPlayer.$refs && vrPlayer.$refs.vrVideo
            if (videoEl) {
              const onLoaded = () => {
                vrPlayer.seekTo(startSeconds)
              }
              videoEl.addEventListener('loadedmetadata', onLoaded, { once: true })
            }
          }
        })
      }
    },
    replaceSceneInList (updatedScene) {
      if (!updatedScene || !updatedScene.id) return
      const idx = this.detailedScenes.findIndex(sc => sc.id === updatedScene.id)
      if (idx >= 0) {
        this.$set(this.detailedScenes, idx, updatedScene)
      }
      if (this.selectedScene && this.selectedScene.id === updatedScene.id) {
        this.selectedScene = updatedScene
      }
      if (this.currentPlayingScene && this.currentPlayingScene.id === updatedScene.id) {
        this.currentPlayingScene = updatedScene
      }
      this.$store.commit('sceneList/updateScene', updatedScene)
    },
    unmatchFile (file) {
      this.$buefy.dialog.confirm({
        title: 'Unmatch file',
        message: `You're about to unmatch the file <strong>${file.filename}</strong> from this scene. Afterwards, it can be matched again to this or any other scene.`,
        type: 'is-info is-wide',
        hasIcon: true,
        id: 'heh',
        onConfirm: () => {
          ky.post('/api/files/unmatch', { json: { file_id: file.id } }).json().then(data => {
            this.replaceSceneInList(data)
          })
        }
      })
    },
    removeFile (file) {
      this.$buefy.dialog.confirm({
        title: 'Remove file',
        message: `You're about to remove file <strong>${file.filename}</strong> from <strong>disk</strong>.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          ky.delete(`/api/files/file/${file.id}`).json().then(data => {
            this.replaceSceneInList(data)
          })
        }
      })
    },
    deleteThumbnail (file) {
      this.$buefy.dialog.confirm({
        title: 'Delete thumbnail',
        message: `You're about to remove thumbnail <strong>${file.filename}</strong> from <strong>disk</strong>.`,
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          ky.delete(`/api_custom/thumbnail/image/${file.id}`).json().then(data => {
            this.replaceSceneInList(data)
          })
        }
      })
    },
    resetProjection (file) {
      this.$buefy.dialog.confirm({
        title: 'Reset projection',
        message: 'Reset projection?',
        confirmText: 'OK',
        cancelText: 'Cancel',
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          ky.post('/api_custom/files/resetprojection', { json: { file_id: file.id, projection: '' } }).json().then(data => {
            this.replaceSceneInList(data)
          })
        }
      })
    },
    resetFileName (scene, file) {
      this.$buefy.dialog.confirm({
        title: 'Reset filename',
        message: 'Reset the file name?',
        confirmText: 'OK',
        cancelText: 'Cancel',
        type: 'is-danger',
        hasIcon: true,
        onConfirm: () => {
          ky.post('/api_custom/files/resetname', { json: { file_id: file.id, scene_id: scene.scene_id } }).json().then(data => {
            this.replaceSceneInList(data)
          })
        }
      })
    },
    renameFile (file) {
      this.$buefy.dialog.prompt({
        title: 'Rename file',
        message: `Please input new filename for <strong>${file.filename}</strong>`,
        animation: 'fade',
        hasAnimation: false,
        type: 'is-danger',
        hasIcon: true,
        inputAttrs: {
          type: 'text',
          placeholder: 'new filename is...',
          value: file.filename
        },
        onConfirm: value => {
          ky.post('/api_custom/files/rename', { json: { file_id: file.id, filename: value } }).json().then(data => {
            this.replaceSceneInList(data)
          })
        }
      })
    },
    async setOverallRating (val) {
      this.$set(this.actor, 'star_rating', val)
      const updatedActor = Object.assign({}, this.actor)
      updatedActor.star_rating = val
      this.$store.commit('actorList/updateActor', updatedActor)
      try {
        await ky.post(`/api_custom/actor/ratings/${this.actor.id}`, {
          json: { star_rating: val }
        }).json()
      } catch (e) {
        console.error('Failed to save overall actor rating:', e)
      }
    },
    setRating (val) {
      ky.post(`/api/actor/rate/${this.actor.id}`, { json: { rating: val } })
      const updatedActor = Object.assign({}, this.actor)
      updatedActor.star_rating = val
      this.actor.star_rating = val
      this.$store.commit('actorList/updateActor', updatedActor)
    },
    toggleGallery () {
      this.activeMedia = this.activeMedia === 0 ? 1 : 0
    },
    handleLeftArrow () {
      if (this.activeMedia === 0 && this.carouselSlide > 0) {
        this.carouselSlide -= 1
      }
    },
    handleRightArrow () {
      if (this.activeMedia === 0 && this.carouselSlide < this.galleryImages.length - 1) {
        this.carouselSlide += 1
      }
    },
    scrollToActiveIndicator (value) {
      const indicators = document.querySelector('.actor-collection-modal .carousel-indicator')
      if (indicators) {
        const active = indicators.children[value]
        if (active) {
          indicators.scrollLeft = active.offsetLeft - indicators.offsetWidth / 2
        }
      }
    },
    nextActor () {
      const data = this.$store.getters['actorList/nextActor'](this.actor)
      if (data !== null) {
        this.$store.commit('overlay/showActorCollection', { actor: data })
      }
    },
    prevActor () {
      const data = this.$store.getters['actorList/prevActor'](this.actor)
      if (data !== null) {
        this.$store.commit('overlay/showActorCollection', { actor: data })
      }
    },
    close () {
      this.$store.commit('overlay/hideActorCollection')
    }
  }
}
</script>

<style lang="less" scoped>
.actor-collection-modal {
  width: 97vw;
  max-width: 97vw;
  height: 95vh;
  max-height: 95vh;
  display: flex;
  flex-direction: row;
  align-items: stretch;
  border-radius: 8px;
  overflow: hidden;
  background: transparent;
}

.actor-collection-modal .modal-card-body {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 0.75rem 1rem;
  overflow: hidden;
  border-radius: 0;
}

.top-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.left-pane-inner {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-right: 0.5rem;
}

.left-pane-inner ::v-deep .b-tabs {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.left-pane-inner ::v-deep section.tab-content {
  flex: 1;
  min-height: 0;
  height: 100% !important;
  overflow-y: auto;
}

.left-pane-inner ::v-deep .tab-item {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.right-pane-inner {
  height: 100%;
  display: flex;
  flex-direction: column;
  padding-left: 0.5rem;
}

.gallery-toolbar {
  display: flex;
  justify-content: center;
  margin-bottom: 0.5rem;
}

.player-container {
  flex: 1;
  min-height: 0;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.player-footer-bar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.player-file-switcher {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  flex-shrink: 0;
}

.player-file-switcher .button.is-current-file {
  font-weight: 700;
  box-shadow: 0 0 0 2px rgba(121, 87, 213, 0.7);
}

.player-scene-title {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.play-all-btn.is-active {
  background: #7957d5;
  border-color: #8b6ce0;
}

.actor-header {
  padding: 0.75rem;
  margin-bottom: 0.6rem;
  background: #22252d;
  border: 1px solid #353945;
  border-radius: 8px;
  color: #e6e8ef;
}

.actor-header-layout {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
}

.actor-avatar {
  width: 118px;
  height: 142px;
  object-fit: cover;
  border-radius: 6px;
  display: block;
}

.actor-header-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.actor-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.actor-name-wrap {
  display: flex;
  align-items: baseline;
  gap: 0.45rem;
  min-width: 0;
  flex-wrap: wrap;
}

.actor-name {
  font-size: 1.25rem;
  font-weight: 700;
  color: #ffffff;
}

.actor-furigana {
  font-size: 0.9rem;
  color: #9ca3af;
}

.actor-top-right {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  flex-wrap: wrap;
}

.overall-rating-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  padding: 3px 9px;
  border-radius: 6px;
  background: rgba(244, 63, 94, 0.14);
  border: 1px solid rgba(244, 63, 94, 0.35);
}

.overall-label {
  font-size: 0.78rem;
  font-weight: 700;
  color: #fda4af;
}

.overall-rating-pill ::v-deep .vue-star-rating-rating-text {
  color: #fecdd3;
  font-size: 0.8rem;
  font-weight: 700;
}

.overall-reset-icon {
  cursor: pointer;
  color: #fda4af;
  opacity: 0.75;
}

.overall-reset-icon:hover {
  opacity: 1;
}

.actor-fav-btns {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.actor-stats-row {
  display: flex;
  align-items: center;
  gap: 0.4rem;
  flex-wrap: wrap;
}

.actor-stat-pill {
  font-size: 0.75rem;
  padding: 1px 8px;
  border-radius: 4px;
  background: #272a34;
  border: 1px solid #353945;
  color: #d1d5db;
}

.actor-avatar-wrapper {
  position: relative;
}

.actor-hover-tooltip {
  position: absolute;
  z-index: 50;
  top: 0;
  left: 100%;
  width: 300px;
  background-color: white;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.35);
  padding: 8px;
  pointer-events: none;
}

.actor-hover-tooltip img {
  max-width: 100%;
  height: auto;
  display: block;
}

.scenes-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.scenes-scroll {
  flex: 1;
  overflow-y: auto;
  padding-right: 0.25rem;
}

.scene-group-card {
  border: 1px solid #353945;
  border-radius: 8px;
  margin-bottom: 0.85rem;
  background: #22252d;
  color: #e6e8ef;
  overflow: hidden;
}

.scene-group-card.is-selected {
  border-color: #7957d5;
  box-shadow: 0 0 0 1px rgba(121, 87, 213, 0.4);
}

.scene-group-header {
  display: flex;
  flex-direction: column;
  padding: 0.75rem;
  background: #272a34;
  border-bottom: 1px solid #353945;
  cursor: pointer;
  gap: 0.55rem;
}

.scene-header-upper {
  display: flex;
  align-items: flex-start;
  gap: 0.85rem;
}

.scene-header-thumb-wrap {
  width: 148px;
  height: 86px;
  flex-shrink: 0;
  border-radius: 6px;
  overflow: hidden;
  background: #181a20;
}

.scene-header-cover {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.scene-header-cover-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #6b7280;
}

.scene-header-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 0.35rem;
}

.scene-header-top-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 0.5rem;
}

.scene-header-title {
  font-weight: 700;
  font-size: 0.96rem;
  color: #ffffff;
  line-height: 1.3;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
  min-width: 0;
}

.spec-badges.is-right {
  justify-content: flex-end;
  flex-shrink: 0;
}

.scene-right-meta-stack {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.3rem;
  flex-shrink: 0;
}

.scene-overall-pill {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(244, 63, 94, 0.14);
  border: 1px solid rgba(244, 63, 94, 0.35);
}

.scene-overall-pill ::v-deep .vue-star-rating-rating-text {
  color: #fecdd3;
  font-size: 0.76rem;
  font-weight: 700;
}

.scene-header-sub {
  font-size: 0.78rem;
  color: #9ca3af;
}

.scene-header-bottom-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.scene-header-actions-bar {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  flex-wrap: wrap;
}

.scene-header-actions-bar ::v-deep .button.is-dark.is-outlined {
  color: #d1d5db;
  border-color: #4b5563;
  background: transparent;
}

.scene-header-actions-bar ::v-deep .button.is-dark.is-outlined:hover {
  color: #fff;
  background: #374151;
}

.play-all-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  background: #363a45;
  color: #fff;
  border: 1px solid #4b505e;
  border-radius: 4px;
  padding: 0.28rem 0.65rem;
  font-size: 0.78rem;
  font-weight: 600;
  cursor: pointer;
}

.play-all-btn:hover {
  background: #464b59;
}

.scene-icon-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #323640;
  color: #d1d5db;
  border: 1px solid #454a57;
  border-radius: 4px;
  padding: 0.28rem 0.45rem;
  cursor: pointer;
}

.scene-icon-btn:hover {
  background: #424754;
  color: #fff;
}

.spec-badges {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  flex-wrap: wrap;
}

.spec-badge {
  font-size: 0.72rem;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  border: 1px solid;
  line-height: 1.35;
}

.spec-badge.is-res {
  color: #facc15;
  border-color: #ca8a04;
  background: rgba(234, 179, 8, 0.12);
}

.spec-badge.is-codec {
  color: #c084fc;
  border-color: #9333ea;
  background: rgba(168, 85, 247, 0.12);
}

.spec-badge.is-proj {
  color: #2dd4bf;
  border-color: #0d9488;
  background: rgba(20, 184, 166, 0.12);
}

.scene-files-body {
  padding: 0.25rem 0.75rem 0.5rem;
  background: #1f222a;
}

.file-table-header {
  display: grid;
  grid-template-columns: 36px minmax(180px, 1fr) 74px 72px auto;
  gap: 0.5rem;
  padding: 0.35rem 0.4rem;
  font-size: 0.75rem;
  color: #9ca3af;
  border-bottom: 1px solid #323642;
}

.file-table-row {
  padding: 0.55rem 0.4rem;
  border-bottom: 1px solid #2b2f3a;
}

.file-row-main {
  display: grid;
  grid-template-columns: 36px minmax(180px, 1fr) 74px 72px auto;
  gap: 0.5rem;
  align-items: center;
}

.file-table-row:last-child {
  border-bottom: none;
}

.file-table-row.is-playing {
  background-color: rgba(121, 87, 213, 0.18);
  border-radius: 4px;
}

.col-play {
  padding-top: 2px;
}

.col-name {
  min-width: 0;
}

.file-name-text {
  font-weight: 600;
  font-size: 0.85rem;
  color: #f3f4f6;
  overflow-wrap: break-word;
}

.file-meta-sub {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-top: 2px;
}

.file-path-text {
  font-size: 0.74rem;
  color: #9ca3af;
  overflow-wrap: break-word;
}

.col-duration,
.col-size {
  font-size: 0.82rem;
  color: #d1d5db;
  white-space: nowrap;
}

.col-actions,
.col-actions-head {
  display: flex;
  justify-content: flex-end;
}

.file-row-previews {
  margin-top: 0.45rem;
  padding-left: 42px;
}

.file-actions-group {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  flex-wrap: nowrap;
}

.file-actions-group .button.is-dark.is-outlined {
  color: #d1d5db;
  border-color: #4b5563;
  background: transparent;
}

.file-actions-group .button.is-dark.is-outlined:hover {
  color: #fff;
  background: #374151;
}

.docked-nav-btn {
  width: 52px;
  flex-shrink: 0;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 8px 4px;
  background: #22252d;
  border: 1px solid #353945;
  color: #e6e8ef;
  font-weight: bold;
  font-size: 16px;
  user-select: none;
  transition: background 0.15s ease, border-color 0.15s ease;
}

.docked-nav-btn.is-prev {
  border-right: none;
  border-radius: 8px 0 0 8px;
}

.docked-nav-btn.is-next {
  border-left: none;
  border-radius: 0 8px 8px 0;
}

.docked-nav-btn:hover:not(.is-disabled) {
  background: #7957d5;
  border-color: #8b6ce0;
  color: #ffffff;
}

.docked-nav-btn.is-disabled {
  opacity: 0.25;
  cursor: default;
  pointer-events: none;
}

.nav-actor-face {
  width: 38px;
  height: 38px;
  object-fit: cover;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.5);
  display: block;
}

.nav-arrow {
  line-height: 1;
}
</style>
