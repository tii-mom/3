import { readdir, readFile, stat } from 'node:fs/promises'
import { join, relative } from 'node:path'
import { fileURLToPath } from 'node:url'

const frontendRoot = fileURLToPath(new URL('..', import.meta.url))
const workspaceRoot = join(frontendRoot, '..')
const roots = ['src/components', 'src/views', 'src/i18n', 'src/stores', 'src/content', 'index.html']
const publicDocs = [
  'README.md',
  'README_CN.md',
  'README_JA.md',
  'DEV_GUIDE.md',
  'CLA.md',
  'docs',
  'deploy/README.md',
  'deploy/DOCKER.md',
  'deploy/APPLE_CONTAINER.md',
  'deploy/CODEX_POOL_PLAYBOOK.md',
  'deploy/DATAMANAGEMENTD_CN.md',
  'deploy/EDGE_SECURITY.md',
]
const excludedSegments = new Set(['__tests__', 'api', 'types'])
const excludedFiles = new Set(['src/constants/brand.ts'])
const legacyBrandPattern = /sub2api|subscription to api conversion platform/i
const legacyPublicDocsPattern = /\bSub2API\b|Subscription to API Conversion Platform/
const findings = []

async function scanFile(path, pattern = legacyBrandPattern) {
  const relativePath = relative(frontendRoot, path)
  if (excludedFiles.has(relativePath)) return
  const content = await readFile(path, 'utf8')
  if (pattern.test(content)) findings.push(relativePath)
}

async function scan(path, pattern = legacyBrandPattern) {
  const entries = await readdir(path, { withFileTypes: true })
  for (const entry of entries) {
    const child = join(path, entry.name)
    const relativePath = relative(frontendRoot, child)
    if (entry.isDirectory()) {
      if (excludedSegments.has(entry.name)) continue
      await scan(child, pattern)
      continue
    }
    if (excludedFiles.has(relativePath) || /\.(png|jpe?g|gif|svg|woff2?)$/i.test(entry.name)) continue
    await scanFile(child, pattern)
  }
}

for (const root of roots) {
  const path = join(frontendRoot, root)
  if (root.includes('.')) await scanFile(path)
  else await scan(path)
}

for (const root of publicDocs) {
  const path = join(workspaceRoot, root)
  const info = await stat(path)
  if (info.isDirectory()) await scan(path, legacyPublicDocsPattern)
  else await scanFile(path, legacyPublicDocsPattern)
}

if (findings.length > 0) {
  console.error(`Legacy brand found in public frontend files:\n${findings.map((file) => `- ${file}`).join('\n')}`)
  process.exit(1)
}

console.log('Public frontend brand check passed: 3API only.')
