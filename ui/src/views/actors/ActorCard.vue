<template>
  <!-- Amazon Prime Video 風カード (Collectionタブ) -->
  <div v-if="isPrimeCollection" class="prime-card" @click="showDetails(actor)">
    <div class="prime-poster-wrap">
      <div
        class="prime-poster-img"
        :style="{
          backgroundImage: `url(${getImageURL(getActorImageURL(actor))})`,
          backgroundSize: actorCardScale,
          opacity: isAvailable(actor) ? 1.0 : isAvailOpacity
        }"
      ></div>

      <!-- 上部バッジ -->
      <div class="prime-top-badges">
        <span v-if="actor.avail_count > 0" class="prime-avail-badge">
          <span class="prime-check">✓</span>
          {{ actor.avail_count }} 本
        </span>
        <span v-else class="prime-scene-badge">
          {{ actor.scenes ? actor.scenes.length : 0 }} 作品
        </span>

        <span v-if="actor.star_rating > 0" class="prime-rating-badge">
          ★ {{ actor.star_rating }}
        </span>
      </div>

      <!-- ホバー時のクイック操作バー (高さ変化なしでオーバーレイ表示) -->
      <div class="prime-hover-actions" @click.stop>
        <button type="button" class="prime-play-btn" @click="showDetails(actor)" title="Open Collection">
          <b-icon pack="mdi" icon="play" size="is-small" />
          <span>開く</span>
        </button>
        <actor-favourite-button :actor="actor" v-if="$store.state.optionsWeb.web.sceneFavourite" />
        <actor-watchlist-button :actor="actor" v-if="$store.state.optionsWeb.web.sceneWatchlist" />
        <actor-edit-button :actor="actor" />
        <actor-edit-image-button :actor="actor" />
      </div>
    </div>

    <!-- 画像下フッタ（画像に被らないタイトル・メタ情報） -->
    <div class="prime-card-footer">
      <div class="prime-actor-name" :title="actor.name">{{ actor.name }}</div>
      <div v-if="actor.furigana" class="prime-actor-furigana" :title="actor.furigana">{{ actor.furigana }}</div>
      <div class="prime-meta-row">
        <span>{{ actor.scenes ? actor.scenes.length : 0 }} scenes</span>
        <span v-if="actor.scene_rating_average > 0" class="prime-scene-avg">
          作品 ★{{ Math.round(actor.scene_rating_average * 4) / 4 }}
        </span>
      </div>
    </div>
  </div>

  <!-- 従来のカード (Actorsタブ) -->
  <div v-else class="card is-shadowless">
    <div class="card-image">
      <!-- Custom Black（顔表示を追加） -->
      <div class="bbox"          
           v-bind:style="{backgroundImage: `url(${getImageURL(getActorImageURL(actor))})`, backgroundSize: actorCardScale, backgroundPosition: 'center', backgroundRepeat: 'no-repeat', opacity:isAvailable(actor) ? 1.0 : isAvailOpacity, aspectRatio: actorCardAspectRatio}"
           @click="showDetails(actor)"
           @mouseover="preview = true"
           @mouseleave="preview = false">
      <!-- Custom End -->
      <!-- バックアップ用
      v-bind:style="{backgroundImage: `url(${getActorImageURL(actor)})`, backgroundSize: 'contain', backgroundPosition: 'center', backgroundRepeat: 'no-repeat', opacity:isAvailable(actor) ? 1.0 : isAvailOpactiy}"
      -->
      </div>
        <div class="overlay align-bottom-left">
         </div>
    </div>

    <div style="padding-top:4px;">
      <div class="scene_title">{{actor.name}}</div>
      <!-- Custom Blakc（フリガナ表示） -->
      <div v-if="actor.furigana!=''" class="furigana">{{actor.furigana}}</div>
      <!-- Custom End -->
      <a v-if="colleague!=undefined" class="button is-info is-outlined is-small"
        @click="showColleague(actor.name,colleague)"
        :title="'Show Scenes with ' + actor.name">
        <b-icon pack="mdi" :icon="'movie-outline'" size="is-small"/>
      </a>
      <actor-favourite-button :actor="actor" v-if="this.$store.state.optionsWeb.web.sceneFavourite"/>
      <actor-watchlist-button :actor="actor" v-if="this.$store.state.optionsWeb.web.sceneWatchlist"/>
      <actor-edit-button :actor="actor"/>
      <!-- Custom Black（actor画像編集ボタン） -->
      <actor-edit-image-button :actor="actor"/>
      <!-- Custom End -->
      <link-stashdb-button :item="actor" objectType="actor" />
      <b-tooltip :label="$t('Your rating')" :delay="500">
      <b-tag type="is-warning" v-if="actor.star_rating != 0 " size="is-small" style="height:30px;">
        <b-icon pack="mdi" icon="star" size="is-small"/>
        {{actor.star_rating}}
      </b-tag>
      </b-tooltip>
      <b-tooltip :label="$t('Average rating of scenes')" :delay="500">
      <b-tag type="is-primary" v-if="actor.scene_rating_average != 0 " style="height:30px;">
        <b-icon pack="mdi" icon="star" size="is-small"/>
        {{Math.round(actor.scene_rating_average * 4) / 4}}
      </b-tag>
      </b-tooltip>

      <span class="is-pulled-right" style="font-size:11px;text-align:right;">                
          <span v-if="actor.birth_date != '0001-01-01T00:00:00Z'">{{format(parseISO(actor.birth_date), "yyyy-MM-dd")}}</span>
          <vue-load-image style="display:inline-block">
            <img slot="image" :src="getImageURL('https://flagcdn.com/' + actor.nationality.toLowerCase() +'.svg')" style="height:10px;border: 1px solid black;margin-left:0.5em" />
          </vue-load-image>
          <div>
          <span v-if="actor.scenes.length == 1">{{actor.scenes.length}} scene</span>
          <span v-if="actor.scenes.length > 1">{{actor.scenes.length}} scenes</span>
          <span v-if="actor.avail_count > 0">, {{actor.avail_count}} available</span>
          </div>
      </span>
    </div>
  </div>
