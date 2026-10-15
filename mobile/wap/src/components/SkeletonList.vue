<script setup>
defineProps({
  variant: { type: String, default: 'grid' }, // grid | row | icon | avatar | feed
  count: { type: Number, default: 6 }
})
</script>

<template>
  <!-- 双列视频卡片骨架 -->
  <div v-if="variant === 'grid'" class="skeleton-grid">
    <div v-for="i in count" :key="i" class="sk-card">
      <div class="sk-cover"></div>
      <div class="sk-info">
        <div class="sk-line sk-line--title"></div>
        <div class="sk-line sk-line--title-short"></div>
        <div class="sk-line sk-line--author"></div>
      </div>
    </div>
  </div>

  <!-- 动态卡片骨架（头像 + 文本 + 图片块） -->
  <div v-else-if="variant === 'feed'" class="skeleton-feed-list">
    <div v-for="i in count" :key="i" class="sk-feed">
      <div class="sk-feed__header">
        <div class="sk-avatar-row__circle"></div>
        <div class="sk-feed__header-text">
          <div class="sk-line sk-line--title"></div>
          <div class="sk-line sk-line--meta"></div>
        </div>
      </div>
      <div class="sk-feed__body">
        <div class="sk-line sk-line--title"></div>
        <div class="sk-line sk-line--title-short"></div>
      </div>
      <div class="sk-feed__images">
        <div class="sk-feed__image"></div>
        <div class="sk-feed__image"></div>
        <div class="sk-feed__image"></div>
      </div>
    </div>
  </div>

  <!-- 横向列表卡片骨架（左封面 + 右信息） -->
  <div v-else-if="variant === 'row'" class="skeleton-row-list">
    <div v-for="i in count" :key="i" class="sk-row">
      <div class="sk-row__cover"></div>
      <div class="sk-row__info">
        <div class="sk-line sk-line--title"></div>
        <div class="sk-line sk-line--title-short"></div>
        <div class="sk-line sk-line--author"></div>
        <div class="sk-line sk-line--meta"></div>
      </div>
    </div>
  </div>

  <!-- 头像列表骨架（好友/消息/通知等） -->
  <div v-else-if="variant === 'avatar'" class="skeleton-avatar-list">
    <div v-for="i in count" :key="i" class="sk-avatar-row">
      <div class="sk-avatar-row__circle"></div>
      <div class="sk-avatar-row__info">
        <div class="sk-line sk-line--title"></div>
        <div class="sk-line sk-line--meta"></div>
      </div>
    </div>
  </div>

  <!-- 图标网格骨架（直播分类等） -->
  <div v-else class="skeleton-icon-grid">
    <div v-for="i in count" :key="i" class="sk-icon">
      <div class="sk-icon__circle"></div>
      <div class="sk-line sk-line--icon-text"></div>
    </div>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.sk-cover,
.sk-line,
.sk-icon__circle,
.sk-avatar-row__circle,
.sk-feed__image {
  background: linear-gradient(90deg, #f1f2f3 25%, #e6e7e8 37%, #f1f2f3 63%);
  background-size: 400% 100%;
  animation: skeleton-loading 1.4s ease infinite;
}

@keyframes skeleton-loading {
  0% { background-position: 100% 50%; }
  100% { background-position: 0 50%; }
}

/* ===== grid 双列卡片 ===== */
.skeleton-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px 8px;
  padding: 12px 8px;
}

.sk-card {
  .sk-cover {
    width: 100%;
    aspect-ratio: 16 / 9;
    border-radius: 6px;
  }

  .sk-info {
    padding: 8px 2px 0;
  }
}

/* ===== row 横向卡片 ===== */
.skeleton-row-list {
  padding: 12px;
}

.sk-row {
  display: flex;
  gap: 12px;
  margin-bottom: 16px;

  .sk-row__cover {
    width: 140px;
    height: 87px;
    border-radius: 6px;
    flex-shrink: 0;
  }

  .sk-row__info {
    flex: 1;
    padding-top: 2px;
  }
}

/* ===== avatar 头像列表 ===== */
.skeleton-avatar-list {
  padding: 12px;
}

.sk-avatar-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f1f2f3;

  .sk-avatar-row__circle {
    width: 48px;
    height: 48px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .sk-avatar-row__info {
    flex: 1;
    min-width: 0;
  }
}

/* ===== feed 动态卡片 ===== */
.skeleton-feed-list {
  padding: 12px;
}

.sk-feed {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  margin-bottom: 12px;

  .sk-feed__header {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 12px;

    .sk-feed__header-text {
      flex: 1;
      min-width: 0;
    }
  }

  .sk-feed__body {
    margin-bottom: 12px;
  }

  .sk-feed__images {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 4px;

    .sk-feed__image {
      aspect-ratio: 16 / 10;
      border-radius: 4px;
    }
  }
}

/* ===== icon 图标网格 ===== */
.skeleton-icon-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  padding: 16px 12px;
}

.sk-icon {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;

  .sk-icon__circle {
    width: 52px;
    height: 52px;
    border-radius: 50%;
  }
}

/* ===== 通用灰条 ===== */
.sk-line {
  border-radius: 4px;

  &--title {
    height: 14px;
    width: 100%;
    margin-bottom: 8px;
  }

  &--title-short {
    height: 14px;
    width: 65%;
    margin-bottom: 8px;
  }

  &--author {
    height: 12px;
    width: 40%;
    margin-bottom: 6px;
  }

  &--meta {
    height: 12px;
    width: 55%;
  }

  &--icon-text {
    height: 12px;
    width: 60%;
    margin-bottom: 0;
  }
}
</style>
