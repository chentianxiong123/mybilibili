const TaskType = {
  TRANSCODE: 'transcode',
  SUBTITLE: 'subtitle',
  AUDIO: 'audio',
  AI_SUBTITLE: 'ai_subtitle'
}

const TaskConfig = {
  [TaskType.TRANSCODE]: { increment: 1, interval: 800, label: '转码' },
  [TaskType.SUBTITLE]: { increment: 1, interval: 800, label: '字幕' },
  [TaskType.AUDIO]: { increment: 2, interval: 600, label: '音频' },
  [TaskType.AI_SUBTITLE]: { increment: 2, interval: 600, label: 'AI字幕' }
}

class ProgressSimulator {
  constructor(options = {}) {
    this.taskType = options.taskType || TaskType.TRANSCODE
    const config = TaskConfig[this.taskType] || TaskConfig[TaskType.TRANSCODE]
    
    this.increment = options.increment ?? config.increment
    this.interval = options.interval ?? config.interval
    this.slowDownAt = options.slowDownAt ?? 90
    this.maxProgress = options.maxProgress ?? 99
    
    this.currentProgress = 0
    this.isRunning = false
    this.isCompleted = false
    this.isPaused = false
    
    this.timer = null
    this.onProgress = null
    this.onComplete = null
  }

  start(callbacks = {}) {
    if (this.isRunning) return this
    
    this.onProgress = callbacks.onProgress || (() => {})
    this.onComplete = callbacks.onComplete || (() => {})
    this.isRunning = true
    this.isCompleted = false
    this.isPaused = false
    
    this._tick()
    return this
  }

  _tick() {
    if (!this.isRunning || this.isPaused || this.isCompleted) return
    
    const { increment, interval } = this._calculateNextProgress()
    
    if (this.currentProgress < this.maxProgress) {
      this.currentProgress = Math.min(
        this.currentProgress + increment,
        this.maxProgress
      )
      this.onProgress(this.currentProgress)
      
      this.timer = setTimeout(() => this._tick(), interval)
    } else {
      this._handleMaxProgress()
    }
  }

  _calculateNextProgress() {
    if (this.currentProgress >= this.slowDownAt) {
      return {
        increment: Math.max(0.5, this.increment / 2),
        interval: this.interval * 2
      }
    }
    return {
      increment: this.increment,
      interval: this.interval
    }
  }

  _handleMaxProgress() {
    if (!this.isCompleted) {
      this.onProgress(this.currentProgress)
    }
  }

  forceComplete() {
    if (this.isCompleted) return
    
    this._clearTimer()
    this.currentProgress = 100
    this.isCompleted = true
    this.isRunning = false
    
    this.onProgress(100)
    this.onComplete()
  }

  pause() {
    if (!this.isRunning || this.isCompleted) return
    this.isPaused = true
    this._clearTimer()
  }

  resume() {
    if (!this.isRunning || this.isCompleted || !this.isPaused) return
    this.isPaused = false
    this._tick()
  }

  reset() {
    this._clearTimer()
    this.currentProgress = 0
    this.isRunning = false
    this.isCompleted = false
    this.isPaused = false
    this.onProgress = null
    this.onComplete = null
  }

  stop() {
    this._clearTimer()
    this.isRunning = false
  }

  _clearTimer() {
    if (this.timer) {
      clearTimeout(this.timer)
      this.timer = null
    }
  }

  getProgress() {
    return this.currentProgress
  }

  getStatus() {
    if (this.isCompleted) return 'completed'
    if (this.isPaused) return 'paused'
    if (this.isRunning) return 'running'
    return 'idle'
  }

  isAtSlowDownPhase() {
    return this.currentProgress >= this.slowDownAt
  }
}

class ProgressSimulatorManager {
  constructor() {
    this.simulators = new Map()
  }

  create(id, options = {}) {
    if (this.simulators.has(id)) {
      const existing = this.simulators.get(id)
      existing.stop()
    }
    
    const simulator = new ProgressSimulator(options)
    this.simulators.set(id, simulator)
    return simulator
  }

  get(id) {
    return this.simulators.get(id)
  }

  start(id, callbacks = {}) {
    const simulator = this.simulators.get(id)
    if (simulator) {
      return simulator.start(callbacks)
    }
    return null
  }

  forceComplete(id) {
    const simulator = this.simulators.get(id)
    if (simulator) {
      simulator.forceComplete()
    }
  }

  stop(id) {
    const simulator = this.simulators.get(id)
    if (simulator) {
      simulator.stop()
    }
  }

  remove(id) {
    const simulator = this.simulators.get(id)
    if (simulator) {
      simulator.stop()
      this.simulators.delete(id)
    }
  }

  clear() {
    this.simulators.forEach(simulator => simulator.stop())
    this.simulators.clear()
  }

  has(id) {
    return this.simulators.has(id)
  }
}

const globalManager = new ProgressSimulatorManager()

export {
  ProgressSimulator,
  ProgressSimulatorManager,
  TaskType,
  TaskConfig,
  globalManager
}

export default ProgressSimulator
