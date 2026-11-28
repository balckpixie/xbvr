<template>
  <aside class="amz-filters">
    <!-- ヘッダ & クリア -->
    <div class="amz-header">
      <div class="amz-header-title">
        <b-icon pack="mdi" icon="filter-variant" size="is-small" />
        <span>絞り込み</span>
        <span v-if="activeFilterCount > 0" class="amz-count-badge">{{ activeFilterCount }}</span>
      </div>
      <button
        v-if="activeFilterCount > 0"
        type="button"
        class="amz-clear-btn"
        @click="clearAllFilters"
      >
        すべてクリア
      </button>
    </div>

    <!-- 適用中のフィルタチップ -->
    <div v-if="activeChips.length > 0" class="amz-active-chips">
      <button
        v-for="chip in activeChips"
        :key="chip.key"
        type="button"
        class="amz-active-chip"
        @click="chip.clear()"
        title="クリックで解除"
      >
        <span>{{ chip.label }}</span>
        <span class="amz-chip-x">×</span>
      </button>
    </div>

    <!-- 並び替え (Sort) -->
    <div class="amz-section">
      <div class="amz-section-head is-static">
        <span class="amz-section-title">並び替え</span>
      </div>
      <div class="amz-section-body">
        <div class="select is-small is-fullwidth amz-select">
          <select v-model="sort">
            <option value="name_asc">↑ {{ $t("Name") }}</option>
            <option value="name_desc">↓ {{ $t("Name") }}</option>
            <option value="furigana_asc">↑ {{ $t("Furigana") }}</option>
            <option value="furigana_desc">↓ {{ $t("Furigana") }}</option>
            <option value="birthday_desc">↓ {{ $t("Birthdate") }}</option>
            <option value="birthday_asc">↑ {{ $t("Birthdate") }}</option>
            <option value="rating_desc">↓ {{ $t("Rating") }}</option>
            <option value="rating_asc">↑ {{ $t("Rating") }}</option>
            <option value="scene_rating_desc">↓ {{ $t("Average Scene Rating") }}</option>
            <option value="added_desc">↓ {{ $t("Added date") }}</option>
            <option value="added_asc">↑ {{ $t("Added date") }}</option>
            <option value="modified_desc">↓ {{ $t("Modified date") }}</option>
            <option value="modified_asc">↑ {{ $t("Modified date") }}</option>
            <option value="scene_release_desc">↓ {{ $t("Scene Released Date") }}</option>
            <option value="scene_added_desc">↓ {{ $t("Scene Added Date") }}</option>
            <option value="file_added_desc">↓ {{ $t("File Added Date") }}</option>
            <option value="scene_available_desc">↓ {{ $t("Available Scene Count") }}</option>
            <option value="scene_count_desc">↓ {{ $t("Scene Count") }}</option>
            <option value="random">↯ {{ $t("Random") }}</option>
          </select>
        </div>
      </div>
    </div>

    <!-- 女優タグ (Genre Tags) -->
    <div class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('actorTags')">
        <span class="amz-section-title">
          女優タグ
          <span v-if="tags.length > 0" class="amz-inline-val">{{ tags.length }}</span>
        </span>
        <span class="amz-chevron">{{ sections.actorTags ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.actorTags" class="amz-section-body">
        <div class="amz-genre-pills">
          <button
            v-for="tag in actorGenreOptions"
            :key="tag"
            type="button"
            class="amz-genre-pill"
            :class="{ 'is-selected': tags.includes(tag) }"
            @click="toggleActorTag(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>
    </div>

    <!-- シーン有無 (ダイレクト選択) -->
    <div class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('scenePresence')">
        <span class="amz-section-title">
          シーン有無
          <span v-if="scenePresenceMode !== 'all'" class="amz-inline-val">
            {{ scenePresenceMode === 'with' ? 'あり' : 'なし' }}
          </span>
        </span>
        <span class="amz-chevron">{{ sections.scenePresence ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.scenePresence" class="amz-section-body">
        <div class="amz-seg-row">
          <button
            type="button"
            class="amz-seg-btn"
            :class="{ 'is-selected': scenePresenceMode === 'all' }"
            @click="setScenePresence('all')"
          >
            すべて
          </button>
          <button
            type="button"
            class="amz-seg-btn"
            :class="{ 'is-selected': scenePresenceMode === 'with' }"
            @click="setScenePresence('with')"
          >
            あり
          </button>
          <button
            type="button"
            class="amz-seg-btn"
            :class="{ 'is-selected': scenePresenceMode === 'without' }"
            @click="setScenePresence('without')"
          >
            なし
          </button>
        </div>
        <label class="amz-check-row mt-1">
          <input type="checkbox" v-model="onlyAvailableScenes" />
          <span>視聴可能シーンありのみ</span>
        </label>
      </div>
    </div>

    <!-- リスト (Watchlist / Favourite) -->
    <div class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('lists')">
        <span class="amz-section-title">リスト</span>
        <span class="amz-chevron">{{ sections.lists ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.lists" class="amz-section-body">
        <label class="amz-check-row">
          <input type="checkbox" value="watchlist" v-model="lists" />
          <b-icon pack="mdi" icon="calendar-check" size="is-small" class="amz-icon-primary" />
          <span>{{ $t('Watchlist') }}</span>
        </label>
        <label class="amz-check-row">
          <input type="checkbox" value="favourite" v-model="lists" />
          <b-icon pack="mdi" icon="heart" size="is-small" class="amz-icon-danger" />
          <span>{{ $t('Favourite') }}</span>
        </label>
      </div>
    </div>

    <!-- 女優評価 (Amazon風スター) -->
    <div class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('rating')">
        <span class="amz-section-title">
          女優評価
          <span v-if="ratings[0] > 0 || ratings[1] < 5" class="amz-inline-val">{{ ratings[0] }}–{{ ratings[1] }}</span>
        </span>
        <span class="amz-chevron">{{ sections.rating ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.rating" class="amz-section-body">
        <button
          v-for="star in [4, 3, 2, 1]"
          :key="`actor-star-${star}`"
          type="button"
          class="amz-star-row"
          :class="{ 'is-selected': ratings[0] === star && ratings[1] === 5 }"
          @click="selectMinRating('ratings', star)"
        >
          <span class="amz-stars">
            <span v-for="n in 5" :key="n" :class="n <= star ? 'is-filled' : 'is-empty'">★</span>
          </span>
          <span class="amz-star-text">以上</span>
        </button>
        <div class="amz-mini-slider">
          <b-slider :min="0" :max="5" :step="0.5" :tooltip="true" size="is-small" v-model="ratings" lazy />
        </div>
      </div>
    </div>

    <!-- 平均作品評価 (Amazon風スター) -->
    <div class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('sceneRating')">
        <span class="amz-section-title">
          作品平均評価
          <span v-if="sceneratings[0] > 0 || sceneratings[1] < 5" class="amz-inline-val">{{ sceneratings[0] }}–{{ sceneratings[1] }}</span>
        </span>
        <span class="amz-chevron">{{ sections.sceneRating ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.sceneRating" class="amz-section-body">
        <button
          v-for="star in [4, 3, 2, 1]"
          :key="`scene-star-${star}`"
          type="button"
          class="amz-star-row"
          :class="{ 'is-selected': sceneratings[0] === star && sceneratings[1] === 5 }"
          @click="selectMinRating('sceneratings', star)"
        >
          <span class="amz-stars">
            <span v-for="n in 5" :key="n" :class="n <= star ? 'is-filled' : 'is-empty'">★</span>
          </span>
          <span class="amz-star-text">以上</span>
        </button>
        <div class="amz-mini-slider">
          <b-slider :min="0" :max="5" :step="0.25" :tooltip="true" size="is-small" v-model="sceneratings" lazy />
        </div>
      </div>
    </div>

    <!-- 女優・サイト・属性フィルタ -->
    <div v-if="Object.keys(filters).length !== 0" class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('tags')">
        <span class="amz-section-title">
          女優・サイト・属性
          <span v-if="cast.length + sites.length + attributes.length > 0" class="amz-inline-val">
            {{ cast.length + sites.length + attributes.length }}
          </span>
        </span>
        <span class="amz-chevron">{{ sections.tags ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.tags" class="amz-section-body amz-tag-fields">
        <div class="amz-field-label">{{ $t('Cast') }}</div>
        <b-taginput
          v-model="cast"
          size="is-small"
          autocomplete
          :data="filteredCast"
          placeholder="女優名を入力..."
          @typing="getFilteredCast"
        >
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">{{ $t("No matching cast") }}</template>
          <template #selected="props">
            <b-tag
              v-for="(tag, index) in props.tags"
              :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
              :key="tag+index"
              size="is-small"
              :tabstop="false"
              closable
              @close="cast=cast.filter(e => e !== tag)"
              @click="toggle2Way(tag,index,'cast')"
            >
              <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon" />
              <b-icon pack="mdi" v-if="tag.charAt(0)=='&'" icon="plus-circle-outline" size="is-small" class="tagicon" />
              {{ removeConditionPrefix(tag) }}
            </b-tag>
          </template>
        </b-taginput>

        <div class="amz-field-label mt-2">{{ $t('Site') }}</div>
        <b-taginput
          v-model="sites"
          size="is-small"
          autocomplete
          :data="filteredSites"
          placeholder="サイト名を入力..."
          @typing="getFilteredSites"
        >
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">{{ $t("No matching sites") }}</template>
          <template #selected="props">
            <b-tag
              v-for="(tag, index) in props.tags"
              :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
              :key="tag+index"
              size="is-small"
              :tabstop="false"
              closable
              @close="sites=sites.filter(e => e !== tag)"
              @click="toggle3way(tag,index,'sites')"
            >
              <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon" />
              {{ removeConditionPrefix(tag) }}
            </b-tag>
          </template>
        </b-taginput>

        <div class="amz-field-label mt-2">{{ $t('Attributes') }}</div>
        <b-taginput
          v-model="attributes"
          size="is-small"
          autocomplete
          :data="filteredAttributes"
          placeholder="属性を入力..."
          @typing="getFilteredAttributes"
        >
          <template slot-scope="props">{{ props.option }}</template>
          <template slot="empty">{{ $t("No matching attributes") }}</template>
          <template #selected="props">
            <b-tag
              v-for="(tag, index) in props.tags"
              :type="tag.charAt(0)=='!' ? 'is-danger': (tag.charAt(0)=='&' ? 'is-success' : '')"
              :key="tag+index"
              size="is-small"
              :tabstop="false"
              closable
              @close="attributes=attributes.filter(e => e !== tag)"
              @click="toggle3way(tag,index,'attributes')"
            >
              <b-icon pack="mdi" v-if="tag.charAt(0)=='!'" icon="minus-circle-outline" size="is-small" class="tagicon" />
              <b-icon pack="mdi" v-if="tag.charAt(0)=='&'" icon="plus-circle-outline" size="is-small" class="tagicon" />
              {{ removeConditionPrefix(tag) }}
            </b-tag>
          </template>
        </b-taginput>
      </div>
    </div>

    <!-- 作品数・保有数 -->
    <div v-if="Object.keys(filters).length !== 0" class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('counts')">
        <span class="amz-section-title">作品数・保有数</span>
        <span class="amz-chevron">{{ sections.counts ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.counts" class="amz-section-body">
        <div class="amz-range-row">
          <div class="amz-range-head">
            <span>{{ $t("Scenes") }}</span>
            <span class="amz-range-val">{{ scenecounts[0] }} – {{ scenecounts[1] }}</span>
          </div>
          <b-slider :min="0" :max="150" :step="1" :tooltip="true" size="is-small" v-model="scenecounts" lazy />
        </div>
        <div class="amz-range-row">
          <div class="amz-range-head">
            <span>{{ $t("Available") }}</span>
            <span class="amz-range-val">{{ avails[0] }} – {{ avails[1] }}</span>
          </div>
          <b-slider :min="0" :max="150" :step="1" :tooltip="true" size="is-small" v-model="avails" lazy />
        </div>
      </div>
    </div>

    <!-- プロフィール範囲 (年齢・身長・体重) -->
    <div v-if="Object.keys(filters).length !== 0" class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('profile')">
        <span class="amz-section-title">年齢・身長・体重</span>
        <span class="amz-chevron">{{ sections.profile ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.profile" class="amz-section-body">
        <div class="amz-range-row">
          <div class="amz-range-head">
            <span>{{ $t("Age") }}</span>
            <span class="amz-range-val">{{ ages[0] }} – {{ ages[1] }}</span>
          </div>
          <b-slider :min="18" :max="100" :step="1" :tooltip="true" size="is-small" v-model="ages" lazy />
        </div>
        <div class="amz-range-row">
          <div class="amz-range-head">
            <span>{{ $t("Height") }}</span>
            <span class="amz-range-val">{{ heights[0] }} – {{ heights[1] }}</span>
          </div>
          <b-slider :min="120" :max="220" :step="1" :tooltip="true" size="is-small" v-model="heights" lazy />
        </div>
        <div class="amz-range-row">
          <div class="amz-range-head">
            <span>{{ $t("Weight") }}</span>
            <span class="amz-range-val">{{ weights[0] }} – {{ weights[1] }}</span>
          </div>
          <b-slider :min="25" :max="150" :step="1" :tooltip="true" size="is-small" v-model="weights" lazy />
        </div>
      </div>
    </div>

    <!-- 保存済み検索 -->
    <div class="amz-section">
      <button type="button" class="amz-section-head" @click="toggleSection('saved')">
        <span class="amz-section-title">保存済み検索</span>
        <span class="amz-chevron">{{ sections.saved ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.saved" class="amz-section-body">
        <SavedSearch />
      </div>
    </div>

    <!-- AKAグループ -->
    <div class="amz-section amz-section-last">
      <button type="button" class="amz-section-head" @click="toggleSection('aka')">
        <span class="amz-section-title">AKA グループ</span>
        <span class="amz-chevron">{{ sections.aka ? '▾' : '▸' }}</span>
      </button>
      <div v-show="sections.aka" class="amz-section-body">
        <b-field>
          <b-tooltip position="is-right" :label="$t('New Aka Group. Select 2 or more actors in the Cast filter')" multilined :delay="200">
            <button class="button is-small is-outlined" @click="createAkaGroup" :disabled="disableNewAkaGroup">
              <b-icon pack="mdi" icon="account-multiple-plus-outline" size="is-small" />
            </button>
          </b-tooltip>
          <b-tooltip position="is-right" :label="$t('Select the Aka Group to delete in the Cast Filter')" multilined :delay="200">
            <button class="button is-small is-outlined" @click="deleteAkaGroup" :disabled="disableDeleteAkaGroup">
              <b-icon pack="mdi" icon="delete-outline" size="is-small" />
            </button>
          </b-tooltip>
          <b-tooltip position="is-bottom" :label="$t('Add Cast to Aka Group. Select the Aka group and Actors to add in the Cast Filter')" multilined :delay="200">
            <button class="button is-small is-outlined" @click="addToAkaGroup" :disabled="disableAddToAkaGroup">
              <b-icon pack="mdi" icon="account-plus-outline" size="is-small" />
            </button>
          </b-tooltip>
          <b-tooltip position="is-bottom" :label="$t('Remove Cast from Aka Group. Select the Aka group and Actors to remove in the Cast Filter')" multilined :delay="200">
            <button class="button is-small is-outlined" @click="removeFromAkaGroup" :disabled="disableRemoveFromAkaGroup">
              <b-icon pack="mdi" icon="account-minus-outline" size="is-small" />
            </button>
          </b-tooltip>
        </b-field>
      </div>
    </div>
  </aside>
</template>

<script>
import SavedSearch from './SavedSearch'
import ky from 'ky'

export default {
  name: 'CollectionFilters',
  components: { SavedSearch },
  mounted () {
    this.$store.dispatch('actorList/filters')
    this.fetchFilters()
  },
  data () {
    return {
      filteredCast: [],
      filteredSites: [],
      filteredAttributes: [],
      actorGenreOptions: ['美少女', 'お姉さん', '癒し', 'グラマー', '爆乳', '人妻', 'ハーフ', '欧米', 'エロ'],
      sections: {
        actorTags: true,
        scenePresence: true,
        lists: true,
        rating: true,
        sceneRating: true,
        tags: true,
        counts: false,
        profile: false,
        saved: false,
        aka: false
      }
    }
  },
  methods: {
    toggleSection (key) {
      this.sections[key] = !this.sections[key]
    },
    selectMinRating (prop, star) {
      const current = this[prop]
      if (current[0] === star && current[1] === 5) {
        this[prop] = [0, 5]
      } else {
        this[prop] = [star, 5]
      }
    },
    setScenePresence (mode) {
      if (mode === 'with') {
        this.scenecounts = [1, 150]
      } else if (mode === 'without') {
        this.scenecounts = [0, 0]
      } else {
        this.scenecounts = [0, 150]
      }
    },
    toggleActorTag (tag) {
      if (this.tags.includes(tag)) {
        this.tags = this.tags.filter(t => t !== tag)
      } else {
        this.tags = [...this.tags, tag]
      }
    },
    clearAllFilters () {
      const f = this.$store.state.actorList.filters
      f.lists = []
      f.tags = []
      f.cast = []
      f.sites = []
      f.attributes = []
      f.jumpTo = ''
      f.min_age = 0
      f.max_age = 100
      f.min_height = 120
      f.max_height = 220
      f.min_weight = 25
      f.max_weight = 150
      f.min_count = 0
      f.max_count = 150
      f.min_avail = 0
      f.max_avail = 150
      f.min_rating = 0
      f.max_rating = 5
      f.min_scene_rating = 0
      f.max_scene_rating = 5
      this.reloadList(true)
    },
    reloadList (keepView = false) {
      if (!keepView) {
        this.$store.commit('actorList/setCollectionView', 'results')
      }
      const targetRoute = this.$route && this.$route.name === 'actorsCollection' ? 'actorsCollection' : 'actors'
      this.$router.push({
        name: targetRoute,
        query: {
          q: this.$store.getters['actorList/filterQueryParams']
        }
      })
    },
    getFilteredCast (text) {
      const castList = this.filters.cast || []
      const startWithText = castList.filter(option => (
        (option.toString().toLowerCase().indexOf(text.toLowerCase()) === 0 || option.toString().toLowerCase().indexOf('aka:' + text.toLowerCase()) === 0) &&
        !this.cast.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
      const containsText = castList.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) > 0 &&
        !this.cast.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
      this.filteredCast = startWithText.concat(containsText)
    },
    getFilteredSites (text) {
      const siteList = this.filters.sites || []
      this.filteredSites = siteList.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.sites.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
    },
    getFilteredAttributes (text) {
      const attrList = this.filters.attributes || []
      this.filteredAttributes = attrList.filter(option => (
        option.toString().toLowerCase().indexOf(text.toLowerCase()) >= 0 &&
        !this.attributes.some(entry => this.removeConditionPrefix(entry.toString()) === option.toString())
      ))
    },
    createAkaGroup () {
      this.$store.state.actorList.isLoading = true
      ky.post('/api/aka/create', { json: { actorList: this.cast } }).json().then(data => {
        this.cast.push(data.akas.aka_actor.name)
        this.$store.dispatch('actorList/filters')
        this.reloadList()
        if (data.status !== '') {
          this.$buefy.toast.open({ message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000 })
        }
        this.$store.state.actorList.isLoading = false
      })
    },
    deleteAkaGroup () {
      this.$store.state.actorList.isLoading = true
      ky.post('/api/aka/delete', { json: { name: this.cast[0] } }).json().then(() => {
        this.cast = []
        this.$store.dispatch('actorList/filters')
        this.reloadList()
        this.$store.state.actorList.isLoading = false
      })
    },
    addToAkaGroup () {
      this.$store.state.actorList.isLoading = true
      ky.post('/api/aka/add', { json: { actorList: this.cast } }).json().then(data => {
        this.cast = this.cast.filter(e => !e.startsWith('aka:'))
        this.cast.push(data.akas.aka_actor.name)
        this.$store.dispatch('actorList/filters')
        this.reloadList()
        if (data.status !== '') {
          this.$buefy.toast.open({ message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000 })
        }
        this.$store.state.actorList.isLoading = false
      })
    },
    removeFromAkaGroup () {
      this.$store.state.actorList.isLoading = true
      ky.post('/api/aka/remove', { json: { actorList: this.cast } }).json().then(data => {
        this.cast = this.cast.filter(e => !e.startsWith('aka:'))
        this.cast.push(data.akas.aka_actor.name)
        this.$store.dispatch('actorList/filters')
        this.reloadList()
        if (data.status !== '') {
          this.$buefy.toast.open({ message: `Warning:  ${data.status}`, type: 'is-warning', duration: 5000 })
        }
        this.$store.state.actorList.isLoading = false
      })
    },
    toggle3way (text, idx, list) {
      let tags = []
      switch (list) {
        case 'cast':
          tags = this.cast
          break
        case 'sites':
          tags = this.sites
          break
        case 'attributes':
          tags = this.attributes
          break
      }
      switch (tags[idx].charAt(0)) {
        case '!':
          tags[idx] = this.removeConditionPrefix(tags[idx])
          break
        case '&':
          tags[idx] = '!' + this.removeConditionPrefix(tags[idx])
          break
        default:
          tags[idx] = '&' + text
      }
      this[list] = tags
    },
    toggle2Way (text, idx, list) {
      const tags = list === 'sites' ? this.sites : this.cast
      if (tags[idx].charAt(0) === '!') {
        tags[idx] = this.removeConditionPrefix(tags[idx])
      } else {
        tags[idx] = '!' + text
      }
      this[list] = tags
    },
    removeConditionPrefix (txt) {
      if (txt.charAt(0) === '!' || txt.charAt(0) === '&') {
        return txt.substring(1)
      }
      return txt
    },
    async fetchFilters () {
      this.filteredAttributes = ['Loading attributes']
      ky.get('/api/actor/filters').json().then(data => {
        this.filteredAttributes = data.attributes
      })
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
    filters () {
      return this.$store.state.actorList.filterOpts
    },
    activeChips () {
      const chips = []
      this.tags.forEach(tag => {
        chips.push({
          key: `genre-${tag}`,
          label: `タグ: ${tag}`,
          clear: () => {
            this.tags = this.tags.filter(t => t !== tag)
          }
        })
      })
      this.lists.forEach(item => {
        chips.push({
          key: `list-${item}`,
          label: item === 'watchlist' ? 'Watchlist' : 'Favourite',
          clear: () => {
            this.lists = this.lists.filter(v => v !== item)
          }
        })
      })
      if (this.ratings[0] > 0 || this.ratings[1] < 5) {
        chips.push({
          key: 'rating',
          label: `女優評価 ${this.ratings[0]}–${this.ratings[1]}`,
          clear: () => { this.ratings = [0, 5] }
        })
      }
      if (this.sceneratings[0] > 0 || this.sceneratings[1] < 5) {
        chips.push({
          key: 'scene-rating',
          label: `作品評価 ${this.sceneratings[0]}–${this.sceneratings[1]}`,
          clear: () => { this.sceneratings = [0, 5] }
        })
      }
      this.cast.forEach(tag => {
        chips.push({
          key: `cast-${tag}`,
          label: `Cast: ${tag}`,
          clear: () => { this.cast = this.cast.filter(e => e !== tag) }
        })
      })
      this.sites.forEach(tag => {
        chips.push({
          key: `site-${tag}`,
          label: `Site: ${tag}`,
          clear: () => { this.sites = this.sites.filter(e => e !== tag) }
        })
      })
      this.attributes.forEach(tag => {
        chips.push({
          key: `attr-${tag}`,
          label: `Attr: ${tag}`,
          clear: () => { this.attributes = this.attributes.filter(e => e !== tag) }
        })
      })
      if (this.scenecounts[0] > 0 || this.scenecounts[1] < 150) {
        chips.push({
          key: 'scenes',
          label: `作品数 ${this.scenecounts[0]}–${this.scenecounts[1]}`,
          clear: () => { this.scenecounts = [0, 150] }
        })
      }
      if (this.avails[0] > 0 || this.avails[1] < 150) {
        chips.push({
          key: 'avail',
          label: `保有数 ${this.avails[0]}–${this.avails[1]}`,
          clear: () => { this.avails = [0, 150] }
        })
      }
      if (this.ages[0] > 18 || this.ages[1] < 100) {
        chips.push({
          key: 'age',
          label: `年齢 ${this.ages[0]}–${this.ages[1]}`,
          clear: () => { this.ages = [0, 100] }
        })
      }
      if (this.heights[0] > 120 || this.heights[1] < 220) {
        chips.push({
          key: 'height',
          label: `身長 ${this.heights[0]}–${this.heights[1]}`,
          clear: () => { this.heights = [120, 220] }
        })
      }
      if (this.weights[0] > 25 || this.weights[1] < 150) {
        chips.push({
          key: 'weight',
          label: `体重 ${this.weights[0]}–${this.weights[1]}`,
          clear: () => { this.weights = [25, 150] }
        })
      }
      return chips
    },
    activeFilterCount () {
      return this.activeChips.length
    },
    scenePresenceMode () {
      const [min, max] = this.scenecounts
      if (min >= 1 && max === 150) return 'with'
      if (min === 0 && max === 0) return 'without'
      if (min === 0 && max === 150) return 'all'
      return 'custom'
    },
    onlyAvailableScenes: {
      get () {
        return this.avails[0] >= 1
      },
      set (checked) {
        this.avails = checked ? [1, 150] : [0, 150]
      }
    },
    tags: {
      get () {
        return this.$store.state.actorList.filters.tags || []
      },
      set (value) {
        this.$store.state.actorList.filters.tags = value
        this.reloadList(true)
      }
    },
    lists: {
      get () {
        return this.$store.state.actorList.filters.lists
      },
      set (value) {
        this.$store.state.actorList.filters.lists = value
        this.reloadList()
      }
    },
    cast: {
      get () {
        return this.$store.state.actorList.filters.cast
      },
      set (value) {
        this.$store.state.actorList.filters.cast = value
        this.reloadList()
      }
    },
    sites: {
      get () {
        return this.$store.state.actorList.filters.sites
      },
      set (value) {
        this.$store.state.actorList.filters.sites = value
        this.reloadList()
      }
    },
    ages: {
      get () {
        return [this.$store.state.actorList.filters.min_age, this.$store.state.actorList.filters.max_age]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_age !== value[0] || this.$store.state.actorList.filters.max_age !== value[1]) {
          this.$store.state.actorList.filters.min_age = value[0]
          this.$store.state.actorList.filters.max_age = value[1]
          this.reloadList()
        }
      }
    },
    heights: {
      get () {
        return [this.$store.state.actorList.filters.min_height, this.$store.state.actorList.filters.max_height]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_height !== value[0] || this.$store.state.actorList.filters.max_height !== value[1]) {
          this.$store.state.actorList.filters.min_height = value[0]
          this.$store.state.actorList.filters.max_height = value[1]
          this.reloadList()
        }
      }
    },
    weights: {
      get () {
        return [this.$store.state.actorList.filters.min_weight, this.$store.state.actorList.filters.max_weight]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_weight !== value[0] || this.$store.state.actorList.filters.max_weight !== value[1]) {
          this.$store.state.actorList.filters.min_weight = value[0]
          this.$store.state.actorList.filters.max_weight = value[1]
          this.reloadList()
        }
      }
    },
    scenecounts: {
      get () {
        return [this.$store.state.actorList.filters.min_count, this.$store.state.actorList.filters.max_count]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_count !== value[0] || this.$store.state.actorList.filters.max_count !== value[1]) {
          this.$store.state.actorList.filters.min_count = value[0]
          this.$store.state.actorList.filters.max_count = value[1]
          this.reloadList()
        }
      }
    },
    avails: {
      get () {
        return [this.$store.state.actorList.filters.min_avail, this.$store.state.actorList.filters.max_avail]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_avail !== value[0] || this.$store.state.actorList.filters.max_avail !== value[1]) {
          this.$store.state.actorList.filters.min_avail = value[0]
          this.$store.state.actorList.filters.max_avail = value[1]
          this.reloadList()
        }
      }
    },
    ratings: {
      get () {
        return [this.$store.state.actorList.filters.min_rating, this.$store.state.actorList.filters.max_rating]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_rating !== value[0] || this.$store.state.actorList.filters.max_rating !== value[1]) {
          this.$store.state.actorList.filters.min_rating = value[0]
          this.$store.state.actorList.filters.max_rating = value[1]
          this.reloadList()
        }
      }
    },
    sceneratings: {
      get () {
        return [this.$store.state.actorList.filters.min_scene_rating, this.$store.state.actorList.filters.max_scene_rating]
      },
      set (value) {
        if (this.$store.state.actorList.filters.min_scene_rating !== value[0] || this.$store.state.actorList.filters.max_scene_rating !== value[1]) {
          this.$store.state.actorList.filters.min_scene_rating = value[0]
          this.$store.state.actorList.filters.max_scene_rating = value[1]
          this.reloadList()
        }
      }
    },
    attributes: {
      get () {
        return this.$store.state.actorList.filters.attributes
      },
      set (value) {
        this.$store.state.actorList.filters.attributes = value
        this.reloadList()
      }
    },
    sort: {
      get () {
        return this.$store.state.actorList.filters.sort
      },
      set (value) {
        this.$store.state.actorList.filters.sort = value
        this.reloadList()
      }
    },
    disableNewAkaGroup () {
      let akaCastCnt = 0
      let actorCnt = 0
      for (let i = 0; i < this.cast.length; i++) {
        if (this.cast[i].startsWith('aka:')) akaCastCnt++
        else actorCnt++
      }
      return !(akaCastCnt === 0 && actorCnt > 1)
    },
    disableDeleteAkaGroup () {
      let akaCastCnt = 0
      let actorCnt = 0
      for (let i = 0; i < this.cast.length; i++) {
        if (this.cast[i].startsWith('aka:')) akaCastCnt++
        else actorCnt++
      }
      return !(akaCastCnt === 1 && actorCnt === 0)
    },
    disableAddToAkaGroup () {
      let akaCastCnt = 0
      let actorCnt = 0
      for (let i = 0; i < this.cast.length; i++) {
        if (this.cast[i].startsWith('aka:')) akaCastCnt++
        else actorCnt++
      }
      return !(akaCastCnt === 1 && actorCnt > 0)
    },
    disableRemoveFromAkaGroup () {
      let akaCastCnt = 0
      let actorCnt = 0
      for (let i = 0; i < this.cast.length; i++) {
        if (this.cast[i].startsWith('aka:')) akaCastCnt++
        else actorCnt++
      }
      return !(akaCastCnt === 1 && actorCnt > 0)
    }
  }
}
</script>

<style lang="less" scoped>
.amz-sidebar-wrap {
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.amz-view-mode-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.45rem 0.65rem;
  background: #16202a;
  border: 1px solid #253342;
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
}

.amz-view-mode-label {
  font-size: 0.74rem;
  font-weight: 700;
  color: #94a3b8;
}

.amz-view-mode-switch {
  display: inline-flex;
  flex: 1;
  background: #0f171e;
  border: 1px solid #2b3b4c;
  border-radius: 6px;
  padding: 2px;
  gap: 2px;
}

.amz-view-mode-btn {
  flex: 1;
  border: none;
  border-radius: 4px;
  padding: 4px 8px;
  background: transparent;
  color: #94a3b8;
  font-size: 0.75rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s ease;
}

.amz-view-mode-btn:hover {
  color: #ffffff;
}

.amz-view-mode-btn.is-active {
  background: #00a8e1;
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 168, 225, 0.35);
}

.amz-filters {
  background: #16202a;
  border: 1px solid #253342;
  border-radius: 10px;
  font-size: 0.78rem;
  color: #e2e8f0;
  overflow: hidden;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.35);
}

.amz-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.55rem 0.7rem;
  background: #1b2836;
  border-bottom: 1px solid #253342;
}

.amz-header-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-weight: 700;
  font-size: 0.82rem;
  color: #ffffff;
}

.amz-count-badge {
  background: #00a8e1;
  color: #ffffff;
  font-size: 0.68rem;
  font-weight: 700;
  padding: 0 6px;
  border-radius: 999px;
  line-height: 1.4;
}

.amz-clear-btn {
  background: none;
  border: none;
  padding: 0;
  font-size: 0.72rem;
  color: #38bdf8;
  cursor: pointer;
  font-weight: 600;
}

.amz-clear-btn:hover {
  text-decoration: underline;
  color: #7dd3fc;
}

.amz-active-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  padding: 0.45rem 0.65rem;
  background: #121b24;
  border-bottom: 1px solid #253342;
}

.amz-active-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 2px 7px;
  background: rgba(0, 168, 225, 0.18);
  color: #7dd3fc;
  border: 1px solid rgba(0, 168, 225, 0.4);
  border-radius: 999px;
  font-size: 0.7rem;
  font-weight: 600;
  cursor: pointer;
}

.amz-active-chip:hover {
  background: rgba(0, 168, 225, 0.3);
  color: #ffffff;
}

.amz-chip-x {
  font-weight: 700;
  line-height: 1;
}

.amz-section {
  border-bottom: 1px solid #22303e;
}

.amz-section-last {
  border-bottom: none;
}

.amz-section-head {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.45rem 0.7rem;
  background: transparent;
  border: none;
  cursor: pointer;
  text-align: left;
  font-size: 0.76rem;
  font-weight: 700;
  color: #f1f5f9;
  transition: background 0.12s ease;
}

.amz-section-head.is-static {
  cursor: default;
  padding-bottom: 0.2rem;
}

.amz-section-head:not(.is-static):hover {
  background: #1d2a38;
}

.amz-section-title {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
}

.amz-inline-val {
  font-size: 0.68rem;
  font-weight: 600;
  color: #38bdf8;
  background: rgba(0, 168, 225, 0.16);
  padding: 0 5px;
  border-radius: 999px;
}

.amz-chevron {
  font-size: 0.75rem;
  color: #8197a4;
}

.amz-section-body {
  padding: 0.25rem 0.7rem 0.55rem;
}

.amz-select select {
  font-size: 0.76rem;
  background: #0f171e;
  color: #f1f5f9;
  border-color: #2e4052;
  border-radius: 6px;
}

.amz-select::after {
  border-color: #38bdf8 !important;
}

.amz-check-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.22rem 0.3rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.78rem;
  color: #cbd5e1;
  user-select: none;
}

.amz-check-row:hover {
  background: #1d2a38;
  color: #ffffff;
}

.amz-check-row input[type="checkbox"] {
  accent-color: #00a8e1;
  cursor: pointer;
}

.amz-icon-primary {
  color: #00a8e1;
}

.amz-icon-danger {
  color: #f43f5e;
}

.amz-genre-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 0.3rem;
}

.amz-genre-pill {
  padding: 2px 9px;
  border-radius: 999px;
  background: #233140;
  border: 1px solid #324457;
  color: #cbd5e1;
  font-size: 0.72rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.14s ease;
}

.amz-genre-pill:hover {
  background: #2d4054;
  color: #ffffff;
}

.amz-genre-pill.is-selected {
  background: #00a8e1;
  border-color: #00a8e1;
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 168, 225, 0.35);
}

.amz-seg-row {
  display: flex;
  gap: 2px;
  background: #0f171e;
  border: 1px solid #2b3b4c;
  border-radius: 6px;
  padding: 2px;
}

.amz-seg-btn {
  flex: 1;
  border: none;
  border-radius: 4px;
  padding: 4px 6px;
  background: transparent;
  color: #94a3b8;
  font-size: 0.73rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.14s ease;
}

.amz-seg-btn:hover {
  color: #ffffff;
}

.amz-seg-btn.is-selected {
  background: #00a8e1;
  color: #ffffff;
  box-shadow: 0 1px 4px rgba(0, 168, 225, 0.35);
}

.amz-star-row {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.35rem;
  padding: 0.18rem 0.35rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: 4px;
  cursor: pointer;
  text-align: left;
}

.amz-star-row:hover {
  background: #1d2a38;
}

.amz-star-row.is-selected {
  background: rgba(245, 158, 11, 0.16);
  border-color: rgba(245, 158, 11, 0.45);
  font-weight: 700;
}

.amz-stars {
  font-size: 0.88rem;
  letter-spacing: 1px;
  line-height: 1;
}

.amz-stars .is-filled {
  color: #f59e0b;
}

.amz-stars .is-empty {
  color: #475569;
}

.amz-star-text {
  font-size: 0.74rem;
  color: #cbd5e1;
}

.amz-mini-slider {
  margin-top: 0.25rem;
  padding: 0 0.25rem;
}

.amz-field-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: #94a3b8;
  margin-bottom: 0.15rem;
}

.amz-tag-fields ::v-deep .taginput .taginput-container {
  min-height: 1.75rem;
  padding: 2px 4px;
  font-size: 0.74rem;
  background: #0f171e;
  border-color: #2e4052;
  color: #f1f5f9;
}

.amz-tag-fields ::v-deep .taginput .taginput-container input {
  font-size: 0.74rem;
  height: 1.4rem;
  color: #f1f5f9;
  background: transparent;
}

.amz-tag-fields ::v-deep .taginput .taginput-container input::placeholder {
  color: #64748b;
}

.amz-range-row {
  margin-bottom: 0.45rem;
}

.amz-range-row:last-child {
  margin-bottom: 0;
}

.amz-range-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.72rem;
  font-weight: 600;
  color: #cbd5e1;
  margin-bottom: 0.1rem;
}

.amz-range-val {
  font-size: 0.7rem;
  color: #38bdf8;
  font-variant-numeric: tabular-nums;
}

.tagicon {
  margin-right: -0.2em !important;
}
</style>
