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
          style="height: 84vh;"
          :gutter-size="10"
          :min-pane-size="25"
          :max-pane-size="100"
          ref="splitpanesRef"
        >
          <!-- 左ペイン: ギャラリー兼プレイヤー -->
          <pane :min-size="25" :max-size="100" :size="hideRightPane ? 100 : 50">
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
                  <div :class="['video-player-wrapper', aspectClass]" ref="videoContainer">
                    <video
                      id="actorCollectionVideoPlayer"
                      ref="player"
                      class="video-js vjs-default-skin video-custom"
                      controls
                      playsinline
                      preload="none"
                      muted
                    />
                  </div>

                  <div class="player-controls-row">
                    <b-select v-model="projectionMode" size="is-small" placeholder="Projection">
                      <option value="NONE">Flat</option>
                      <option value="180_LR">180_LR</option>
                      <option value="180_MONO">180_MONO</option>
                      <option value="360_TB">360</option>
                      <option value="AUTO">AUTO</option>
                    </b-select>
                    <b-select v-model="aspectRatio" size="is-small" placeholder="Aspect">
                      <option value="1:1">1:1</option>
                      <option value="4:3">4:3</option>
                      <option value="16:9">16:9</option>
                    </b-select>
                    <span v-if="currentPlayingScene" class="is-size-7 has-text-grey">
                      {{ currentPlayingScene.title }}
                    </span>
                  </div>
                </b-tab-item>
              </b-tabs>
            </div>
          </pane>

          <!-- 右ペイン: Actor 情報 + シーンごとの作品・動画ファイル一覧 -->
          <pane v-if="!hideRightPane" :min-size="25" :max-size="75" :size="50">
            <div class="right-pane-inner">
              <!-- Actor ヘッダ -->
              <div class="actor-header box is-shadowless">
                <div class="columns is-vcentered is-mobile">
                  <div class="column is-narrow" v-if="actor.face_image_url || actor.image_url">
                    <img
                      :src="getImageURL(actor.face_image_url || actor.image_url, '300,fit')"
                      class="actor-avatar"
                      alt="Actor"
                    />
                  </div>
                  <div class="column">
                    <h3 class="title is-4 mb-1">
                      {{ actor.name }}
                      <small v-if="actor.furigana" class="is-size-6 has-text-grey ml-2">
                        ({{ actor.furigana }})
                      </small>
                    </h3>
                    <div class="is-flex is-align-items-center" style="gap: 0.75rem; flex-wrap: wrap;">
                      <star-rating
                        :key="actor.id"
                        v-model="actor.star_rating"
                        :rating="actor.star_rating"
                        @rating-selected="setRating"
                        :increment="0.5"
                        :star-size="18"
                        :show-rating="true"
                      />
                      <div>
                        <actor-favourite-button :actor="actor" />
                        <actor-watchlist-button :actor="actor" />
                      </div>
                    </div>
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
                  <!-- シーンヘッダ -->
                  <div class="scene-group-header" @click="toggleSceneExpand(scene)">
                    <div class="scene-header-left">
                      <b-icon
                        pack="mdi"
                        :icon="isSceneExpanded(scene.id) ? 'chevron-down' : 'chevron-right'"
                        size="is-small"
                      />
                      <img
                        v-if="scene.cover_url"
                        :src="getImageURL(scene.cover_url, '120,fit')"
                        class="scene-mini-cover"
                        alt="Cover"
                      />
                      <div class="scene-header-meta">
                        <div class="scene-header-title">{{ scene.title }}</div>
                        <div class="is-size-7 has-text-grey">
                          <span>{{ scene.site }}</span>
                          <span v-if="scene.release_date && scene.release_date !== '0001-01-01T00:00:00Z'">
                            • {{ format(parseISO(scene.release_date), 'yyyy-MM-dd') }}
                          </span>
                          <span>• {{ getVideoFiles(scene).length }} video(s)</span>
                        </div>
                      </div>
                    </div>
                    <div class="scene-header-actions" @click.stop>
                      <button
                        class="button is-small is-light"
                        title="Show scene gallery"
                        @click="selectSceneGallery(scene)"
                      >
                        <b-icon pack="mdi" icon="image-multiple" size="is-small" />
                      </button>
                      <button
                        class="button is-small is-light"
                        title="Open Scene Details"
                        @click="openSceneDetails(scene)"
                      >
                        <b-icon pack="mdi" icon="open-in-new" size="is-small" />
                      </button>
                    </div>
                  </div>

                  <!-- シーン内ファイルリスト (Scenes Details と同等の構成 + 抜粋サムネイル) -->
                  <div v-show="isSceneExpanded(scene.id)" class="scene-files-body">
                    <div
                      v-if="!scene.files || scene.files.length === 0"
                      class="is-size-7 has-text-grey px-3 py-2"
                    >
                      No files linked to this scene.
                    </div>

                    <div
                      class="content media is-small scene-file-row"
                      v-for="f in sortedFiles(scene)"
                      :key="f.id"
                      :class="{ 'is-playing': currentPlayingFile && currentPlayingFile.id === f.id }"
                    >
                      <div class="media-left">
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

                      <div class="media-content" style="overflow-wrap: break-word; min-width: 0;">
                        <strong>{{ f.filename }}</strong>
                        <br />
                        <small>
                          <span class="pathDetails">{{ f.path }}</span>
                          <br />
                          {{ prettyBytes(f.size) }}
                          <span v-if="f.type === 'video'">
                            ({{ prettyBytes(f.video_bitrate, { bits: true }) }}/s),
                            <span class="videosize">{{ f.video_width }}x{{ f.video_height }} {{ f.video_codec_name }}</span>,
                            {{ f.projection }},&nbsp;
                          </span>
                          <span v-if="f.duration > 1">{{ humanizeSeconds(f.duration) }},</span>
                          {{ format(parseISO(f.created_time), 'yyyy-MM-dd') }}
                        </small>

                        <div v-if="f.type === 'script' && f.has_heatmap" class="heatmapFunscript">
                          <img :src="getHeatmapURL(f.id)" />
                        </div>

                        <!-- サムネイル抜粋表示 -->
                        <FileThumbnailStrip
                          v-if="f.type === 'video' && f.has_thumbnail"
                          :file="f"
                          :max-count="5"
                          @thumbnail-clicked="sec => playFileAt(scene, f, sec)"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </pane>
        </splitpanes>
      </section>
    </div>

    <button class="modal-close is-large" aria-label="close" @click="close"></button>
    <a class="prev" @click="prevActor" title="Keyboard shortcut: O">&#10094;</a>
    <a class="next" @click="nextActor" title="Keyboard shortcut: P">&#10095;</a>
  </div>
