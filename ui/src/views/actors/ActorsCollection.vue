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

    <!-- 画面左下常時ホバー表示 (Face/Body + Card Size + Topへ戻る) -->
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

  .collection-floating-dock {
    position: fixed;
    bottom: 16px;
    left: 16px;
    z-index: 40;
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    padding: 0.38rem 0.5rem;
    background: rgba(15, 23, 30, 0.9);
    backdrop-filter: blur(8px);
    border: 1px solid #2b3f54;
    border-radius: 10px;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.55);
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
