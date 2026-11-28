<template>
  <div class="container is-fluid collection-page">
    <div class="collection-layout">

      <div class="collection-sidebar">
        <CollectionFilters/>
      </div>

      <div class="collection-main">
        <List/>
      </div>

    </div>

    <!-- 画面左下常時ホバー表示 (選択中フィルタ + Face/Body + Card Size + Topへ戻る) -->
    <div class="collection-floating-stack">
      <div v-if="activeChips.length > 0" class="floating-active-chips">
        <button
          v-for="chip in activeChips"
          :key="chip.key"
          type="button"
          class="floating-active-chip"
          @click="chip.clear()"
          title="クリックで解除"
        >
          <span>{{ chip.label }}</span>
          <span class="floating-chip-x">×</span>
        </button>
        <button
          type="button"
          class="floating-clear-chip"
          @click="clearAllFilters"
          title="すべての条件を解除"
        >
          すべてクリア
        </button>
      </div>

      <div class="collection-floating-dock">
      <div class="dock-group">
        <button
          type="button"
          class="dock-btn"
          :class="{ 'is-active': showFace === true }"
          @click="showFace = true"
        >
          Face
        </button>
        <button
          type="button"
          class="dock-btn"
          :class="{ 'is-active': showFace === false }"
          @click="showFace = false"
        >
          Body
        </button>
      </div>

      <div class="dock-divider"></div>

      <div class="dock-group">
        <button
          v-for="sz in cardSizeOptions"
          :key="sz.value"
          type="button"
          class="dock-btn"
          :class="{ 'is-active': cardSize === sz.value }"
          @click="cardSize = sz.value"
        >
          {{ sz.label }}
        </button>
      </div>

      <template v-if="showToTop">
        <div class="dock-divider"></div>
        <button
          type="button"
          class="dock-top-btn"
          @click="scrollToTop"
          title="Topへ戻る"
        >
          <b-icon pack="mdi" icon="arrow-up" size="is-small" />
        </button>
      </template>
      </div>
    </div>
  </div>
</template>

<script>
import CollectionFilters from './CollectionFilters'
import List from './List'