</template>

<script>
import ky from 'ky'
import videojs from 'video.js'
import 'videojs-vr/dist/videojs-vr.min.js'
import 'videojs-thumbnail-sprite-custom'
import prettyBytes from 'pretty-bytes'
import { format, parseISO } from 'date-fns'
import VueLoadImage from 'vue-load-image'
import GlobalEvents from 'vue-global-events'
import StarRating from 'vue-star-rating'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'

import ActorFavouriteButton from '../../components/ActorFavouriteButton'
import ActorWatchlistButton from '../../components/ActorWatchlistButton'
import FileThumbnailStrip from '../../components/FileThumbnailStrip'

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
    FileThumbnailStrip
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
      player: null,
      projectionMode: '180_LR',
      aspectRatio: '1:1'
    }
  },
  computed: {
    actor () {
      return this.$store.state.overlay.actorcollection.actor
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
    aspectClass () {
      return {
        'aspect-1-1': this.aspectRatio === '1:1',
        'aspect-4-3': this.aspectRatio === '4:3',
        'aspect-16-9': this.aspectRatio === '16:9'
      }
    }
  },
  watch: {
    'actor.id': {
      immediate: true,
      handler () {
        this.loadActorScenes()
      }
    },
    projectionMode () {
      if (this.player && this.currentPlayingFile) {
        this.$nextTick(() => {
          this.restartPlayer()
        })
      }
    },
    aspectRatio () {
      this.$nextTick(() => {
        if (this.player && typeof this.player.aspectRatio === 'function') {
          this.player.aspectRatio(this.aspectRatio)
        }
        window.dispatchEvent(new Event('resize'))
      })
    }
  },
  mounted () {
    this.setupPlayer()
  },
  beforeDestroy () {
    if (this.player) {
      this.player.dispose()
      this.player = null
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
      if (!scene || !scene.files) return []
      return scene.files.filter(f => f.type === 'video')
    },
    sortedFiles (scene) {
      if (!scene || !scene.files) return []
      const order = { video: 0, script: 1, hsp: 2, subtitles: 3 }
      return [...scene.files].sort((a, b) => (order[a.type] ?? 9) - (order[b.type] ?? 9))
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
    setupPlayer () {
      this.player = videojs('actorCollectionVideoPlayer', {
        aspectRatio: this.aspectRatio,
        fluid: true,
        loop: true,
        restoreEl: true
      })
      const videoElement = this.player.el()
      videoElement.addEventListener('wheel', (event) => {
        event.preventDefault()
        const vr = this.player.vr()
        if (!vr || !vr.camera) return
        let fov = vr.camera.fov + (event.deltaY < 0 ? -1 : 1)
        fov = Math.max(30, Math.min(130, fov))
        vr.camera.fov = fov
        vr.camera.updateProjectionMatrix()
      })
    },
    updatePlayer (src, projection, posterUrl) {
      if (!this.player) return
      this.player.reset()
      this.player.vr({
        projection: projection,
        forceCardboard: false
      })
      if (src) {
        this.player.src({ src, type: 'video/mp4' })
      }
      if (posterUrl) {
        setTimeout(() => {
          if (this.player) {
            this.player.poster(this.getImageURL(posterUrl, '700,fit'))
          }
        }, 500)
      }
    },
    playFile (scene, file) {
      this.playFileAt(scene, file, null)
    },
    playFileAt (scene, file, startSeconds) {
      this.selectedScene = scene
      this.currentPlayingScene = scene
      this.currentPlayingFile = file
      this.activeMedia = 1

      const proj = file.projection === 'flat' ? 'NONE' : '180'
      this.$nextTick(() => {
        this.updatePlayer(`/api/dms/file/${file.id}?dnt=true`, proj, scene.cover_url)
        this.setupSprite(file)
        if (startSeconds !== null && startSeconds !== undefined) {
          this.player.one('loadedmetadata', () => {
            this.player.currentTime(Math.floor(startSeconds))
          })
        }
        this.player.play()
      })
    },
    restartPlayer () {
      if (!this.player) return
      const currentTime = this.player.currentTime()
      const currentSrc = this.player.currentSrc()
      const wasPlaying = !this.player.paused()
      const poster = this.currentPlayingScene ? this.currentPlayingScene.cover_url : ''

      this.player.dispose()
      this.player = null
      this.setupPlayer()
      this.updatePlayer(currentSrc, this.projectionMode, poster)

      this.player.ready(() => {
        this.player.load()
        if (wasPlaying) {
          this.player.play()
          this.player.currentTime(Math.floor(currentTime))
          this.setupSprite(this.currentPlayingFile)
        } else {
          this.player.one('playing', () => {
            this.player.pause()
            this.player.currentTime(Math.floor(currentTime))
            this.setupSprite(this.currentPlayingFile)
          })
          this.player.play()
        }
      })
    },
    setupSprite (file) {
      const videoPlayer = this.player
      if (!videoPlayer) return
      if (!file || !file.has_thumbnail) {
        if (typeof videoPlayer.thumbnailSprite === 'function') {
          const plugin = videoPlayer.thumbnailSprite()
          if (plugin && typeof plugin.reset === 'function') {
            plugin.reset()
          }
        }
        return
      }
      let parsed = {}
      try {
        parsed = typeof file.thumbnail_parameters === 'string'
          ? JSON.parse(file.thumbnail_parameters || '{}')
          : (file.thumbnail_parameters || {})
      } catch (e) {
        parsed = {}
      }
      const params = {
        url: `/api_custom/thumbnail/image/${file.id}`,
        start: parsed.start ?? 15,
        interval: parsed.interval ?? 15,
        width: parsed.resolution ?? 200,
        height: parsed.resolution ?? 200
      }
      if (typeof videoPlayer.thumbnailSprite === 'function') {
        videoPlayer.thumbnailSprite({ sprites: [params] })
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
      if (this.player) {
        this.player.dispose()
        this.player = null
      }
      this.$store.commit('overlay/hideActorCollection')
    }
  }
}
</script>

<style lang="less" scoped>
.actor-collection-modal {
  width: 92vw;
  max-width: 1600px;
}

.top-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.left-pane-inner {
  height: 100%;
  overflow-y: auto;
  padding-right: 0.5rem;
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

.player-controls-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
  flex-wrap: wrap;
}

.actor-header {
  padding: 0.65rem 0.85rem;
  margin-bottom: 0.5rem;
  background-color: #fafafa;
  border: 1px solid #eaeaea;
}

.actor-avatar {
  width: 64px;
  height: 64px;
  object-fit: cover;
  border-radius: 6px;
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
  border: 1px solid #e2e2e2;
  border-radius: 6px;
  margin-bottom: 0.65rem;
  background: #fff;
}

.scene-group-card.is-selected {
  border-color: #7957d5;
}

.scene-group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.65rem;
  background: #f7f7f9;
  border-bottom: 1px solid #ececec;
  cursor: pointer;
  gap: 0.5rem;
}

.scene-header-left {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
  flex: 1;
}

.scene-mini-cover {
  width: 48px;
  height: 36px;
  object-fit: cover;
  border-radius: 3px;
  flex-shrink: 0;
}

.scene-header-meta {
  min-width: 0;
  flex: 1;
}

.scene-header-title {
  font-weight: 600;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.scene-header-actions {
  display: flex;
  gap: 0.25rem;
  flex-shrink: 0;
}

.scene-files-body {
  padding: 0.35rem 0.65rem;
}

.scene-file-row {
  padding: 0.4rem 0.25rem;
  margin-bottom: 0 !important;
}

.scene-file-row.is-playing {
  background-color: #f3effd;
  border-radius: 4px;
}

.pathDetails {
  color: #b5b5b5;
}

.prev,
.next {
  cursor: pointer;
  position: fixed;
  top: 50%;
  width: auto;
  padding: 16px;
  margin-top: -50px;
  color: white;
  font-weight: bold;
  font-size: 24px;
  user-select: none;
}

.prev {
  left: 0;
}

.next {
  right: 0;
}
</style>
