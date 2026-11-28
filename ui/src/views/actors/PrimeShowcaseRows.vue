<template>
  <div class="prime-showcase">
    <!-- 1. ピックアップ -->
    <section v-if="pickupActors.length > 0" class="prime-row-section">
      <div class="prime-row-header">
        <div class="prime-row-title-wrap">
          <span class="prime-row-badge is-pickup">PICKUP</span>
          <h3 class="prime-row-title">ピックアップ</h3>
          <span class="prime-row-sub">おすすめサジェスト（最大50件）</span>
        </div>
        <div class="prime-row-nav-btns">
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('pickup', -1)" title="前へ">&#9664;</button>
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('pickup', 1)" title="次へ">&#9654;</button>
        </div>
      </div>
      <div class="prime-row-viewport">
        <button type="button" class="prime-edge-btn is-left" @click="scrollRow('pickup', -1)" aria-label="Previous">&#10094;</button>
        <div ref="pickupTrack" class="prime-row-track">
          <div v-for="actor in pickupActors" :key="`pickup-${actor.id}`" class="prime-row-card-item" :style="rowCardStyle">
            <ActorCard :actor="actor" :showFace="showFace" />
          </div>
        </div>
        <button type="button" class="prime-edge-btn is-right" @click="scrollRow('pickup', 1)" aria-label="Next">&#10095;</button>
      </div>
    </section>

    <!-- 2. ランキング：総合 -->
    <section v-if="overallRankingRow && overallRankingRow.actors.length > 0" class="prime-row-section">
      <div class="prime-row-header">
        <div class="prime-row-title-wrap">
          <span class="prime-row-badge is-ranking">RANKING</span>
          <h3 class="prime-row-title">ランキング：総合</h3>
          <span class="prime-row-sub">総合評価順（最大50件）</span>
        </div>
        <div class="prime-row-nav-btns">
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('rank_overall', -1)" title="前へ">&#9664;</button>
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('rank_overall', 1)" title="次へ">&#9654;</button>
        </div>
      </div>
      <div class="prime-row-viewport">
        <button type="button" class="prime-edge-btn is-left" @click="scrollRow('rank_overall', -1)" aria-label="Previous">&#10094;</button>
        <div ref="track-rank_overall" class="prime-row-track">
          <div v-for="(actor, idx) in overallRankingRow.actors" :key="`rank_overall-${actor.id}`" class="prime-row-card-item" :style="rowCardStyle">
            <div class="prime-rank-num" :class="{ 'is-top3': idx < 3 }">{{ idx + 1 }}</div>
            <ActorCard :actor="actor" :showFace="showFace" />
          </div>
        </div>
        <button type="button" class="prime-edge-btn is-right" @click="scrollRow('rank_overall', 1)" aria-label="Next">&#10095;</button>
      </div>
    </section>

    <!-- 3. ランキング：詳細 (顔 / 胸 / スタイル / 雰囲気 切替) -->
    <section v-if="activeDetailRankingRow && activeDetailRankingRow.actors.length > 0" class="prime-row-section">
      <div class="prime-row-header">
        <div class="prime-row-title-wrap">
          <span class="prime-row-badge is-ranking">RANKING</span>
          <h3 class="prime-row-title">ランキング：詳細</h3>
          <div class="prime-detail-tabs">
            <button
              v-for="tab in detailRankTabs"
              :key="tab.key"
              type="button"
              class="prime-detail-tab-btn"
              :class="{ 'is-active': activeDetailRankKey === tab.key }"
              @click="selectDetailRank(tab.key)"
            >
              {{ tab.label }}
            </button>
          </div>
          <span class="prime-row-sub">{{ activeDetailRankingRow.sub }}（最大50件）</span>
        </div>
        <div class="prime-row-nav-btns">
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('rank_detail', -1)" title="前へ">&#9664;</button>
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('rank_detail', 1)" title="次へ">&#9654;</button>
        </div>
      </div>
      <div class="prime-row-viewport">
        <button type="button" class="prime-edge-btn is-left" @click="scrollRow('rank_detail', -1)" aria-label="Previous">&#10094;</button>
        <div ref="track-rank_detail" class="prime-row-track">
          <div v-for="(actor, idx) in activeDetailRankingRow.actors" :key="`${activeDetailRankKey}-${actor.id}`" class="prime-row-card-item" :style="rowCardStyle">
            <div class="prime-rank-num" :class="{ 'is-top3': idx < 3 }">{{ idx + 1 }}</div>
            <ActorCard :actor="actor" :showFace="showFace" />
          </div>
        </div>
        <button type="button" class="prime-edge-btn is-right" @click="scrollRow('rank_detail', 1)" aria-label="Next">&#10095;</button>
      </div>
    </section>

    <!-- 4. ランキング：シーン総合 -->
    <section v-if="sceneRankingRow && sceneRankingRow.actors.length > 0" class="prime-row-section">
      <div class="prime-row-header">
        <div class="prime-row-title-wrap">
          <span class="prime-row-badge is-ranking">RANKING</span>
          <h3 class="prime-row-title">ランキング：シーン総合</h3>
          <span class="prime-row-sub">シーン総合点の平均値順（最大50件）</span>
        </div>
        <div class="prime-row-nav-btns">
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('rank_scene', -1)" title="前へ">&#9664;</button>
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('rank_scene', 1)" title="次へ">&#9654;</button>
        </div>
      </div>
      <div class="prime-row-viewport">
        <button type="button" class="prime-edge-btn is-left" @click="scrollRow('rank_scene', -1)" aria-label="Previous">&#10094;</button>
        <div ref="track-rank_scene" class="prime-row-track">
          <div v-for="(actor, idx) in sceneRankingRow.actors" :key="`rank_scene-${actor.id}`" class="prime-row-card-item" :style="rowCardStyle">
            <div class="prime-rank-num" :class="{ 'is-top3': idx < 3 }">{{ idx + 1 }}</div>
            <ActorCard :actor="actor" :showFace="showFace" />
          </div>
        </div>
        <button type="button" class="prime-edge-btn is-right" @click="scrollRow('rank_scene', 1)" aria-label="Next">&#10095;</button>
      </div>
    </section>

    <!-- 3. 最近チェック -->
    <section v-if="visibleRecentActors.length > 0" class="prime-row-section">
      <div class="prime-row-header">
        <div class="prime-row-title-wrap">
          <span class="prime-row-badge is-recent">HISTORY</span>
          <h3 class="prime-row-title">最近チェック</h3>
          <span class="prime-row-sub">チェックしたActor（最大50件）</span>
        </div>
        <div class="prime-row-nav-btns">
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('recent', -1)" title="前へ">&#9664;</button>
          <button type="button" class="prime-row-arrow-sm" @click="scrollRow('recent', 1)" title="次へ">&#9654;</button>
        </div>
      </div>
      <div class="prime-row-viewport">
        <button type="button" class="prime-edge-btn is-left" @click="scrollRow('recent', -1)" aria-label="Previous">&#10094;</button>
        <div ref="recentTrack" class="prime-row-track">
          <div v-for="actor in visibleRecentActors" :key="`recent-${actor.id}`" class="prime-row-card-item" :style="rowCardStyle">
            <ActorCard :actor="actor" :showFace="showFace" />
          </div>
        </div>
        <button type="button" class="prime-edge-btn is-right" @click="scrollRow('recent', 1)" aria-label="Next">&#10095;</button>
      </div>
    </section>
  </div>
