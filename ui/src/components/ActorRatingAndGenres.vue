<template>
  <div class="actor-rating-genres" :class="{ 'is-compact': compact, 'is-dark': dark, 'is-grid-2x2': grid2x2 }">
    <div class="ratings-grid">
      <div v-for="item in visibleRatingItems" :key="item.key" class="rating-row">
        <span class="rating-label">{{ item.label }}</span>
        <star-rating
          :key="`${actor.id}-${item.key}`"
          :rating="actor[item.key] || 0"
          :read-only="item.key === 'star_rating'"
          @rating-selected="val => setRating(item.key, val)"
          :increment="item.key === 'star_rating' ? 0.1 : 0.5"
          :star-size="compact ? 15 : 18"
          :show-rating="true"
          :active-color="item.key === 'star_rating' ? '#ffd055' : '#48c774'"
        />
        <b-tooltip v-if="item.key !== 'star_rating'" :label="$t('Reset Rating')" position="is-right" :delay="250">
          <b-icon
            pack="mdi"
            icon="autorenew"
            size="is-small"
            class="reset-icon"
            @click.native="setRating(item.key, 0)"
          />
        </b-tooltip>
      </div>
    </div>

    <div class="genre-tags-section">
      <div class="genre-buttons">
        <button
          v-for="tag in genreOptions"
          :key="tag"
          type="button"
          class="genre-pill-btn"
          :class="{ 'is-selected': selectedGenres.includes(tag) }"
          @click="toggleGenre(tag)"
        >
          {{ tag }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
import ky from 'ky'
import StarRating from 'vue-star-rating'

const GENRE_OPTIONS = ['美少女', 'お姉さん', '癒し', 'グラマー', '人妻', 'ハーフ', 'エロ']

export default {
  name: 'ActorRatingAndGenres',
  components: { StarRating },
  props: {
    actor: {
      type: Object,
      required: true
    },
    compact: {
      type: Boolean,
      default: false
    },
    dark: {
      type: Boolean,
      default: false
    },
    excludeOverall: {
      type: Boolean,
      default: false
    },
    grid2x2: {
      type: Boolean,
      default: false
    }
  },
  data () {
    return {
      genreOptions: GENRE_OPTIONS,
      ratingItems: [
        { key: 'rating_face', label: '顔' },
        { key: 'rating_style', label: 'スタイル' },
        { key: 'rating_bust', label: '胸' },
        { key: 'rating_vibe', label: '雰囲気' },
        { key: 'star_rating', label: '総合' }
      ]
    }
  },
  computed: {
    visibleRatingItems () {
      if (this.excludeOverall) {
        return this.ratingItems.filter(item => item.key !== 'star_rating')
      }
      return this.ratingItems
    },
    selectedGenres () {
      if (!this.actor || !this.actor.genre_tags) return []
      try {
        const parsed = JSON.parse(this.actor.genre_tags)
        return Array.isArray(parsed) ? parsed : []
      } catch (e) {
        return []
      }
    }
  },
  methods: {
    async setRating (field, val) {
      if (field === 'star_rating') return

      const payload = {}
      payload[field] = val
      this.$set(this.actor, field, val)

      // 4項目の平均から総合(star_rating)を自動算出
      const subKeys = ['rating_face', 'rating_style', 'rating_bust', 'rating_vibe']
      const vals = subKeys.map(k => Number(this.actor[k]) || 0).filter(v => v > 0)
      const overall = vals.length > 0
        ? Math.round((vals.reduce((a, b) => a + b, 0) / vals.length) * 10) / 10
        : 0
      this.$set(this.actor, 'star_rating', overall)

      const updatedActor = Object.assign({}, this.actor)
      updatedActor[field] = val
      updatedActor.star_rating = overall
      this.$store.commit('actorList/updateActor', updatedActor)

      try {
        await ky.post(`/api_custom/actor/ratings/${this.actor.id}`, { json: payload }).json()
      } catch (e) {
        console.error('Failed to save actor rating:', e)
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
      this.$set(this.actor, 'genre_tags', jsonStr)

      const updatedActor = Object.assign({}, this.actor)
      updatedActor.genre_tags = jsonStr
      this.$store.commit('actorList/updateActor', updatedActor)

      try {
        await ky.post(`/api_custom/actor/genre_tags/${this.actor.id}`, {
          json: { genre_tags: current }
        }).json()
      } catch (e) {
        console.error('Failed to save actor genre tags:', e)
      }
    }
  }
}
</script>

<style scoped>
.actor-rating-genres {
  margin-top: 0.35rem;
}

.ratings-grid {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.is-compact .ratings-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
  gap: 0.25rem 0.75rem;
}

.is-grid-2x2 .ratings-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.35rem 0.75rem;
  padding: 0.45rem 0.65rem;
  border-radius: 6px;
  background: #272a34;
  border: 1px solid #353945;
}

.rating-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.rating-label {
  font-weight: 600;
  font-size: 0.8rem;
  min-width: 4.5em;
}

.is-dark .rating-label,
.is-dark .genre-label {
  color: #d1d5db;
}

.is-dark ::v-deep .vue-star-rating-rating-text {
  color: #d1d5db;
  font-size: 0.78rem;
}

.is-dark .reset-icon {
  color: #9ca3af;
}

.reset-icon {
  cursor: pointer;
  opacity: 0.65;
}

.reset-icon:hover {
  opacity: 1;
}

.genre-tags-section {
  margin-top: 0.5rem;
}

.genre-label {
  font-weight: 600;
  font-size: 0.8rem;
  margin-bottom: 0.25rem;
}

.genre-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 0.35rem;
}

.genre-pill-btn {
  padding: 3px 11px;
  border-radius: 999px;
  font-size: 0.76rem;
  font-weight: 600;
  cursor: pointer;
  border: 1px solid #d1d5db;
  background: #f5f5f5;
  color: #363636;
  transition: all 0.15s ease;
}

.genre-pill-btn.is-selected {
  background: #7957d5;
  border-color: #7957d5;
  color: #ffffff;
}

.is-dark .genre-pill-btn {
  background: #2c303a;
  border-color: #454a57;
  color: #d1d5db;
}

.is-dark .genre-pill-btn:hover {
  background: #383d4a;
  color: #ffffff;
}

.is-dark .genre-pill-btn.is-selected {
  background: #7957d5;
  border-color: #8b6ce0;
  color: #ffffff;
}
</style>
