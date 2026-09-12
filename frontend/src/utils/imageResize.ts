/**
 * 浏览器端图片压缩 / 裁剪工具：上传前把图处理成「首页实际需要的尺寸」。
 *
 * 为什么放在客户端：后端没有图像处理依赖，而首页商品图只是 46px 的方块
 * （PlanCard → BrandLogo），原图动辄几千像素纯属浪费带宽与存储。
 * 这里用 canvas 一次性输出目标尺寸，上传体积可稳定压到百 KB 以内。
 *
 * 失败即降级：浏览器不支持 canvas / toBlob 时原样返回，绝不因为压缩失败而阻断上传。
 */

export interface ImageResizeOptions {
  /** 输出为正方形（居中裁剪），单位像素。用于商品图这类方块展示位。 */
  square?: number
  /** 非正方形时的最大宽度（等比缩放，不放大）。用于轮播这类横向展示位。 */
  maxWidth?: number
  /** 非正方形时的最大高度（等比缩放，不放大）。用于轮播这类横向展示位。 */
  maxHeight?: number
  /** 期望的输出体积上限（字节）；超出则继续降画质、再降尺寸。0 表示不限制。 */
  maxBytes?: number
}

/** 画质档位：从高到低，先保画质、再保体积。 */
const QUALITY_STEPS = [0.92, 0.85, 0.78, 0.68]
/** 尺寸档位：画质降到底仍超标时，再按比例缩小边长。 */
const SIZE_STEPS = [1, 0.8, 0.64, 0.5]

function readAsDataURL(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(typeof reader.result === 'string' ? reader.result : '')
    reader.onerror = () => reject(reader.error ?? new Error('图片读取失败'))
    reader.readAsDataURL(file)
  })
}

function loadImage(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new Image()
    image.onload = () => resolve(image)
    image.onerror = () => reject(new Error('图片解析失败'))
    image.src = src
  })
}

function canvasToBlob(canvas: HTMLCanvasElement, quality: number): Promise<Blob | null> {
  return new Promise((resolve) => {
    canvas.toBlob((blob) => resolve(blob), 'image/webp', quality)
  })
}

function renameExt(name: string, fallback: string): string {
  const base = name.replace(/\.[^.]+$/, '') || fallback
  return base
}

/** 计算等比缩放后的目标边长（永不放大）。 */
function fitWithin(width: number, height: number, maxWidth?: number, maxHeight?: number) {
  const limitW = maxWidth && maxWidth > 0 ? maxWidth : width
  const limitH = maxHeight && maxHeight > 0 ? maxHeight : height
  const scale = Math.min(1, limitW / width, limitH / height)
  return {
    width: Math.max(1, Math.round(width * scale)),
    height: Math.max(1, Math.round(height * scale))
  }
}

/**
 * 按目标展示尺寸压缩图片，返回新的 File。
 *
 * - `square`：居中裁剪成正方形（首页商品图是方块位，避免用户随手传的宽图被压扁）；
 * - 否则按 `maxWidth` / `maxHeight` 等比缩放；
 * - 输出统一 WebP，保留透明通道，商品 logo 不会带上一圈白底。
 */
export async function resizeImageForUpload(file: File, options: ImageResizeOptions): Promise<File> {
  const square = options.square && options.square > 0 ? Math.round(options.square) : 0
  const maxBytes = options.maxBytes && options.maxBytes > 0 ? options.maxBytes : 0

  let image: HTMLImageElement
  try {
    image = await loadImage(await readAsDataURL(file))
  } catch {
    return file
  }

  const sourceWidth = image.naturalWidth || image.width
  const sourceHeight = image.naturalHeight || image.height
  if (!sourceWidth || !sourceHeight) return file

  const canvas = document.createElement('canvas')
  const ctx = canvas.getContext('2d')
  if (!ctx || typeof canvas.toBlob !== 'function') return file

  const base = square
    ? { width: square, height: square }
    : fitWithin(sourceWidth, sourceHeight, options.maxWidth, options.maxHeight)

  // 正方形裁剪时，从原图中心取一块最大的正方形贴到输出画布上（cover 语义）
  const cropSide = square ? Math.min(sourceWidth, sourceHeight) : 0
  const cropX = square ? (sourceWidth - cropSide) / 2 : 0
  const cropY = square ? (sourceHeight - cropSide) / 2 : 0

  let smallest: Blob | null = null

  for (const sizeScale of SIZE_STEPS) {
    const width = Math.max(1, Math.round(base.width * sizeScale))
    const height = Math.max(1, Math.round(base.height * sizeScale))
    canvas.width = width
    canvas.height = height
    ctx.clearRect(0, 0, width, height)
    if (square) {
      ctx.drawImage(image, cropX, cropY, cropSide, cropSide, 0, 0, width, height)
    } else {
      ctx.drawImage(image, 0, 0, width, height)
    }

    for (const quality of QUALITY_STEPS) {
      const blob = await canvasToBlob(canvas, quality)
      if (!blob) continue
      if (!smallest || blob.size < smallest.size) smallest = blob
      if (!maxBytes || blob.size <= maxBytes) {
        return toWebpFile(blob, file.name)
      }
    }
  }

  // 所有档位都超标时，交出体积最小的那一版，让后端的上限去兜底报错
  return smallest ? toWebpFile(smallest, file.name) : file
}

function toWebpFile(blob: Blob, originalName: string): File {
  // 浏览器若不支持 WebP 编码会回退成 PNG，这里按真实类型给扩展名，避免名实不符
  const isWebp = blob.type === 'image/webp'
  const ext = isWebp ? 'webp' : 'png'
  return new File([blob], `${renameExt(originalName, 'image')}.${ext}`, {
    type: blob.type || 'image/webp'
  })
}
