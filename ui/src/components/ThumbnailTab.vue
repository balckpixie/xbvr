<template>
  <div ref="thumbContainer" class="thumbnail-container"></div>
</template>

<script setup>
import { ref, defineExpose, defineEmits } from 'vue'

const props = defineProps({
  file: {
    type: Object,
    required: true
  },
  displayWidth: {
    type: Number,
    default: 120   // 表示用（thumbsImageのCSS幅）
  }
})

const emit = defineEmits(['thumbnailClicked'])
const thumbContainer = ref(null)
const thumbnails = ref([])

function loadThumbnails() {
  const canvasContainer = thumbContainer.value
  if (!canvasContainer) {
    console.warn('Container not ready.')
    return
  }
  clearVideThumbnails()
  if (!props.file.has_thumbnail)
  {
    return
  }
  const thumbnailUrl = '/api_custom/thumbnail/image/' + props.file.id
  fetchAndDisplayThumbnails(thumbnailUrl, canvasContainer, props.file)
}

function fetchAndDisplayThumbnails(imageUrl, container, file) {
  // thumbnail_parameters をパースして値を取得
  let parsed = {}
  try {
    // 文字列なら JSON.parse、すでにオブジェクトならそのまま
    parsed = (typeof file.thumbnail_parameters === 'string')
      ? JSON.parse(file.thumbnail_parameters || '{}')
      : (file.thumbnail_parameters || {})
  } catch (e) {
    console.error('Failed to parse thumbnail_parameters:', e)
    parsed = {}
  }

  const start = parsed.start ?? 5
  const interval = parsed.interval ?? 30
  const tileWidthSetting = parsed.resolution ?? 200

  loadImage(imageUrl)
    .then((img) => {
      const canvas = drawImageToCanvas(img)
      const { tileWidth, tileHeight, rows, cols } = calculateTileGrid(canvas, file, tileWidthSetting)

      let duration = start

      for (let row = 0; row < rows; row++) {
        for (let col = 0; col < cols; col++) {
          const currentDuration = duration
          const thumbnailCanvas = createThumbnailCanvas(
            canvas,
            row,
            col,
            tileWidth,
            tileHeight,
            file.projection
          )
          if (!thumbnailCanvas) continue

          const ctx = thumbnailCanvas.getContext('2d')
          if (!ctx) continue

          if (!isImageBlack(ctx)) {
            // displayWidth を適用
            thumbnailCanvas.style.width = props.displayWidth + 'px'
            thumbnailCanvas.classList.add('thumb-wrapper')

            // --- ホバーイベント追加 ---
            const popup = document.createElement('div')
            popup.className = 'thumb-popup'
            popup.innerText = formatTime(currentDuration)
            popup.style.display = 'none'
            thumbnailCanvas.parentElement?.appendChild(popup)

            thumbnailCanvas.addEventListener('mouseenter', () => {
              thumbnailCanvas.classList.add('hovered')
              popup.style.display = 'block'
            })
            thumbnailCanvas.addEventListener('mouseleave', () => {
              thumbnailCanvas.classList.remove('hovered')
              popup.style.display = 'none'
            })

            thumbnailCanvas.addEventListener('click', () => {
              emit('thumbnailClicked', currentDuration)
            })

            // サムネイルをラップする要素を作成して配置
            const wrapper = document.createElement('div')
            wrapper.className = 'thumb-container'
            wrapper.appendChild(thumbnailCanvas)
            wrapper.appendChild(popup)

            container.appendChild(wrapper)
          }
          duration += interval
        }
      }
    })
    .catch((error) => {
      console.error('Failed to load image:', error)
    })
}

function formatTime(seconds) {
  const m = Math.floor(seconds / 60)
  const s = seconds % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

function loadImage(url) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.crossOrigin = 'Anonymous'
    img.src = url
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('Image load error'))
  })
}

function drawImageToCanvas(img) {
  const canvas = document.createElement('canvas')
  canvas.width = img.width
  canvas.height = img.height
  const ctx = canvas.getContext('2d')
  ctx.drawImage(img, 0, 0)
  return canvas
}

function calculateTileGrid(canvas, file, tileWidthSetting) {
  const tileWidth = tileWidthSetting
  let tileHeight = tileWidthSetting
  if (file?.projection === 'flat') {
    tileHeight = (file.video_height / file.video_width) * tileWidth
  }
  const rows = Math.floor(canvas.height / tileHeight)
  const cols = Math.floor(canvas.width / tileWidth)
  return {
    tileWidth,
    tileHeight,
    rows,
    cols
  }
}

function createThumbnailCanvas(canvas, row, col, tileWidth, tileHeight, projection) {
  try {
    const thumbnailCanvas = document.createElement('canvas')
    thumbnailCanvas.width = tileWidth
    thumbnailCanvas.height = tileHeight
    thumbnailCanvas.classList.add('thumbsImage')

    const ctx = thumbnailCanvas.getContext('2d')
    if (!ctx) throw new Error("getContext('2d') returned null")

    if (projection === 'flat') {
      ctx.drawImage(
        canvas,
        col * tileWidth,
        row * tileHeight,
        tileWidth,
        tileHeight,
        0,
        0,
        tileWidth,
        tileHeight
      )
    } else {
      ctx.drawImage(
        canvas,
        col * tileWidth + 20,
        row * tileHeight + 20,
        tileWidth - 40,
        tileHeight - 40,
        0,
        0,
        tileWidth,
        tileHeight
      )
    }

    return thumbnailCanvas
  } catch (error) {
    console.error('createThumbnailCanvas error:', error)
    return null
  }
}

function isImageBlack(ctx) {
  const imageData = ctx.getImageData(0, 0, ctx.canvas.width, ctx.canvas.height)
  const data = imageData.data

  for (let i = 0; i < data.length; i += 4) {
    if (data[i] > 10 || data[i + 1] > 10 || data[i + 2] > 10) {
      return false
    }
  }
  return true
}

function clearVideThumbnails() {
  const canvasContainer = thumbContainer.value
  if (canvasContainer) {
    canvasContainer.innerHTML = ''
  }
  thumbnails.value = []
}

defineExpose({
  loadThumbnails,
  fetchAndDisplayThumbnails,
  clearVideThumbnails
})
</script>

<style>
.thumbnail-container {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.thumb-container {
  position: relative;
  display: inline-block;
}

.thumbsImage {
  transition: transform 0.2s ease;
  border-radius: 6px;
  cursor: pointer;
}

/* ホバー時に拡大 */
.thumbsImage.hovered {
  transform: scale(1.3);
  z-index: 2;
}

.thumb-popup {
  position: absolute;
  top: -10px;          /* 上寄せ */
  right: -10px;        /* 右寄せ */
  background: rgba(0,0,0,0.75);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  pointer-events: none;
}
</style>
