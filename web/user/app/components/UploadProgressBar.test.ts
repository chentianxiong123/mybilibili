import { describe, it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'

vi.mock('element-plus', () => ({
  ElMessage: { info: vi.fn(), success: vi.fn(), error: vi.fn(), warning: vi.fn() },
  ElMessageBox: { confirm: vi.fn() },
}))

import UploadProgressBar from './UploadProgressBar.vue'

const UPLOAD_STAGES = { UPLOADING: 'UPLOADING', COMPLETED: 'COMPLETED', FAILED: 'FAILED' }

function createWrapper(props: Record<string, any> = {}) {
  return mount(UploadProgressBar, {
    props: { UPLOAD_STAGES, ...props },
    global: {
      stubs: {
        'el-dialog': {
          template: '<div v-if="modelValue" class="el-dialog-stub"><div class="dialog-title-stub">{{ title }}</div><slot /><slot name="footer" /></div>',
          props: ['modelValue', 'title', 'width'],
        },
        'el-progress': {
          template: '<div class="el-progress-stub">{{ percentage }}%</div>',
          props: ['percentage', 'status', 'strokeWidth', 'showText'],
        },
        'el-icon': { template: '<span class="el-icon-stub"><slot /></span>' },
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(\'click\')"><slot /></button>',
        },
      },
    },
  })
}

describe('UploadProgressBar', () => {
  it('show=false 时弹窗不渲染内容', () => {
    const w = createWrapper({ show: false })
    expect(w.find('.dialog-title-stub').exists()).toBe(false)
  })

  it('show=true 时渲染弹窗标题', () => {
    const w = createWrapper({ show: true })
    expect(w.find('.dialog-title-stub').text()).toBe('稿件上传进度')
  })

  it('formatFileSize 在 stats 中展示', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      uploadedBytes: 1024,
      totalBytes: 1024 * 1024,
      speed: 512,
    })
    expect(w.text()).toContain('1 KB / 1 MB')
    expect(w.text()).toContain('512 B/s')
  })

  it('formatEta 显示 "秒"', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      etaSeconds: 30,
      uploadedBytes: 0,
      totalBytes: 100,
      speed: 1,
    })
    expect(w.text()).toContain('剩余 30秒')
  })

  it('formatEta 显示 "分秒"', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      etaSeconds: 90,
      uploadedBytes: 0,
      totalBytes: 100,
      speed: 1,
    })
    expect(w.text()).toContain('剩余 1分30秒')
  })

  it('formatEta 显示 "时分"', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      etaSeconds: 3700,
      uploadedBytes: 0,
      totalBytes: 100,
      speed: 1,
    })
    expect(w.text()).toContain('剩余 1时1分')
  })

  it('etaSeconds=0 不显示剩余时间', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      etaSeconds: 0,
      uploadedBytes: 0,
      totalBytes: 100,
      speed: 1,
    })
    expect(w.text()).not.toContain('剩余')
  })

  it('partProgress > 1 时显示分片进度', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      partProgress: [
        { title: 'P1', uploaded: 50, total: 100 },
        { title: 'P2', uploaded: 0, total: 100 },
      ],
      uploadedBytes: 50,
      totalBytes: 200,
      speed: 1,
    })
    expect(w.text()).toContain('P1')
    expect(w.text()).toContain('P2')
  })

  it('partProgress total=0 + length=1 不渲染分片列表', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
      partProgress: [{ title: 'P1', uploaded: 0, total: 0 }],
    })
    expect(w.text()).not.toContain('P1')
  })

  it('error prop 渲染错误文案', () => {
    const w = createWrapper({
      show: true,
      isUploading: false,
      isFinished: true,
      error: '上传失败',
      stage: UPLOAD_STAGES.FAILED,
    })
    expect(w.text()).toContain('上传失败')
  })

  it('isFinished=true 且 COMPLETED 显示提示文案', () => {
    const w = createWrapper({
      show: true,
      isFinished: true,
      stage: UPLOAD_STAGES.COMPLETED,
    })
    expect(w.text()).toContain('稿件已提交')
  })

  it('isUploading=true 显示 "取消上传" 按钮', () => {
    const w = createWrapper({
      show: true,
      isUploading: true,
    })
    expect(w.text()).toContain('取消上传')
  })

  it('isFinished=true 显示 "关闭" 按钮', () => {
    const w = createWrapper({
      show: true,
      isFinished: true,
      stage: UPLOAD_STAGES.COMPLETED,
    })
    expect(w.text()).toContain('关闭')
  })

  it('stageLabel prop 渲染', () => {
    const w = createWrapper({
      show: true,
      stageLabel: '正在上传中…',
    })
    expect(w.text()).toContain('正在上传中…')
  })

  it('百分比 prop 透传到 el-progress', () => {
    const w = createWrapper({
      show: true,
      percentage: 42,
    })
    expect(w.find('.el-progress-stub').text()).toBe('42%')
  })

  it('close 事件在 dialog modelValue=false 时触发', async () => {
    const w = createWrapper({ show: true })
    const vm: any = (w.vm as any)
    vm.dialogVisible = false
    await new Promise(r => setTimeout(r, 0))
    expect(w.emitted('close')).toBeTruthy()
  })
})
