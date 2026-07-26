export const PRODUCT_NAME = '3API'
export const DEFAULT_SITE_SUBTITLE = 'AI API gateway for unified model access'

/** Keep the historical source default from leaking into public UI. */
export function normalizeSiteName(value: string | null | undefined): string {
  const name = String(value ?? '').trim()
  return !name || name.toLowerCase() === 'sub2api' ? PRODUCT_NAME : name
}

/** Normalize the old generated subtitle while preserving custom operator copy. */
export function normalizeSiteSubtitle(value: string | null | undefined): string {
  const subtitle = String(value ?? '').trim()
  return subtitle.toLowerCase() === 'subscription to api conversion platform' ? DEFAULT_SITE_SUBTITLE : subtitle
}
