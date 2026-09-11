import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { Router } from 'vue-router'
import {
  buildChunkReloadUrl,
  CHUNK_RELOAD_QUERY,
  CHUNK_RELOAD_STORAGE_KEY,
  clearChunkReloadAttempt,
  clearChunkReloadQuery,
  isChunkLoadError,
  reloadForChunkError
} from '../chunkLoadRecovery'

function createRouterStub(href: string): Router {
  return {
    resolve: vi.fn(() => ({ href }))
  } as unknown as Router
}

describe('chunk load recovery', () => {
  beforeEach(() => {
    window.history.replaceState({}, '', '/')
    window.sessionStorage.clear()
    clearChunkReloadAttempt()
  })

  afterEach(() => {
    window.history.replaceState({}, '', '/')
  })

  it('识别动态模块和 chunk 加载错误', () => {
    expect(isChunkLoadError(new Error('Failed to fetch dynamically imported module'))).toBe(true)
    expect(isChunkLoadError(new Error('Loading CSS chunk 123 failed'))).toBe(true)

    const error = new Error('Loading chunk failed')
    error.name = 'ChunkLoadError'
    expect(isChunkLoadError(error)).toBe(true)
    expect(isChunkLoadError(new Error('Request failed'))).toBe(false)
  })

  it('为目标路由添加缓存刷新参数并保留原查询和 hash', () => {
    const url = buildChunkReloadUrl('/login?redirect=%2Fdashboard#form', 12345)

    expect(url).toContain('/login?redirect=%2Fdashboard')
    expect(url).toContain(`${CHUNK_RELOAD_QUERY}=12345`)
    expect(url).toContain('#form')
  })

  it('chunk 错误重载目标路由，并阻止短时间内重复重载', () => {
    const router = createRouterStub('/login?redirect=%2Fdashboard')
    const reload = vi.fn()

    expect(reloadForChunkError(router, '/login', 1_000, reload)).toBe(true)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(reload.mock.calls[0][0]).toContain('/login?redirect=%2Fdashboard')
    expect(reload.mock.calls[0][0]).toContain(`${CHUNK_RELOAD_QUERY}=1000`)
    expect(window.sessionStorage.getItem(CHUNK_RELOAD_STORAGE_KEY)).toBe('1000')

    expect(reloadForChunkError(router, '/login', 5_000, reload)).toBe(false)
    expect(reload).toHaveBeenCalledTimes(1)
    expect(reloadForChunkError(router, '/login', 11_001, reload)).toBe(true)
    expect(reload).toHaveBeenCalledTimes(2)
  })

  it('成功加载后清除 URL 标记和重载状态', () => {
    window.history.replaceState({}, '', '/login?redirect=%2Fdashboard&__chunk_reload=123#form')
    window.sessionStorage.setItem(CHUNK_RELOAD_STORAGE_KEY, '123')

    clearChunkReloadQuery()
    expect(window.location.pathname).toBe('/login')
    expect(window.location.search).toBe('?redirect=%2Fdashboard')
    expect(window.location.hash).toBe('#form')
    expect(window.sessionStorage.getItem(CHUNK_RELOAD_STORAGE_KEY)).toBe('123')

    clearChunkReloadAttempt()
    expect(window.sessionStorage.getItem(CHUNK_RELOAD_STORAGE_KEY)).toBeNull()
  })
})
