<script setup>
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getHotwords } from '../../api/search'

const router = useRouter()
const hotwords = ref([])

onMounted(async () => {
  const res = await getHotwords()
  if (res.code === '1') hotwords.value = res.data || []
})

const goSearch = (keyword) => {
  router.push({ path: '/m/search', query: { keyword } })
}
</script>

<template>
  <div class="hot-rank-page">
    <section class="hero">
      <button class="nav-btn left" @click="router.back()" aria-label="返回">
        <svg viewBox="0 0 24 24"><path d="M15 18l-6-6 6-6" /></svg>
      </button>
      <button class="nav-btn right" aria-label="分享">
        <svg viewBox="0 0 24 24"><path d="M4 12v7a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-7" /><path d="M16 6l-4-4-4 4M12 2v14" /></svg>
      </button>
      <div class="hero-graphic">
        <div class="lens"></div>
        <h1>bilibili 热搜</h1>
      </div>
    </section>

    <main class="rank-panel">
      <button class="top-topic" @click="goSearch(hotwords[0]?.keyword || '')">
        <b>↑</b>
        <span>{{ hotwords[0]?.keyword || '暂无热搜' }}</span>
        <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg>
      </button>

      <button
        v-for="(item, index) in hotwords.slice(0, 30)"
        :key="item.keyword + index"
        class="rank-row"
        @click="goSearch(item.keyword)"
      >
        <strong :class="{ top: index < 3 }">{{ index + 1 }}</strong>
        <span>{{ item.keyword }}</span>
        <i v-if="index === 0 || index === 8 || index === 13">新</i>
        <i v-else-if="index === 3 || index === 11" class="hot">热</i>
        <i v-else-if="index === 2" class="rise">▥</i>
        <i v-else-if="index === 4 || index === 5 || index === 7 || index === 9" class="exclusive">独家</i>
        <svg viewBox="0 0 24 24"><path d="M9 18l6-6-6-6" /></svg>
      </button>
    </main>
  </div>
</template>

<style scoped lang="scss">
.hot-rank-page {
  min-height: 100vh;
  background: #f4f6fb;
}

.hero {
  position: relative;
  height: 224px;
  overflow: hidden;
  color: #fff;
  background:
    radial-gradient(circle at 88% 74%, rgba(255, 155, 216, 0.85), transparent 20%),
    radial-gradient(circle at 14% 34%, rgba(210, 192, 255, 0.95), transparent 28%),
    linear-gradient(128deg, #8ea1ff 0%, #3565d8 62%, #2450c3 100%);
}

.nav-btn {
  position: absolute;
  top: 58px;
  z-index: 3;
  width: 44px;
  height: 44px;
  border: 0;
  padding: 0;
  background: transparent;
  color: #fff;

  &.left {
    left: 20px;
  }

  &.right {
    right: 22px;
  }

  svg {
    width: 38px;
    height: 38px;
    fill: none;
    stroke: currentColor;
    stroke-width: 2.2;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
}

.hero-graphic {
  position: absolute;
  left: 56px;
  right: 28px;
  bottom: 44px;
  display: flex;
  align-items: center;
  gap: 16px;

  h1 {
    margin: 0;
    color: #fff;
    font-size: 46px;
    font-weight: 800;
    letter-spacing: 0;
  }
}

.lens {
  position: relative;
  width: 110px;
  height: 110px;
  border-radius: 50%;
  background: radial-gradient(circle at 42% 40%, rgba(255,255,255,0.28), rgba(110,142,239,0.1) 42%, rgba(67,94,209,0.95) 70%);
  box-shadow: inset 0 0 26px rgba(255,255,255,0.38);

  &::after {
    content: '';
    position: absolute;
    left: -35px;
    bottom: -25px;
    width: 76px;
    height: 28px;
    border-radius: 18px;
    background: rgba(67, 94, 209, 0.82);
    transform: rotate(-43deg);
  }
}

.rank-panel {
  position: relative;
  z-index: 5;
  margin-top: -28px;
  padding: 10px 18px 20px;
  border-radius: 16px 16px 0 0;
  background: #151617;
  color: #f6f7f9;
}

.top-topic,
.rank-row {
  width: 100%;
  min-width: 0;
  border: 0;
  border-bottom: 1px solid rgba(255,255,255,0.08);
  display: grid;
  grid-template-columns: 42px 1fr auto 28px;
  align-items: center;
  gap: 10px;
  background: transparent;
  color: inherit;
  text-align: left;

  span {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  svg {
    width: 24px;
    height: 24px;
    fill: none;
    stroke: #777b83;
    stroke-width: 2.2;
  }
}

.top-topic {
  height: 58px;

  b {
    color: #f05555;
    font-size: 31px;
  }

  span {
    font-size: 22px;
  }
}

.rank-row {
  height: 57px;

  strong {
    color: #b8bcc5;
    font-size: 23px;
    font-style: italic;
    font-weight: 700;

    &.top {
      color: #ffbf3b;
    }
  }

  span {
    color: #f6f7f9;
    font-size: 22px;
  }

  i {
    min-width: 21px;
    height: 21px;
    border-radius: 4px;
    display: inline-grid;
    place-items: center;
    color: #fff;
    background: #ffbd2f;
    font-size: 14px;
    font-style: normal;
    font-weight: 700;

    &.hot {
      background: #f65c5c;
    }

    &.rise,
    &.exclusive {
      min-width: 34px;
      background: #f25d9c;
    }
  }
}

@media (max-width: 390px) {
  .hero-graphic h1 {
    font-size: 38px;
  }

  .lens {
    width: 88px;
    height: 88px;
  }

  .rank-row span,
  .top-topic span {
    font-size: 19px;
  }
}
</style>
