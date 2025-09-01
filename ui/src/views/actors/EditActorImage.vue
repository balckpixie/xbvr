<template>
  <div class="modal is-active">
    <GlobalEvents :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)" @keyup.esc="close" @keyup.s="save" />

    <div class="modal-background"></div>

    <div class="modal-card">
      <header class="modal-card-head">
        <p class="modal-card-title">{{ $t('Edit actor Image') }} - {{ actor.name }}</p>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>

      <section class="modal-card-body">
        <b-tabs position="is-centered" :animated="false">

          <b-tab-item :label="$t('Actor')">
            <div class="columns is-gapless" style="height: 64vh;overflow: hidden;">

              <!-- カルーセル：スクロール可能 -->
              <div class="column" style="overflow-y: auto; padding: 1rem;">
                <b-carousel v-model="carouselSlide" @change="scrollToActiveIndicator" :autoplay="false"
                  :indicator-inside="false">
                  <b-carousel-item v-for="(carousel, i) in images" :key="i">
                    <div class="image is-1by1 is-full"
                      v-bind:style="{ backgroundImage: `url(${getImageURL(carousel, '700,fit')})`, backgroundSize: 'contain', backgroundPosition: 'center', backgroundRepeat: 'no-repeat' }">
                    </div>
                  </b-carousel-item>
                  <template slot="indicators" slot-scope="props">
                    <span class="al image" style="width:max-content;">
                      <vue-load-image>
                        <img slot="image" :src="getIndicatorURL(props.i)" style="height:85px;" />
                        <img slot="preloader" :src="getImageURL('https://i.stack.imgur.com/kOnzy.gif')"
                          style="height:25px;" />
                        <img slot="error" src="/ui/images/blank_female_profile.png" style="height:85px;" />
                      </vue-load-image>
                    </span>
                  </template>
                </b-carousel>
              </div>
              
              <!-- サイドバー（固定） -->
              <div class="column is-2" style="background-color: rgb(147 134 134 / 14%);">
                <div class="p-4">
                  <div class="columns is-mobile is-multiline">
                    <div class="column is-half card-container">
                      <b-button @click="changeActorImage()" class="is-primary is-fullwidth"
                        style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Set Main') }}</b-button>
                    </div>
                    <div class="column is-half card-container">
                      <b-button @click="changeActorFaceImage()" class="is-primary is-fullwidth"
                        style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Set Face') }}</b-button>
                    </div>
                    <div class="column is-full card-container">
                      <b-button @click="deleteActorImage()" class="is-danger is-fullwidth"
                        style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Delete') }}</b-button>
                    </div>
                  </div>
                  <span style="display: flex; justify-content: center;">{{ $t('Main Image') }}</span>
                  <div class="vue-load-image" style="display: flex; justify-content: center;">
                    <img :src="actor.image_url" style="max-height: 23vh">
                  </div>
                  <span style="display: flex; justify-content: center;">{{ $t('Face Image') }}</span>
                  <div class="vue-load-image" style="display: flex; justify-content: center;">
                    <img :src="actor.face_image_url" style="max-height: 23vh">
                  </div>
                </div>
              </div>

            </div>

          </b-tab-item>

          <b-tab-item :label="$t('Search')">
            <div class="columns is-gapless" style="height: 66vh;overflow: hidden;">


              <!-- カルーセル：スクロール可能 -->
              <div class="column" style="overflow-y: auto; padding: 1rem;">
                <div>
                  <vue-select-image ref="vueSelectImage" :data-images="getImages" :is-multiple="true"
                    :selected-images="initialSelected" @onselectmultipleimage="onSelectMultipleImage" />
                </div>
              </div>
              
              <!-- サイドバー（固定） -->
              <div class="column is-2" style="background-color: rgb(147 134 134 / 14%);;">
                <div class="p-4">
                  <div class="columns is-mobile is-multiline">
                    <div class="column is-half card-container"><b-button :disabled="SelectMultipleImage.length != 1"
                        @click="setActorImage()" class="is-primary is-fullwidth"
                        style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Set Main') }}</b-button>
                    </div>
                    <div class="column is-half card-container"><b-button :disabled="SelectMultipleImage.length != 1"
                        @click="setActorFaceImage()" class="is-primary is-fullwidth"
                        style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Set Face') }}</b-button>
                    </div>
                    <div class="column is-half card-container"><b-button :disabled="SelectMultipleImage.length === 0"
                        @click="addActorImages()" class="is-primary is-fullwidth"
                        style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Add Images')
                        }}</b-button></div>
                    <div class="column is-half card-container"><b-button @click="resetSelection"
                        class="button is-fullwidth" style="display:flex; justify-content:center; margin-bottom:5px;">{{
                          $t('Clear') }}</b-button></div>
                  </div>
                  <span style="display: flex; justify-content: center;">Scrape</span>

                  <div class="toggle-container">
                    <label v-for="(label, key) in SiteEnum" :key="key"
                      :class="['toggle-button', { active: selectedSite === label }]">
                      <input type="radio" :value="label" v-model="selectedSite" />
                      {{ label }}
                    </label>
                  </div>

                  <div>
                    <div class="columns is-multiline is-mobile" style="margin-inline: -2px;">
                      <div v-for="category in categories" :key="category" class="column card-container" :class="{
                        'is-half': $t(category).length >= 7,
                        'is-4': $t(category).length < 7
                      }">
                        <div class="card selectable-card"
                          :class="{ 'is-selected': selectedKeywords.includes(category) }"
                          @click="toggleKeyword(category)" role="button" tabindex="0"
                          @keydown.space.prevent="toggleKeyword(category)"
                          @keydown.enter.prevent="toggleKeyword(category)"
                          :aria-pressed="selectedKeywords.includes(category).toString()">
                          <div class="card-content">
                            {{ $t(category) }}
                          </div>
                        </div>
                      </div>
                    </div>

                    <div class="columns is-mobile is-multiline">
                      <div class="column is-half card-container"><b-button :disabled="selectedKeywords.length === 1000"
                          @click="searchWithSelectedKeywords()" class="is-primary is-fullwidth"
                          style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Search') }}</b-button>
                      </div>
                      <div class="column is-half card-container"><b-button :disabled="selectedKeywords.length === 0"
                          @click="clearSelection()" class="is-fullwidth"
                          style="display:flex; justify-content:center; margin-bottom:5px;">{{ $t('Clear All')
                          }}</b-button>
                      </div>
                    </div>

                    <!-- 追加キーワード入力 -->
                    <div class="field has-addons" style="margin-bottom: 10px;">
                      <div class="control is-expanded">
                        <input v-model.trim="newKeyword" class="input" type="text" placeholder="キーワードを追加"
                          @keyup.enter="addKeyword" />
                      </div>
                      <div class="control">
                        <button class="button is-primary" @click="addKeyword">
                          追加
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

            </div>
          </b-tab-item>


          <b-tab-item :label="$t('Images')">
            <ListEditor :list="this.actor.imageArray" type="image_arr" :blurFn="() => blur('image_arr')"
              :showUrl="true" />
          </b-tab-item>
        </b-tabs>

      </section>

      <footer class="modal-card-foot">
        <b-field>
          <b-button type="is-primary" @click="save">{{ $t('Save Details') }}</b-button>
        </b-field>
      </footer>
    </div>
  </div>
