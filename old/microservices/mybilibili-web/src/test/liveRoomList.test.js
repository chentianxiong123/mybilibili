import { beforeEach, describe, expect, it, vi } from 'vitest'
import { useLiveRoomList } from '../composables/useLiveRoomList.js'

const createDeferred = () => {
  let resolve
  let reject
  const promise = new Promise((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

const createLiveRoomList = (options = {}) => {
  const api = options.api || {
    getLiveList: vi.fn(async () => ({ code: 200, data: options.liveRooms || [] })),
    getMyRoom: vi.fn(async () => ({ code: 200, data: options.myRoom || null }))
  }
  const message = options.message || { success: vi.fn() }
  const logger = options.logger || { error: vi.fn() }
  let pollHandler = null
  const setIntervalFn = vi.fn(handler => {
    pollHandler = handler
    return 'poll-timer'
  })
  const clearIntervalFn = vi.fn()

  const liveList = useLiveRoomList({
    api,
    message,
    logger,
    pollDelay: options.pollDelay ?? 10000,
    setIntervalFn,
    clearIntervalFn
  })

  return {
    api,
    message,
    logger,
    setIntervalFn,
    clearIntervalFn,
    runPoll: () => pollHandler?.(),
    ...liveList
  }
}

describe('useLiveRoomList', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('loads live rooms and my room while toggling loading', async () => {
    const liveListRequest = createDeferred()
    const myRoomRequest = createDeferred()
    const api = {
      getLiveList: vi.fn(() => liveListRequest.promise),
      getMyRoom: vi.fn(() => myRoomRequest.promise)
    }
    const rooms = [
      { id: 1, userId: 10, roomName: '直播间 A', viewerCount: 12 },
      { id: 2, userId: 11, roomName: '直播间 B', viewerCount: 3 }
    ]
    const myRoom = { id: 9, userId: 99, roomName: '我的直播间' }
    const liveList = createLiveRoomList({ api })

    const request = liveList.loadData()

    expect(api.getLiveList).toHaveBeenCalled()
    expect(api.getMyRoom).toHaveBeenCalled()
    expect(liveList.loading.value).toBe(true)

    liveListRequest.resolve({ code: 200, data: rooms })
    myRoomRequest.resolve({ code: 200, data: myRoom })
    await request

    expect(liveList.liveRooms.value).toEqual(rooms)
    expect(liveList.myRoom.value).toEqual(myRoom)
    expect(liveList.totalLive.value).toBe(2)
    expect(liveList.loading.value).toBe(false)
  })

  it('filters out my room, searches, and sorts by hot or newest', () => {
    const liveList = createLiveRoomList()
    liveList.liveRooms.value = [
      { id: 1, userId: 10, roomName: '我的游戏直播', category: '游戏', viewerCount: 100 },
      { id: 2, userId: 11, roomName: 'Study With Me', category: '学习', viewerCount: 20 },
      { id: 3, userId: 12, roomName: '游戏挑战', category: '游戏', viewerCount: 70 },
      { id: 4, userId: 13, roomName: '闲聊', category: '娱乐', viewerCount: 0 }
    ]
    liveList.myRoom.value = { userId: 10 }

    expect(liveList.filteredRooms.value.map(room => room.id)).toEqual([3, 2, 4])

    liveList.selectedCategory.value = '游戏'
    expect(liveList.filteredRooms.value.map(room => room.id)).toEqual([3])

    liveList.keyword.value = '12'
    expect(liveList.filteredRooms.value.map(room => room.id)).toEqual([3])

    liveList.selectedCategory.value = '全部'
    liveList.keyword.value = 'study'
    expect(liveList.filteredRooms.value.map(room => room.id)).toEqual([2])

    liveList.keyword.value = ''
    liveList.sortBy.value = 'new'
    expect(liveList.filteredRooms.value.map(room => room.id)).toEqual([4, 3, 2])
  })

  it('refreshes data and announces the refresh immediately', async () => {
    const liveListRequest = createDeferred()
    const api = {
      getLiveList: vi.fn(() => liveListRequest.promise),
      getMyRoom: vi.fn(async () => ({ code: 200, data: null }))
    }
    const message = { success: vi.fn() }
    const liveList = createLiveRoomList({ api, message })

    const request = liveList.refreshData()

    expect(api.getLiveList).toHaveBeenCalled()
    expect(message.success).toHaveBeenCalledWith('已刷新')
    expect(liveList.loading.value).toBe(true)

    liveListRequest.resolve({
      code: 200,
      data: [{ id: 5, userId: 15, roomName: '刷新后的直播间' }]
    })
    await request

    expect(liveList.liveRooms.value).toEqual([
      { id: 5, userId: 15, roomName: '刷新后的直播间' }
    ])
    expect(liveList.loading.value).toBe(false)
  })

  it('polls silently and cleans up the interval', async () => {
    const liveList = createLiveRoomList({ pollDelay: 2500 })

    liveList.startPolling()

    expect(liveList.setIntervalFn).toHaveBeenCalledWith(expect.any(Function), 2500)
    expect(liveList.api.getLiveList).not.toHaveBeenCalled()

    const pollRequest = liveList.runPoll()
    expect(liveList.loading.value).toBe(false)
    await pollRequest

    expect(liveList.api.getLiveList).toHaveBeenCalledTimes(1)

    liveList.cleanupLiveRoomList()
    liveList.cleanupLiveRoomList()

    expect(liveList.clearIntervalFn).toHaveBeenCalledTimes(1)
    expect(liveList.clearIntervalFn).toHaveBeenCalledWith('poll-timer')
  })
})
