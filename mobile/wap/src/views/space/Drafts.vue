<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goBack = () => router.back()

const goWebDrafts = () => {
  const origin = window.location.origin.replace(':5174', ':5173')
  window.location.href = `${origin}/create-center/drafts`
}
</script>

<template>
  <div class="drafts-page">
    <header class="drafts-header">
      <button class="back-btn" @click="goBack" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <h1>草稿箱</h1>
      <span class="header-spacer"></span>
    </header>

    <main class="drafts-content">
      <div class="placeholder-box">
        <div class="icon">
          <svg viewBox="0 0 24 24">
            <path d="M12 3v18M5 8h14M5 16h14" />
          </svg>
        </div>
        <p class="title">移动端草稿箱暂未开放</p>
        <p class="desc">
          草稿保存在浏览器本地<br />
          请前往 Web 创作中心使用草稿箱
        </p>
        <button class="web-btn" @click="goWebDrafts">前往 Web 创作中心</button>
      </div>
    </main>
  </div>
</template>

<style scoped lang="scss">
.drafts-page {
  min-height: 100dvh;
  background: #f4f4f4;
}

.drafts-header {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 8px 0 4px;
  background: #fff;
  border-bottom: 1px solid #e3e5e7;

  h1 {
    flex: 1;
    text-align: center;
    font-size: 16px;
    font-weight: 600;
    color: #18191c;
  }

  .back-btn {
    width: 40px;
    height: 40px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: none;
    border: none;
    svg {
      width: 22px;
      height: 22px;
      fill: none;
      stroke: #18191c;
      stroke-width: 2;
      stroke-linecap: round;
      stroke-linejoin: round;
    }
  }

  .header-spacer {
    width: 40px;
  }
}

.drafts-content {
  display: flex;
  justify-content: center;
  padding-top: 90px;
}

.placeholder-box {
  width: 260px;
  text-align: center;
}

.icon {
  width: 72px;
  height: 72px;
  margin: 0 auto 18px;
  border-radius: 50%;
  background: #e3e5e7;
  display: flex;
  align-items: center;
  justify-content: center;
  svg {
    width: 34px;
    height: 34px;
    fill: none;
    stroke: #9499a0;
    stroke-width: 1.5;
    stroke-linecap: round;
  }
}

.title {
  font-size: 15px;
  font-weight: 600;
  color: #18191c;
  margin-bottom: 8px;
}

.desc {
  font-size: 12px;
  color: #9499a0;
  line-height: 1.8;
  margin-bottom: 22px;
}

.web-btn {
  padding: 9px 24px;
  font-size: 14px;
  color: #fff;
  background: #fb7299;
  border: none;
  border-radius: 20px;
  cursor: pointer;
  &:active {
    opacity: 0.85;
  }
}
</style>