</template>

<script>
import { format, parseISO } from 'date-fns'
import ActorFavouriteButton from '../../components/ActorFavouriteButton'
import ActorWatchlistButton from '../../components/ActorWatchlistButton'
import ActorEditButton from '../../components/ActorEditButton'
import LinkStashdbButton from '../../components/LinkStashdbButton'
import VueLoadImage from 'vue-load-image'
import { tr } from 'date-fns/locale'

// Custom Black(Actor 画像編集ボタン追加)
import ActorEditImageButton from '../../components/ActorEditImageButton.vue'
// Custom End

export default {
  name: 'ActorCard',
  // Custom Black(顔表示用のprops追加)
  props: { actor: Object, colleague: String, showFace: Boolean },
  // Custom End

  // Custom Black(Actor 画像編集ボタン追加)
  components: {ActorFavouriteButton, ActorWatchlistButton, VueLoadImage, ActorEditButton, LinkStashdbButton,
               ActorEditImageButton,
   },
   // Custom End
  data () {
    return {
      preview: false,
      format,
      parseISO
    }
  },
  computed: {
    isPrimeCollection () {
      return this.$route && this.$route.name === 'actorsCollection'
    },
    isAvailOpacity () {      
      if (this.$store.state.optionsWeb.web.isAvailOpacity == undefined) {
        return .4
      }
      return this.$store.state.optionsWeb.web.isAvailOpacity / 100
    },
    actorCardAspectRatio () {
      if (this.$store.state.optionsWeb.web.actorCardAspectRatio == "2:3") {
        return 2 / 3
      } else if (this.$store.state.optionsWeb.web.actorCardAspectRatio == "9:16") {
        return 9 / 16
      } else {
        return 1
      }
    },
    actorCardScale () {
      if (this.$store.state.optionsWeb.web.actorCardScaleToFit) {
        return "contain"
      } else {
        return "cover"
      }
    },
  },
  methods: {
    // Custom Black(顔表示用のメソッド追加)
    getActorImageURL(actor) {
      if (!actor) return ''
      return (this.showFace && actor.face_image_url) ? actor.face_image_url : actor.image_url
    },
    // Custom END
    getImageURL (u) {
      if (u=='' || u == undefined) {
        return "/ui/images/blank_female_profile.png"
      }
      if (u.startsWith('http')) {
        return '/img/700x/' + u.replace('://', ':/')
      } else {
        return u
      }
    },
    recordRecentCollectionActor (actor) {
      if (!actor || !actor.id) return
      const key = 'xbvr_recent_collection_actors'
      try {
        const raw = localStorage.getItem(key)
        const list = raw ? JSON.parse(raw) : []
        const filtered = Array.isArray(list) ? list.filter(a => a && a.id !== actor.id) : []
        const snapshot = {
          id: actor.id,
          actor_id: actor.actor_id,
          name: actor.name,
          furigana: actor.furigana,
          image_url: actor.image_url,
          face_image_url: actor.face_image_url,
          star_rating: actor.star_rating,
          scene_rating_average: actor.scene_rating_average,
          avail_count: actor.avail_count,
          nationality: actor.nationality,
          birth_date: actor.birth_date,
          watchlist: actor.watchlist,
          favourite: actor.favourite,
          genre_tags: actor.genre_tags || '',
          scenes: Array.isArray(actor.scenes)
            ? actor.scenes.map(s => ({ id: s.id, is_available: s.is_available }))
            : []
        }
        filtered.unshift(snapshot)
        localStorage.setItem(key, JSON.stringify(filtered.slice(0, 50)))
        window.dispatchEvent(new CustomEvent('xbvr-recent-actors-updated'))
      } catch (e) {
        // ignore storage quota errors
      }
    },
    showDetails (actor) {
      if (this.$route && this.$route.name === 'actorsCollection') {
        this.recordRecentCollectionActor(actor)
        this.$store.commit('overlay/showActorCollection', { actor: actor })
      } else {
        this.$store.commit('overlay/showActorDetails', { actor: actor })
      }
    },
    isAvailable(actor) {
      let index = actor.scenes.findIndex(scene => scene.is_available == 1);
      if (index == -1) {
        return false
      }
      return true
    },
    showColleague (main_actor, colleague) {      
      this.$store.state.sceneList.filters.cast = ["&" + main_actor , "&"+ colleague]
      this.$store.state.sceneList.filters.sites = []
      this.$store.state.sceneList.filters.tags = []
      this.$store.state.sceneList.filters.attributes = []
      this.$store.state.actorList.filters.dlState = "Any"
      this.$router.push({
        name: 'scenes',
        query: { q: this.$store.getters['sceneList/filterQueryParams'] }
      })
      this.$store.commit('overlay/hideActorDetails')
    },

  },
}
</script>