</template>

<script>
import ky from 'ky'
import GlobalEvents from 'vue-global-events'
import ListEditor from '../../components/ListEditor'
import VueLoadImage from 'vue-load-image'

import VueSelectImage from 'vue-select-image'
// require('vue-select-image/dist/vue-select-image.css')

const SiteEnum = Object.freeze({
  GOOGLE: 'Google',
  BING: 'Bing'
})

export default {
  name: 'EditActorImage',
  components: { VueLoadImage, ListEditor, GlobalEvents, VueSelectImage },
  data() {
    const actor = Object.assign({}, this.$store.state.overlay.actoreditimage.actor)

    let tmp_images;
    try {
      tmp_images = JSON.parse(actor.image_arr)
    } catch {
      tmp_images = []
    }
    actor.imageArray = tmp_images.map(i => i)
    try {
      actor.aliasArray = JSON.parse(actor.aliases)
    } catch {
      actor.aliasArray = []
    }

    return {
      actor,
      // A shallow copy won't work, need a deep copy
      source: JSON.parse(JSON.stringify(actor)),
      changesMade: false,
      extrefsChangesMade: false,
      countryList: [],
      countries: [],
      selectedCountry: '',
      filteredCountries: [],
      extrefsArray: [],
      extrefsSource: '',
      getImages: [],
      // initialSelected:[],
      SelectImage: '',
      SelectMultipleImage: [],
      // selectOne: false

      SiteEnum,
      selectedSite: SiteEnum.GOOGLE,
      categories: ['AV', 'PORN', 'SEXY', 'ERO', 'NAKED', 'CUTE', 'NIPPLE', 'OPPAI', 'BOOBS', 'BUSTY', 'FACE', 'FANZA', 'GRAVURE', 'FULL BODY', 'CLOSE UP', 'BODY SHOT'],
      selectedKeywords: [],

      carouselSlide: 0,

    }
  },
  computed: {
    images() {
      if (this.actor.image_arr == undefined || this.actor.image_arr == "") {
        return []
      }
      return JSON.parse(this.actor.image_arr).filter(im => im != "")
    },
    chunkedCategories() {
      const chunkSize = 2;
      const chunks = [];
      for (let i = 0; i < this.categories.length; i += chunkSize) {
        chunks.push(this.categories.slice(i, i + chunkSize));
      }
      return chunks;
    },

  },
  mounted() {
    // ky.get(`/api/actor/extrefs/${this.actor.id}`)
    //   .json()
    //   .then(list => {
    //     this.extrefsArray = []
    //     list.forEach(extref => {
    //       this.extrefsArray.push(extref.external_reference.external_url)
    //     }      
    //     )
    //     this.extrefsSource = JSON.parse(JSON.stringify(this.extrefsArray))
    //     this.extrefsChangesMade=false
    //   })

    //NotFound画像を非表示にする
    const observer = new MutationObserver(() => {
      const imgs = this.$el.querySelectorAll('img');
      imgs.forEach(img => {
        if (img._errorHandlerAttached) return;
        img._errorHandlerAttached = true;

        img.addEventListener('error', () => {
          const li = img.closest('li');
          if (li) li.style.display = 'none';
        });
      });
    });

    observer.observe(this.$el, {
      childList: true,
      subtree: true
    });

  },

  methods: {
    // Custom Black
    addKeyword() {
      if (this.newKeyword && !this.categories.includes(this.newKeyword)) {
        this.categories.push(this.newKeyword)
      }
      this.newKeyword = ''
    },
    toggleKeyword(category) {
      const index = this.selectedKeywords.indexOf(category);
      if (index === -1) {
        this.selectedKeywords.push(category);
      } else {
        this.selectedKeywords.splice(index, 1);
      }
    },
    searchWithSelectedKeywords() {
      const keywordString = this.selectedKeywords
        .map(kw => (kw.includes(' ') ? `"${kw}"` : kw))
        .join(' ');
      this.scrapeActorImage(this.selectedSite, keywordString);
    },
    clearSelection() {
      this.selectedKeywords = []
    },
    changeActorImage(val) {
      ky.post('/api/actor/setimage', {
        json: {
          actor_id: this.actor.id,
          url: this.images[this.carouselSlide]
        }
      }).json().then(data => {
        this.actor = data
        this.$store.state.overlay.actoreditimage.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide = 0
      })
    },
    changeActorFaceImage(val) {
      ky.post('/api_custom/actor/setfaceimage', {
        json: {
          actor_id: this.actor.id,
          url: this.images[this.carouselSlide]
        }
      }).json().then(data => {
        this.actor = data
        this.$store.state.overlay.actoreditimage.actor = data
        this.$store.state.overlay.actordetails.actor = data
        // this.carouselSlide = 0
      })
    },
    deleteActorImage(val) {
      ky.delete('/api/actor/delimage', {
        json: {
          actor_id: this.actor.id,
          url: this.images[this.carouselSlide]
        }
      }).json().then(data => {
        this.actor = data
        this.$store.state.overlay.actoreditimage.actor = data
        this.$store.state.overlay.actordetails.actor = data
        // this.carouselSlide = 0
      })
    },

    setActorImage(val) {
      const selected = this.SelectMultipleImage[0];
      const selectedId = selected?.id;

      ky.post('/api/actor/setimage', {
        json: {
          actor_id: this.actor.id,
          url: selected?.url
        }
      }).json().then(data => {
        this.actor = data
        this.$store.state.overlay.actoreditimage.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide = 0
        if (selectedId != null && this.getImages[selectedId - 1]) {
          this.getImages[selectedId - 1].disabled = true;
        }
        this.resetSelection();
      })
    },
    setActorFaceImage(val) {
      const selected = this.SelectMultipleImage[0];
      const selectedId = selected?.id;

      ky.post('/api_custom/actor/setfaceimage', {
        json: {
          actor_id: this.actor.id,
          url: selected?.url
        }
      }).json().then(data => {
        this.actor = data
        this.$store.state.overlay.actoreditimage.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide = 0
        if (selectedId != null && this.getImages[selectedId - 1]) {
          this.getImages[selectedId - 1].disabled = true;
        }
        this.resetSelection();
      })
    },
    addActorImages(val) {
      ky.post('/api_custom/actor/addimages', {
        json: {
          actor_id: this.actor.id,
          urls: this.SelectMultipleImage.map(img => img.url)
        }
      }).json().then(data => {
        this.actor = data
        this.$store.state.overlay.actoredit.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide = 0
        this.SelectMultipleImage.forEach(selected => {
          const target = this.getImages.find(img => img.id === selected.id);
          if (target) target.disabled = true;
        });
        this.resetSelection();
      })
    },
    resetSelection() {
      this.$refs.vueSelectImage.resetMultipleSelection();
      this.SelectMultipleImage = [];
    },
    onSelectImage(selected) {
      this.SelectImage = "id:" + selected.id
      selectOne = true
    },
    onSelectMultipleImage(selected) {
      let arr = [];
      for (let i = 0; i < selected.length; i++) {
        const id = selected[i].id;
        const url = this.getImages[id - 1].src;
        arr.push({ id: id, url: url });
      }
      this.SelectMultipleImage = arr;
    },
    getImageURL(u, size) {
      if (u.startsWith('http') || u.startsWith('https')) {
        return '/img/' + size + '/' + u.replace('://', ':/')
      } else {
        return u
      }
    },
    getIndicatorURL(idx) {
      if (this.images[idx] !== undefined) {
        return this.getImageURL(this.images[idx], 'x85')
      } else {
        return '/ui/images/blank_female_profile.png'
      }
    },
    scrapeActorImage(site, val) {
      this.resetSelection();
      this.getImages = [];
      ky.post('/api_custom/images/searchImage', {
        json: {
          actor_id: this.actor.id,
          url: "",
          keyword: val,
          site: site
        }
      }).json().then(data => {
        this.getImages = data.images
      })
    },

    // Custom End
    close() {
      if (this.changesMade || this.extrefsChangesMade) {
        this.$buefy.dialog.confirm({
          title: 'Close without saving',
          message: 'Are you sure you want to close before saving your changes?',
          confirmText: 'Close',
          type: 'is-warning',
          hasIcon: true,
          onConfirm: () => this.$store.commit('overlay/hideActorEditImage')
        })
        return
      }
      this.$store.commit('overlay/hideActorEditImage')
    },
    async save() {
      this.$store.state.actorList.isLoading = true
      // this.actor.image_arr = JSON.stringify(this.actor.imageArray)  

      await ky.post(`/api/actor/edit/${this.actor.id}`, { json: { ...this.actor } })
      // await ky.post(`/api/actor/edit_extrefs/${this.actor.id}`, { json: this.extrefsArray  })
      await ky.get('/api/actor/' + this.actor.id).json().then(data => {
        if (data.id != 0) {
          this.$store.state.overlay.actordetails.actor = data
        }
      })

      this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
      this.changesMade = false
      this.extrefsChangesMade = false
      this.$store.state.actorList.isLoading = false
      this.close()
    },

    blur(field) {
      if (this.changesMade) return // Changes have already been made. No point to check any further   
      if (['image_arr', 'tattoos', 'piercings', 'aliases', 'urls'].includes(field)) {
        if (this.actor[field].length !== this.source[field].length) {
          this.changesMade = true
        } else {
          // change to actor and use foreah 
          for (let i = 0; i < this.actor[field].length; i++) {
            if (this.actor[field][i] !== this.source[field][i]) {
              this.changesMade = true
              break
            }
          }
        }
      } else if (this.actor[field] !== this.source[field]) {
        this.changesMade = true
      }
    },
    extrefBlur() {
      if (this.extrefsChangesMade) return // Changes have already been made. No point to check any further         
      if (this.extrefsArray.length !== this.extrefsSource.length) {
        this.extrefsChangesMade = true
      } else {
        // change to actor and use foreah 
        for (let i = 0; i < this.extrefsArray.length; i++) {
          if (this.extrefsArray[i] !== this.extrefsSource[i]) {
            this.extrefsChangesMade = true
            break
          }
        }
      }
    },

  },
}
</script>

