<template>
  <div v-if="samples.length > 0" class="file-thumb-strip">
    <div
      v-for="(sample, idx) in samples"
      :key="idx"
      class="file-thumb-item"
      :title="`Play from ${sample.timeLabel}`"
      @click.stop="$emit('thumbnail-clicked', sample.duration)"
    >
      <img :src="sample.dataUrl" :alt="sample.timeLabel" />
      <span class="file-thumb-time">{{ sample.timeLabel }}</span>
    </div>
  </div>
</template>

<script>
import { extractSampleThumbnails } from '../utils/thumbnailSprite'

export default {
  name: 'FileThumbnailStrip',
  props: {
    file: {
      type: Object,
      required: true
    },
    maxCount: {
      type: Number,
      default: 5
    }
  },
  data () {
    return {
      samples: []
    }
  },
  watch: {
    'file.id': {
      immediate: true,
      handler () {
        this.loadSamples()
      }
    }
  },
  methods: {
    async loadSamples () {
      this.samples = []
      if (!this.file || !this.file.has_thumbnail) return
      try {
        this.samples = await extractSampleThumbnails(this.file, this.maxCount)
      } catch (e) {
        this.samples = []
      }
    }
  }
}
</script>

<style scoped>
.file-thumb-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 6px;
}

.file-thumb-item {
  position: relative;
  cursor: pointer;
  border-radius: 4px;
  overflow: hidden;
  border: 1px solid rgba(0, 0, 0, 0.15);
  transition: transform 0.15s ease, box-shadow 0.15s ease;
}

.file-thumb-item:hover {
  transform: scale(1.08);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.25);
  z-index: 2;
}

.file-thumb-item img {
  display: block;
  height: 54px;
  width: auto;
  object-fit: cover;
}

.file-thumb-time {
  position: absolute;
  right: 2px;
  bottom: 2px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  font-size: 10px;
  line-height: 1.2;
  padding: 1px 4px;
  border-radius: 3px;
}
</style>
