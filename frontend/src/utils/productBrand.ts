/**
 * 商品品牌 logo 解析器（前端方案，不改后端 / 不改库结构）。
 *
 * 商城系统商品原本只显示一个文字品类 chip，或空图时回退成突兀的「3」字样。
 * 这里按品类映射一个品牌底 + 内联 SVG 字形，渲染成精致的「品牌图标方块」，
 * 直接消除「图标不高级、突兀」的问题。若后台给商品配了 image_url，调用方优先用图。
 *
 * 字形统一用**准确品牌符号**：
 * - GPT 系（代充 / 成品号 / 使用服务 / Codex，含后台自建 gpt_*）→ 官方 OpenAI 结线稿图片
 * - Gemini  → 官方四角星 spark，自带多彩渐变（凹边星，非直线星）
 * - X       → X 官方 logo
 * 单色品牌用「品牌色方块 + 白色字形」，多彩品牌（Gemini）用「浅色方块 + 多彩字形」。
 */
import type { PublicProduct } from '@/api/publicShop'
import { normalizeShopCategory, type ShopCategory } from '@/constants/shop'

export interface BrandGlyph {
  /** 'fill' = 用 currentColor 填充 paths；'stroke' = 用 currentColor 描边 paths */
  mode: 'fill' | 'stroke'
  paths: string[]
  /** 描边宽度（仅 stroke 模式生效，默认 1.6） */
  strokeWidth?: number
  /**
   * 多彩品牌专用：字形自带渐变填充，替代 currentColor（如 Gemini 四角星）。
   * 坐标用 objectBoundingBox（0%~100%）。
   */
  gradient?: {
    x1: string
    y1: string
    x2: string
    y2: string
    stops: Array<{ offset: string; color: string }>
  }
}

export interface BrandMeta {
  key: string
  label: string
  /** logo 方块的 CSS 背景 */
  gradient: string
  /** 方块内字形颜色（默认 #fff）；浅底彩色 logo 可显式覆盖 */
  glyphColor?: string
  /** 浅色底方块（如 Gemini）：需要一层内描边才能在白色卡片上看出边界 */
  tileLight?: boolean
  /**
   * 图片型 logo（相对 public/ 的路径）。
   * 有值时优先于内联 SVG 字形渲染 —— OpenAI 结使用官方线稿图片。
   */
  image?: string
  glyph: BrandGlyph
}

// OpenAI 官方「结」形标志（ChatGPT）
const OPENAI_KNOT =
  'M22.2819 9.8211a5.9847 5.9847 0 0 0-.5157-4.9108 6.0462 6.0462 0 0 0-6.5098-2.9A6.0651 6.0651 0 0 0 4.9807 4.1818a5.9847 5.9847 0 0 0-3.9977 2.9 6.0462 6.0462 0 0 0 .7427 7.0966 5.98 5.98 0 0 0 .511 4.9107 6.051 6.051 0 0 0 6.5146 2.9001A5.9847 5.9847 0 0 0 13.2599 24a6.0557 6.0557 0 0 0 5.7718-4.2058 5.9894 5.9894 0 0 0 3.9977-2.9001 6.0557 6.0557 0 0 0-.7475-7.0729zm-9.022 12.6081a4.4755 4.4755 0 0 1-2.8764-1.0408l.1419-.0804 4.7783-2.7582a.7948.7948 0 0 0 .3927-.6813v-6.7369l2.02 1.1686a.071.071 0 0 1 .038.052v5.5826a4.504 4.504 0 0 1-4.4945 4.4944zm-9.6607-4.1254a4.4708 4.4708 0 0 1-.5346-3.0137l.142.0852 4.783 2.7582a.7712.7712 0 0 0 .7806 0l5.8428-3.3685v2.3324a.0804.0804 0 0 1-.0332.0615L9.74 19.9502a4.4992 4.4992 0 0 1-6.1408-1.6464zM2.3408 7.8956a4.485 4.485 0 0 1 2.3655-1.9728V11.6a.7664.7664 0 0 0 .3879.6765l5.8144 3.3543-2.0201 1.1685a.0757.0757 0 0 1-.071 0l-4.8303-2.7865A4.504 4.504 0 0 1 2.3408 7.872zm16.5963 3.8558L13.1038 8.364 15.1192 7.2a.0757.0757 0 0 1 .071 0l4.8303 2.7913a4.4944 4.4944 0 0 1-.6765 8.1042v-5.6772a.79.79 0 0 0-.407-.667zm2.0107-3.0231l-.142-.0852-4.7735-2.7818a.7759.7759 0 0 0-.7854 0L9.409 9.2297V6.8974a.0662.0662 0 0 1 .0284-.0615l4.8303-2.7866a4.4992 4.4992 0 0 1 6.6802 4.66zM8.3065 12.863l-2.02-1.1638a.0804.0804 0 0 1-.038-.0567V6.0742a4.4992 4.4992 0 0 1 7.3757-3.4537l-.142.0805L8.704 5.459a.7948.7948 0 0 0-.3927.6813zm1.0976-2.3654l2.602-1.4998 2.6069 1.4998v2.9994l-2.5974 1.4997-2.6067-1.4997Z'

