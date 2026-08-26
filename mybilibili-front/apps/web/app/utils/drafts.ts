/**
 * 投稿草稿 - 纯客户端 localStorage 持久化（与 wap 端共用同一存储结构）
 *
 * 草稿保存投稿发布中途填写的表单数据，用户可稍后继续编辑或删除。
 * 不依赖后端接口（后端无草稿接口）。
 *
 * 注意：视频 File 对象无法序列化进 localStorage，因此草稿中仅保存
 * 分集元信息（标题/排序）。继续编辑时需重新选择本地视频文件。
 */

import { safeStorage } from '@/utils/safeStorage'

export interface DraftVideoPart {
  title: string
  description?: string
  videoId?: number
  playUrl?: string
  duration?: number
  sortOrder: number
}

export interface ManuscriptDraft {
  id: string
  title: string
  categoryId: number | null
  tags: string[]
  description: string
  type: string
  coverPreview?: string
  videoParts: DraftVideoPart[]
  hasLocalVideoFiles: boolean
  updatedAt: number
  createdAt: number
}

const DRAFTS_KEY = 'manuscript:drafts'
const DRAFT_MAX = 20

function loadAll(): ManuscriptDraft[] {
  try {
    const raw = safeStorage.getItem(DRAFTS_KEY)
    if (!raw) return []
    const list = JSON.parse(raw)
    return Array.isArray(list) ? list : []
  } catch (e) {
    console.error('读取草稿失败:', e)
    return []
  }
}

function saveAll(list: ManuscriptDraft[]) {
  try {
    safeStorage.setItem(DRAFTS_KEY, JSON.stringify(list))
  } catch (e) {
    console.error('保存草稿失败:', e)
  }
}

function genId(): string {
  return 'd_' + Date.now().toString(36) + '_' + Math.random().toString(36).slice(2, 8)
}

/** 列出所有草稿（按更新时间倒序） */
export function listDrafts(): ManuscriptDraft[] {
  return loadAll().sort((a, b) => b.updatedAt - a.updatedAt)
}

/** 读取单个草稿 */
export function getDraft(id: string): ManuscriptDraft | null {
  return loadAll().find((d) => d.id === id) || null
}

/** 新增或更新草稿。传 id 则更新，不传则新建 */
export function saveDraft(draft: Partial<ManuscriptDraft> & { id?: string }): ManuscriptDraft | null {
  try {
    let list = loadAll()
    const now = Date.now()
    if (draft.id) {
      const idx = list.findIndex((d) => d.id === draft.id)
      if (idx >= 0) {
        list[idx] = { ...list[idx], ...draft, updatedAt: now } as ManuscriptDraft
        saveAll(list)
        return list[idx]
      }
    }
    const newDraft: ManuscriptDraft = {
      id: genId(),
      title: '',
      categoryId: null,
      tags: [],
      description: '',
      type: 'original',
      videoParts: [],
      hasLocalVideoFiles: false,
      createdAt: now,
      updatedAt: now,
      ...draft
    } as ManuscriptDraft
    list.unshift(newDraft)
    if (list.length > DRAFT_MAX) list = list.slice(0, DRAFT_MAX)
    saveAll(list)
    return newDraft
  } catch (e) {
    console.error('保存草稿失败:', e)
    return null
  }
}

/** 删除单个草稿 */
export function deleteDraft(id: string): boolean {
  const list = loadAll()
  const next = list.filter((d) => d.id !== id)
  if (next.length === list.length) return false
  saveAll(next)
  return true
}

/** 清空所有草稿 */
export function clearAllDrafts(): void {
  try {
    safeStorage.removeItem(DRAFTS_KEY)
  } catch (e) {
    console.error('清空草稿失败:', e)
  }
}