export default {
  name: 'ActorsCollection',
  components: { CollectionFilters, List },
  data () {
    return {
      showToTop: false,
      cardSizeOptions: [
        { value: '1', label: 'XS' },
        { value: '2', label: 'S' },
        { value: '3', label: 'M' },
        { value: '4', label: 'L' }
      ]
    }
  },
  computed: {
    showFace: {
      get () {
        return this.$store.state.actorList.showFace
      },
      set (value) {
        this.$store.commit('actorList/setShowFace', value)
      }
    },
    cardSize: {
      get () {
        return this.$store.state.actorList.filters.cardSize
      },
      set (value) {
        this.$store.state.actorList.filters.cardSize = value
      }
    },
    activeChips () {
      const f = this.$store.state.actorList.filters || {}
      const chips = []
      const tags = Array.isArray(f.tags) ? f.tags : []
      const cupSizes = Array.isArray(f.cup_sizes) ? f.cup_sizes : []
      const lists = Array.isArray(f.lists) ? f.lists : []
      const cast = Array.isArray(f.cast) ? f.cast : []
      const sites = Array.isArray(f.sites) ? f.sites : []
      const attributes = Array.isArray(f.attributes) ? f.attributes : []

      tags.forEach(tag => {
        chips.push({
          key: `genre-${tag}`,
          label: `タグ: ${tag}`,
          clear: () => {
            this.$set(f, 'tags', tags.filter(t => t !== tag))
            this.reloadFilters(true)
          }
        })
      })
      cupSizes.forEach(cup => {
        chips.push({
          key: `cup-${cup}`,
          label: `カップ: ${cup}`,
          clear: () => {
            this.$set(f, 'cup_sizes', cupSizes.filter(c => c !== cup))
            this.reloadFilters()
          }
        })
      })
      lists.forEach(item => {
        chips.push({
          key: `list-${item}`,
          label: item === 'watchlist' ? 'Watchlist' : 'Favourite',
          clear: () => {
            this.$set(f, 'lists', lists.filter(v => v !== item))
            this.reloadFilters()
          }
        })
      })
      const minRating = f.min_rating || 0
      const maxRating = f.max_rating !== undefined ? f.max_rating : 5
      if (minRating > 0 || maxRating < 5) {
        chips.push({
          key: 'rating',
          label: `女優評価 ${minRating}–${maxRating}`,
          clear: () => {
            this.$set(f, 'min_rating', 0)
            this.$set(f, 'max_rating', 5)
            this.reloadFilters()
          }
        })
      }
      const minSceneRating = f.min_scene_rating || 0
      const maxSceneRating = f.max_scene_rating !== undefined ? f.max_scene_rating : 5
      if (minSceneRating > 0 || maxSceneRating < 5) {
        chips.push({
          key: 'scene-rating',
          label: `作品評価 ${minSceneRating}–${maxSceneRating}`,
          clear: () => {
            this.$set(f, 'min_scene_rating', 0)
            this.$set(f, 'max_scene_rating', 5)
            this.reloadFilters()
          }
        })
      }
      cast.forEach(tag => {
        chips.push({
          key: `cast-${tag}`,
          label: `Cast: ${tag}`,
          clear: () => {
            this.$set(f, 'cast', cast.filter(e => e !== tag))
            this.reloadFilters()
          }
        })
      })
      sites.forEach(tag => {
        chips.push({
          key: `site-${tag}`,
          label: `Site: ${tag}`,
          clear: () => {
            this.$set(f, 'sites', sites.filter(e => e !== tag))
            this.reloadFilters()
          }
        })
      })
      attributes.forEach(tag => {
        chips.push({
          key: `attr-${tag}`,
          label: `Attr: ${tag}`,
          clear: () => {
            this.$set(f, 'attributes', attributes.filter(e => e !== tag))
            this.reloadFilters()
          }
        })
      })
      const minCount = f.min_count || 0
      const maxCount = f.max_count !== undefined ? f.max_count : 150
      if (minCount > 0 || maxCount < 150) {
        chips.push({
          key: 'scenes',
          label: `作品数 ${minCount}–${maxCount}`,
          clear: () => {
            this.$set(f, 'min_count', 0)
            this.$set(f, 'max_count', 150)
            this.reloadFilters()
          }
        })
      }
      const minAvail = f.min_avail || 0
      const maxAvail = f.max_avail !== undefined ? f.max_avail : 150
      if (minAvail > 0 || maxAvail < 150) {
        chips.push({
          key: 'avail',
          label: `保有数 ${minAvail}–${maxAvail}`,
          clear: () => {
            this.$set(f, 'min_avail', 0)
            this.$set(f, 'max_avail', 150)
            this.reloadFilters()
          }
        })
      }
      const minAge = f.min_age || 0
      const maxAge = f.max_age !== undefined ? f.max_age : 100
      if (minAge > 18 || maxAge < 100) {
        chips.push({
          key: 'age',
          label: `年齢 ${minAge}–${maxAge}`,
          clear: () => {
            this.$set(f, 'min_age', 0)
            this.$set(f, 'max_age', 100)
            this.reloadFilters()
          }
        })
      }
      const minHeight = f.min_height !== undefined ? f.min_height : 120
      const maxHeight = f.max_height !== undefined ? f.max_height : 220
      if (minHeight > 120 || maxHeight < 220) {
        chips.push({
          key: 'height',
          label: `身長 ${minHeight}–${maxHeight}`,
          clear: () => {
            this.$set(f, 'min_height', 120)
            this.$set(f, 'max_height', 220)
            this.reloadFilters()
          }
        })
      }
      const minWeight = f.min_weight !== undefined ? f.min_weight : 25
      const maxWeight = f.max_weight !== undefined ? f.max_weight : 150
      if (minWeight > 25 || maxWeight < 150) {
        chips.push({
          key: 'weight',
          label: `体重 ${minWeight}–${maxWeight}`,
          clear: () => {
            this.$set(f, 'min_weight', 25)
            this.$set(f, 'max_weight', 150)
            this.reloadFilters()
          }
        })
      }
      return chips
    }
  },
  mounted () {
    window.addEventListener('scroll', this.handleWindowScroll, { passive: true })
    this.handleWindowScroll()
  },
  beforeDestroy () {
    window.removeEventListener('scroll', this.handleWindowScroll)
  },
  methods: {
    handleWindowScroll () {
      const y = window.scrollY || document.documentElement.scrollTop || document.body.scrollTop || 0
      this.showToTop = y > 120
    },
    scrollToTop () {
      window.scrollTo({ top: 0, behavior: 'smooth' })
    },
    reloadFilters (keepView = false) {
      if (!keepView) {
        this.$store.commit('actorList/setCollectionView', 'results')
      }
      const st = Object.assign({}, this.$store.state.actorList.filters)
      delete st.cardSize
      const q = Buffer.from(JSON.stringify(st)).toString('base64')
      this.$router.push({
        name: 'actorsCollection',
        query: { q }
      }).catch(() => {
        this.$store.dispatch('actorList/load', { offset: 0 })
      })
    },
    clearAllFilters () {
      const f = this.$store.state.actorList.filters
      this.$set(f, 'lists', [])
      this.$set(f, 'tags', [])
      this.$set(f, 'cup_sizes', [])
      this.$set(f, 'cast', [])
      this.$set(f, 'sites', [])
      this.$set(f, 'attributes', [])
      this.$set(f, 'jumpTo', '')
      this.$set(f, 'min_age', 0)
      this.$set(f, 'max_age', 100)
      this.$set(f, 'min_height', 120)
      this.$set(f, 'max_height', 220)
      this.$set(f, 'min_weight', 25)
      this.$set(f, 'max_weight', 150)
      this.$set(f, 'min_count', 0)
      this.$set(f, 'max_count', 150)
      this.$set(f, 'min_avail', 0)
      this.$set(f, 'max_avail', 150)
      this.$set(f, 'min_rating', 0)
      this.$set(f, 'max_rating', 5)
      this.$set(f, 'min_scene_rating', 0)
      this.$set(f, 'max_scene_rating', 5)
      this.reloadFilters(true)
    }
  },
  beforeRouteEnter (to, from, next) {
    next(vm => {
      if (to.query !== undefined) {
        vm.$store.commit('actorList/stateFromQuery', to.query)
      }
      vm.$store.dispatch('optionsWeb/load')
      vm.$store.dispatch('actorList/load', { offset: 0 })
      vm.$store.dispatch('optionsAdvanced/load')
    })
  },
  beforeRouteUpdate (to, from, next) {
    if (to.query !== undefined) {
      this.$store.commit('actorList/stateFromQuery', to.query)
    }
    this.$store.dispatch('actorList/load', { offset: 0 })
    next()
  }
}
</script>