</template>

<script>
import ky from 'ky'
import ActorCard from './ActorCard'

const RECENT_STORAGE_KEY = 'xbvr_recent_collection_actors'

const RANKING_DEFS = [
  { key: 'rank_overall', sort: 'rating_desc', field: 'star_rating', title: 'ランキング：総合', sub: '総合評価順' },
  { key: 'rank_face', sort: 'rating_face_desc', field: 'rating_face', title: 'ランキング：顔', sub: '顔評価順' },
  { key: 'rank_bust', sort: 'rating_bust_desc', field: 'rating_bust', title: 'ランキング：胸', sub: '胸評価順' },
  { key: 'rank_style', sort: 'rating_style_desc', field: 'rating_style', title: 'ランキング：スタイル', sub: 'スタイル評価順' },
  { key: 'rank_vibe', sort: 'rating_vibe_desc', field: 'rating_vibe', title: 'ランキング：雰囲気', sub: '雰囲気評価順' },
  { key: 'rank_scene', sort: 'scene_rating_desc', field: 'scene_rating_average', title: 'ランキング：シーン総合', sub: 'シーン総合点の平均値順' }
]

export default {
  name: 'PrimeShowcaseRows',
  components: { ActorCard },
  props: {
    showFace: {
      type: Boolean,
      default: true
    }
  },
  data () {
    return {
      pickupActors: [],
      rankingRows: RANKING_DEFS.map(def => ({ ...def, actors: [] })),
      recentActors: [],
      activeDetailRankKey: 'rank_face',
      detailRankTabs: [
        { key: 'rank_face', label: '顔' },
        { key: 'rank_bust', label: '胸' },
        { key: 'rank_style', label: 'スタイル' },
        { key: 'rank_vibe', label: '雰囲気' }
      ]
    }
  },
  computed: {
    cardSize () {
      return this.$store.state.actorList.filters.cardSize || '3'
    },
    rowCardStyle () {
      const widthMap = { '1': '135px', '2': '170px', '3': '210px', '4': '260px' }
      const w = widthMap[this.cardSize] || '210px'
      return { flex: `0 0 ${w}`, width: w }
    },
    selectedTags () {
      return this.$store.state.actorList.filters.tags || []
    },
    overallRankingRow () {
      return this.rankingRows.find(r => r.key === 'rank_overall') || null
    },
    activeDetailRankingRow () {
      return this.rankingRows.find(r => r.key === this.activeDetailRankKey) || null
    },
    sceneRankingRow () {
      return this.rankingRows.find(r => r.key === 'rank_scene') || null
    },
    visibleRecentActors () {
      if (!this.selectedTags.length) return this.recentActors
      return this.recentActors.filter(actor => {
        if (!actor || !actor.genre_tags) return false
        try {
          const parsed = typeof actor.genre_tags === 'string' ? JSON.parse(actor.genre_tags) : actor.genre_tags
          if (!Array.isArray(parsed)) return false
          return this.selectedTags.every(t => parsed.includes(t))
        } catch (e) {
          return false
        }
      })
    }
  },
  watch: {
    selectedTags: {
      deep: true,
      handler () {
        this.loadShowcaseRows()
      }
    }
  },
  mounted () {
    this.loadShowcaseRows()
    this.loadRecentActors()
    window.addEventListener('xbvr-recent-actors-updated', this.loadRecentActors)
  },
  beforeDestroy () {
    window.removeEventListener('xbvr-recent-actors-updated', this.loadRecentActors)
  },
  methods: {
    selectDetailRank (key) {
      this.activeDetailRankKey = key
      this.$nextTick(() => {
        const el = this.$refs['track-rank_detail']
        const track = Array.isArray(el) ? el[0] : el
        if (track) track.scrollLeft = 0
      })
    },
    baseQuery (overrides) {
      return Object.assign({
        dlState: 'available',
        lists: [],
        cast: [],
        sites: [],
        tags: this.selectedTags,
        attributes: [],
        jumpTo: '',
        min_age: 0,
        max_age: 100,
        min_height: 120,
        max_height: 220,
        min_weight: 25,
        max_weight: 150,
        min_count: 0,
        max_count: 150,
        min_avail: 0,
        max_avail: 150,
        min_rating: 0,
        max_rating: 5,
        min_scene_rating: 0,
        max_scene_rating: 5,
        offset: 0,
        limit: 50
      }, overrides)
    },
    async loadShowcaseRows () {
      try {
        const pickupPromise = ky.post('/api/actor/list', {
          json: this.baseQuery({ sort: 'random', min_avail: 1, limit: 50 })
        }).json()
        const rankingPromises = RANKING_DEFS.map(def =>
          ky.post('/api/actor/list', {
            json: this.baseQuery({ sort: def.sort, limit: 50 })
          }).json()
        )
        const [pickupRes, ...rankingResList] = await Promise.all([pickupPromise, ...rankingPromises])
        this.pickupActors = (pickupRes && Array.isArray(pickupRes.actors)) ? pickupRes.actors.slice(0, 50) : []
        this.rankingRows = RANKING_DEFS.map((def, idx) => {
          const res = rankingResList[idx]
          const rawActors = (res && Array.isArray(res.actors)) ? res.actors : []
          return {
            ...def,
            actors: rawActors.slice(0, 50)
          }
        })
      } catch (e) {
        // ignore network errors
      }
    },
    loadRecentActors () {
      try {
        const raw = localStorage.getItem(RECENT_STORAGE_KEY)
        if (!raw) {
          this.recentActors = []
          return
        }
        const parsed = JSON.parse(raw)
        this.recentActors = Array.isArray(parsed) ? parsed.slice(0, 50) : []
      } catch (e) {
        this.recentActors = []
      }
    },
    scrollRow (key, dir) {
      let el = null
      if (key === 'pickup') {
        el = this.$refs.pickupTrack
      } else if (key === 'recent') {
        el = this.$refs.recentTrack
      } else {
        const refVal = this.$refs[`track-${key}`]
        el = Array.isArray(refVal) ? refVal[0] : refVal
      }
      if (!el) return
      const step = Math.max(300, Math.floor(el.clientWidth * 0.85))
      el.scrollBy({
        left: dir * step,
        behavior: 'smooth'
      })
    }
  }
}
</script>

