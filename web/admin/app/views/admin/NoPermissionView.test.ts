import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import NoPermissionView from './NoPermissionView.vue'

const logoutMock = vi.hoisted(() => vi.fn())
vi.mock('@/stores/admin', async () => {
  const { defineStore } = await vi.importActual<any>('pinia')
  return {
    useAdminStore: defineStore('admin-no-permission-test', {
      state: () => ({ role: '', permissions: [] as string[], userInfo: null }),
      actions: { logout: logoutMock },
    }),
  }
})

const makeWrapper = () => {
  const pinia = createPinia()
  setActivePinia(pinia)
  return mount(NoPermissionView, {
    global: {
      plugins: [pinia],
      stubs: {
        'el-empty': { template: '<div class="el-empty-stub"><slot /></div>' },
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
          emits: ['click'],
        },
      },
    },
  })
}

beforeEach(() => {
  vi.clearAllMocks()
  logoutMock.mockReset()
  logoutMock.mockResolvedValue(undefined)
})

describe('NoPermissionView.vue', () => {
  it('渲染空状态与退出登录按钮', () => {
    const w = makeWrapper()
    expect(w.text()).toContain('退出登录')
    expect(w.find('button').exists()).toBe(true)
  })

  it('点击退出登录 → 交给 store 处理（store 内部负责清凭证与跳 /login）', async () => {
    const w = makeWrapper()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(logoutMock).toHaveBeenCalledTimes(1)
  })

  it('组件自身不直接操作 router，仅委托 store', async () => {
    const w = makeWrapper()
    await w.vm.handleLogout()
    expect(logoutMock).toHaveBeenCalled()
  })

  it('store logout 抛错时错误向上冒泡，不被组件静默吞掉', async () => {
    logoutMock.mockRejectedValueOnce(new Error('network'))
    const w = makeWrapper()
    await expect(w.vm.handleLogout()).rejects.toThrow('network')
  })
})
