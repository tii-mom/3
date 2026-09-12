import { afterEach, describe, expect, it, vi } from 'vitest'
import { resizeImageForUpload } from '../imageResize'

const originalFileReader = globalThis.FileReader
const originalImage = globalThis.Image
const originalCreateElement = document.createElement.bind(document)

interface CanvasProbe {
  widths: number[]
  heights: number[]
  drawCalls: number[][]
}

interface MockOptions {
  naturalWidth?: number
  naturalHeight?: number
  /** 每次 toBlob 返回的体积（字节）；用尽后沿用最后一个值 */
  blobSizes?: number[]
  /** 模拟浏览器不支持 canvas 2d */
  noContext?: boolean
  /** 模拟图片解码失败 */
  decodeFails?: boolean
}

function installMocks(options: MockOptions = {}): CanvasProbe {
  const probe: CanvasProbe = { widths: [], heights: [], drawCalls: [] }
  const blobSizes = options.blobSizes ?? [1024]

  class MockFileReader {
    result: string | ArrayBuffer | null = null
    onload: (() => void) | null = null
    onerror: (() => void) | null = null

    readAsDataURL() {
      this.result = 'data:image/png;base64,ZmFrZQ=='
      this.onload?.()
    }
  }

  class MockImage {
    naturalWidth = options.naturalWidth ?? 3000
    naturalHeight = options.naturalHeight ?? 2000
    onload: (() => void) | null = null
    onerror: (() => void) | null = null

    set src(_value: string) {
      if (options.decodeFails) this.onerror?.()
      else this.onload?.()
    }
  }

  globalThis.FileReader = MockFileReader as unknown as typeof FileReader
  globalThis.Image = MockImage as unknown as typeof Image

  vi.spyOn(document, 'createElement').mockImplementation(((tagName: string, opts?: ElementCreationOptions) => {
    if (tagName !== 'canvas') return originalCreateElement(tagName, opts)

    let blobIndex = 0
    const canvas: Record<string, unknown> = {
      getContext: () =>
        options.noContext
          ? null
          : {
              clearRect: vi.fn(),
              drawImage: (...args: number[]) => {
                probe.drawCalls.push(args)
              }
            },
      toBlob: (callback: BlobCallback, type?: string) => {
        const size = blobSizes[Math.min(blobIndex, blobSizes.length - 1)]
        blobIndex += 1
        callback(new Blob([new Uint8Array(size)], { type: type || 'image/webp' }))
      }
    }
    // 记录每次写入的边长，用于断言「输出尺寸确实是目标尺寸」
    let width = 0
    let height = 0
    Object.defineProperty(canvas, 'width', {
      get: () => width,
      set: (next: number) => {
        width = next
        probe.widths.push(next)
      }
    })
    Object.defineProperty(canvas, 'height', {
      get: () => height,
      set: (next: number) => {
        height = next
        probe.heights.push(next)
      }
    })
    return canvas as unknown as HTMLCanvasElement
  }) as typeof document.createElement)

  return probe
}

function imageFile(name = 'shot.png', type = 'image/png'): File {
  return new File([new Uint8Array(4096)], name, { type })
}

afterEach(() => {
  globalThis.FileReader = originalFileReader
  globalThis.Image = originalImage
  vi.restoreAllMocks()
})

describe('resizeImageForUpload', () => {
  it('把宽图居中裁成正方形并输出 WebP', async () => {
    const probe = installMocks({ naturalWidth: 3000, naturalHeight: 2000 })

    const result = await resizeImageForUpload(imageFile('photo.png'), {
      square: 512,
      maxBytes: 200 * 1024
    })

    expect(probe.widths[0]).toBe(512)
    expect(probe.heights[0]).toBe(512)
    // 9 参数绘制 = 从原图中心取 2000×2000 的正方形贴到输出画布（首参是源 Image）
    expect(probe.drawCalls[0].slice(1)).toEqual([500, 0, 2000, 2000, 0, 0, 512, 512])
    expect(result.name).toBe('photo.webp')
    expect(result.type).toBe('image/webp')
    expect(result.size).toBe(1024)
  })

  it('等比缩放不裁剪（轮播图场景）', async () => {
    const probe = installMocks({ naturalWidth: 4000, naturalHeight: 3000 })

    await resizeImageForUpload(imageFile('banner.jpg', 'image/jpeg'), {
      maxWidth: 1600,
      maxBytes: 200 * 1024
    })

    expect(probe.widths[0]).toBe(1600)
    expect(probe.heights[0]).toBe(1200)
    expect(probe.drawCalls[0].slice(1)).toEqual([0, 0, 1600, 1200])
  })

  it('体积超标时继续降画质、降尺寸', async () => {
    // 4 档画质都超标 → 尺寸缩到 0.8 倍时才达标
    const probe = installMocks({ naturalWidth: 3000, naturalHeight: 2000, blobSizes: [9000, 9000, 9000, 9000, 100] })

    const result = await resizeImageForUpload(imageFile('big.png'), {
      square: 512,
      maxBytes: 500
    })

    expect(probe.widths).toContain(Math.round(512 * 0.8))
    expect(result.size).toBe(100)
  })

  it('canvas 不可用时原样返回，不阻断上传', async () => {
    installMocks({ noContext: true })
    const file = imageFile('plain.png')

    const result = await resizeImageForUpload(file, { square: 512 })

    expect(result).toBe(file)
  })

  it('图片无法解码时原样返回', async () => {
    installMocks({ decodeFails: true })
    const file = imageFile('broken.png')

    const result = await resizeImageForUpload(file, { square: 512 })

    expect(result).toBe(file)
  })
})
