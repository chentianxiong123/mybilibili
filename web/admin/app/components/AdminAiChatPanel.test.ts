import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import AdminAiChatPanel from './AdminAiChatPanel.vue'

const sendMessage = vi.hoisted(() => vi.fn())
vi.mock('~/api/adminAi', () => ({ adminAiApi: { sendMessage } }))

const errorMock = vi.hoisted(() => vi.fn())
vi.mock('element-plus', () => ({
  ElMessage: { error: errorMock, success: vi.fn(), warning: vi.fn() },
}))

vi.mock('@element-plus/icons-vue', () => {
  const stub = { template: '<i />' }
  return {
    ChatDotRound: stub,
    Close: stub,
    Tools: stub,
    Download: stub,
    TrendCharts: stub,
    DataLine: stub,
    PieChart: stub,
  }
})

const makeWrapper = () =>
  mount(AdminAiChatPanel, {
    props: { visible: true },
    global: {
      stubs: {
        'el-button': {
          template: '<button class="el-button-stub" @click="$emit(`click`)"><slot /></button>',
          emits: ['click'],
        },
        'el-input': {
          template:
            '<textarea class="el-input-stub" :value="modelValue" @input="e=>$emit(`update:modelValue`, e.target.value)" @keydown="e=>$emit(`keydown`, e)" />',
          props: ['modelValue'],
        },
        'el-icon': { template: '<span><slot /></span>' },
        'el-table': { template: '<div class="el-table-stub"><slot /></div>' },
        'el-table-column': { template: '<div class="el-table-column-stub"><slot /></div>' },
        'el-drawer': { template: '<div class="el-drawer-stub"><slot /></div>' },
      },
    },
  })

/** 拿到 sendMessage 的回调集合，模拟服务端推送 */
function captureCallbacks() {
  const cbs: any = {}
  sendMessage.mockImplementation((_content: string, handlers: any) => {
    Object.assign(cbs, handlers)
    return { abort: vi.fn() }
  })
  return cbs
}

beforeEach(() => {
  vi.clearAllMocks()
  errorMock.mockReset()
  sendMessage.mockReset()
  sendMessage.mockImplementation(() => ({ abort: vi.fn() }))
})

describe('AdminAiChatPanel.vue — 发送消息', () => {
  it('内容为空或纯空白：不发送请求', () => {
    const w = makeWrapper()
    w.vm.inputText = ''
    w.vm.handleSend()
    w.vm.inputText = '   '
    w.vm.handleSend()
    expect(sendMessage).not.toHaveBeenCalled()
  })

  it('发送成功：清空输入框、推入用户消息、进入流式态', () => {
    const w = makeWrapper()
    w.vm.inputText = '  查一下投稿量  '
    w.vm.handleSend()
    expect(sendMessage).toHaveBeenCalledTimes(1)
    expect(sendMessage.mock.calls[0][0]).toBe('查一下投稿量')
    expect(w.vm.inputText).toBe('')
    expect(w.vm.isStreaming).toBe(true)
    expect(w.vm.messages).toHaveLength(1)
    expect(w.vm.messages[0]).toMatchObject({ role: 'user', content: '查一下投稿量' })
  })

  it('流式过程中重复发送被忽略', () => {
    const w = makeWrapper()
    w.vm.inputText = 'a'
    w.vm.handleSend()
    w.vm.inputText = 'b'
    w.vm.handleSend()
    expect(sendMessage).toHaveBeenCalledTimes(1)
    expect(w.vm.messages).toHaveLength(1)
  })

  it('onData 累加流式内容，并通过 displayMessages 合并出一条临时消息', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onData('你')
    cbs.onData('好')
    expect(w.vm.streamingContent).toBe('你好')
    const display = w.vm.displayMessages
    expect(display).toHaveLength(2)
    expect(display[1]).toMatchObject({ id: 'streaming', role: 'assistant', content: '你好' })
  })

  it('onToolCall：置起工具调用态并记录工具名', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onToolCall('query_db')
    expect(w.vm.toolCalling).toBe(true)
    expect(w.vm.toolCallingName).toBe('query_db')
  })

  it('onToolCall 缺工具名时回落"处理中"', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onToolCall('')
    expect(w.vm.toolCallingName).toBe('处理中')
  })

  it('onDone 收到字符串：直接作为回复并结束流式', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onDone('纯文本回复')
    expect(w.vm.messages).toHaveLength(2)
    expect(w.vm.messages[1]).toMatchObject({ role: 'assistant', content: '纯文本回复' })
    expect(w.vm.isStreaming).toBe(false)
    expect(w.vm.streamingContent).toBe('')
    expect(w.vm.toolCalling).toBe(false)
  })

  it('onDone 收到对象：兼容 content 与 reply 两种字段', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'a'
    w.vm.handleSend()
    cbs.onDone({ content: 'C' })
    w.vm.inputText = 'b'
    w.vm.handleSend()
    cbs.onDone({ reply: 'R' })
    expect(w.vm.messages[1].content).toBe('C')
    expect(w.vm.messages[3].content).toBe('R')
  })

  it('onDone 的 render 为 JSON 字符串时解析成对象', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onDone({ content: 'x', render: '{"type":"table","data":{"rows":[{"a":1}]}}' })
    expect(w.vm.messages[1].render).toEqual({ type: 'table', data: { rows: [{ a: 1 }] } })
  })

  it('onDone 的 render 已是对象时原样使用', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onDone({ content: 'x', render: { type: 'pie' } })
    expect(w.vm.messages[1].render).toEqual({ type: 'pie' })
  })

  it('render 是非法 JSON 字符串时兜底保留原始字符串，不抛异常', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    expect(() => cbs.onDone({ content: 'x', render: '{not json' })).not.toThrow()
    expect(w.vm.messages[1].render).toBe('{not json')
  })

  it('onError：提示错误并复位流式与工具调用状态', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onToolCall('t')
    cbs.onError('后端炸了')
    expect(errorMock).toHaveBeenCalledWith('后端炸了')
    expect(w.vm.isStreaming).toBe(false)
    expect(w.vm.streamingContent).toBe('')
    expect(w.vm.toolCalling).toBe(false)
  })

  it('onError 未带错误信息时回落通用文案', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onError()
    expect(errorMock).toHaveBeenCalledWith('回复失败，请重试')
  })

  it('结束后可以再次发送', () => {
    const cbs = captureCallbacks()
    const w = makeWrapper()
    w.vm.inputText = 'a'
    w.vm.handleSend()
    cbs.onDone('ok')
    w.vm.inputText = 'b'
    w.vm.handleSend()
    expect(sendMessage).toHaveBeenCalledTimes(2)
  })
})

