<template>
  <div class="scene-ratings" :class="{ 'is-dark': dark }">
    <!-- 総合 (Details.vue など showOverall=true の場合のみ表示) -->
    <div v-if="showOverall" class="scene-rating-item scene-overall-pill">
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

    <!-- 女優 (主演女優の総合評価の平均、ユーザー編集不可) -->
    <div class="scene-rating-item">
      <span class="scene-rating-label">女優</span>
      <star-rating
        :key="`${item.id}-actor-avg`"
        :rating="actorAverageRating"
        read-only
        :increment="0.1"
        :star-size="starSize"
        :show-rating="true"
        active-color="#60a5fa"
      />
    </div>

    <!-- 画質・プレイ・抜け感 -->
    <div v-for="r in editableItems" :key="r.key" class="scene-rating-item">
      <span class="scene-rating-label">{{ r.label }}</span>
      <star-rating
        :key="`${item.id}-${r.key}`"
        :rating="item[r.key] || 0"
        @rating-selected="val => setRating(r.key, val)"
        :increment="0.5"
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
          @click.native="setRating(r.key, 0)"
        />
      </b-tooltip>
    </div>
  </div>
</template>

<script>
import ky from 'ky'
import StarRating from 'vue-star-rating'

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
    }
  },
  data () {
    return {
      editableItems: [
        { key: 'rating_video_quality', label: '画質' },
        { key: 'rating_play', label: 'プレイ' },
        { key: 'rating_appeal', label: '抜け感' }
      ]
    }
  },
  computed: {
    actorAverageRating () {
      const cast = this.item && Array.isArray(this.item.cast) ? this.item.cast : []
      const vals = cast.map(a => Number(a.star_rating) || 0).filter(v => v > 0)
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
    }
  },
  methods: {
    async setRating (field, val) {
      const payload = {}
      payload[field] = val
      this.$set(this.item, field, val)

      const overall = this.effectiveOverallRating
      this.$set(this.item, 'star_rating', overall)

      const updatedScene = Object.assign({}, this.item)
      updatedScene[field] = val
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
    }
  }
}
</script>

<style scoped>
.scene-ratings {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.35rem 0.85rem;
}

.scene-rating-item {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.scene-rating-label {
  font-size: 0.78rem;
  font-weight: 600;
}

.is-dark .scene-rating-label {
  color: #d1d5db;
}

.is-dark ::v-deep .vue-star-rating-rating-text {
  color: #d1d5db;
  font-size: 0.75rem;
}

.scene-overall-pill {
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
</style>
