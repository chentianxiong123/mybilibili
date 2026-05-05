<template>
  <div class="account-diagnosis">
    <div class="diagnosis-summary">
      <h3 class="section-title">表现总结</h3>
      <div class="summary-tip">
        <span>以下基于您近期的表现进行分析，关注提升空间。</span>
      </div>
      <div class="summary-content">
        <p>• 播放量处于上升趋势，播放速率可观</p>
        <p>• 点赞量稳定提升，证明内容质量不错</p>
        <p>• 收藏数增长较快，建议保持现有节奏</p>
      </div>
    </div>

    <div class="diagnosis-ranking">
      <div class="ranking-box">
        <h4 class="ranking-title">播放排行</h4>
        <div class="ranking-value"># -</div>
        <div class="ranking-change">注意：与同类UP主比较</div>
      </div>
      <div class="ranking-box">
        <h4 class="ranking-title">互动排行</h4>
        <div class="ranking-value"># -</div>
        <div class="ranking-change">注意：评论、弹幕等互动排行</div>
      </div>
    </div>

    <div class="diagnosis-metrics">
      <h3 class="section-title">核心指标</h3>
      <div class="metrics-grid">
        <div class="metric-item">
          <span class="metric-label">播放完成率</span>
          <span class="metric-value">-</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">互动率</span>
          <span class="metric-value">-</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">粉丝转化率</span>
          <span class="metric-value">-</span>
        </div>
        <div class="metric-item">
          <span class="metric-label">活跃度</span>
          <span class="metric-value">-</span>
        </div>
      </div>
    </div>

    <div class="diagnosis-suggestions">
      <h3 class="section-title">优化建议</h3>
      <div class="suggestions-list">
        <div class="suggestion-item">
          <span class="suggestion-icon">💡</span>
          <span class="suggestion-text">建议增加视频发布频率，保持粉丝活跃度</span>
        </div>
        <div class="suggestion-item">
          <span class="suggestion-icon">📈</span>
          <span class="suggestion-text">可以尝试在热门时段发布视频</span>
        </div>
        <div class="suggestion-item">
          <span class="suggestion-icon">🎯</span>
          <span class="suggestion-text">加强与粉丝互动，回复评论提升粘性</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
</script>

<style scoped>
.account-diagnosis {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin: 0 0 12px 0;
}
.diagnosis-summary {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.summary-tip {
  background: #fff7e6;
  border-radius: 6px;
  padding: 10px 14px;
  margin-bottom: 14px;
  color: #d48806;
  font-size: 13px;
}
.summary-content {
  color: #666;
  font-size: 14px;
  line-height: 1.8;
}
.diagnosis-ranking {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}
.ranking-box {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
  text-align: center;
}
.ranking-title {
  font-size: 14px;
  color: #999;
  margin: 0 0 10px 0;
}
.ranking-value {
  font-size: 28px;
  font-weight: 700;
  color: #00aeec;
}
.ranking-change {
  font-size: 12px;
  color: #999;
  margin-top: 6px;
}
.diagnosis-metrics {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.metrics-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
}
.metric-item {
  text-align: center;
  padding: 16px;
  background: #f9f9f9;
  border-radius: 8px;
}
.metric-label {
  display: block;
  font-size: 13px;
  color: #999;
  margin-bottom: 8px;
}
.metric-value {
  font-size: 20px;
  font-weight: 600;
  color: #333;
}
.diagnosis-suggestions {
  background: #fff;
  border-radius: 10px;
  border: 1px solid #f0f0f0;
  padding: 20px;
}
.suggestions-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.suggestion-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #f9f9f9;
  border-radius: 8px;
}
.suggestion-icon {
  font-size: 18px;
}
.suggestion-text {
  font-size: 14px;
  color: #666;
}
</style>
