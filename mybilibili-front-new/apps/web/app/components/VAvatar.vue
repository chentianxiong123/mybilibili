<script setup lang="ts">
withDefaults(defineProps<{
  src?: string
  size?: number
  auth?: 0 | 1 | 2
  alt?: string
  rounded?: boolean
}>(), {
  src: '',
  size: 40,
  auth: 0,
  alt: 'avatar',
  rounded: true
})

const authColor = (auth: number) => {
  if (auth === 1) return '#FFC62E'
  if (auth === 2) return '#4AC7FF'
  return 'transparent'
}
</script>

<template>
  <div
    class="v-avatar"
    :class="{ 'is-rounded': rounded }"
    :style="{
      width: `${size}px`,
      height: `${size}px`
    }"
  >
    <img v-if="src" :src="src" :alt="alt" class="v-avatar-img">
    <span v-else class="v-avatar-placeholder"></span>
    <span v-if="auth !== 0" class="v-avatar-badge" :style="{ background: authColor(auth) }">
      <svg viewBox="0 0 24 24" class="v-avatar-badge-icon">
        <path d="M9 16.17 4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" fill="#fff"/>
      </svg>
    </span>
  </div>
</template>

<style scoped>
.v-avatar {
  position: relative;
  display: inline-block;
  flex-shrink: 0;
  background: var(--v-bg3, #F1F2F3);
  overflow: visible;
}

.v-avatar.is-rounded {
  border-radius: 50%;
}

.v-avatar-img {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: inherit;
  display: block;
}

.v-avatar-placeholder {
  position: absolute;
  inset: 0;
  background: var(--v-bg3, #F1F2F3);
  border-radius: inherit;
}

.v-avatar-placeholder::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 40%;
  height: 40%;
  transform: translate(-50%, -50%);
  border-radius: 50%;
  background: var(--v-bg4, #C9CCD0);
}

.v-avatar-badge {
  position: absolute;
  right: 0;
  bottom: 0;
  width: 30%;
  height: 30%;
  min-width: 12px;
  min-height: 12px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1.5px solid var(--v-bg1, #FFFFFF);
  box-sizing: border-box;
}

.v-avatar-badge-icon {
  width: 70%;
  height: 70%;
  display: block;
}
</style>
