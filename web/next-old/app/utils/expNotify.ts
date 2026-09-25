import { userApi } from '@/api/client'
import { getCurrentUserId, getStoredUser, setAuthSession } from '@/utils/auth.ts'
import { ElMessage } from 'element-plus'

function maxExpForLevel(level: number): number {
  return Math.floor(100 * Math.pow(level, 1.8))
}

export async function refreshUserWithNotify() {
  const uid = getCurrentUserId()
  if (!uid) return
  const oldUser = getStoredUser()
  const oldLevel = oldUser?.level || 0
  const oldExp = oldUser?.experience || 0
  try {
    const r = await userApi.getUserById(uid)
    if (r.code !== 200) return
    const newLevel = r.data.level || 0
    const newExp = r.data.experience || 0
    if (newLevel > oldLevel) {
      ElMessage.success(`🎉 恭喜升级到 LV${newLevel}！`)
    } else if (newExp > oldExp) {
      ElMessage.success(`获得经验 +${newExp - oldExp}`)
    }
    setAuthSession({ user: r.data })
  } catch {}
}