describe('AdminAiChatPanel.vue — 关闭与生命周期', () => {
  it('关闭面板：中断 EventSource 并把 visible 置 false', async () => {
    const abort = vi.fn()
    sendMessage.mockImplementation(() => ({ abort }))
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    expect(w.vm.eventSource).not.toBeNull()
    w.vm.handleClose()
    expect(abort).toHaveBeenCalled()
    expect(w.vm.eventSource).toBeNull()
    expect(w.vm.visible).toBe(false)
  })

  it('visible 变 false：中断流式并清空所有瞬时状态', async () => {
    const abort = vi.fn()
    const cbs: any = {}
    sendMessage.mockImplementation((_c: string, h: any) => {
      Object.assign(cbs, h)
      return { abort }
    })
    const w = makeWrapper()
    w.vm.inputText = 'hi'
    w.vm.handleSend()
    cbs.onToolCall('t')
    cbs.onData('partial')
    await w.setProps({ visible: false })
    expect(abort).toHaveBeenCalled()
    expect(w.vm.isStreaming).toBe(false)
    expect(w.vm.streamingContent).toBe('')
    expect(w.vm.toolCalling).toBe(false)
  })

  it('visible 变 true：不做任何中断动作', async () => {
    const abort = vi.fn()
    sendMessage.mockImplementation(() => ({ abort }))
    const w = makeWrapper()
    await w.setProps({ visible: true })
    expect(abort).not.toHaveBeenCalled()
  })

  it('Enter 发送、Shift+Enter 换行', () => {
    const w = makeWrapper()
    w.vm.inputText = 'x'
    w.vm.handleKeydown({ key: 'Enter', shiftKey: false, preventDefault: vi.fn() })
    expect(sendMessage).toHaveBeenCalledTimes(1)

    w.vm.inputText = 'y'
    w.vm.handleKeydown({ key: 'Enter', shiftKey: true, preventDefault: vi.fn() })
    expect(sendMessage).toHaveBeenCalledTimes(1)

    w.vm.inputText = 'z'
    w.vm.handleKeydown({ key: 'a', shiftKey: false, preventDefault: vi.fn() })
    expect(sendMessage).toHaveBeenCalledTimes(1)
  })
})

describe('AdminAiChatPanel.vue — render 解析辅助函数', () => {
  it('getTableRows：数组直接返回，对象取第一个数组字段', () => {
    const w = makeWrapper()
    expect(w.vm.getTableRows(null)).toEqual([])
    expect(w.vm.getTableRows('str')).toEqual([])
    expect(w.vm.getTableRows([{ a: 1 }])).toEqual([{ a: 1 }])
    expect(w.vm.getTableRows({ total: 3, rows: [{ a: 1 }] })).toEqual([{ a: 1 }])
    expect(w.vm.getTableRows({ total: 3 })).toEqual([])
  })

  it('getTableColumns：从首行推断列，空行返回空数组', () => {
    const w = makeWrapper()
    expect(w.vm.getTableColumns([])).toEqual([])
    expect(w.vm.getTableColumns(null)).toEqual([])
    expect(w.vm.getTableColumns([{ id: 1, name: 'x' }])).toEqual([
      { prop: 'id', label: 'id' },
      { prop: 'name', label: 'name' },
    ])
  })

  it('formatNumberStat：数字加千分位，非数字转字符串', () => {
    const w = makeWrapper()
    expect(w.vm.formatNumberStat(null)).toBeNull()
    expect(w.vm.formatNumberStat('x')).toBeNull()
    expect(w.vm.formatNumberStat({ total: 1234567, label: '稿件' })).toEqual([
      { label: 'total', value: (1234567).toLocaleString() },
      { label: 'label', value: '稿件' },
    ])
  })

  it('getChartIcon：未知类型回落到 DataLine', () => {
    const w = makeWrapper()
    expect(w.vm.getChartIcon('unknown')).toBeTruthy()
    expect(w.vm.getChartIcon('line')).toBeTruthy()
    expect(w.vm.getChartIcon('pie')).toBeTruthy()
    expect(w.vm.getChartIcon('table')).toBeTruthy()
  })
})
