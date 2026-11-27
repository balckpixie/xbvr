export function parseThumbnailParams (rawParams) {
  let parsed = {}
  try {
    parsed = (typeof rawParams === 'string')
      ? JSON.parse(rawParams || '{}')
      : (rawParams || {})
  } catch (e) {
    parsed = {}
  }
  return {
    start: parsed.start ?? 15,
    interval: parsed.interval ?? 15,
    resolution: parsed.resolution ?? 200
  }
}

export function formatDurationSeconds (seconds) {
  const sec = Math.floor(seconds || 0)
  const m = Math.floor(sec / 60)
  const s = sec % 60
  return `${m}:${s.toString().padStart(2, '0')}`
}

export function loadImage (url) {
  return new Promise((resolve, reject) => {
    const img = new Image()
    img.crossOrigin = 'Anonymous'
    img.src = url
    img.onload = () => resolve(img)
    img.onerror = () => reject(new Error('Image load error'))
  })
}

export function drawImageToCanvas (img) {
  const canvas = document.createElement('canvas')
  canvas.width = img.width
  canvas.height = img.height
  const ctx = canvas.getContext('2d')
  ctx.drawImage(img, 0, 0)
  return canvas
}

export function calculateTileGrid (canvas, file, tileWidthSetting) {
  const tileWidth = tileWidthSetting
  let croppedWidth = file.video_width || 1920
  let croppedHeight = file.video_height || 1080

  switch (file?.projection) {
    case '180_sbs':
    case '360_sbs':
      croppedWidth = croppedWidth / 2
      break
    case '180_tb':
    case '360_tb':
      croppedHeight = croppedHeight / 2
      break
    default:
      break
  }

  const tileHeight = Math.max(1, Math.round((croppedHeight / croppedWidth) * tileWidth))
  const rows = Math.floor(canvas.height / tileHeight)
  const cols = Math.floor(canvas.width / tileWidth)

  return { tileWidth, tileHeight, rows, cols }
}

export function createThumbnailCanvas (canvas, row, col, tileWidth, tileHeight, projection) {
  try {
    const thumbnailCanvas = document.createElement('canvas')
    thumbnailCanvas.width = tileWidth
    thumbnailCanvas.height = tileHeight
    const ctx = thumbnailCanvas.getContext('2d')
    if (!ctx) return null

    if (projection === 'flat' || tileWidth <= 40 || tileHeight <= 40) {
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
  } catch (e) {
    return null
  }
}

export function isImageBlack (ctx) {
  const width = ctx.canvas.width
  const height = ctx.canvas.height
  const imageData = ctx.getImageData(0, 0, width, height)
  const data = imageData.data
  const excludeXStart = width * (5 / 6)
  const excludeYEnd = height * (1 / 6)

  for (let y = 0; y < height; y += 2) {
    for (let x = 0; x < width; x += 2) {
      if (x >= excludeXStart && y < excludeYEnd) continue
      const index = (y * width + x) * 4
      const brightness = (data[index] + data[index + 1] + data[index + 2]) / 3
      if (brightness > 20) return false
    }
  }
  return true
}

/**
 * Extract up to `maxCount` evenly spaced non-black thumbnail frames as data URLs + timestamps.
 */
export async function extractSampleThumbnails (file, maxCount = 5) {
  if (!file || !file.has_thumbnail) return []

  const { start, interval, resolution } = parseThumbnailParams(file.thumbnail_parameters)
  const imageUrl = `/api_custom/thumbnail/image/${file.id}`
  const img = await loadImage(imageUrl)
  const spriteCanvas = drawImageToCanvas(img)
  const { tileWidth, tileHeight, rows, cols } = calculateTileGrid(spriteCanvas, file, resolution)

  if (rows <= 0 || cols <= 0) return []

  const candidates = []
  let duration = start
  for (let row = 0; row < rows; row++) {
    for (let col = 0; col < cols; col++) {
      if (file.duration > 0 && duration > file.duration) break
      candidates.push({ row, col, duration })
      duration += interval
    }
    if (file.duration > 0 && duration > file.duration) break
  }

  if (candidates.length === 0) return []

  // Pick evenly spaced candidate indices (plus neighbors if black)
  const sampleCount = Math.min(maxCount, candidates.length)
  const selected = []
  const usedIndices = new Set()

  for (let i = 0; i < sampleCount; i++) {
    const targetIdx = sampleCount === 1
      ? 0
      : Math.round((i * (candidates.length - 1)) / (sampleCount - 1))

    // Search outward from targetIdx for a non-black frame
    let chosen = null
    for (let offset = 0; offset < candidates.length; offset++) {
      const tryIndices = offset === 0 ? [targetIdx] : [targetIdx + offset, targetIdx - offset]
      for (const idx of tryIndices) {
        if (idx < 0 || idx >= candidates.length || usedIndices.has(idx)) continue
        const c = candidates[idx]
        const tile = createThumbnailCanvas(spriteCanvas, c.row, c.col, tileWidth, tileHeight, file.projection)
        if (!tile) continue
        const ctx = tile.getContext('2d')
        if (!ctx || isImageBlack(ctx)) continue
        usedIndices.add(idx)
        chosen = {
          duration: c.duration,
          timeLabel: formatDurationSeconds(c.duration),
          dataUrl: tile.toDataURL('image/jpeg', 0.85)
        }
        break
      }
      if (chosen) break
    }
    if (chosen) selected.push(chosen)
  }

  return selected
}
