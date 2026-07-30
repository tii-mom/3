import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = fileURLToPath(new URL('..', import.meta.url))
const distRoot = join(frontendRoot, '..', 'backend', 'internal', 'web', 'dist')
const seoPagesConfig = JSON.parse(await readFile(join(frontendRoot, 'src', 'content', 'seo-pages.json'), 'utf8'))
const seoPageSlugs = Object.values(seoPagesConfig).map((page) => page.path.replace(/^\//, ''))

function assert(condition, message) {
  if (!condition) {
    throw new Error(message)
  }
}

async function readDistFile(name) {
  return readFile(join(distRoot, name), 'utf8')
}

const indexHtml = await readDistFile('index.html')
const sitemapXml = await readDistFile('sitemap.xml')
const robotsTxt = await readDistFile('robots.txt')
assert(!indexHtml.includes('id="seo-snapshot"'), 'Homepage index.html must not embed a visible SEO snapshot.')
assert(!indexHtml.includes('id="seo-snapshot-style"'), 'Homepage index.html must not include SEO snapshot styles.')
assert(indexHtml.includes('<div id="app"></div>'), 'Homepage index.html must remain a clean Vue app shell.')
assert(indexHtml.includes('<title>3API - AI API 中转站与多模型统一接入</title>'), 'Homepage title must use 3API.')
assert(!/Sub2API|Subscription to API Conversion Platform/i.test(indexHtml), 'Homepage must not expose legacy Sub2API branding.')
assert(robotsTxt.includes('Disallow: /console'), 'robots.txt must disallow console routes.')
assert(robotsTxt.includes('Disallow: /payment'), 'robots.txt must disallow payment routes.')

for (const page of seoPageSlugs) {
  const html = await readDistFile(`${page}.html`)
  assert(html.includes('id="seo-snapshot"'), `${page}.html must include a crawlable SEO snapshot.`)
  assert(html.includes('3API'), `${page}.html must include 3API branding.`)
  assert(!/Sub2API|Subscription to API Conversion Platform/i.test(html), `${page}.html must not expose legacy Sub2API branding.`)
  assert(html.includes('application/ld+json'), `${page}.html must include structured data.`)
  assert(sitemapXml.includes(`https://3api.shop/${page}`), `sitemap.xml must include ${page}.`)
}

console.log('Built SEO files check passed: clean homepage shell, crawlable topic snapshots, 3API public brand.')
