<template>
  <div class="danmu-manager">
    <div class="danmu-management">
      <!-- 顶部导航标签和搜索栏 -->
      <div class="danmu-tabs-with-search">
        <!-- 顶部导航标签 -->
        <div class="danmu-tabs">
          <el-tabs v-model="activeDanmuTab" type="card">
            <el-tab-pane label="稿件弹幕" name="article"></el-tab-pane>
            <el-tab-pane label="弹幕设置" name="settings"></el-tab-pane>
            <el-tab-pane label="弹幕反馈" name="feedback"></el-tab-pane>
          </el-tabs>
        </div>
        
        <!-- 右侧搜索栏 -->
        <div class="danmu-search">
          <el-input
            v-model="danmuSearchText"
            placeholder="搜索弹幕关键字"
            size="small"
            style="width: 200px;"
            clearable
          >
            <template #append>
              <el-button size="small" @click="searchDanmu"><el-icon><Search /></el-icon></el-button>
            </template>
          </el-input>
        </div>
      </div>
      
      <!-- 操作栏和筛选条件（仅在稿件弹幕标签页显示） -->
      <div v-if="activeDanmuTab === 'article'" class="danmu-header">
        <!-- 左侧操作按钮 -->
        <div class="danmu-actions">
          <el-button size="small" type="danger">删除弹幕</el-button>
          <el-button size="small" type="primary" plain>弹幕保护</el-button>
          <el-button size="small" type="primary" plain>取消保护</el-button>
          <el-button size="small" type="primary" plain>字幕</el-button>
          <el-button size="small" type="primary" plain>普通</el-button>
          <el-button size="small" type="primary" plain>
            <el-icon><Refresh /></el-icon>刷新
          </el-button>
          <el-button size="small" type="primary" plain>弹幕转移</el-button>
        </div>
        
        <!-- 右侧筛选条件 -->
        <div class="danmu-filters">
          <el-select v-model="danmuTypeFilter" placeholder="全部弹幕" size="small" style="min-width: 120px; margin-right: 15px;">
            <el-option label="全部弹幕" value="all"></el-option>
            <el-option label="普通弹幕" value="normal"></el-option>
            <el-option label="字幕弹幕" value="subtitle"></el-option>
          </el-select>
          
          <el-select v-model="danmuTimeFilter" placeholder="最近弹幕" size="small" style="min-width: 120px; margin-right: 15px;">
            <el-option label="最近弹幕" value="latest"></el-option>
            <el-option label="本周弹幕" value="week"></el-option>
            <el-option label="本月弹幕" value="month"></el-option>
          </el-select>
          
          <el-select v-model="danmuVideoFilter" placeholder="全部视频" size="small" style="min-width: 120px;">
            <el-option label="全部视频" value="all"></el-option>
            <el-option label="视频1" value="video1"></el-option>
            <el-option label="视频2" value="video2"></el-option>
          </el-select>
        </div>
      </div>
      
      <!-- 弹幕列表（仅在稿件弹幕标签页显示） -->
      <div v-if="activeDanmuTab === 'article'" class="danmu-list">
        <el-table :data="danmus" stripe style="width: 100%">
          <el-table-column type="selection" width="55"></el-table-column>
          <el-table-column prop="sender" label="发送者" width="120"></el-table-column>
          <el-table-column prop="playTime" label="播放时间" width="100" sortable></el-table-column>
          <el-table-column prop="content" label="弹幕内容" min-width="300">
            <template #default="scope">
              <div>{{ scope.row.content }}</div>
              <div class="danmu-video-info">视频: {{ scope.row.videoTitle }}</div>
            </template>
          </el-table-column>
          <el-table-column prop="likes" label="点赞" width="80" sortable>
            <template #default="scope">
              <el-icon><StarFilled /></el-icon>{{ scope.row.likes }}
            </template>
          </el-table-column>
          <el-table-column prop="type" label="属性" width="80">
            <template #default="scope">
              <el-tag size="small">{{ scope.row.type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sendTime" label="发送时间" width="160" sortable></el-table-column>
          <el-table-column label="操作" width="80">
            <template #default="scope">
              <el-dropdown>
                <el-button size="small" plain>
                  <el-icon><MoreFilled /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item>删除</el-dropdown-item>
                    <el-dropdown-item>举报</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </template>
          </el-table-column>
        </el-table>
      </div>
      
      <!-- 弹幕设置（仅在弹幕设置标签页显示） -->
      <div v-else-if="activeDanmuTab === 'settings'" class="danmu-settings">
        <!-- 发送弹幕的类型 -->
        <div class="setting-section">
          <h4 class="section-title">发送弹幕的类型</h4>
          <div class="radio-group">
            <el-radio-group v-model="danmuSendType" size="large">
              <el-radio value="all">允许发送所有类型的弹幕</el-radio>
              <el-radio value="specified">允许发送指定类型的弹幕</el-radio>
            </el-radio-group>
          </div>
        </div>
        
        <!-- 高级弹幕请求 -->
        <div class="setting-section">
          <h4 class="section-title">高级弹幕请求</h4>
          <el-select v-model="advancedDanmuRequest" placeholder="任何人" style="min-width: 200px;">
            <el-option label="任何人" value="anyone"></el-option>
            <el-option label="仅粉丝" value="fans"></el-option>
            <el-option label="仅关注3天以上粉丝" value="fans3d"></el-option>
            <el-option label="禁止所有人" value="none"></el-option>
          </el-select>
        </div>
        
        <!-- 黑名单 -->
        <div class="setting-section">
          <h4 class="section-title">黑名单</h4>
          <div class="setting-description">
            <p>添加方式：</p>
            <p>(1) 在网页播放器的弹幕列表上，右击弹幕选择"up主视频中禁言此用户"</p>
            <p>(2) 在创作中心弹幕管理的删除弹幕按钮上，点击下拉按钮选择"删除并拉黑该用户"</p>
          </div>
        </div>
        
        <!-- 关键词过滤 -->
        <div class="setting-section">
          <h4 class="section-title">关键词过滤</h4>
          <div class="filter-input-group">
            <el-input
              v-model="keywordFilterText"
              placeholder="输入关键词进行过滤。例如mdzz"
              style="flex: 1;"
            ></el-input>
            <el-button type="primary" style="margin-left: 10px;">添加</el-button>
          </div>
          <div class="setting-description">
            <p>输入关键词进行过滤，例如mdzz。观众将不能在你的视频中发送包含指定关键词的弹幕</p>
          </div>
        </div>
        
        <!-- 正则表达式过滤 -->
        <div class="setting-section">
          <h4 class="section-title">正则表达式过滤</h4>
          <div class="filter-input-group">
            <el-input
              v-model="regexFilterText"
              placeholder="输入正则表达式进行过滤"
              style="flex: 1;"
            ></el-input>
            <el-button type="primary" style="margin-left: 10px;">添加</el-button>
          </div>
          <div class="setting-description">
            <p>观众将不能在你的视频中发送匹配指定正则表达式的弹幕</p>
          </div>
        </div>
      </div>
      
      <!-- 弹幕反馈（仅在弹幕反馈标签页显示） -->
      <div v-else-if="activeDanmuTab === 'feedback'" class="danmu-feedback">
        <!-- 弹幕反馈子标签 -->
        <div class="danmu-feedback-tabs">
          <el-tabs v-model="activeDanmuFeedbackTab" type="card" size="small">
            <el-tab-pane label="弹幕举报" name="report"></el-tab-pane>
            <el-tab-pane label="高级请求" name="advanced"></el-tab-pane>
            <el-tab-pane label="弹幕保护" name="protection"></el-tab-pane>
          </el-tabs>
        </div>
        
        <!-- 弹幕反馈操作栏 -->
        <div class="danmu-feedback-header">
          <div class="feedback-actions">
            <el-button size="small" type="danger">删除弹幕</el-button>
            <el-button size="small" type="primary" plain>忽略举报</el-button>
          </div>
          
          <div class="feedback-filter">
            <el-select v-model="feedbackVideoFilter" placeholder="全部视频" size="small" style="min-width: 150px;">
              <el-option label="全部视频" value="all"></el-option>
              <el-option label="在byrut下载过游戏的可能会中挖矿病毒" value="video1"></el-option>
              <el-option label="陈梦曾经在贴吧的帖子和评论" value="video2"></el-option>
              <el-option label="NB化学实验室卡bug不用VIP做VIP实验" value="video3"></el-option>
            </el-select>
          </div>
        </div>
        
        <!-- 弹幕反馈列表 -->
        <div class="danmu-feedback-list">
          <el-table :data="danmuFeedbackList" stripe style="width: 100%">
            <el-table-column type="selection" width="55"></el-table-column>
            <el-table-column prop="content" label="弹幕内容" min-width="200"></el-table-column>
            <el-table-column prop="video" label="弹幕视频" min-width="250"></el-table-column>
            <el-table-column prop="sender" label="发送者" width="120"></el-table-column>
            <el-table-column prop="sendTime" label="发送时间" width="180" sortable></el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="scope">
                <el-button size="small" link class="text-danger">删除弹幕</el-button>
                <el-button size="small" link>忽略</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
        
        <!-- 弹幕列表显示上限提示 -->
        <div class="danmu-list-limit">
          弹幕列表显示上限10000条
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { Search, Refresh, MoreFilled, StarFilled } from '@element-plus/icons-vue'

// 弹幕管理相关数据
const activeDanmuTab = ref('article') // article: 稿件弹幕, settings: 弹幕设置, feedback: 弹幕反馈
const danmuSearchText = ref('') // 搜索文本
const danmuTypeFilter = ref('all') // all: 全部弹幕, normal: 普通弹幕, subtitle: 字幕弹幕
const danmuTimeFilter = ref('latest') // latest: 最近弹幕, week: 本周弹幕, month: 本月弹幕
const danmuVideoFilter = ref('all') // all: 全部视频, 其他: 具体视频

// 弹幕设置相关数据
const danmuSendType = ref('all') // all: 允许发送所有类型的弹幕, specified: 允许发送指定类型的弹幕
const advancedDanmuRequest = ref('anyone') // anyone: 任何人, fans: 仅粉丝, fans3d: 仅关注3天以上粉丝, none: 禁止所有人
const keywordFilterText = ref('') // 关键词过滤输入框
const regexFilterText = ref('') // 正则表达式过滤输入框

// 弹幕反馈相关数据
const activeDanmuFeedbackTab = ref('report') // report: 弹幕举报, advanced: 高级请求, protection: 弹幕保护
const feedbackVideoFilter = ref('all') // all: 全部视频, 其他: 具体视频
const danmuFeedbackList = ref([
  { id: 1, content: '该', video: '在byrut下载过游戏的可能会中挖矿病毒', sender: '为梦绘色', sendTime: '2025-11-18 16:32:19' },
  { id: 2, content: '这不是很正常？很多正规软件也...', video: '在byrut下载过游戏的可能会中挖矿病毒', sender: '骨痂骨痂', sendTime: '2025-09-28 10:17:30' },
  { id: 3, content: '受着', video: '在byrut下载过游戏的可能会中挖矿病毒', sender: '爷的昵称值6个币', sendTime: '2025-09-19 10:50:24' },
  { id: 4, content: '这吐舌真难绷', video: '陈梦曾经在贴吧的帖子和评论', sender: '蒙古上单限定版', sendTime: '2025-05-20 23:52:37' },
  { id: 5, content: '这表情没绷住', video: '陈梦曾经在贴吧的帖子和评论', sender: 'AEFf_', sendTime: '2025-05-20 23:52:35' },
  { id: 6, content: 'byd吸嗨了', video: 'NB化学实验室卡bug不用VIP做VIP实验', sender: 'bili_114514881', sendTime: '2025-04-17 16:21:38' },
  { id: 7, content: '弹幕美术服', video: 'NB化学实验室卡bug不用VIP做VIP实验', sender: '失侍大王', sendTime: '2025-03-08 08:55:57' }
])

// 模拟弹幕数据
const danmus = ref([
  {
    id: 1,
    selected: false,
    sender: '浅忆Official',
    playTime: '03:43',
    content: '还不如删文件呢',
    likes: 0,
    type: '普通',
    sendTime: '2026-01-24 17:59:44',
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\ttaskhost.exe'
  },
  {
    id: 2,
    selected: false,
    sender: '無所見即我',
    playTime: '03:58',
    content: '360很容易啊，除了小白',
    likes: 0,
    type: '普通',
    sendTime: '2026-01-21 10:17:48',
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\ttaskhost.exe'
  },
  {
    id: 3,
    selected: false,
    sender: '夜月txllwhmc',
    playTime: '01:49',
    content: '这玩意儿还装着观赏吗',
    likes: 0,
    type: '普通',
    sendTime: '2026-01-10 03:15:34',
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\ttaskhost.exe'
  },
  {
    id: 4,
    selected: false,
    sender: '此人懒到不想取名字',
    playTime: '01:15',
    content: '另外下载的啥？',
    likes: 0,
    type: '普通',
    sendTime: '2026-01-09 16:20:59',
    videoTitle: 'proxypin 开源免费用的...'
  },
  {
    id: 5,
    selected: false,
    sender: '富春渔夫',
    playTime: '00:22',
    content: '好中二',
    likes: 0,
    type: '普通',
    sendTime: '2025-12-23 00:30:22',
    videoTitle: '符月华 贴吧 没有上某...'
  },
  {
    id: 6,
    selected: false,
    sender: '上紧固SFU',
    playTime: '01:32',
    content: '666',
    likes: 0,
    type: '普通',
    sendTime: '2025-12-19 17:33:50',
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\ttaskhost.exe'
  }
])

// 搜索弹幕
const searchDanmu = () => {
  // 实际项目中这里会调用API进行搜索
  console.log('搜索弹幕:', danmuSearchText.value)
}
</script>

<style scoped>
.danmu-manager {
  width: 100%;
}

.danmu-management {
  width: 100%;
}

/* 弹幕标签页和搜索组合样式 */
.danmu-tabs-with-search {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

/* 弹幕标签页样式 */
.danmu-tabs {
  flex: 1;
}

/* 弹幕搜索样式 */
.danmu-search {
  display: flex;
  align-items: center;
  margin-left: 20px;
}

/* 弹幕头部样式（操作栏和筛选条件） */
.danmu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #e0e0e0;
  margin-bottom: 15px;
}

/* 弹幕筛选条件样式 */
.danmu-filters {
  display: flex;
  align-items: center;
  margin-left: auto;
}

/* 弹幕操作按钮样式 */
.danmu-actions {
  display: flex;
}

/* 弹幕设置样式 */
.danmu-settings {
  width: 100%;
}

/* 设置区块样式 */
.setting-section {
  margin-bottom: 30px;
  padding: 20px;
  background-color: #fafafa;
  border-radius: 8px;
  border: 1px solid #e0e0e0;
}

/* 区块标题样式 */
.setting-section .section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e0e0e0;
  display: block;
}

/* 单选按钮组样式 */
.radio-group {
  margin-top: 10px;
}

/* 过滤输入框组样式 */
.filter-input-group {
  display: flex;
  align-items: center;
  margin-bottom: 10px;
}

/* 设置描述样式 */
.setting-description {
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
  margin-top: 10px;
  padding: 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

/* 弹幕反馈样式 */
.danmu-feedback {
  width: 100%;
}

/* 弹幕反馈子标签样式 */
.danmu-feedback-tabs {
  margin-bottom: 20px;
}

/* 弹幕反馈操作栏样式 */
.danmu-feedback-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #e0e0e0;
  margin-bottom: 15px;
}

/* 反馈操作按钮样式 */
.feedback-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 反馈筛选样式 */
.feedback-filter {
  display: flex;
  align-items: center;
}

/* 弹幕反馈列表样式 */
.danmu-feedback-list {
  margin-bottom: 20px;
}

/* 弹幕列表显示上限提示样式 */
.danmu-list-limit {
  text-align: right;
  color: #909399;
  font-size: 12px;
  margin-top: 10px;
}

/* 文本样式 - 危险色 */
.text-danger {
  color: #f56c6c;
}

.danmu-actions .el-button {
  min-width: auto;
  padding: 0 8px;
  height: 28px;
  line-height: 28px;
  font-size: 12px;
}

/* 弹幕列表样式 */
.danmu-list {
  margin-top: 20px;
}

/* 弹幕列表中的视频信息样式 */
.danmu-video-info {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

/* 表格样式调整 */
.danmu-list .el-table {
  border: 1px solid #ebeef5;
  border-radius: 4px;
}

.danmu-list .el-table__header-wrapper th {
  background-color: #f5f7fa;
  font-weight: 500;
  color: #303133;
}

.danmu-list .el-table__body-wrapper td {
  padding: 12px 0;
}
</style>