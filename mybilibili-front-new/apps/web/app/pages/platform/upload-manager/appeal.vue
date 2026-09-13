<template>
    <div class="page-placeholder">
        <i class="iconfont icon-jubao"></i>
        <h2>申诉管理</h2>
        <p>该功能正在建设中，敬请期待。</p>
        <p class="hint">如有违规下架等问题，请通过 <a @click="$router.push('/message/system')">系统通知</a> 查看详情</p>
    </div>
</template>

<style scoped>
.page-placeholder {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 80px 24px;
    color: var(--text3);
    text-align: center;
}

.page-placeholder .iconfont {
    font-size: 64px;
    color: var(--graph_weak);
    margin-bottom: 16px;
}

.page-placeholder h2 {
    font-size: 20px;
    color: var(--text2);
    font-weight: 500;
    margin: 0 0 8px 0;
}

.page-placeholder p {
    font-size: 14px;
    margin: 4px 0;
}

.page-placeholder .hint {
    margin-top: 16px;
    color: var(--text3);
}

.page-placeholder .hint a {
    color: var(--brand_pink);
    cursor: pointer;
    text-decoration: none;
}

.page-placeholder .hint a:hover {
    text-decoration: underline;
}
</style>
