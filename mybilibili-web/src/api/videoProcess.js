import { globalManager, TaskType, TaskConfig } from '../utils/progressSimulator.js'

class VideoProcessService {
  constructor() {
    this.sseConnections = new Map()
    this.taskStates = new Map()
  }

  startProcess(taskId, taskType, callbacks = {}) {
    const {
      onProgress = () => {},
      onComplete = () => {},
      onError = () => {}
    } = callbacks

    if (globalManager.has(taskId)) {
      const existing = globalManager.get(taskId)
      if (existing.getStatus() === 'running') {
        return existing
      }
    }

    const simulator = globalManager.create(taskId, { taskType })

    this.taskStates.set(taskId, {
      taskType,
      progress: 0,
      status: 'running',
      startTime: Date.now()
    })

    simulator.start({
      onProgress: (progress) => {
        const state = this.taskStates.get(taskId)
        if (state) {
          state.progress = progress
        }
        onProgress(progress, {
          taskType,
          isSlowDown: progress >= 90,
          status: progress >= 90 ? '即将完成' : '处理中'
        })
      },
      onComplete: () => {
        const state = this.taskStates.get(taskId)
        if (state) {
          state.status = 'completed'
          state.progress = 100
        }
        this.closeSSE(taskId)
        onComplete()
      }
    })

    this.connectSSE(taskId, {
      onComplete: () => {
        this.forceComplete(taskId)
        onComplete()
      },
      onError: (error) => {
        console.warn(`[SSE] 连接错误，继续使用模拟进度: ${error}`)
        onError(error)
      }
    })

    return simulator
  }

  connectSSE(taskId, callbacks = {}) {
    const { onComplete = () => {}, onError = () => {} } = callbacks

    if (this.sseConnections.has(taskId)) {
      return
    }

    const baseURL = window.location.origin
    const url = `${baseURL}/api/video/process/sse/${taskId}`
    const token = localStorage.getItem('token')

    const controller = new AbortController()

    fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': token ? `Bearer ${token}` : '',
        'Accept': 'text/event-stream'
      },
      signal: controller.signal
    })
    .then(response => {
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`)
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      const readStream = () => {
        reader.read().then(({ done, value }) => {
          if (done) {
            return
          }

          buffer += decoder.decode(value, { stream: true })

          const lines = buffer.split('\n\n')
          buffer = lines.pop() || ''

          lines.forEach(line => {
            const event = this._parseSSEEvent(line)
            if (event) {
              this._handleSSEEvent(taskId, event, { onComplete })
            }
          })

          readStream()
        }).catch(error => {
          if (error.name !== 'AbortError') {
            console.warn('[SSE] 读取流错误:', error)
          }
        })
      }

      readStream()
    })
    .catch(error => {
      if (error.name !== 'AbortError') {
        onError(error.message)
      }
    })

    this.sseConnections.set(taskId, controller)
  }

  _parseSSEEvent(raw) {
    const lines = raw.split('\n')
    let event = ''
    let data = ''

    lines.forEach(line => {
      if (line.startsWith('event:')) {
        event = line.substring(6).trim()
      } else if (line.startsWith('data:')) {
        data = line.substring(5).trim()
      }
    })

    if (!event && !data) return null
    return { event: event || 'message', data }
  }

  _handleSSEEvent(taskId, { event, data }, callbacks = {}) {
    const { onComplete = () => {} } = callbacks

    switch (event) {
      case 'complete':
      case 'done':
        this.forceComplete(taskId)
        onComplete()
        break
      case 'progress':
        try {
          const parsed = JSON.parse(data)
          if (parsed.done || parsed.type === 'complete' || parsed.eventName === 'complete' || parsed.status === 5) {
            this.forceComplete(taskId)
            onComplete()
          }
        } catch (e) {
          console.warn('[SSE] 解析进度数据失败:', e)
        }
        break
      case 'error':
        console.warn('[SSE] 收到错误事件:', data)
        break
    }
  }

  forceComplete(taskId) {
    const simulator = globalManager.get(taskId)
    if (simulator && !simulator.isCompleted) {
      simulator.forceComplete()
    }

    const state = this.taskStates.get(taskId)
    if (state) {
      state.status = 'completed'
      state.progress = 100
    }

    this.closeSSE(taskId)
  }

  closeSSE(taskId) {
    const controller = this.sseConnections.get(taskId)
    if (controller) {
      controller.abort()
      this.sseConnections.delete(taskId)
    }
  }

  stopProcess(taskId) {
    globalManager.stop(taskId)
    this.closeSSE(taskId)

    const state = this.taskStates.get(taskId)
    if (state) {
      state.status = 'stopped'
    }
  }

  removeProcess(taskId) {
    globalManager.remove(taskId)
    this.closeSSE(taskId)
    this.taskStates.delete(taskId)
  }

  getProgress(taskId) {
    const simulator = globalManager.get(taskId)
    return simulator ? simulator.getProgress() : 0
  }

  getStatus(taskId) {
    const simulator = globalManager.get(taskId)
    return simulator ? simulator.getStatus() : 'idle'
  }

  getState(taskId) {
    return this.taskStates.get(taskId)
  }

  clear() {
    this.sseConnections.forEach((controller, taskId) => {
      controller.abort()
    })
    this.sseConnections.clear()
    globalManager.clear()
    this.taskStates.clear()
  }
}

const videoProcessService = new VideoProcessService()

export const videoProcessApi = {
  startProcess: (taskId, taskType, callbacks) => 
    videoProcessService.startProcess(taskId, taskType, callbacks),
  
  forceComplete: (taskId) => 
    videoProcessService.forceComplete(taskId),
  
  stopProcess: (taskId) => 
    videoProcessService.stopProcess(taskId),
  
  removeProcess: (taskId) => 
    videoProcessService.removeProcess(taskId),
  
  getProgress: (taskId) => 
    videoProcessService.getProgress(taskId),
  
  getStatus: (taskId) => 
    videoProcessService.getStatus(taskId),
  
  getState: (taskId) => 
    videoProcessService.getState(taskId),
  
  clear: () => 
    videoProcessService.clear(),
  
  TaskType,
  TaskConfig
}

export default videoProcessApi
