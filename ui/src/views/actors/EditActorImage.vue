<template>
  <div class="modal is-active">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keyup.esc="close"
      @keyup.s="save"/>

    <div class="modal-background"></div>

    <div class="modal-card">
      <header class="modal-card-head">
        <p class="modal-card-title">{{ $t('Edit actor Image') }} - {{ actor.name }}</p>
        <button class="delete" @click="close" aria-label="close"></button>
      </header>

      <section class="modal-card-body">
        <b-tabs position="is-centered" :animated="false">

          <b-tab-item :label="$t('Actor Images')">
                <b-carousel v-model="carouselSlide" @change="scrollToActiveIndicator" :autoplay="false" :indicator-inside="false">
                  <b-carousel-item v-for="(carousel, i) in images" :key="i">
                    <div class="image is-1by1 is-full"
                         v-bind:style="{backgroundImage: `url(${getImageURL(carousel, '700,fit')})`, backgroundSize: 'contain', backgroundPosition: 'center', backgroundRepeat: 'no-repeat'}"></div>
                  </b-carousel-item>
                  <template slot="indicators" slot-scope="props">
                      <span class="al image" style="width:max-content;">
                        <vue-load-image>
                          <img slot="image" :src="getIndicatorURL(props.i)" style="height:85px;"/>
                          <img slot="preloader" :src="getImageURL('https://i.stack.imgur.com/kOnzy.gif')" style="height:25px;"/>
                          <img slot="error" src="/ui/images/blank_female_profile.png" style="height:85px;"/>
                        </vue-load-image>
                      </span>
                  </template>
                </b-carousel>
                <div class="flexcentre">
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;" v-on:click="setActorImage()">{{$t('Set Main')}}</b-button>
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="setActorFaceImage()">{{$t('Set Face')}}</b-button>
                <b-button v-if="images.length != 0" class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="deleteActorImage()">{{$t('Delete')}}</b-button>
</div>
          </b-tab-item>

<b-tab-item :label="$t('Search')">
  <div class="columns is-gapless" style="height: 66vh;overflow: hidden;">

    <!-- 左カラム：サイドバー（固定） -->
    <div class="column is-2" style="background-color: #f5f5f5;">
      <div class="p-4">
        <b-button :disabled="this.SelectMultipleImage.length != 1" @click="setActorImage()" class="is-primary is-fullwidth" style="display:flex; justify-content:center; margin-bottom: 5px;">{{ $t('Set Main') }}</b-button>
        <b-button :disabled="this.SelectMultipleImage.length != 1" @click="setActorFaceImage()" class="is-primary is-fullwidth" style="display:flex; justify-content:center; margin-bottom: 5px;">{{ $t('Set Face') }}</b-button>
        <b-button :disabled="this.SelectMultipleImage.length === 0" @click="addActorImages()" class="is-primary is-fullwidth" style="display:flex; justify-content:center; margin-bottom: 5px;">{{ $t('Add Images') }}</b-button>
        <span style="display: flex; justify-content: center;" >Scrape</span>
        <b-button class="button is-fullwidth" style="display: flex; justify-content: center;" v-on:click="resetSelection">{{$t('Reset Selection')}}</b-button>
        <b-button class="button is-fullwidth" style="display: flex; justify-content: center;" v-on:click="scrapeActorImage('b', 'エロ')">{{$t('Bing')}}</b-button>
        <b-button class="button is-fullwidth" style="display: flex; justify-content: center;" v-on:click="scrapeActorImage('g', 'エロ')">{{$t('Google')}}</b-button>
        <b-button class="button is-fullwidth" style="display: flex; justify-content: center;" v-on:click="scrapeActorImage('g', 'セクシー女優 全裸')">{{$t('Google2')}}</b-button>
        <b-button class="button is-fullwidth" style="display: flex; justify-content: center;" v-on:click="scrapeActorImage('g', 'グラビア')">{{$t('Gravia')}}</b-button>
        <b-button class="button is-fullwidth" style="display: flex; justify-content: center;" v-on:click="scrapeActorImage('g', '顔')">{{$t('Face')}}</b-button>
                
      </div>
    </div>

    <!-- 右カラム：スクロール可能 -->
    <div class="column" style="overflow-y: auto; padding: 1rem;">
        <div>
          <vue-select-image
            ref="vueSelectImage"
            :data-images="getImages"
            :is-multiple="true"
            :selected-images="initialSelected"
            @onselectmultipleimage="onSelectMultipleImage"
          />
          <p>選択中: {{ selectMultipleImage }}</p>
        </div>
    </div>

  </div>
