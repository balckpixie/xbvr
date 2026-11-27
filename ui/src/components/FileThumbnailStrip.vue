<template>
  <div v-if="samples.length > 0" class="file-thumb-strip">
    <div
      v-for="(sample, idx) in samples"
      :key="idx"
      class="file-thumb-item"
      @click.stop="$emit('thumbnail-clicked', sample.duration)"
    >
      <img class="thumbsImage" :src="sample.dataUrl" :alt="sample.timeLabel" />
      <div class="thumb-popup">{{ sample.timeLabel }}</div>
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
      default: 8
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
  display: flex;
  overflow: visible;
  cursor: pointer;
}

.thumbsImage {
  width: 90px;
  height: 90px;
  object-fit: cover;
  display: block;
  border-radius: 6px;
  transition: transform 0.2s ease;
}

.file-thumb-item:hover {
  z-index: 10;
}

.file-thumb-item:hover .thumbsImage {
  transform: scale(1.3);
  z-index: 2;
}

.thumb-popup {
  display: none;
  position: absolute;
  top: -5px;
  right: -10px;
  background: rgba(0, 0, 0, 0.75);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  pointer-events: none;
  z-index: 999;
}

.file-thumb-item:hover .thumb-popup {
  display: block;
}
</style>
