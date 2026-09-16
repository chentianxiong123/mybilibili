import { describe, it, expect, vi, beforeEach } from 'vitest'
import { listDrafts, getDraft, saveDraft, deleteDraft, clearAllDrafts } from './drafts'

describe('drafts', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('listDrafts', () => {
    it('returns empty array when no drafts', () => {
      expect(listDrafts()).toEqual([])
    })

    it('returns drafts sorted by updatedAt descending', () => {
      saveDraft({ title: 'first', updatedAt: 1 })
      saveDraft({ title: 'second', updatedAt: 2 })
      const list = listDrafts()
      expect(list.length).toBe(2)
      expect(list[0].title).toBe('second')
      expect(list[1].title).toBe('first')
    })
  })

  describe('getDraft', () => {
    it('returns null for nonexistent id', () => {
      expect(getDraft('nonexistent')).toBeNull()
    })

    it('returns draft for existing id', () => {
      const draft = saveDraft({ title: 'Test' })
      expect(draft).not.toBeNull()
      expect(getDraft(draft.id)).not.toBeNull()
      expect(getDraft(draft.id).title).toBe('Test')
    })
  })

  describe('saveDraft', () => {
    it('creates a new draft when no id', () => {
      const draft = saveDraft({ title: 'New' })
      expect(draft).not.toBeNull()
      expect(draft.id).toMatch(/^d_/)
      expect(draft.title).toBe('New')
    })

    it('updates existing draft when id provided', () => {
      const draft = saveDraft({ title: 'Original' })
      const updated = saveDraft({ id: draft.id, title: 'Updated' })
      expect(updated.title).toBe('Updated')
    })

    it('truncates to 20 drafts', () => {
      for (let i = 0; i < 25; i++) {
        saveDraft({ title: `draft ${i}` })
      }
      const list = listDrafts()
      expect(list.length).toBe(20)
    })
  })

  describe('deleteDraft', () => {
    it('returns true when deleting existing draft', () => {
      const draft = saveDraft({ title: 'To Delete' })
      expect(deleteDraft(draft.id)).toBe(true)
      expect(getDraft(draft.id)).toBeNull()
    })

    it('returns false when deleting nonexistent draft', () => {
      expect(deleteDraft('nonexistent')).toBe(false)
    })
  })

  describe('clearAllDrafts', () => {
    it('clears all drafts', () => {
      saveDraft({ title: 'A' })
      saveDraft({ title: 'B' })
      expect(listDrafts().length).toBe(2)
      clearAllDrafts()
      expect(listDrafts()).toEqual([])
    })
  })
})
