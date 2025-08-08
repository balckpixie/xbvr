<template>
  <div class="container">
    <b-loading :is-full-page="false" :active.sync="isLoading"></b-loading>
    <div class="content">
      <h3>{{$t("Custom Settings")}}</h3>
      <hr/>
      <b-tabs v-model="activeTab" size="medium" type="is-boxed" style="margin-left: 0px" id="importexporttab">
            <b-tab-item label="Thumbnail Generation"/>
            <b-tab-item label="DMM Access"/>
      </b-tabs>
      <div class="columns">
        <div class="column">

          <section>
          <!-- Actor Related Settings -->
          <div v-if="activeTab == 0">
            <section>
              <b-field>
                <b-switch v-model="thumbnailEnabled">Enable schedule</b-switch>
              </b-field>
              <b-field v-if="thumbnailEnabled">
                <b-slider v-model="thumbnailHourInterval" :min="1" :max="23" :step="1" ></b-slider>
                <div class="column is-one-third" style="margin-left:.75em">{{`Run every ${this.thumbnailHourInterval} hour${this.thumbnailHourInterval > 1 ? 's': ''}`}}</div>
              </b-field>
              <b-field>
                <b-switch v-if="thumnbnailEnabled" v-model="useThumnbnailTimeRange">Limit time of day</b-switch>
              </b-field>
              <div v-if="useThumnbnailTimeRange && thumbnailEnabled">
                <b-field>
                  <b-slider v-model="thumbnailTimeRange" :min="0" :max="48" :step="1" :custom-formatter="val => timeRange[val]" @input="restrictThumbnailTo24Hours">
                    <b-slider-tick :value="0">00:00</b-slider-tick>
                    <b-slider-tick :value="6">06:00</b-slider-tick>
                    <b-slider-tick :value="12">12:00</b-slider-tick>
                    <b-slider-tick :value="18">18:00</b-slider-tick>
                    <b-slider-tick :value="24">Midnight</b-slider-tick>
                    <b-slider-tick :value="30">06:00</b-slider-tick>
                    <b-slider-tick :value="36">12:00</b-slider-tick>
                    <b-slider-tick :value="42">18:00</b-slider-tick>
                    <b-slider-tick :value="48">00:00</b-slider-tick>
                  </b-slider>
                  <div class="column is-one-third" style="margin-left:.75em">{{`${this.timeRange[this.thumbnailTimeRange[0]]} - ${this.timeRange[this.thumbnailTimeRange[1]]}`}}</div>
                </b-field>
                <b-field>
                  <b-slider v-model="thumbnailMinuteStart" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third" style="margin-left:.75em">{{ minutesStartMsg(thumbnailMinuteStart) }}</div>
                </b-field>
                <p>
                  Thumbnail Generation of a scene will not start after the Time Window Ends
                </p>
              </div>
              <br/>
              <b-field label="Startup">
                  <b-slider v-model="thumbnailStartDelay" :min="0" :max="60" :step="1" ></b-slider>
                  <div class="column is-one-third" style="margin-left:.75em">{{ delayStartMsg(thumbnailStartDelay) }}</div>
              </b-field>
              <p>
                BETA NOTE: Please note this is CPU-heavy process, if approriate limit the Time of Day the task runs                  
              </p>                  
              
          </div>

          <!-- Actor Related Settings -->
           <div v-if="activeTab == 1">
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
  name: 'CustomOption',
  data () {
    return {
      isLoading: true,
      activeTab: 0,
      
      dmmApiId: '',
      dmmAffiliateId: '',

      thumbnailEnabled: false,
      thumbnailTimeRange:[0,23],
      thumbnailHourInterval: 0,
      thumbnailMinuteStart: 0,
      thumbnailStartDelay: 0,
      lastThumbnailTimeRange: [0,23],
      useThumbnailTimeRange: false,     

      timeRange: ['00:00', '01:00', '02:00', '03:00', '04:00', '05:00', '06:00', '07:00', '08:00', '09:00', '10:00', '11:00',
        '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00', '21:00', '22:00', '23:00',
        '00:00', '01:00', '02:00', '03:00', '04:00', '05:00', '06:00', '07:00', '08:00', '09:00', '10:00', '11:00',
        '12:00', '13:00', '14:00', '15:00', '16:00', '17:00', '18:00', '19:00', '20:00', '21:00', '22:00', '23:00', '00:00']
    }
  },
  async mounted () {
    await this.loadState()
  },
  computed: {
    // dmmApiId: {
    //   get () {
    //     return this.$store.state.optionsAdvanced.advanced.dmmApiId
    //   },
    //   set (value) {
    //     this.$store.state.optionsAdvanced.advanced.dmmApiId = value
    //   }
    // },
    // dmmAffiliateId: {
    //   get () {
    //     return this.$store.state.optionsAdvanced.advanced.dmmAffiliateId
    //   },
    //   set (value) {
    //     this.$store.state.optionsAdvanced.advanced.dmmAffiliateId = value
    //   }
    // }
  },
  methods: {
    restrictThumbnailTo24Hours () {
      this.thumbnailTimeRange = this.restrictTo24Hours(this.thumbnailTimeRange, this.lastThumbnailTimeRange)
      this.lastThumbnailTimeRange = this.thumbnailTimeRange
    },
    restrictTo24Hours (timeRange, lastTimeRange) {
      // check the first time is not in the second 24 hours, no need, should be in the first 24 hours
      if (timeRange[0] > 23) {
        timeRange[0] = 23
        timeRange = [timeRange[0], timeRange[1]]
      }
      // check they are not trying to select more than a 24 hour range
      if ((timeRange[1] - timeRange[0]) > 23 ) {
        if (timeRange[0] === lastTimeRange[0] || timeRange[0] === lastTimeRange[1]) {
          timeRange = [timeRange[1] - 23, timeRange[1]]
        } else {
          timeRange = [timeRange[0], timeRange[0] + 23]
        }
      }
      return timeRange
    },
    async loadState () {
      this.isLoading = true
      await ky.get('/api_custom/options/state')
        .json()
        .then(data => {

          this.dmmAffiliateId = data.config.custom.dmmAffiliateId
          this.dmmApiId = data.config.custom.dmmApiId
          this.thumbnailEnabled = data.config.cron.thumbnailSchedule.enabled
          this.thumbnailHourInterval = data.config.cron.thumbnailSchedule.hourInterval
          this.useThumbnailTimeRange = data.config.cron.thumbnailSchedule.useRange
          this.thumbnailMinuteStart = data.config.cron.thumbnailSchedule.minuteStart
          if (data.config.cron.thumbnailSchedule.hourStart > data.config.cron.thumbnailSchedule.hourEnd) {
            this.thumbnailTimeRange = [data.config.cron.thumbnailSchedule.hourStart, data.config.cron.thumbnailSchedule.hourEnd + 24]
          } else {
            this.thumbnailTimeRange = [data.config.cron.thumbnailSchedule.hourStart, data.config.cron.thumbnailSchedule.hourEnd]            
          }
          this.thumbnailStartDelay = data.config.cron.thumbnailSchedule.runAtStartDelay

          this.isLoading = false
        })
    },
    minutesStartMsg (start) {
      if (start === 0) {
        return 'Start on the hour'
      }
      if (start === 1) {
        return 'Start at 1 minute past the hour'
      }
      return `Start at ${start} minutes past the hour`
    },
    delayStartMsg (start) {
      if (start === 0) {
        return 'Do not run at statup'
      }else{
        if (start === 1) {
          return `Run at 1 minute after startup`
        }else{
          return `Run at ${start} minutes after startup`
        }
      }
    },
    async saveSettings () {
      this.isLoading = true
      await ky.post('/api_custom/options/save', {
        json: {          
          dmmAffiliateId: this.dmmAffiliateId,
          dmmApiId: this.dmmApiId,

          thumbnailEnabled: this.thumbnailEnabled,
          thumbnailHourInterval: this.thumbnailHourInterval,
          thumbnailUseRange: this.useThumbnailTimeRange,
          thumbnailMinuteStart: this.thumbnailMinuteStart,
          thumbnailHourStart: this.thumbnailTimeRange[0],
          thumbnailHourEnd: this.thumbnailTimeRange[1],
          thumbnailStartDelay:this.thumbnailStartDelay,
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
