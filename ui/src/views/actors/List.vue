<template>
  <div>
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keydown.left="prevpage"
      @keydown.right="nextpage"
      @keydown.o="prevpage"
      @keydown.p="nextpage"
    />
    <b-loading :is-full-page="true" :active.sync="isLoading"></b-loading>

    <div class="columns is-multiline is-full">
      <div class="column">
        <strong>{{total}} results</strong>
      </div>
      <!-- Custom Black（ひらがな選択時は非表示） -->
      <div class="column" v-if="!hideHiragana">
      <!-- Custom End -->
        <b-tooltip :label="$t('Press o/left arrow to page back, p/right arrow to page forward')" :delay="500" position="is-top">
          <b-pagination
              :total="total"
              v-model="current"
              range-before=1
              range-after=3    
              size="is-small"                                           
              :per-page="limit"
              aria-next-label="Next page"
              aria-previous-label="Previous page"
              aria-page-label="Page"
              aria-current-label="Current page"
              :page-input=true
              @change="pageChanged"
              debounce-page-input="250"
              >
          </b-pagination>
        </b-tooltip>
        <span v-show="show_actor_id==='never show, just need the computed show_actor_id to trigger '">{{show_actor_id}}</span>
      </div>

      <!-- Custom Black（ひらがな選択時はカスタムコンポーネントを表示） -->
      <div v-if="hideHiragana">
        <CustomPagination
          :total="total"
          :limit="limit"
          :offset="offset"
          @change="onOffsetChange"
        />
      </div>
      <!-- Custom End -->
       
      <div class="column">
        <div class="is-pulled-right">
          <b-field>
            <span class="list-header-label">{{$t('Card size')}}</span>
            <b-radio-button v-model="cardSize" native-value="1" size="is-small">
              XS
            </b-radio-button>
            <b-radio-button v-model="cardSize" native-value="2" size="is-small">
              S
            </b-radio-button>
            <b-radio-button v-model="cardSize" native-value="3" size="is-small">
              M
            </b-radio-button>
            <b-radio-button v-model="cardSize" native-value="4" size="is-small">
              L
            </b-radio-button>
          </b-field>
        </div>

        <!-- Custom Black(顔表示/体表示のトグルボタン追加) -->
        <div class="is-pulled-right">
          <div class="toggle-container">
            <label class="toggle-button" :class="{ active: showFace === true }">
              <input type="radio" :value="true" v-model="showFace" />
              Face
            </label>
            <label class="toggle-button" :class="{ active: showFace === false }">
              <input type="radio" :value="false" v-model="showFace" />
              Body
            </label>
          </div>
        </div>
        <!-- Custom End -->

      </div>
    </div>
        <!-- 頭文字検索 (画像①レイアウト + 画像②デザイン) -->
        <div class="initial-search-card">
          <div class="initial-group-row">
            <button
              v-for="grp in initialGroups"
              :key="grp.key"
              type="button"
              class="initial-pill"
              :class="{ 'is-active': activeInitialGroup === grp.key }"
              @click="selectInitialGroup(grp)"
            >
              {{ grp.label }}
            </button>
          </div>
          <div v-if="activeInitialGroupObj" class="initial-sub-row">
            <button
              v-for="letter in activeInitialGroupObj.letters"
              :key="letter"
              type="button"
              class="initial-pill initial-sub-pill"
              :class="{ 'is-active': jumpTo === letter }"
              @click="selectInitialLetter(activeInitialGroupObj, letter)"
            >
              {{ letter }}
            </button>
            <button
              v-if="jumpTo"
              type="button"
              class="initial-pill initial-clear-pill"
              @click="clearInitialJump"
            >
              クリア
            </button>
          </div>
        </div>
         
    <div class="is-clearfix"></div>

    <div class="columns is-multiline">
      <div :class="['column', 'is-multiline', cardSizeClass]"
           v-for="actor in actors" :key="actor.id">
        <!-- Custom Black(顔表示用のprops追加) -->
        <ActorCard :actor="actor" :showFace="showFace"/>
        <!-- Custom End -->
      </div>
    </div>
      <div class="columns is-gapless is-centered" v-if="hideLetters">
        <b-radio-button v-model="jumpTo" native-value="" size="is-small"></b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="A" size="is-small">A</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="B" size="is-small">B</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="C" size="is-small">C</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="D" size="is-small">D</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="E" size="is-small">E</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="F" size="is-small">F</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="G" size="is-small">G</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="H" size="is-small">H</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="I" size="is-small">I</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="J" size="is-small">J</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="K" size="is-small">K</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="L" size="is-small">L</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="M" size="is-small">M</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="N" size="is-small">N</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="O" size="is-small">O</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="P" size="is-small">P</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="Q" size="is-small">Q/R</b-radio-button>        
        <b-radio-button v-model="jumpTo" native-value="S" size="is-small">S</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="T" size="is-small">T</b-radio-button>
        <b-radio-button v-model="jumpTo" native-value="U" size="is-small">U/V</b-radio-button>        
        <b-radio-button v-model="jumpTo" native-value="W" size="is-small">W/X/Y/Z</b-radio-button>
      </div>
      <!-- Custom Black（ひらがな選択時は非表示） -->
      <div class="columns is-gapless is-centered" v-if="!hideHiragana">          
      <!-- Custom End -->
        <b-tooltip :label="$t('Press k to page back, l to page forward')" :delay="500" position="is-top">
          <b-pagination
            :total="total"
            v-model="current"
            range-before=2
            range-after=3 
            size="is-small"                                              
            :per-page="limit"
            aria-next-label="Next page"
            aria-previous-label="Previous page"
            aria-page-label="Page"
            aria-current-label="Current page"
            :page-input=true
            @change="pageChanged"
            debounce-page-input="250"
            >
        </b-pagination>
      </b-tooltip>
      </div>
      <!-- Custom Black（ひらがな選択時はカスタムコンポーネントを表示） -->
      <div v-if="hideHiragana">
        <CustomPagination
          :total="total"
          :limit="limit"
          :offset="offset"
          @change="onOffsetChange"
        />
      </div>
      <!-- Custom End -->
  </div>
