<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getLocalUserId, getMyInfo, updateMyInfo } from '../../api/user'
import noface from '../../assets/noface.gif'

const router = useRouter()
const userInfo = ref<any>(null)
const editing = ref<string | null>(null)
const draftValue = ref('')
const saving = ref(false)

const displayName = computed(() => userInfo.value?.nickname || userInfo.value?.username || '未命名用户')
const genderText = computed(() => {
  const g = Number(userInfo.value?.gender)
  if (g === 1) return '男'
  if (g === 2) return '女'
  return '保密'
})

const loadProfile = async () => {
  const res = await getMyInfo()
  if (res.code === '1') userInfo.value = res.data
}

const goBack = () => router.back()

const openEditor = (field: string, value = '') => {
  editing.value = field
  draftValue.value = value || ''
}

const saveField = async () => {
  const userId = getLocalUserId()
  if (!userId || !editing.value) return
  saving.value = true
  const payload: Record<string, any> = {}
  payload[editing.value] = draftValue.value
  const res = await updateMyInfo(Number(userId), payload)
  if (res.code === '1' && res.data) {
    userInfo.value = res.data
  }
  saving.value = false
  editing.value = null
}

onMounted(loadProfile)
</script>

<template>
  <div class="profile-edit-page">
    <header class="page-header">
      <button class="back-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <h1>账号资料</h1>
    </header>

    <main v-if="userInfo" class="profile-sections">
      <section class="row-group">
        <button class="profile-row avatar-row">
          <span>头像</span>
          <span class="row-value">
            <img :src="userInfo.avatar || noface" alt="" />
            <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg>
          </span>
        </button>
        <button class="profile-row" @click="openEditor('nickname', displayName)">
          <span>昵称</span>
          <span class="row-value text">{{ displayName }}<svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
        <button class="profile-row">
          <span>性别</span>
          <span class="row-value text">{{ genderText }}<svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
        <button class="profile-row" @click="openEditor('birthdate', userInfo.birthdate || '')">
          <span>出生年月</span>
          <span class="row-value text">{{ userInfo.birthdate || '未设置' }}<svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
        <button class="profile-row" @click="openEditor('signature', userInfo.signature || userInfo.bio || '')">
          <span>个性签名</span>
          <span class="row-value text">{{ userInfo.signature || userInfo.bio || '未设置' }}<svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
        <button class="profile-row">
          <span>学校</span>
          <span class="row-value text muted">填写学校发现更多校友~<svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
      </section>

      <section class="row-group">
        <button class="profile-row">
          <span>头像挂件</span>
          <span class="row-value"><svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
      </section>

      <section class="row-group">
        <div class="profile-row static">
          <span>UID</span>
          <span class="row-value text">{{ userInfo.id }}</span>
        </div>
        <button class="profile-row">
          <span>二维码名片</span>
          <span class="row-value qr">
            <svg class="qr-icon" viewBox="0 0 24 24"><path d="M3 3h7v7H3zM14 3h7v7h-7zM3 14h7v7H3zM14 14h3v3h-3zM18 18h3v3h-3zM18 14h3M14 20h3" /></svg>
            <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg>
          </span>
        </button>
        <button class="profile-row">
          <span>购买邀请码</span>
          <span class="row-value"><svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
      </section>

      <section class="row-group">
        <button class="profile-row">
          <span>哔哩哔哩认证</span>
          <span class="row-value"><svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg></span>
        </button>
      </section>
    </main>

    <div v-else class="loading">加载中...</div>

    <div v-if="editing" class="editor-mask" @click.self="editing = null">
      <div class="editor-panel">
        <div class="editor-title">修改{{ editing === 'nickname' ? '昵称' : editing === 'birthdate' ? '出生年月' : '个性签名' }}</div>
        <textarea v-model="draftValue" rows="4" />
        <div class="editor-actions">
          <button @click="editing = null">取消</button>
          <button class="primary" :disabled="saving" @click="saveField">{{ saving ? '保存中' : '保存' }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.profile-edit-page {
  min-height: 100vh;
  background: #f6f7f9;
  color: #18191c;
}

.page-header {
  position: sticky;
  top: 0;
  z-index: 10;
  height: 74px;
  display: flex;
  align-items: center;
  gap: 18px;
  padding: 18px 14px 0;
  background: #fff;
  border-bottom: 1px solid #f1f2f3;

  h1 {
    margin: 0;
    font-size: 21px;
    font-weight: 600;
    letter-spacing: 0;
  }
}

.back-btn {
  width: 36px;
  height: 36px;
  border: 0;
  padding: 0;
  display: grid;
  place-items: center;
  background: transparent;
  color: #61666d;

  svg {
    width: 30px;
    height: 30px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.profile-sections {
  padding-top: 20px;
}

.row-group {
  margin-bottom: 20px;
  background: #fff;
}

.profile-row {
  width: 100%;
  min-height: 52px;
  border: 0;
  border-bottom: 1px solid #f1f2f3;
  padding: 0 16px;
  background: transparent;
  color: #18191c;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 18px;
  font-size: 18px;
  text-align: left;

  &:last-child {
    border-bottom: 0;
  }

  &.static {
    cursor: default;
  }
}

.avatar-row {
  min-height: 88px;
}

.row-value {
  min-width: 0;
  display: inline-flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
  color: #9499a0;

  img {
    width: 64px;
    height: 64px;
    border-radius: 50%;
    object-fit: cover;
  }

  svg {
    flex: 0 0 auto;
    width: 23px;
    height: 23px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.row-value.text {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 17px;
}

.muted {
  color: #9499a0;
}

.qr .qr-icon {
  width: 24px;
  height: 24px;
}

.loading {
  padding: 80px 0;
  text-align: center;
  color: #9499a0;
}

.editor-mask {
  position: fixed;
  inset: 0;
  z-index: 200;
  background: rgba(0,0,0,0.55);
  display: flex;
  align-items: flex-end;
}

.editor-panel {
  width: 100%;
  padding: 18px 16px calc(20px + env(safe-area-inset-bottom));
  border-radius: 14px 14px 0 0;
  background: #fff;
}

.editor-title {
  margin-bottom: 12px;
  font-size: 17px;
  font-weight: 600;
}

textarea {
  width: 100%;
  box-sizing: border-box;
  border: 1px solid #e3e5e7;
  border-radius: 8px;
  padding: 10px;
  resize: none;
  outline: 0;
  background: #f6f7f9;
  color: #18191c;
  font-size: 15px;
}

.editor-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 14px;

  button {
    height: 36px;
    border: 0;
    border-radius: 18px;
    padding: 0 18px;
    background: #f1f2f3;
    color: #61666d;
    font-size: 15px;
  }

  .primary {
    background: #ff6aa2;
    color: #fff;
  }
}
</style>