<style scoped>
  .collection-page {
    padding: 0.9rem 1.25rem 2rem;
    min-height: calc(100vh - 52px);
    background: radial-gradient(circle at 20% 0%, #16222f 0%, #0f171e 60%);
    color: #f1f5f9;
  }

  .collection-layout {
    display: flex;
    align-items: flex-start;
    gap: 1.15rem;
  }

  .collection-sidebar {
    width: 248px;
    flex-shrink: 0;
  }

  .collection-main {
    flex: 1;
    min-width: 0;
  }

  @media (max-width: 768px) {
    .collection-layout {
      flex-direction: column;
    }

    .collection-sidebar {
      width: 100%;
    }
  }

  .collection-floating-stack {
    position: fixed;
    bottom: 16px;
    left: 16px;
    z-index: 40;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.4rem;
    max-width: calc(100vw - 32px);
    pointer-events: none;
  }

  .floating-active-chips {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.3rem;
    padding: 0.38rem 0.5rem;
    background: rgba(15, 23, 30, 0.92);
    backdrop-filter: blur(8px);
    border: 1px solid #2b3f54;
    border-radius: 10px;
    box-shadow: 0 6px 20px rgba(0, 0, 0, 0.5);
    pointer-events: auto;
    max-width: 520px;
  }

  .floating-active-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    padding: 2px 8px;
    background: rgba(0, 168, 225, 0.18);
    color: #7dd3fc;
    border: 1px solid rgba(0, 168, 225, 0.45);
    border-radius: 999px;
    font-size: 0.7rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.14s ease;
  }

  .floating-active-chip:hover {
    background: rgba(239, 68, 68, 0.22);
    border-color: rgba(239, 68, 68, 0.55);
    color: #fecaca;
  }

  .floating-chip-x {
    font-weight: 700;
    line-height: 1;
  }

  .floating-clear-chip {
    border: 1px solid #334155;
    border-radius: 999px;
    padding: 2px 8px;
    background: #1b2836;
    color: #94a3b8;
    font-size: 0.68rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.14s ease;
  }

  .floating-clear-chip:hover {
    background: #334155;
    color: #ffffff;
  }

  .collection-floating-dock {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.38rem 0.5rem;
    background: rgba(15, 23, 30, 0.9);
    backdrop-filter: blur(8px);
    border: 1px solid #2b3f54;
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.55);
    pointer-events: auto;
  }

  .dock-group {
    display: inline-flex;
    align-items: center;
    background: #16202a;
    border: 1px solid #253342;
    border-radius: 6px;
    padding: 2px;
    gap: 2px;
  }

  .dock-btn {
    border: none;
    border-radius: 4px;
    padding: 3px 8px;
    background: transparent;
    color: #94a3b8;
    font-size: 0.72rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dock-btn:hover {
    color: #ffffff;
  }

  .dock-btn.is-active {
    background: #00a8e1;
    color: #ffffff;
    box-shadow: 0 1px 4px rgba(0, 168, 225, 0.35);
  }

  .dock-divider {
    width: 1px;
    height: 18px;
    background: #2b3b4c;
  }

  .dock-top-btn {
    border: 1px solid #2b3f54;
    border-radius: 6px;
    width: 28px;
    height: 26px;
    background: #1b2836;
    color: #00a8e1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .dock-top-btn:hover {
    background: #00a8e1;
    border-color: #00a8e1;
    color: #ffffff;
  }
</style>
