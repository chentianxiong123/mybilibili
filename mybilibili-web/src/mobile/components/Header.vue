<script setup>
</script>

<template>
  <div class="mobile-header">
    <a href="/m/index" class="logo">
      <img src="../assets/images/logo.png" alt="bilibili" />
    </a>
    <router-link to="/m/search" class="search-box">
      <i class="icon-search" />
      <span class="placeholder">搜索...</span>
    </router-link>
    <router-link to="/m/space" class="user-avatar">
      <img src="../assets/images/noface.gif" alt="user" />
    </router-link>
  </div>
</template>

<style scoped lang="scss">
@import '../styles/variables';

.mobile-header {
  display: flex;
  align-items: center;
  height: $header-height;
  padding: 0 12px;
  background: $theme-pink;

  .logo {
    width: 72px;
    margin-right: 12px;
    img { width: 100%; }
  }

  .search-box {
    flex: 1;
    display: flex;
    align-items: center;
    height: 30px;
    background: rgba(255,255,255,0.2);
    border-radius: 15px;
    padding: 0 12px;
    color: rgba(255,255,255,0.8);
    font-size: 13px;
    .placeholder { margin-left: 6px; }
  }

  .user-avatar {
    width: 30px;
    height: 30px;
    margin-left: 12px;
    border-radius: 50%;
    overflow: hidden;
    img { width: 100%; height: 100%; }
  }
}
</style>