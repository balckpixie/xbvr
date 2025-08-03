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

          <b-tab-item :label="$t('Search Images')">
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
                <span style="display: flex; justify-content: center;margin-left: 1em;" >Scrape</span>
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="scrapeActorImage('b', 'エロ')">{{$t('Bing')}}</b-button>
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="scrapeActorImage('g', 'エロ')">{{$t('Google')}}</b-button>
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="scrapeActorImage('g', 'セクシー女優 全裸')">{{$t('Google2')}}</b-button>
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="scrapeActorImage('g', 'グラビア')">{{$t('Gravia')}}</b-button>
                <b-button class="button is-primary is-small" style="display: flex; justify-content: center;margin-left: 1em;" v-on:click="scrapeActorImage('g', '顔')">{{$t('Face')}}</b-button>
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
          <b-button v-if="actor.scenes.length == 0 && !actor.name.startsWith('aka:')" type="is-danger" outlined @click="deleteactor">{{ $t('Delete Actor') }}</b-button>
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

export default {
  name: 'EditActorImage',
  components: { VueLoadImage, ListEditor, GlobalEvents },
  data () {
    const actor = Object.assign({}, this.$store.state.overlay.actoreditimage.actor)
    let images;
    try {
      images = JSON.parse(actor.image_arr)
    } catch {
      images = []
    }    
    actor.imageArray = images.map(i => i)    
    try {
      actor.aliasArray = JSON.parse(actor.aliases)
    } catch {
      actor.aliasArray = []
    }
    // try {
    //   actor.tattooArray = JSON.parse(actor.tattoos)
    // } catch {
    //   actor.tattooArray = []
    // }
    // try {
    //   actor.piercingArray = JSON.parse(actor.piercings)
    // } catch {
    //   actor.piercingArray = []
    // }
    // actor.measurements = Math.round(actor.band_size / 2.54) + actor.cup_size + '-' + Math.round(actor.waist_size / 2.54) + '-' + Math.round(actor.hip_size / 2.54)
    // this.convertCountryCodeToName()
    // let urls;
    // try {
    //   urls = JSON.parse(actor.urls)
    // } catch {
    //   urls = []
    // }    
    // actor.urlArray = urls.map(i => i.url)    

    // const totalInches = Math.round(actor.height / 2.54)
    // const  feet = Math.floor(totalInches / 12)
    // const inches =  Math.round(totalInches - (feet*12))      
    // const lbs = Math.round(actor.weight * 220462 / 100000);
    // actor.feet = feet
    // actor.inches = inches
    // actor.lbs = lbs

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
      getimages: [],
    }
  },
  computed: {
    images () {
      if (this.actor.image_arr==undefined || this.actor.image_arr=="") {
        return []
      }      
      return JSON.parse(this.actor.image_arr).filter(im => im != "")      
    },
    birthdate: {
      get () {        
        if (this.actor.birth_date=='0001-01-01T00:00:00Z') {
          return new Date()
        }
        return new Date(this.actor.birth_date)
      },
      set (value) {        
        if (value==null){
          this.actor.birth_date=null
        }else{
        // remove the time offset, or toISOString may result in a different date
        let adjustedDate = new Date(value.getTime() - (value.getTimezoneOffset() * 60000))
        this.actor.birth_date = adjustedDate.toISOString().split('.')[0] + 'Z'        
        }
      }
    },
    useImperialEntry () {
      return this.$store.state.optionsAdvanced.advanced.useImperialEntry
    },
  },
  mounted () {
    ky.get('/api/actor/countrylist')
    .json()
    .then(list => {
      this.countryList = list
      this.convertCountryCodeToName()
    })  

  ky.get(`/api/actor/extrefs/${this.actor.id}`)
    .json()
    .then(list => {
      this.extrefsArray = []
      list.forEach(extref => {
        this.extrefsArray.push(extref.external_reference.external_url)
      }      
      )
      this.extrefsSource = JSON.parse(JSON.stringify(this.extrefsArray))
      this.extrefsChangesMade=false
    })
  },
  methods: {
    // Custom Black
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
      if (this.useImperialEntry) {
        this.actor.height = Math.round(((this.actor.feet * 12) + this.actor.inches) * 2.54)
        this.actor.weight = Math.round(this.actor.lbs * 453592 / 1000000);
      }
      this.actor.aliases = JSON.stringify(this.actor.aliasArray)      
      this.actor.tattoos = JSON.stringify(this.actor.tattooArray)         
      this.actor.piercings = JSON.stringify(this.actor.piercingArray)
      if (this.countries.length==0){
        this.actor.nationality=""
      } else {
        this.actor.nationality=this.countries[0]
      }

      let  dataArray = []
      if (this.actor.urls != "") {
        const existingurls = JSON.parse(this.actor.urls)      
        this.actor.urlArray.forEach(url => {        
          let t = ''
          existingurls.forEach(u => {
            if (u.url==url) {
              t=u.type
            }
          })
          dataArray.push({
            url,
            type: t
          })
        })
      }
      this.actor.height = parseInt(this.actor.height)
      this.actor.weight = parseInt(this.actor.weight)
      this.actor.start_year = parseInt(this.actor.start_year)
      this.actor.end_year = parseInt(this.actor.end_year)

      this.actor.urls = JSON.stringify(dataArray)

      this.actor.image_arr = JSON.stringify(this.actor.imageArray)  

      await ky.post(`/api/actor/edit/${this.actor.id}`, { json: { ...this.actor } })
      await ky.post(`/api/actor/edit_extrefs/${this.actor.id}`, { json: this.extrefsArray  })
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
    deleteactor () {
      this.$buefy.dialog.confirm({
        title: 'Delete actor',
        message: `Do you really want to delete <strong>${this.actor.name}</strong>`,
        type: 'is-info is-wide',
        hasIcon: true,
        id: 'heh',
        onConfirm: () => {
          ky.delete(`/api/actor/delete/${this.actor.id}`).json().then(data => {
            this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset - this.$store.state.actorList.limit })
            this.$store.commit('overlay/hideActorEditImage')
            this.$store.commit('overlay/hideActorDetails')
          })
        }
      })
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
    getFilteredCountries (text) {
      const filtered = this.countryList.filter(option => (
        option.name.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0        
      ))
      this.filteredCountries=[]
      filtered.forEach(item => this.filteredCountries.push(item.name))      
    },
    getYear (text) {
      if (text==0) {
        return ""
      }
      return year
    },
    convertCountryCodeToName() {      
      if (this.countryList != undefined && this.actor != undefined && this.actor.nationality.length == 2) {
        this.countryList.forEach(country => {
          if (country.code == this.actor.nationality) {
            this.actor.nationality=country.name
          }
        })
      }

      if (this.actor != undefined){      
        this.countries = [this.actor.nationality]
      }      
    },
  },
}
</script>

<style scoped>
.modal-card {
  width: 65%;
}

.tab-item {
  height: 40vh;
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