</template>

<script>
import ActorCard from './ActorCard'
import ky from 'ky'
import GlobalEvents from 'vue-global-events'
// Custom Black（ひらがな選択用のデータ追加）
import CustomPagination from '../../components/CustomPagination.vue';

export default {
  name: 'List',
  components: { ActorCard, GlobalEvents, CustomPagination },
  data () {
    return {      
      current: 0,      
      showFace:true,
      groups: ['あ', 'か', 'さ', 'た', 'な', 'は', 'ま', 'や', 'ら', 'わ'],
      hiragana: {
        'あ': ['あ', 'い', 'う', 'え', 'お'],
        'か': ['か', 'き', 'く', 'け', 'こ'],
        'さ': ['さ', 'し', 'す', 'せ', 'そ'],
        'た': ['た', 'ち', 'つ', 'て', 'と'],
        'な': ['な', 'に', 'ぬ', 'ね', 'の'],
        'は': ['は', 'ひ', 'ふ', 'へ', 'ほ'],
        'ま': ['ま', 'み', 'む', 'め', 'も'],
        'や': ['や', 'ゆ', 'よ'],
        'ら': ['ら', 'り', 'る', 'れ', 'ろ'],
        'わ': ['わ', 'を', 'ん']
      },
      selectedGroup: 'あ',
      selectedLetter: '',
      activeInitialGroup: 'あ',
      initialGroups: [
        { key: 'あ', label: 'あ行', kind: 'kana', letters: ['あ', 'い', 'う', 'え', 'お'] },
        { key: 'か', label: 'か行', kind: 'kana', letters: ['か', 'き', 'く', 'け', 'こ'] },
        { key: 'さ', label: 'さ行', kind: 'kana', letters: ['さ', 'し', 'す', 'せ', 'そ'] },
        { key: 'た', label: 'た行', kind: 'kana', letters: ['た', 'ち', 'つ', 'て', 'と'] },
        { key: 'な', label: 'な行', kind: 'kana', letters: ['な', 'に', 'ぬ', 'ね', 'の'] },
        { key: 'は', label: 'は行', kind: 'kana', letters: ['は', 'ひ', 'ふ', 'へ', 'ほ'] },
        { key: 'ま', label: 'ま行', kind: 'kana', letters: ['ま', 'み', 'む', 'め', 'も'] },
        { key: 'や', label: 'や行', kind: 'kana', letters: ['や', 'ゆ', 'よ'] },
        { key: 'ら', label: 'ら行', kind: 'kana', letters: ['ら', 'り', 'る', 'れ', 'ろ'] },
        { key: 'わ', label: 'わ行', kind: 'kana', letters: ['わ', 'を', 'ん'] },
        { key: 'A-G', label: 'A-G', kind: 'alpha', letters: ['A', 'B', 'C', 'D', 'E', 'F', 'G'] },
        { key: 'H-N', label: 'H-N', kind: 'alpha', letters: ['H', 'I', 'J', 'K', 'L', 'M', 'N'] },
        { key: 'O-T', label: 'O-T', kind: 'alpha', letters: ['O', 'P', 'Q', 'R', 'S', 'T'] },
        { key: 'U-Z', label: 'U-Z', kind: 'alpha', letters: ['U', 'V', 'W', 'X', 'Y', 'Z'] },
        { key: '0-9', label: '数字/記号', kind: 'alpha', letters: ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'] }
      ]
    }
  },
  watch: {
    // selectedGroupが変更されたとき、JumpToのデフォルト値を設定する
    selectedGroup(newVal, oldVal) {
      if (newVal != "") {
        this.jumpToKana = this.hiragana[newVal][0];
      } else {
        this.jumpToKana = ""
      }
// Custom End
    }
  },
  computed: {
    // Custom Black（ひらがな選択用のデータ追加）
    offset: {
      get () {
        return this.$store.state.actorList.offset - this.$store.state.actorList.limit  
      },
      set (value) {
        this.$store.state.actorList.offset =  value
        this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset })
      } 
    },
    jumpToKana : {
      get () {
        return this.selectedLetter
      },
      set (value) {
        if (value !== this.$store.state.actorList.filters.jumpTo) {
          this.$store.state.actorList.filters.jumpTo = value
          this.selectedLetter = value
          this.reloadList()
        }
      }
    },
    hideHiragana: {
      get () {        
        switch (this.$store.state.actorList.filters.sort) {
          case "furigana_asc":
            return true
          case "furigana_desc":
            return true
        }
        return false
      },
    },
    // Custom End
    cardSize: {
      get () {
        return this.$store.state.actorList.filters.cardSize
      },
      set (value) {
        this.$store.state.actorList.filters.cardSize = value
        switch (value){
          case "1":
            this.limit=36
            break
          case "2":
            this.limit=18
            break
          case "3":
            this.limit=10
            break
          case "4":
            this.limit=8
            break
            }            
        }      
    },
    limit: {
      get(){
        return this.$store.state.actorList.limit
      },
      set(newLimit){
        // find the position of the first actor
        let currentOffset = this.$store.state.actorList.offset - this.$store.state.actorList.limit + 1
        // what is the new page number, based on the new limit
        this.current = Math.floor(currentOffset / newLimit) + 1
        if (this.current<1)
          this.current=1
        this.$store.state.actorList.limit = newLimit
        // what is the the first actor based on the new page size
        this.$store.state.actorList.offset = (this.current -1) * this.$store.state.actorList.limit          
        this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset })
      }
    },
    jumpTo: {
      get () {
        return this.$store.state.actorList.filters.jumpTo
      },
      set (value) {
        // Custom Black（ひらがな選択時はjumpToをクリアする）
        if (value !== this.$store.state.actorList.filters.jumpTo) {
          this.$store.state.actorList.filters.jumpTo = value
          this.reloadList()
        }
        // Custom End
      }
    },
    cardSizeClass () {
      switch (this.$store.state.actorList.filters.cardSize) {
        case '1':
          return 'is-1'
        case '2':
          return 'is-2'
        case '3':
          return 'is-one-fifth'
        case '4':
          return 'is-one-quarter'
        default:
          return 'is-2'
      }
    },
    isLoading () {
      this.current = this.$store.state.actorList.offset / this.$store.state.actorList.limit
      return this.$store.state.actorList.isLoading
    },
    actors () {
      return this.$store.state.actorList.actors
    },
    total () {
      return this.$store.state.actorList.total
    },
    show_actor_id() {
      if (this.$store.state.actorList.show_actor_id != undefined && this.$store.state.actorList.show_actor_id !='')
      {
        ky.get('/api/actor/'+this.$store.state.actorList.show_actor_id).json().then(data => {
          if (data.id != 0){
            this.$store.commit('overlay/showActorDetails', { actor: data })
          }          
        })
        this.$store.state.actorList.show_actor_id = ''
      }
      
      return this.$store.state.actorList.show_actor_id
    },
    activeInitialGroupObj () {
      return this.initialGroups.find(g => g.key === this.activeInitialGroup) || null
    },
    hideLetters: {
      get () {        
        switch (this.$store.state.actorList.filters.sort) {
          case "":
            return true
          case "name_asc":
            return true
          case "name_desc":
            return true
        }
        return false
        },
    },
  },
  methods: {
    selectInitialGroup (grp) {
      this.activeInitialGroup = grp.key
      if (grp.kind === 'kana') {
        if (this.$store.state.actorList.filters.sort !== 'furigana_asc' && this.$store.state.actorList.filters.sort !== 'furigana_desc') {
          this.$store.state.actorList.filters.sort = 'furigana_asc'
        }
      } else {
        if (this.$store.state.actorList.filters.sort !== 'name_asc' && this.$store.state.actorList.filters.sort !== 'name_desc') {
          this.$store.state.actorList.filters.sort = 'name_asc'
        }
      }
      this.jumpTo = grp.letters[0]
    },
    selectInitialLetter (grp, letter) {
      if (this.jumpTo === letter) {
        this.clearInitialJump()
        return
      }
      if (grp.kind === 'kana') {
        if (this.$store.state.actorList.filters.sort !== 'furigana_asc' && this.$store.state.actorList.filters.sort !== 'furigana_desc') {
          this.$store.state.actorList.filters.sort = 'furigana_asc'
        }
      } else {
        if (this.$store.state.actorList.filters.sort !== 'name_asc' && this.$store.state.actorList.filters.sort !== 'name_desc') {
          this.$store.state.actorList.filters.sort = 'name_asc'
        }
      }
      this.jumpTo = letter
    },
    clearInitialJump () {
      this.jumpTo = ''
    },
    // Custom Black（ひらがな選択用の関数追加）
    onOffsetChange(newOffset) {
      this.offset = newOffset;
      console.log('Offset changed:', newOffset);
    },
    // Custom End
    reloadList () {
      const targetRoute = this.$route && this.$route.name === 'actorsCollection' ? 'actorsCollection' : 'actors'
      this.$router.push({
        name: targetRoute,
        query: {
          q: this.$store.getters['actorList/filterQueryParams']
        }
      })
    },
    async pageChanged () {      
      // Custom Black（Offset設定処理を修正）
      // this.$store.state.actorList.offset = (this.current -1) * this.$store.state.actorList.limit
      this.$store.state.actorList.offset =  parseInt((this.current - 1) * this.$store.state.actorList.limit, 10)
      // Custom End
      this.$store.dispatch('actorList/load', { offset: this.$store.state.actorList.offset })
    },
    nextpage () {
      if (this.$store.state.overlay.actordetails.show){
        return 
      }
      if (this.$store.state.overlay.details.show){
        return 
      }
      if (this.current * this.limit >= this.total) {
        this.current = 1
      } else {
        this.current += 1
      }      
      this.pageChanged()
    },
    prevpage () {      
      if (this.$store.state.overlay.actordetails.show){
        return 
      }
      if (this.$store.state.overlay.details.show){
        return 
      }
      if (this.current > 1) {
        this.current -= 1
      } else {
        this.current = Math.floor(this.total / this.limit) + 1        
      }      
      this.pageChanged()
    },
  }
}
</script>

