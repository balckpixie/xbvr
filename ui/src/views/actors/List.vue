<template>
  <div :class="{ 'is-prime-list': isPrimeCollection }">
    <GlobalEvents
      :filter="e => !['INPUT', 'TEXTAREA'].includes(e.target.tagName)"
      @keydown.left="prevpage"
      @keydown.right="nextpage"
      @keydown.o="prevpage"
      @keydown.p="nextpage"
    />
    <b-loading :is-full-page="true" :active.sync="isLoading"></b-loading>

    <!-- Collectionタブ: 2レイアウト切替バー -->
    <div v-if="isPrimeCollection" class="prime-view-tabs-bar">
      <div class="prime-view-tabs">
        <button
          type="button"
          class="prime-view-tab-btn"
          :class="{ 'is-active': collectionView === 'home' }"
          @click="collectionView = 'home'"
        >
          <b-icon pack="mdi" icon="home-outline" size="is-small" />
          <span>ホーム（ピックアップ・ランキング）</span>
        </button>
        <button
          type="button"
          class="prime-view-tab-btn"
          :class="{ 'is-active': collectionView === 'results' }"
          @click="collectionView = 'results'"
        >
          <b-icon pack="mdi" icon="magnify" size="is-small" />
          <span>検索結果（{{ total }} 件）</span>
        </button>
      </div>
    </div>

    <div v-if="!isPrimeCollection" class="columns is-multiline is-full">
      <div class="column">
        <strong>{{total}} results</strong>
      </div>
      <!-- Custom Black（ひらがな選択時は非表示） -->
      <div class="column" v-if="!isPrimeCollection && !hideHiragana">
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
      <div v-if="!isPrimeCollection && hideHiragana">
        <CustomPagination
          :total="total"
          :limit="limit"
          :offset="offset"
          @change="onOffsetChange"
        />
      </div>
      <!-- Custom End -->
       
      <div class="column">
        <div v-if="!isPrimeCollection" class="is-pulled-right">
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

        <!-- Custom Black(顔表示/体表示のトグルボタン追加 - Collectionタブでは左サイドバーに配置) -->
        <div v-if="!isPrimeCollection" class="is-pulled-right">
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
        <!-- 頭文字検索 (Actorsタブでは上部に表示) -->
        <div v-if="!isPrimeCollection" class="initial-search-card">
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

    <!-- 1つ目: 初期表示 (ピックアップ / 6ランキング / 最近チェック) -->
    <PrimeShowcaseRows v-if="isPrimeCollection && collectionView === 'home'" :showFace="showFace" />

    <!-- 2つ目: 検索結果表示 (すべての検索結果 + 頭文字検索 + 無限スクロール) -->
    <template v-if="isPrimeCollection && collectionView === 'results'">
      <div class="prime-results-header">
        <span class="prime-results-title">すべての検索結果</span>
        <span class="prime-results-count">{{ actors.length }} / {{ total }} 件表示中</span>
      </div>

      <!-- 頭文字検索 (画像①レイアウト + 画像②デザイン / スクロール時上部ホバー) -->
      <div
        ref="initialSearchWrap"
        class="initial-search-sticky-wrap"
        :style="isInitialSticky ? { height: initialSearchHeight + 'px' } : {}"
      >
        <div
          ref="initialSearchCard"
          class="initial-search-card"
          :class="{ 'is-fixed-hover': isInitialSticky }"
          :style="isInitialSticky ? { left: initialStickyLeft + 'px', width: initialStickyWidth + 'px' } : {}"
        >
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
            <button
              v-if="activeInitialGroup || selectedInitialLetter"
              type="button"
              class="initial-pill initial-clear-pill"
              @click="clearInitialJump"
            >
              クリア
            </button>
          </div>
          <div v-if="activeInitialGroupObj" class="initial-sub-row">
            <button
              v-for="letter in activeInitialGroupObj.letters"
              :key="letter"
              type="button"
              class="initial-pill initial-sub-pill"
              :class="{ 'is-active': selectedInitialLetter === letter }"
              @click="selectInitialLetter(activeInitialGroupObj, letter)"
            >
              {{ letter }}
            </button>
          </div>
        </div>
      </div>

      <div ref="primeGridWrap" class="prime-virtual-grid">
        <div v-if="topSpacerHeight > 0" :style="{ height: topSpacerHeight + 'px' }"></div>

        <div
          v-for="row in visibleActorRows"
          :key="`vrow-${row.rowIndex}`"
          class="prime-virtual-row"
          :style="{ gridTemplateColumns: `repeat(${colsPerRow}, minmax(0, 1fr))` }"
        >
          <div v-for="actor in row.items" :key="actor.id" class="prime-virtual-cell">
            <ActorCard :actor="actor" :showFace="showFace" />
          </div>
        </div>

        <div v-if="bottomSpacerHeight > 0" :style="{ height: bottomSpacerHeight + 'px' }"></div>
      </div>

      <div ref="infiniteSentinel" class="prime-infinite-sentinel">
        <span v-if="isLoadingMore" class="prime-loading-more">
          <b-icon pack="mdi" icon="loading" custom-class="mdi-spin" size="is-small" />
          <span>読み込み中...</span>
        </span>
        <span v-else-if="actors.length >= total && total > 0" class="prime-end-text">
          全 {{ total }} 件を表示しました
        </span>
      </div>
    </template>

    <!-- 従来のActorsタブ一覧 -->
    <template v-else>
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
    </template>
  </div>