// Gemini 官方「四角星 / spark」：凹边曲线星（不是直线四角星）
const GEMINI_SPARK =
  'M12 0C12 6.627 6.627 12 0 12c6.627 0 12 5.373 12 12 0-6.627 5.373-12 12-12-6.627 0-12-5.373-12-12z'

/**
 * public/ 下静态资源的实际 URL。
 * 用 BASE_URL 而不是硬编码 '/'，这样部署到子路径时也不会 404。
 */
const publicAsset = (path: string) => `${import.meta.env.BASE_URL}${path}`.replace(/\/{2,}/g, '/')

const BRANDS: Record<string, BrandMeta> = {
  chatgpt: {
    key: 'chatgpt',
    label: 'ChatGPT',
    // 白底（用户要求：不要用绿色）；配官方 OpenAI 结线稿图片
    gradient: 'linear-gradient(140deg, #ffffff 0%, #f4f5f7 100%)',
    tileLight: true,
    glyphColor: '#0a0a0a',
    image: publicAsset('logos/openai-knot.png'),
    glyph: {
      mode: 'fill',
      paths: [OPENAI_KNOT]
    }
  },
  x: {
    key: 'x',
    label: 'X',
    gradient: 'linear-gradient(140deg, #111827 0%, #000000 100%)',
    glyph: {
      mode: 'fill',
      paths: [
        'M18.244 2.25h3.308l-7.227 8.26 8.502 11.24h-6.656l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z'
      ]
    }
  },
  gemini: {
    key: 'gemini',
    label: 'Gemini',
    // 多彩品牌：浅底 + 多彩星（对齐官方图标观感）
    gradient: 'linear-gradient(140deg, #ffffff 0%, #e9edf5 100%)',
    tileLight: true,
    glyph: {
      mode: 'fill',
      paths: [GEMINI_SPARK],
      gradient: {
        x1: '0%',
        y1: '0%',
        x2: '100%',
        y2: '100%',
        stops: [
          { offset: '0%', color: '#EA4335' },
          { offset: '30%', color: '#9B72CB' },
          { offset: '55%', color: '#4285F4' },
          { offset: '80%', color: '#34A853' },
          { offset: '100%', color: '#FBBC04' }
        ]
      }
    }
  },
  codex: {
    key: 'codex',
    label: 'Codex',
    gradient: 'linear-gradient(140deg, #1f2937 0%, #0f172a 100%)',
    glyph: {
      mode: 'stroke',
      paths: [
        'M9 7l-4 5 4 5',
        'M15 7l4 5-4 5'
      ]
    }
  },
  generic: {
    key: 'generic',
    label: '服务',
    gradient: 'linear-gradient(140deg, #6b7280 0%, #9ca3af 100%)',
    glyph: {
      mode: 'fill',
      paths: [
        'M12 4l1.6 4.4L18 10l-4.4 1.6L12 16l-1.6-4.4L6 10l4.4-1.6z'
      ]
    }
  }
}

/**
 * 品类 → 品牌。
 *
 * GPT 系（代充 / 成品号 / 使用服务 / Codex，以及后台新建的 gpt_* 品类）
 * 统一使用 OpenAI 结图标；其余按品类取对应品牌，未知品类兜底 generic。
 */
export function productBrandMeta(product: PublicProduct): BrandMeta {
  const cat = normalizeShopCategory(product.category) as ShopCategory
  switch (cat) {
    case 'x_premium':
      return BRANDS.x
    case 'gemini':
      return BRANDS.gemini
    default:
      break
  }
  // gpt 前缀覆盖后台自建品类（如 gpt_usage、gpt_api），避免新增品类掉进 generic
  if (cat.startsWith('gpt') || cat === 'codex') {
    return BRANDS.chatgpt
  }
  return BRANDS.generic
}

/** 后台是否配了真实商品图（优先于品牌字形使用）。 */
export function hasProductImage(product: PublicProduct): boolean {
  return !!product.image_url
}
