/**
 * 投稿草稿 - 纯客户端 localStorage 持久化
 *
 * 草稿保存投稿发布中途填写的表单数据，用户可稍后继续编辑或删除。
 * 不依赖后端接口（后端无草稿接口）。视频/封面等大文件无法序列化，
 * 仅保存文本字段与已上传分集的元信息。
 */

import storage, { K } from './storage_layer'

export interface DraftVideoPart {
  title: string
  description?: string
  // 已上传到服务端的分集会带 videoId / playUrl
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
  type: string // 'original' | 'reprint' ...
  // 封面仅保存预览 base64（体积可控）
  coverPreview?: string
  // 已上传分集的元信息（不含 File 对象）
  videoParts: DraftVideoPart[]
  // 仅标记：是否有未上传的本地视频文件需重新选择
  hasLocalVideoFiles: boolean
  updatedAt: number
  createdAt: number
}

const DRAFTS_KEY = 'manuscript:drafts'
const DRAFT_MAX = 20

function loadAll(): ManuscriptDraft[] {
  const list = storage.get<ManuscriptDraft[]>(DRAFTS_KEY)
  return Array.isArray(list) ? list : []
}

function saveAll(list: ManuscriptDraft[]) {
  storage.set(DRAFTS_KEY, list)
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
    // 新建
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
  storage.remove(DRAFTS_KEY)
}

/** 草稿数量 */
export function draftCount(): number {
  return loadAll().length
}

// 暴露 storage key 供外部引用（保持与 K 风格一致）
export const DRAFT_STORAGE_KEY = DRAFTS_KEY

// 避免 TS 报 K 未使用（K 可能被其他地方引用模式使用）
void K