<style scoped>
.prime-showcase {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  margin-bottom: 1.35rem;
}

.prime-row-section {
  position: relative;
}

.prime-row-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.45rem;
  padding: 0 0.15rem;
}

.prime-row-title-wrap {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.prime-row-badge {
  font-size: 0.64rem;
  font-weight: 800;
  letter-spacing: 0.04em;
  padding: 2px 7px;
  border-radius: 4px;
  color: #ffffff;
}

.prime-row-badge.is-pickup {
  background: #00a8e1;
}

.prime-row-badge.is-ranking {
  background: #d97706;
}

.prime-row-badge.is-recent {
  background: #475569;
}

.prime-row-title {
  font-size: 1rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
}

.prime-row-sub {
  font-size: 0.72rem;
  color: #8197a4;
}

.prime-detail-tabs {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 2px;
  background: #16202a;
  border: 1px solid #2b3b4c;
  border-radius: 6px;
}

.prime-detail-tab-btn {
  border: none;
  border-radius: 4px;
  padding: 2px 10px;
  background: transparent;
  color: #94a3b8;
  font-size: 0.74rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.14s ease;
}

.prime-detail-tab-btn:hover {
  color: #ffffff;
}

.prime-detail-tab-btn.is-active {
  background: #00a8e1;
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 168, 225, 0.35);
}