<style scoped>
.modal-card {
  width: 90%;
  height: 90%;
}

.tab-item {
  height: 40vh;
}

.carousel-item {
  max-width: 37%;
  margin-left: auto;
  margin-right: auto;
}

:deep(.carousel .carousel-indicator) {
  justify-content: flex-start;
  width: 100%;
  max-width: min-content;
  margin-left: auto;
  margin-right: auto;
  overflow: auto;
}

:deep(.carousel .carousel-indicator .indicator-item:not(.is-active)) {
  opacity: 0.5;
}

.toggle-container {
  display: flex;
  width: 100%;
  max-width: 400px;
  /* 任意で制限（不要なら削除OK） */
  margin-bottom: 1rem;
}

.toggle-button {
  flex: 1 1 50%;
  padding: 10px 0;
  text-align: center;
  cursor: pointer;
  border: 1px solid #ccc;
  background-color: #f5f5f5;
  font-weight: bold;
  user-select: none;
  transition: background-color 0.2s;
}

.toggle-button+.toggle-button {
  border-left: none;
}

/* ラジオボタン自体は非表示 */
.toggle-button input {
  display: none;
}

/* アクティブ状態 */
.toggle-button.active {
  background-color: #42b983;
  color: white;
  border-color: #42b983;
}

/* .card-container {
  padding: 0.25rem;
}

.selectable-card {
  cursor: pointer;
  border: 1px solid #ccc;
  transition: border-color 0.3s, background-color 0.3s;
  border-radius: 6px;
}

.card-content {
  background-color: transparent;
  padding: 0.5rem !important;
}

.selectable-card:hover {
  border-color: #999;
}

.selectable-card.is-selected {
  border-color: #3273dc;
  background-color: #f0f8ff;
} */
.card-container {
  padding: 0.25rem;
}

