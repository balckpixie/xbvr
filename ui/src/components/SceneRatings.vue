<template>
  <div class="scene-ratings-wrap" :class="{ 'is-dark': dark, 'is-grid-2x2': grid2x2 }">
    <!-- 総合 (Details.vue など showOverall=true の場合のみ表示) -->
    <div v-if="showOverall" class="scene-rating-item scene-overall-pill mb-1">
      <span class="overall-label">総合</span>
      <star-rating
        :key="`${item.id}-overall`"
        :rating="effectiveOverallRating"
        read-only
        :increment="0.1"
        :star-size="starSize"
        :show-rating="true"
        active-color="#f43f5e"
      />
    </div>

    <div class="scene-ratings-box">
      <!-- 2x2 の順: 女優 / プレイ / 画質 / 抜け感 -->
      <div class="scene-rating-item">
        <span class="scene-rating-label">女優</span>
        <star-rating
          :key="`${item.id}-actor-avg`"
          :rating="actorAverageRating"
          read-only
          :increment="0.1"
          :star-size="starSize"
          :show-rating="true"
          active-color="#48c774"
        />
      </div>

      <div class="scene-rating-item">
        <span class="scene-rating-label">プレイ</span>
        <star-rating
          :key="`${item.id}-rating_play`"
          :rating="item.rating_play || 0"
          @rating-selected="val => setRating('rating_play', val)"
          :increment="1"
          :star-size="starSize"
          :show-rating="true"
          active-color="#48c774"
        />
        <b-tooltip :label="$t('Reset Rating')" position="is-right" :delay="250">
          <b-icon
            pack="mdi"
            icon="autorenew"
            size="is-small"
            class="reset-icon"
            @click.native="setRating('rating_play', 0)"
          />
        </b-tooltip>
      </div>

      <div class="scene-rating-item">
        <span class="scene-rating-label">画質</span>
        <star-rating
          :key="`${item.id}-rating_video_quality`"
          :rating="item.rating_video_quality || 0"
          @rating-selected="val => setRating('rating_video_quality', val)"
          :increment="1"
          :star-size="starSize"
          :show-rating="true"
          active-color="#48c774"
        />
        <b-tooltip :label="$t('Reset Rating')" position="is-right" :delay="250">
          <b-icon
            pack="mdi"
            icon="autorenew"
            size="is-small"
            class="reset-icon"
            @click.native="setRating('rating_video_quality', 0)"
          />
        </b-tooltip>
      </div>

      <div class="scene-rating-item">
        <span class="scene-rating-label">抜け感</span>
        <star-rating
          :key="`${item.id}-rating_appeal`"
          :rating="item.rating_appeal || 0"
          @rating-selected="val => setRating('rating_appeal', val)"
          :increment="1"
          :star-size="starSize"
          :show-rating="true"
          active-color="#48c774"
        />
        <b-tooltip :label="$t('Reset Rating')" position="is-right" :delay="250">
          <b-icon
            pack="mdi"
            icon="autorenew"
            size="is-small"
            class="reset-icon"
            @click.native="setRating('rating_appeal', 0)"
          />
        </b-tooltip>
      </div>
    </div>

    <!-- シーンジャンルタグ -->
    <div v-if="showGenres" class="scene-genre-buttons">
      <button
        v-for="tag in genreOptions"
        :key="tag"
        type="button"
        class="scene-genre-pill"
        :class="{ 'is-selected': selectedGenres.includes(tag) }"
        @click="toggleGenre(tag)"
      >
        {{ tag }}
      </button>
    </div>
  </div>
</template>

<script>
import ky from 'ky'
import StarRating from 'vue-star-rating'

const SCENE_GENRE_OPTIONS = ['ハーレム', 'ピンサロ', 'ソープ・エステ', 'マッサージ', '温泉', '人妻', 'イチャラブ', 'コスプレ', 'ナース', '肉感', 'ハード系', 'AR']