.prime-row-nav-btns {
  display: flex;
  gap: 0.3rem;
}

.prime-row-arrow-sm {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  border: 1px solid #2b3b4c;
  background: #16202a;
  color: #cbd5e1;
  font-size: 0.7rem;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s ease;
}

.prime-row-arrow-sm:hover {
  background: #00a8e1;
  border-color: #00a8e1;
  color: #ffffff;
}

.prime-row-viewport {
  position: relative;
}

.prime-row-track {
  display: flex;
  gap: 0.75rem;
  overflow-x: auto;
  scroll-behavior: smooth;
  padding: 0.35rem 0.15rem 0.55rem;
  scrollbar-width: none;
}

.prime-row-track::-webkit-scrollbar {
  display: none;
}

.prime-row-card-item {
  position: relative;
  flex: 0 0 210px;
  width: 210px;
}

.prime-rank-num {
  position: absolute;
  top: -4px;
  left: -4px;
  z-index: 12;
  min-width: 24px;
  height: 24px;
  padding: 0 6px;
  border-radius: 6px;
  background: rgba(15, 23, 30, 0.9);
  border: 1px solid #475569;
  color: #e2e8f0;
  font-size: 0.74rem;
  font-weight: 800;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  pointer-events: none;
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.5);
}

.prime-rank-num.is-top3 {
  background: linear-gradient(135deg, #f59e0b 0%, #d97706 100%);
  border-color: #fbbf24;
  color: #111827;
}

.prime-edge-btn {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  z-index: 15;
  width: 34px;
  height: 68px;
  border: none;
  border-radius: 6px;
  background: rgba(15, 23, 30, 0.78);
  color: #ffffff;
  font-size: 1.15rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.18s ease, background 0.15s ease;
}

.prime-row-viewport:hover .prime-edge-btn {
  opacity: 1;
}

.prime-edge-btn:hover {
  background: rgba(0, 168, 225, 0.9);
}

.prime-edge-btn.is-left {
  left: 4px;
}

.prime-edge-btn.is-right {
  right: 4px;
}
</style>