</b-tab-item>


          <b-tab-item :label="$t('Images')">
            <ListEditor :list="this.actor.imageArray" type="image_arr" :blurFn="() => blur('image_arr')" :showUrl="true"/>
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

export default {
  name: 'EditActorImage',
  components: { VueLoadImage, ListEditor, GlobalEvents, VueSelectImage },
  data () {
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
      SelectMultipleImage:[],
      // selectOne: false
    }
  },
  computed: {
    images () {
      if (this.actor.image_arr==undefined || this.actor.image_arr=="") {
        return []
      }      
      return JSON.parse(this.actor.image_arr).filter(im => im != "")      
    },

  },
  mounted () {
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
    setActorImage (val) {
      const selected = this.SelectMultipleImage[0];
      const selectedId = selected?.id;

      ky.post('/api/actor/setimage', {
      json: {
        actor_id: this.actor.id,
        url: selected?.url
      }}).json().then(data => {        
        this.actor = data
        this.$store.state.overlay.actoredit.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide=0
        if (selectedId != null && this.getImages[selectedId - 1]) {
          this.getImages[selectedId - 1].disabled = true;
        }
        this.resetSelection();
      })    
    },
    setActorFaceImage (val) {
      const selected = this.SelectMultipleImage[0];
      const selectedId = selected?.id;

      ky.post('/api_custom/actor/setfaceimage', {
      json: {
        actor_id: this.actor.id,
        url: selected?.url
      }}).json().then(data => {        
        this.actor = data
        this.$store.state.overlay.actoredit.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide=0
        if (selectedId != null && this.getImages[selectedId - 1]) {
          this.getImages[selectedId - 1].disabled = true;
        }
        this.resetSelection();
      })     
    },
    addActorImages (val) {
      ky.post('/api_custom/actor/addimages', {
      json: {
        actor_id: this.actor.id,
        urls: this.SelectMultipleImage.map(img => img.url)
      }}).json().then(data => {        
        this.actor = data
        this.$store.state.overlay.actoredit.actor = data
        this.$store.state.overlay.actordetails.actor = data
        this.carouselSlide=0
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
    onSelectImage(selected){
      this.SelectImage = "id:" + selected.id
      selectOne = true
    },
    onSelectMultipleImage(selected){
      let arr = [];
      for(let i=0; i<selected.length; i++){
        const id = selected[i].id;
        const url = this.getImages[id - 1].src;
        arr.push({ id: id, url: url });
      }
      this.SelectMultipleImage = arr;
    },
    getImageURL (u, size) {
      if (u.startsWith('http') || u.startsWith('https')) {
        return '/img/' + size + '/' + u.replace('://', ':/')
      } else {
        return u
      }
    },
    getIndicatorURL (idx) {      
      if (this.images[idx] !== undefined) {
        return this.getImageURL(this.images[idx], 'x85')
      } else {
        return '/ui/images/blank_female_profile.png'
      }
    },
    scrapeActorImage (site,val) {
      ky.post('/api_custom/images/searchImage', {
      json: {
        actor_id: this.actor.id,
        url: "",
        keyword: val,
        site: site
      }}).json().then(data => {
        this.getImages =data.images
      })    
    },

    // Custom End
    close () {
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
    async save () {
      this.$store.state.actorList.isLoading = true
      this.actor.image_arr = JSON.stringify(this.actor.imageArray)  

      await ky.post(`/api/actor/edit/${this.actor.id}`, { json: { ...this.actor } })
      // await ky.post(`/api/actor/edit_extrefs/${this.actor.id}`, { json: this.extrefsArray  })
      await ky.get('/api/actor/'+this.actor.id).json().then(data => {
        if (data.id != 0){
          this.$store.state.overlay.actordetails.actor = data          
        }          
      })

      this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
      this.changesMade = false
      this.extrefsChangesMade = false
      this.$store.state.actorList.isLoading = false
      this.close()
    },

    blur (field) {
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
    extrefBlur () {      
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
    width: 30% !important;
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
</style>

<style>
.vue-select-image__wrapper {
  overflow: auto;
  list-style-image: none;
  list-style-position: outside;
  list-style-type: none;
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); /* 自動で列数調整 */
  gap: 10px; /* 画像の間隔 */
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
  padding: 4px;
  line-height: 20px;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.055);
  transition: all 0.2s ease-in-out;
}

.vue-select-image__thumbnail--selected {
  background: #08c;
}

.vue-select-image__thumbnail--disabled {
  background: #b9b9b9;
  cursor: not-allowed;
}

.vue-select-image__thumbnail--disabled > .vue-select-image__img {
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