export default {
  name: 'SceneRatings',
  components: { StarRating },
  props: {
    item: {
      type: Object,
      required: true
    },
    dark: {
      type: Boolean,
      default: false
    },
    starSize: {
      type: Number,
      default: 15
    },
    showOverall: {
      type: Boolean,
      default: false
    },
    grid2x2: {
      type: Boolean,
      default: false
    },
    showGenres: {
      type: Boolean,
      default: false
    }
  },
  data () {
    return {
      genreOptions: SCENE_GENRE_OPTIONS
    }
  },
  computed: {
    actorAverageRating () {
      const cast = this.item && Array.isArray(this.item.cast) ? this.item.cast : []
      const storeActors = this.$store.state.actorList.actors || []
      const activeColActor = this.$store.state.overlay.actorcollection.actor
      const activeDetActor = this.$store.state.overlay.actordetails.actor

      const vals = cast.map(a => {
        if (activeColActor && activeColActor.id === a.id) {
          return Number(activeColActor.star_rating) || 0
        }
        if (activeDetActor && activeDetActor.id === a.id) {
          return Number(activeDetActor.star_rating) || 0
        }
        const found = storeActors.find(sa => sa.id === a.id)
        return Number(found ? found.star_rating : a.star_rating) || 0
      }).filter(v => v > 0)

      if (vals.length === 0) return 0
      return Math.round((vals.reduce((a, b) => a + b, 0) / vals.length) * 10) / 10
    },
    effectiveOverallRating () {
      const vals = [
        this.actorAverageRating,
        Number(this.item.rating_video_quality) || 0,
        Number(this.item.rating_play) || 0,
        Number(this.item.rating_appeal) || 0
      ].filter(v => v > 0)
      if (vals.length === 0) return 0
      return Math.round((vals.reduce((a, b) => a + b, 0) / vals.length) * 10) / 10
    },
    selectedGenres () {
      if (!this.item || !this.item.genre_tags) return []
      try {
        const parsed = JSON.parse(this.item.genre_tags)
        return Array.isArray(parsed) ? parsed : []
      } catch (e) {
        return []
      }
    }
  },
  methods: {
    async setRating (field, val) {
      const intVal = Math.round(Number(val) || 0)
      const payload = {}
      payload[field] = intVal
      this.$set(this.item, field, intVal)

      const overall = this.effectiveOverallRating
      this.$set(this.item, 'star_rating', overall)

      const updatedScene = Object.assign({}, this.item)
      updatedScene[field] = intVal
      updatedScene.star_rating = overall
      this.$store.commit('sceneList/updateScene', updatedScene)

      try {
        const saved = await ky.post(`/api_custom/scene/ratings/${this.item.id}`, { json: payload }).json()
        if (saved && typeof saved.star_rating === 'number') {
          this.$set(this.item, 'star_rating', saved.star_rating)
        }
      } catch (e) {
        console.error('Failed to save scene rating:', e)
      }
    },
    async toggleGenre (tag) {
      const current = [...this.selectedGenres]
      const idx = current.indexOf(tag)
      if (idx >= 0) {
        current.splice(idx, 1)
      } else {
        current.push(tag)
      }

      const jsonStr = JSON.stringify(current)
      this.$set(this.item, 'genre_tags', jsonStr)

      const updatedScene = Object.assign({}, this.item)
      updatedScene.genre_tags = jsonStr
      this.$store.commit('sceneList/updateScene', updatedScene)

      try {
        await ky.post(`/api_custom/scene/genre_tags/${this.item.id}`, {
          json: { genre_tags: current }
        }).json()
      } catch (e) {
        console.error('Failed to save scene genre tags:', e)
      }
    }
  }
}
</script>

<style scoped>
.scene-ratings-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
}

.scene-ratings-box {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.85rem;
}

.is-grid-2x2 .scene-ratings-box {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.3rem 0.75rem;
  padding: 0.45rem 0.7rem;
  border-radius: 6px;
  background: #2e323d;
  border: 1px solid #3d4250;
}

.scene-rating-item {
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.scene-rating-label {
  font-size: 0.78rem;
  font-weight: 600;
  min-width: 3.4em;
}

.is-dark .scene-rating-label {
  color: #d1d5db;
}

.is-dark ::v-deep .vue-star-rating-rating-text {
  color: #d1d5db;
  font-size: 0.76rem;
}

.scene-overall-pill {
  align-self: flex-start;
  padding: 2px 8px;
  border-radius: 6px;
  background: rgba(244, 63, 94, 0.14);
  border: 1px solid rgba(244, 63, 94, 0.35);
}

.overall-label {
  font-size: 0.76rem;
  font-weight: 700;
  color: #e11d48;
}

.is-dark .overall-label {
  color: #fda4af;
}

.reset-icon {
  cursor: pointer;
  opacity: 0.65;
}

.reset-icon:hover {
  opacity: 1;
}

.is-dark .reset-icon {
  color: #9ca3af;
}

.scene-genre-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.scene-genre-pill {
  padding: 2px 7px;
  border-radius: 999px;
  font-size: 0.68rem;
  font-weight: 600;
  line-height: 1.3;
  white-space: nowrap;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: #f5f5f5;
  color: #363636;
  transition: all 0.15s ease;
}

.scene-genre-pill.is-selected {
  background: #6366f1;
  border-color: #818cf8;
  color: #ffffff;
}

.is-dark .scene-genre-pill {
  background: #2c303a;
  border-color: #454a57;
  color: #d1d5db;
}

.is-dark .scene-genre-pill:hover {
  background: #383d4a;
  color: #ffffff;
}

.is-dark .scene-genre-pill.is-selected {
  background: #6366f1;
  border-color: #818cf8;
  color: #ffffff;
}
</style>
