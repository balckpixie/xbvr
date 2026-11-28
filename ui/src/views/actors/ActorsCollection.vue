<template>
  <div class="container is-fluid collection-page">
    <div class="collection-layout">

      <div class="collection-sidebar">
        <CollectionFilters/>

        <a id="toTop">
          <b-icon pack="mdi" icon="navigation" />
        </a>
      </div>

      <div class="collection-main">
        <List/>
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
  mounted () {
    const toTop = document.getElementById('toTop')
    addEventListener('scroll', function () {
      if (!toTop) return
      toTop.style.display = document.body.scrollTop > 20 || document.documentElement.scrollTop > 20
        ? 'block'
        : 'none'
    })
    toTop.addEventListener('click', function () {
      scrollToTop()
    })

    const scrollToTop = () => {
      const c = document.documentElement.scrollTop || document.body.scrollTop
      if (c > 0) {
        window.requestAnimationFrame(scrollToTop)
        window.scrollTo(0, c - c / 16)
      }
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
    padding-top: 0.65rem;
  }

  .collection-layout {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
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

  #toTop {
    display: none;
    position: fixed;
    bottom: 20px;
    left: 30px;
    background-color: #f0f0f0;
    color: #4a4a4a;
    padding: 15px;
    border-radius: 10px;
    font-size: 18px;
  }

  #toTop:hover {
    background-color: #BDBDBD;
  }
</style>
