<template>
  <div class="container">
    <b-loading :is-full-page="false" :active.sync="isLoading" />
    <div class="content">
      <h3>{{ $t('Web UI') }}</h3>
      <hr />
      <div class="columns">
        <div class="column">
          <section>
            <b-field label="Tag Sort">
              <div class="block">
                <b-radio v-model="tagSort" name="tagSort" native-value="by-tag-count">
                  By Tag Count
                </b-radio>
                <b-radio v-model="tagSort" name="tagSort" native-value="alphabetically">
                  Alphabetically
                </b-radio>
              </div>
            </b-field>

            <b-field label="Buttons in Scene List">
            </b-field>
            <b-field>
              <b-switch v-model="sceneHidden" type="is-danger">
                show Toggle Hidden Status button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneWatchlist" type="is-default">
                show Add/Remove from Watchlist button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneFavourite" type="is-danger">
                show Add/Remove from Favourites button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneWishlist" type="is-info">
                show Add/Remove from Wishlist button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneTrailerlist" type="is-default">
                show Add/Remove from Trailer list button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneWatched" type="is-dark">
                show Toggle Watched Status button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneEdit" type="is-dark">
                show Edit Scene button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneDuration" type="is-dark">
                show Duration button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="sceneCuepoint" type="is-dark">
                show Cuepoints button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="hspFile" type="is-dark">
                show Hsp File button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="subtitlesFile" type="is-dark">
                show subtitles File button
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="ScriptHeatmap" type="is-dark">
                show Script Heatmap
              </b-switch>
            </b-field>
            <b-field v-if="ScriptHeatmap">
              <b-switch v-model="AllHeatmaps" type="is-dark">
                show All Heatmaps
              </b-switch>
            </b-field>
            <b-field>
              <b-switch v-model="openInNewWindow" type="is-dark">
                show Open Tag in New Window
              </b-switch>
            </b-field>
            <b-field label="Opacity of unavailable scenes">
              <div class="columns">
                <div class="column is-two-thirds">
                  <b-slider :min="0" :max="100" :step="10" :tooltip="false" v-model="isAvailOpacity" opacity:isAvailOpacity></b-slider>
                </div>
              </div>
            </b-field>

            <b-field label="Automatically Check for Updates">
              <b-switch v-model="updateCheck">
                Enabled
              </b-switch>
            </b-field>

            <b-field>
              <b-button type="is-primary" @click="save">Save</b-button>
            </b-field>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'InterfaceWeb',
  mounted () {
    this.$store.dispatch('optionsWeb/load')
  },
  methods: {
    save () {
      this.$store.dispatch('optionsWeb/save')
    }
  },
  computed: {
    tagSort: {
      get () {
        return this.$store.state.optionsWeb.web.tagSort
      },
      set (value) {
        this.$store.state.optionsWeb.web.tagSort = value
      }
    },
    sceneHidden: {
      get () {
        return this.$store.state.optionsWeb.web.sceneHidden
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneHidden = value
      }
    },
    sceneWatchlist: {
      get () {
        return this.$store.state.optionsWeb.web.sceneWatchlist
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneWatchlist = value
      }
    },
    sceneFavourite: {
      get () {
        return this.$store.state.optionsWeb.web.sceneFavourite
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneFavourite = value
      }
    },
    sceneWishlist: {
      get () {
        return this.$store.state.optionsWeb.web.sceneWishlist
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneWishlist = value
      }
    },
    sceneTrailerlist: {
      get () {
        return this.$store.state.optionsWeb.web.sceneTrailerlist
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneTrailerlist = value
      }
    },
    sceneWatched: {
      get () {
        return this.$store.state.optionsWeb.web.sceneWatched
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneWatched = value
      }
    },
    sceneEdit: {
      get () {
        return this.$store.state.optionsWeb.web.sceneEdit
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneEdit = value
      }
    },
    ScriptHeatmap: {
      get () {
        return this.$store.state.optionsWeb.web.showScriptHeatmap
      },
      set (value) {
        this.$store.state.optionsWeb.web.showScriptHeatmap = value
      }
    },
    AllHeatmaps: {
      get () {
        return this.$store.state.optionsWeb.web.showAllHeatmaps
      },
      set (value) {
        this.$store.state.optionsWeb.web.showAllHeatmaps = value
      }
    },
    updateCheck: {
      get () {
        return this.$store.state.optionsWeb.web.updateCheck
      },
      set (value) {
        this.$store.state.optionsWeb.web.updateCheck = value
      }
    },
    sceneDuration: {
      get () {
        return this.$store.state.optionsWeb.web.sceneDuration
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneDuration = value
      }
    },
    sceneCuepoint: {
      get () {
        return this.$store.state.optionsWeb.web.sceneCuepoint
      },
      set (value) {
        this.$store.state.optionsWeb.web.sceneCuepoint = value
      }
    },
    hspFile: {
      get () {
        return this.$store.state.optionsWeb.web.showHspFile
      },
      set (value) {
        this.$store.state.optionsWeb.web.showHspFile = value
      }
    },
    subtitlesFile: {
      get () {
        return this.$store.state.optionsWeb.web.showSubtitlesFile
      },
      set (value) {
        this.$store.state.optionsWeb.web.showSubtitlesFile = value
      }
    },
    openInNewWindow: {
      get () {
        return this.$store.state.optionsWeb.web.showOpenInNewWindow
      },
      set (value) {
        this.$store.state.optionsWeb.web.showOpenInNewWindow = value
      }
    },
    isAvailOpacity: {
      get () {
        if  (this.$store.state.optionsWeb.web.isAvailOpacity == undefined) {
          return 40
        }
        return this.$store.state.optionsWeb.web.isAvailOpacity
      },
      set (value) {
        this.$store.state.optionsWeb.web.isAvailOpacity = value
      }
    },
    isLoading: function () {
      return this.$store.state.optionsWeb.loading
    }    
  }
}
</script>

<style scoped>

</style>
