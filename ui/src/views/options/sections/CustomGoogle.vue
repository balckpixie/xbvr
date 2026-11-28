<template>
  <div class="container">
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <h3>{{$t("Custom > Google Image Search API Settings")}}</h3>
      <hr/>
      <div class="columns">
        <div class="column">
          <section>
            <b-field :label="$t('Google API Key')" label-position="on-border">
              <b-input
                v-model="googleApiKey"
                placeholder="Google Cloud Console で発行した Custom Search JSON API キー"
                type="password"
                password-reveal
              />
            </b-field>
            <b-field :label="$t('Search Engine ID (cx)')" label-position="on-border">
              <b-input
                v-model="googleSearchEngineId"
                placeholder="Programmable Search Engine の検索エンジン ID (cx)"
              />
            </b-field>
            <hr/>
            <b-field grouped>
              <b-button type="is-primary" @click="saveSettings" style="margin-right:1em">Save settings</b-button>
            </b-field>
          </section>
          <hr/>
          <section>
            <p class="is-size-7 has-text-grey">
              Google Programmable Search Engine（画像検索をオンに設定）の検索エンジン ID（cx）と、Google Cloud の Custom Search JSON API キーを設定してください。
            </p>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ky from 'ky'

export default {
  name: 'CustomGoogle',
  data () {
    return {
      isLoading: true,
      googleApiKey: '',
      googleSearchEngineId: ''
    }
  },
  async mounted () {
    await this.loadState()
  },
  methods: {
    async loadState () {
      this.isLoading = true
      try {
        const data = await ky.get('/api_custom/options/state').json()
        const g = data && data.config && data.config.custom && data.config.custom.googleapikey
        if (g) {
          this.googleApiKey = g.googleApiKey || ''
          this.googleSearchEngineId = g.googleSearchEngineId || ''
        }
      } finally {
        this.isLoading = false
      }
    },
    async saveSettings () {
      this.isLoading = true
      try {
        await ky.post('/api_custom/options/save', {
          json: {
            saveGoogleSettings: true,
            googleApiKey: this.googleApiKey,
            googleSearchEngineId: this.googleSearchEngineId
          }
        }).json()
        this.$buefy.toast.open({ message: 'Google Image Search API settings saved', type: 'is-success' })
      } catch (e) {
        this.$buefy.toast.open({ message: 'Failed to save Google API settings', type: 'is-danger' })
      } finally {
        this.isLoading = false
      }
    }
  }
}
</script>