<style scoped>
  .button {
    margin-right: 3px;
  }

  .bbox {
    flex: 1 0 calc(25%);
    background: #f0f0f0;
    display: flex;
    align-items: center;
    justify-content: center;
    overflow: hidden;
    padding: 0;
    line-height: 0;
  }

  .bbox:not(:hover) > video {
    display: none;
  }

  video {
    object-fit: cover;
    position: absolute;
    width: 100%;
    height: 100%;
  }

  .overlay {
   position: absolute;
  bottom: 0;
  right: 0;
  display: flex;  
  padding: 5px;
  max-width: 5px;
  }

  .align-bottom-left {
    align-items: flex-end;
    justify-content: flex-end;
  }

  .bbox:after {
    content: '';
    display: block;
    padding-bottom: 100%;
  }

  .tag {
    margin-left: 0.1em;
  }

  .scene_title {
    font-size: 12px;
    text-align: right;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  /* Custom Black（フリガナ用CSS追加） */
  .furigana {
    font-size: 11px;
    text-align: right;
    color: #aa0a7d;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  /* Custom END */

  /* Amazon Prime Video 風カードスタイル */
  .prime-card {
    position: relative;
    border-radius: 8px;
    overflow: hidden;
    background: #19242e;
    border: 1px solid rgba(255, 255, 255, 0.08);
    cursor: pointer;
    transition: transform 0.22s cubic-bezier(0.22, 1, 0.36, 1), box-shadow 0.22s ease, border-color 0.22s ease;
  }

  .prime-card:hover {
    transform: translateY(-4px) scale(1.03);
    z-index: 10;
    border-color: rgba(0, 168, 225, 0.7);
    box-shadow: 0 14px 30px rgba(0, 0, 0, 0.65), 0 0 0 1px rgba(0, 168, 225, 0.35);
  }

  .prime-poster-wrap {
    position: relative;
    width: 100%;
    overflow: hidden;
  }

  .prime-poster-img {
    width: 100%;
    background-position: center top;
    background-repeat: no-repeat;
    background-color: #111921;
    transition: transform 0.35s ease;
  }

  .prime-poster-img:after {
    content: '';
    display: block;
    padding-bottom: 130%;
  }

  .prime-card:hover .prime-poster-img {
    transform: scale(1.05);
  }

  .prime-top-badges {
    position: absolute;
    top: 8px;
    left: 8px;
    right: 8px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    pointer-events: none;
  }

  .prime-avail-badge {
    display: inline-flex;
    align-items: center;
    gap: 3px;
    padding: 2px 7px;
    border-radius: 4px;
    background: rgba(0, 168, 225, 0.9);
    color: #ffffff;
    font-size: 0.68rem;
    font-weight: 700;
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.4);
  }

  .prime-check {
    font-weight: 900;
  }

  .prime-scene-badge {
    padding: 2px 7px;
    border-radius: 4px;
    background: rgba(15, 23, 30, 0.78);
    color: #cbd5e1;
    font-size: 0.68rem;
    font-weight: 600;
  }

  .prime-rating-badge {
    padding: 2px 7px;
    border-radius: 4px;
    background: rgba(15, 23, 30, 0.82);
    color: #f59e0b;
    font-size: 0.7rem;
    font-weight: 700;
    border: 1px solid rgba(245, 158, 11, 0.35);
  }

  .prime-card-footer {
    padding: 0.48rem 0.58rem 0.52rem;
    background: #16202a;
    border-top: 1px solid rgba(255, 255, 255, 0.06);
  }

  .prime-actor-name {
    font-size: 0.86rem;
    font-weight: 700;
    color: #ffffff;
    line-height: 1.25;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .prime-actor-furigana {
    font-size: 0.68rem;
    color: #94a3b8;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .prime-meta-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 2px;
    font-size: 0.68rem;
    color: #8197a4;
  }

  .prime-scene-avg {
    color: #fbbf24;
    font-weight: 600;
  }

  .prime-hover-actions {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 5;
    display: flex;
    align-items: center;
    gap: 0.2rem;
    padding: 0.38rem 0.42rem;
    background: rgba(15, 23, 30, 0.88);
    backdrop-filter: blur(4px);
    border-top: 1px solid rgba(255, 255, 255, 0.1);
    opacity: 0;
    transform: translateY(100%);
    pointer-events: none;
    transition: opacity 0.18s ease, transform 0.2s cubic-bezier(0.22, 1, 0.36, 1);
  }

  .prime-card:hover .prime-hover-actions {
    opacity: 1;
    transform: translateY(0);
    pointer-events: auto;
  }

  .prime-play-btn {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 3px 7px;
    border: none;
    border-radius: 4px;
    background: #00a8e1;
    color: #ffffff;
    font-size: 0.72rem;
    font-weight: 700;
    white-space: nowrap;
    flex-shrink: 0;
    cursor: pointer;
    margin-right: auto;
  }

  .prime-play-btn:hover {
    background: #008ec2;
  }
</style>
