<template>
  <div class="container">
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <h3>{{$t("Custom > DMM API Settings")}}</h3>
      <hr/>
      <b-tabs v-model="activeTab" size="medium" type="is-boxed" style="margin-left: 0px" id="importexporttab">
            <b-tab-item label="DMM API Access"/>
      </b-tabs>
      <div class="columns">
        <div class="column">

          <section>
          <!-- DMM API Settings -->
           <div v-if="activeTab == 0">
                <b-field :label="$t('DMM Api id')" label-position="on-border">
                  <b-input v-model="dmmApiId" placeholder="Visit https://affiliate.dmm.com/api/ to sign up to DMM-api service" type="password"></b-input>
                </b-field>
                <b-field :label="$t('DMM Affiliate id')" label-position="on-border">
                  <b-input v-model="dmmAffiliateId" placeholder="Visit https://affiliate.dmm.com/api/ to sign up to DMM-api service" type="password"></b-input>
                </b-field>
          </div>
            <hr/>
              <b-field grouped>
                <b-button type="is-primary" @click="saveSettings" style="margin-right:1em">Save settings</b-button>
              </b-field>
          </section>
          <hr/>
          <section>
            <p>
              Restart XBVR to use new schedule settings
            </p>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import ky from 'ky'
import prettyBytes from 'pretty-bytes'

export default {
  name: 'CustomDMM',
  data () {
    return {
      isLoading: true,
      activeTab: 0,
      
      dmmApiId: '',
      dmmAffiliateId: '',
    }
  },
  async mounted () {
    await this.loadState()
  },
  computed: {
  },
  methods: {
    async loadState () {
      this.isLoading = true
      await ky.get('/api_custom/options/state')
        .json()
        .then(data => {

          this.dmmAffiliateId = data.config.custom.dmmapikey.dmmAffiliateId
          this.dmmApiId = data.config.custom.dmmapikey.dmmApiId

          this.isLoading = false
        })
    },
    async saveSettings () {
      this.isLoading = true
      await ky.post('/api_custom/options/save', {
        json: {          
          dmmAffiliateId: this.dmmAffiliateId,
          dmmApiId: this.dmmApiId,
        }
      })
        .json()
        .then(data => {
          this.isLoading = false
        })
    },
    prettyBytes
  }
}
</script>