</template>

<script>
import ActorCard from './ActorCard'
import ky from 'ky'
import GlobalEvents from 'vue-global-events'
// Custom Black（ひらがな選択用のデータ追加）
import CustomPagination from '../../components/CustomPagination.vue';
import PrimeShowcaseRows from './PrimeShowcaseRows.vue';

export default {
  name: 'List',
  components: { ActorCard, GlobalEvents, CustomPagination, PrimeShowcaseRows },
  data () {
    return {      
      current: 0,      
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
      activeInitialGroup: '',
      selectedInitialLetter: '',
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
      ],
      colsPerRow: 6,
      scrollTop: 0,
      viewportHeight: 800,
      gridTopOffset: 0,
      rowHeightEstimate: 220,
      gridResizeObserver: null,
      infiniteObserver: null,
      scrollRafId: null,
      isInitialSticky: false,
      initialStickyLeft: 0,
      initialStickyWidth: 0,
      initialSearchHeight: 0
    }
  },
  mounted () {
    if (this.isPrimeCollection && this.collectionView === 'results') {
      this.setupCollectionObservers()
    }
  },
  beforeDestroy () {
    this.teardownCollectionObservers()
  },
  watch: {
    collectionView (val) {
      if (!this.isPrimeCollection) return
      if (val === 'results') {
        this.setupCollectionObservers()
      } else {
        this.teardownCollectionObservers()
      }
    },
    cardSize () {
      if (this.isPrimeCollection) {
        this.$nextTick(() => this.updateGridMetrics())
      }
    },
    actors () {
      if (this.isPrimeCollection) {
        this.$nextTick(() => this.updateGridMetrics())
      }
    },
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
    isPrimeCollection () {
      return this.$route && this.$route.name === 'actorsCollection'
    },
    collectionView: {
      get () {
        return this.$store.state.actorList.collectionView
      },
      set (value) {
        this.$store.commit('actorList/setCollectionView', value)
      }
    },
    showFace: {
      get () {
        return this.$store.state.actorList.showFace
      },
      set (value) {
        this.$store.commit('actorList/setShowFace', value)
      }
    },
    isLoadingMore () {
      return this.$store.state.actorList.isLoadingMore
    },
    actorRows () {
      const cols = Math.max(1, this.colsPerRow)
      const rows = []
      for (let i = 0; i < this.actors.length; i += cols) {
        rows.push({
          rowIndex: Math.floor(i / cols),
          items: this.actors.slice(i, i + cols)
        })
      }
      return rows
    },
    visibleRowRange () {
      const totalRows = this.actorRows.length
      if (totalRows === 0) return { start: 0, end: 0 }
      const rowH = Math.max(120, this.rowHeightEstimate)
      const relTop = Math.max(0, this.scrollTop - this.gridTopOffset)
      const start = Math.max(0, Math.floor(relTop / rowH) - 3)
      const end = Math.min(totalRows, Math.ceil((relTop + this.viewportHeight) / rowH) + 4)
      return { start, end }
    },
    visibleActorRows () {
      const { start, end } = this.visibleRowRange
      return this.actorRows.slice(start, end)
    },
    topSpacerHeight () {
      return this.visibleRowRange.start * this.rowHeightEstimate
    },
    bottomSpacerHeight () {
      const remaining = Math.max(0, this.actorRows.length - this.visibleRowRange.end)
      return remaining * this.rowHeightEstimate
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
    setupCollectionObservers () {
      this.$nextTick(() => {
        this.updateGridMetrics()
        window.addEventListener('scroll', this.onWindowScroll, { passive: true })
        window.addEventListener('resize', this.updateGridMetrics, { passive: true })

        if (typeof ResizeObserver !== 'undefined' && this.$refs.primeGridWrap) {
          this.gridResizeObserver = new ResizeObserver(() => {
            this.updateGridMetrics()
          })
          this.gridResizeObserver.observe(this.$refs.primeGridWrap)
        }

        if (typeof IntersectionObserver !== 'undefined' && this.$refs.infiniteSentinel) {
          this.infiniteObserver = new IntersectionObserver((entries) => {
            if (entries.some(e => e.isIntersecting)) {
              this.loadMoreActors()
            }
          }, { rootMargin: '600px 0px' })
          this.infiniteObserver.observe(this.$refs.infiniteSentinel)
        }
        this.checkSentinelAndLoad()
      })
    },
    teardownCollectionObservers () {
      window.removeEventListener('scroll', this.onWindowScroll)
      window.removeEventListener('resize', this.updateGridMetrics)
      this.isInitialSticky = false
      if (this.scrollRafId) {
        cancelAnimationFrame(this.scrollRafId)
        this.scrollRafId = null
      }
      if (this.gridResizeObserver) {
        this.gridResizeObserver.disconnect()
        this.gridResizeObserver = null
      }
      if (this.infiniteObserver) {
        this.infiniteObserver.disconnect()
        this.infiniteObserver = null
      }
    },
    onWindowScroll () {
      if (this.scrollRafId) return
      this.scrollRafId = requestAnimationFrame(() => {
        this.scrollRafId = null
        this.scrollTop = window.scrollY || document.documentElement.scrollTop || 0
        this.viewportHeight = window.innerHeight || 800
        if (this.$refs.primeGridWrap) {
          const rect = this.$refs.primeGridWrap.getBoundingClientRect()
          this.gridTopOffset = rect.top + this.scrollTop
        }
        this.updateInitialStickyState()
        this.checkSentinelAndLoad()
      })
    },
    updateInitialStickyState () {
      const wrap = this.$refs.initialSearchWrap
      const card = this.$refs.initialSearchCard
      if (!wrap || !card) {
        this.isInitialSticky = false
        return
      }
      const wrapRect = wrap.getBoundingClientRect()
      if (wrapRect.top < 12) {
        this.initialStickyLeft = wrapRect.left
        this.initialStickyWidth = wrapRect.width
        this.initialSearchHeight = card.offsetHeight
        this.isInitialSticky = true
      } else {
        this.isInitialSticky = false
      }
    },
    checkSentinelAndLoad () {
      const sentinel = this.$refs.infiniteSentinel
      if (!sentinel) return
      const rect = sentinel.getBoundingClientRect()
      if (rect.top <= (window.innerHeight || 800) + 600) {
        this.loadMoreActors()
      }
    },
    updateGridMetrics () {
      this.viewportHeight = window.innerHeight || 800
      this.scrollTop = window.scrollY || document.documentElement.scrollTop || 0
      this.updateInitialStickyState()
      const wrap = this.$refs.primeGridWrap
      if (!wrap) return
      const rect = wrap.getBoundingClientRect()
      this.gridTopOffset = rect.top + this.scrollTop
      const w = wrap.clientWidth || 1000
      const targetMinWidthMap = { '1': 125, '2': 165, '3': 210, '4': 260 }
      const minCardW = targetMinWidthMap[this.cardSize] || 165
      const gap = 14
      const cols = Math.max(2, Math.floor((w + gap) / (minCardW + gap)))
      this.colsPerRow = cols
      const cardW = (w - gap * (cols - 1)) / cols
      const firstRowEl = wrap.querySelector('.prime-virtual-row')
      if (firstRowEl && firstRowEl.offsetHeight > 40) {
        this.rowHeightEstimate = firstRowEl.offsetHeight + gap
      } else {
        this.rowHeightEstimate = Math.round(cardW + gap)
      }
    },
    async loadMoreActors () {
      if (!this.isPrimeCollection || this.collectionView !== 'results') return
      if (this.isLoading || this.isLoadingMore) return
      if (this.actors.length >= this.total) return
      await this.$store.dispatch('actorList/load', {
        offset: this.actors.length,
        limit: 36,
        append: true
      })
      this.$nextTick(() => {
        this.updateGridMetrics()
        this.checkSentinelAndLoad()
      })
    },
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
      this.selectedInitialLetter = grp.letters[0]
      this.jumpTo = grp.letters[0]
    },
    selectInitialLetter (grp, letter) {
      if (this.selectedInitialLetter === letter) {
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
      this.selectedInitialLetter = letter
      this.jumpTo = letter
    },
    clearInitialJump () {
      this.activeInitialGroup = ''
      this.selectedInitialLetter = ''
      this.$store.state.actorList.filters.jumpTo = ''
      this.reloadList()
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

  .initial-search-sticky-wrap {
    margin-bottom: 0.9rem;
  }

  .is-prime-list .initial-search-card {
    background: #16202a;
    border-color: #253342;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
    margin-bottom: 0;
  }

  .is-prime-list .initial-search-card.is-fixed-hover {
    position: fixed;
    top: 12px;
    z-index: 35;
    background: rgba(22, 32, 42, 0.95);
    backdrop-filter: blur(8px);
    border-color: #2b3f54;
    box-shadow: 0 10px 26px rgba(0, 0, 0, 0.58);
  }

  .is-prime-list .initial-sub-row {
    border-top-color: #22303e;
  }

  .is-prime-list .initial-pill {
    background: #233140;
    color: #cbd5e1;
  }

  .is-prime-list .initial-pill:hover {
    background: #2d4054;
    color: #ffffff;
  }

  .is-prime-list .initial-pill.is-active {
    background: #00a8e1;
    color: #ffffff;
    box-shadow: 0 2px 8px rgba(0, 168, 225, 0.35);
  }

  .is-prime-list .initial-clear-pill {
    background: #1b2836;
    color: #94a3b8;
  }

  .is-prime-list .initial-clear-pill:hover {
    background: rgba(244, 63, 94, 0.22);
    color: #fda4af;
  }

  .is-prime-list ::v-deep .pagination-link,
  .is-prime-list ::v-deep .pagination-previous,
  .is-prime-list ::v-deep .pagination-next {
    background-color: #16202a;
    border-color: #2b3b4c;
    color: #cbd5e1;
  }

  .is-prime-list ::v-deep .pagination-link.is-current {
    background-color: #00a8e1;
    border-color: #00a8e1;
    color: #ffffff;
  }

  .is-prime-list .toggle-button {
    background-color: #16202a;
    border-color: #2b3b4c;
    color: #cbd5e1;
  }

  .is-prime-list .toggle-button.active {
    background-color: #00a8e1;
    border-color: #00a8e1;
    color: #ffffff;
  }

  .is-prime-list ::v-deep .b-radio.button {
    background-color: #16202a;
    border-color: #2b3b4c;
    color: #cbd5e1;
  }

  .is-prime-list ::v-deep .b-radio.button.is-selected,
  .is-prime-list ::v-deep .b-radio.button.is-primary {
    background-color: #00a8e1;
    border-color: #00a8e1;
    color: #ffffff;
  }

  .prime-view-tabs-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  .prime-view-tabs {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.28rem;
    background: #16202a;
    border: 1px solid #253342;
    border-radius: 10px;
  }

  .prime-view-tab-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    padding: 0.42rem 0.9rem;
    border: none;
    border-radius: 7px;
    background: transparent;
    color: #94a3b8;
    font-size: 0.8rem;
    font-weight: 700;
    cursor: pointer;
    transition: all 0.15s ease;
  }

  .prime-view-tab-btn:hover {
    color: #ffffff;
    background: rgba(255, 255, 255, 0.06);
  }

  .prime-view-tab-btn.is-active {
    background: #00a8e1;
    color: #ffffff;
    box-shadow: 0 2px 8px rgba(0, 168, 225, 0.35);
  }

  .prime-results-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 0.6rem;
    padding: 0 0.15rem;
  }

  .prime-search-controls-row {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
    margin-bottom: 0.9rem;
    flex-wrap: wrap;
  }

  .prime-search-card-flex {
    flex: 1;
    min-width: 0;
    margin-bottom: 0;
  }

  .prime-card-size-box {
    display: flex;
    align-items: center;
    gap: 0.45rem;
    background: #16202a;
    border: 1px solid #253342;
    border-radius: 10px;
    padding: 0.52rem 0.7rem;
    box-shadow: 0 4px 14px rgba(0, 0, 0, 0.35);
    flex-shrink: 0;
  }

  .prime-card-size-label {
    font-size: 0.74rem;
    font-weight: 700;
    color: #94a3b8;
    white-space: nowrap;
  }

  .prime-results-title {
    font-size: 1rem;
    font-weight: 700;
    color: #ffffff;
  }

  .prime-results-count {
    font-size: 0.76rem;
    color: #8197a4;
    font-variant-numeric: tabular-nums;
  }

  .prime-virtual-grid {
    display: flex;
    flex-direction: column;
  }

  .prime-virtual-row {
    display: grid;
    gap: 14px;
    margin-bottom: 14px;
  }

  .prime-virtual-cell {
    min-width: 0;
  }

  .prime-infinite-sentinel {
    min-height: 44px;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 0.75rem 0 1.25rem;
  }

  .prime-loading-more {
    display: inline-flex;
    align-items: center;
    gap: 0.45rem;
    font-size: 0.82rem;
    color: #00a8e1;
    font-weight: 600;
  }

  .prime-end-text {
    font-size: 0.76rem;
    color: #64748b;
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