<style scoped>
  .list-header-label {
    padding-right: 1em;
  }

  .initial-search-card {
    background: #ffffff;
    border: 1px solid #e5e7eb;
    border-radius: 10px;
    box-shadow: 0 2px 10px rgba(15, 23, 42, 0.06);
    padding: 0.65rem 0.85rem;
    margin-bottom: 0.9rem;
  }

  .initial-group-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
  }

  .initial-sub-row {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.35rem;
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid #f1f5f9;
  }

  .initial-pill {
    appearance: none;
    border: none;
    border-radius: 999px;
    background: #e9edf2;
    color: #374151;
    font-size: 0.78rem;
    font-weight: 500;
    line-height: 1.2;
    padding: 0.34rem 0.8rem;
    cursor: pointer;
    transition: background-color 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;
  }

  .initial-pill:hover {
    background: #dde3eb;
    color: #111827;
  }

  .initial-pill.is-active {
    background: #2b4c73;
    color: #ffffff;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(43, 76, 115, 0.3);
  }

  .initial-sub-pill {
    padding: 0.26rem 0.72rem;
    font-size: 0.76rem;
  }

  .initial-clear-pill {
    background: #f3f4f6;
    color: #6b7280;
    font-size: 0.72rem;
    padding: 0.24rem 0.65rem;
    margin-left: auto;
  }

  .initial-clear-pill:hover {
    background: #fee2e2;
    color: #b91c1c;
  }

  /* Custom Black */
  .toggle-container {
    margin-right: 1.5rem;
    margin-top: 0.25rem;
  }
  .toggle-button {
    flex: 1 1 50%;
    padding: 5px 10px;
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
  .toggle-button input {
    display: none;
  }
  .toggle-button.active {
    background-color: #42b983;
    color: white;
    border-color: #42b983;
  }
  /* Custom END */
</style>