.selectable-card {
  border-radius: 999px;
  /* 角丸 */
  border: 1px solid #e6e6e6;
  background: #fff;
  cursor: pointer;
  user-select: none;
  transition: box-shadow 0.16s ease, transform 0.12s ease, background-color 0.12s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 36px;
}

.selectable-card .card-content {
  width: 100%;
  text-align: center;
  padding: 0.5rem 0.75rem;
  font-weight: 600;
  font-size: smaller;
}

.selectable-card.is-selected {
  /* background: #3273dc22;
  border-color: #3273dc; */
  background: #0044ff29;
  border-color: #3273dc;
  box-shadow: 0 2px 6px rgba(50, 115, 220, 0.12);
  transform: translateY(-1px);
}

.selectable-card:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(50, 115, 220, 0.12);
}
</style>

<style>
.vue-select-image__wrapper {
  overflow: auto;
  list-style-image: none;
  list-style-position: outside;
  list-style-type: none;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  /* 自動で列数調整 */
  gap: 10px;
  /* 画像の間隔 */
  list-style: none;
  padding: 0;
  margin: 0;
}

.vue-select-image__item {
  margin: 0 0 10px 0;
  float: left;
}

.vue-select-image__thumbnail {
  cursor: pointer;
  display: block;
  padding: 6px;
  line-height: 20px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.055);
  transition: all 0.2s ease-in-out;
}

.vue-select-image__thumbnail--selected {
  background: #ab3d99;
}

.vue-select-image__thumbnail--disabled {
  background: #9e459a;
  cursor: not-allowed;
}

.vue-select-image__thumbnail--disabled>.vue-select-image__img {
  opacity: 0.5;
}

.vue-select-image__img {
  -webkit-user-drag: none;
  display: block;
  max-width: 100%;
  margin-right: auto;
  margin-left: auto;
}

.vue-select-image__lbl {
  line-height: 3;
}

@media only screen and (min-width: 1200px) {
  .vue-select-image__item {
    margin-left: 0px;
  }
}
</style>
