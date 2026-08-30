import { beforeEach, describe, expect, it, vi } from 'vitest'
import { ref } from 'vue'
import { useMeetingPresenceSignals } from '../composables/useMeetingPresenceSignals.js'

const createSignals = (options = {}) => {
  const currentRoom = ref(options.currentRoom || { roomCode: 'room-1' })
  const peerConnections = options.peerConnections || {}
  let timerHandler = null
  const setTimer = vi.fn(handler => {
    timerHandler = handler
    return 'speaker-timer'
  })
  const clearTimer = vi.fn()
  let qualityTimerHandler = null
  const setIntervalFn = vi.fn(handler => {
    qualityTimerHandler = handler
    return 'quality-timer'
  })
  const clearIntervalFn = vi.fn()
  let nowValue = options.nowValue || 1000
  const signals = useMeetingPresenceSignals({
    currentRoom,
    peerConnections,
    now: () => nowValue,
    setTimer,
    clearTimer,
    setIntervalFn,
    clearIntervalFn,
    ...options
  })

  return {
    currentRoom,
    peerConnections,
    setTimer,
    clearTimer,
    setIntervalFn,
    clearIntervalFn,
    tickNow: value => { nowValue = value },
    runTimer: () => timerHandler?.(),
    runQualityTimer: () => qualityTimerHandler?.(),
    ...signals
  }
}

describe('useMeetingPresenceSignals', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('polls peer quality and maps quality levels', async () => {
    const signals = createSignals({
      peerConnections: {
        a: {
          getStats: vi.fn(async () => new Map([
            ['candidate', {
              type: 'candidate-pair',
              state: 'succeeded',
              currentRoundTripTime: 0.31
            }],
            ['inbound', {
              type: 'inbound-rtp',
              packetsLost: 2,
              packetsReceived: 18
            }]
          ]))
        },
        b: {
          getStats: vi.fn(async () => new Map([
            ['candidate', {
              type: 'candidate-pair',
              state: 'succeeded',
              currentRoundTripTime: 0.61
            }]
          ]))
        }
      }
    })

    await signals.pollPeerQuality()

    expect(signals.peerQuality.a).toEqual({ rtt: 310, lossRate: 10 })
    expect(signals.peerQuality.b).toEqual({ rtt: 610, lossRate: 0 })
    expect(signals.getQualityLevel('a')).toBe(2)
    expect(signals.getQualityLevel('b')).toBe(1)
    expect(signals.getQualityLevel('missing')).toBe(0)
  })

  it('tracks active speakers and clears them on timeout or mute', () => {
    const signals = createSignals()

    signals.updateAudioActivity('peer-1', true)
    expect(signals.activeSpeakerId.value).toBeNull()

    signals.tickNow(3201)
    signals.updateAudioActivity('peer-1', true)
    expect(signals.activeSpeakerId.value).toBe('peer-1')
    expect(signals.setTimer).toHaveBeenCalledWith(expect.any(Function), 3000)

    signals.runTimer()
    expect(signals.activeSpeakerId.value).toBeNull()

    signals.updateAudioActivity('peer-1', false)
    expect(signals.activeSpeakerId.value).toBeNull()
  })

  it('cleans up timers and quality state', () => {
    const signals = createSignals()

    signals.peerQuality.a = { rtt: 100, lossRate: 1 }
    signals.updateAudioActivity('peer-1', true)
    signals.tickNow(3201)
    signals.updateAudioActivity('peer-1', true)

    signals.cleanupPresenceSignals()

    expect(signals.clearTimer).toHaveBeenCalledWith('speaker-timer')
    expect(signals.activeSpeakerId.value).toBeNull()
    expect(signals.peerQuality.a).toBeUndefined()
  })

  it('owns the peer quality polling interval lifecycle', () => {
    const signals = createSignals()

    expect(signals.startQualityPolling()).toBe(true)
    expect(signals.startQualityPolling()).toBe(false)
    expect(signals.setIntervalFn).toHaveBeenCalledWith(expect.any(Function), 5000)

    signals.cleanupPresenceSignals()

    expect(signals.clearIntervalFn).toHaveBeenCalledWith('quality-timer')
    expect(signals.stopQualityPolling()).toBe(false)
  })
})
