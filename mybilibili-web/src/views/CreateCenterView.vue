<template>
  <div class="create-center-container">
    <!-- 顶部导航栏 -->
    <header class="create-center-header">
      <div class="header-left">
        <el-button class="center-title-btn" @click="goToCreateCenterHome">
          Bilibili创作中心
        </el-button>
        <el-button class="main-site-btn" @click="goToMainSite">
          <el-icon><House /></el-icon>
          <span>主站</span>
        </el-button>
      </div>
      <div class="header-right">
        <el-avatar 
          class="user-avatar" 
          :size="40" 
          :src="currentUser?.avatar || 'https://i0.hdslb.com/bfs/face/3378829f555891d2d5a4537e10264593a1d076b1.jpg@50w_50h_1c_1s_!web-avatar-nav.avif'"
          @click="goToUserProfile"
          style="cursor: pointer;"
        ></el-avatar>
        <div class="up-day-box">
          成为UP主的第123天
        </div>
      </div>
    </header>

    <!-- 主体内容区域 -->
    <div class="create-center-main">
      <!-- 侧边导航栏 -->
      <aside class="sidebar">
        <!-- 独立的投稿按钮 -->
        <el-button type="primary" class="upload-btn-large" @click="goToUpload">
          <el-icon><Upload /></el-icon>
          <span>投稿</span>
        </el-button>
        
        <!-- 侧边导航菜单 -->
        <el-menu
          ref="menuRef"
          default-active="home"
          class="sidebar-menu"
          :unique-opened="false"
          @select="handleMenuSelect"
        >
          <el-menu-item index="home">
            <el-icon><House /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-sub-menu index="content">
            <template #title>
              <el-icon><Document /></el-icon>
              <span>内容管理</span>
            </template>
            <el-menu-item index="content-articles">
              <el-icon><Menu /></el-icon>
              <span>稿件管理</span>
            </el-menu-item>

          </el-sub-menu>
          <el-menu-item index="data">
            <el-icon><DataAnalysis /></el-icon>
            <span>数据中心</span>
          </el-menu-item>
          <el-menu-item index="fans">
            <el-icon><UserFilled /></el-icon>
            <span>粉丝管理</span>
          </el-menu-item>
          <el-sub-menu index="interaction">
            <template #title>
              <el-icon><ChatDotRound /></el-icon>
              <span>互动管理</span>
            </template>
            <el-menu-item index="interaction-comment">
              <el-icon><Comment /></el-icon>
              <span>评论管理</span>
            </el-menu-item>
          </el-sub-menu>


        </el-menu>
      </aside>

      <!-- 主内容区域 -->
      <main class="content-area">
        <!-- 首页内容 -->
        <div v-if="currentActive === 'home'" class="content-section">
          <div class="content-body">
            <p>视频数据</p>
            
            <!-- Loading状态 -->
            <div v-if="homeLoading" class="loading-container" v-loading="homeLoading" element-loading-text="加载中..."></div>
            
            <!-- 错误提示 -->
            <el-alert v-else-if="homeError" :title="homeError" type="error" show-icon :closable="false" style="margin-bottom: 20px;" />
            
            <template v-else>
            <!-- 第一行统计数据：粉丝总数、播放量、评论、弹幕 -->
            <div class="dashboard-stats">
              <div class="stat-card">
                <div class="stat-number">{{ statsData.followerCount }}</div>
                <div class="stat-label">粉丝总数</div>
              </div>
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalViewCount }}</div>
                <div class="stat-label">总播放量</div>
              </div>
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalCommentCount }}</div>
                <div class="stat-label">总评论数</div>
              </div>
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalDanmuCount }}</div>
                <div class="stat-label">总弹幕数</div>
              </div>
            </div>
            
            <!-- 第二行统计数据：点赞、分享、收藏、投币 -->
            <div class="dashboard-stats">
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalLikeCount }}</div>
                <div class="stat-label">总点赞数</div>
              </div>
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalShareCount }}</div>
                <div class="stat-label">总分享数</div>
              </div>
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalFavoriteCount }}</div>
                <div class="stat-label">总收藏数</div>
              </div>
              <div class="stat-card">
                <div class="stat-number">{{ statsData.totalCoinCount }}</div>
                <div class="stat-label">总投币数</div>
              </div>
            </div>
            </template>
            
            <!-- 评论/弹幕选择栏 -->
            <div class="comment-danmu-section">
              <div class="section-header">
                <div class="tab-buttons">
                  <el-button 
                    type="primary" 
                    :plain="activeCommentTab !== 'comment'"
                    @click="activeCommentTab = 'comment'"
                    size="small"
                  >
                    评论
                  </el-button>
                  <el-button 
                    type="primary" 
                    :plain="activeCommentTab !== 'danmu'"
                    @click="activeCommentTab = 'danmu'"
                    size="small"
                  >
                    弹幕
                  </el-button>
                </div>
              </div>
              
              <!-- 评论列表 -->
              <div v-if="activeCommentTab === 'comment'" class="interaction-list">
                <div v-for="comment in latestComments.slice(0, 5)" :key="comment.id" class="interaction-item">
                  <div class="user-info">
                    <el-avatar :size="32" :src="comment.avatar"></el-avatar>
                    <span class="username">{{ comment.username }}</span>
                  </div>
                  <div class="interaction-content">
                    {{ comment.content }}
                  </div>
                  <div class="interaction-time">{{ comment.time }}</div>
                  <div class="interaction-actions">
                    <el-button type="danger" size="small" link @click="handleDeleteComment(comment.id)">删除</el-button>
                  </div>
                </div>
              </div>
              
              <!-- 弹幕列表 -->
              <div v-else-if="activeCommentTab === 'danmu'" class="interaction-list" v-loading="danmakuLoading">
                <div v-if="danmakuList.length === 0" class="developing-tip">
                  <el-empty description="暂无弹幕数据" :image-size="100" />
                </div>
                <div v-else class="danmaku-list">
                  <div v-for="item in danmakuList" :key="item.id" class="danmaku-card">
                    <div class="danmaku-body">
                      <div class="danmaku-content">
                        {{ item.content }}
                      </div>
                    </div>
                    <div class="danmaku-footer">
                      <div class="danmaku-info">
                        <el-link 
                          type="primary" 
                          underline="never"
                          @click="goToVideo(item.manuscriptId, item.time, item.videoOrder)"
                          class="danmaku-link"
                        >
                          {{ item.videoName || '未知视频' }}
                        </el-link>
                        <el-tag size="small" :type="getDanmakuModeType(item.mode)" effect="plain">
                          {{ formatTime(item.time) }}
                        </el-tag>
                        <el-tag size="small" type="info" effect="plain">
                          {{ getDanmakuModeText(item.mode) }}
                        </el-tag>
                      </div>
                      <div class="danmaku-date">{{ item.createTime }}</div>
                      <div class="danmaku-actions">
                        <el-button type="danger" size="small" link @click="deleteDanmaku(item.id)">删除</el-button>
                      </div>
                    </div>
                  </div>
                </div>
                <el-pagination
                  v-if="danmakuTotal > danmakuPageSize"
                  v-model:current-page="danmakuCurrentPage"
                  :page-size="danmakuPageSize"
                  :total="danmakuTotal"
                  layout="prev, pager, next"
                  @current-change="fetchDanmakuList"
                />
              </div>
            </div>
            
            <!-- 观看排行和互动排行选择栏 -->
            <div class="ranking-section">
              <div class="section-header">
                <div class="tab-buttons">
                  <el-button 
                    type="primary" 
                    :plain="activeRankingTab !== 'view'"
                    @click="activeRankingTab = 'view'"
                    size="small"
                  >
                    观看排行
                  </el-button>
                  <el-button 
                    type="primary" 
                    :plain="activeRankingTab !== 'interaction'"
                    @click="activeRankingTab = 'interaction'"
                    size="small"
                  >
                    互动排行
                  </el-button>
                </div>
              </div>
              
              <!-- 观看排行列表 -->
              <div v-if="activeRankingTab === 'view'" class="ranking-list-horizontal">
                <div v-for="(user, index) in viewRanking" :key="user.id" class="ranking-item-horizontal">
                  <div class="user-info">
                    <el-avatar :size="32" :src="user.avatar" :class="getRankingClass(index)"></el-avatar>
                    <span class="username" :class="getRankingClass(index)">{{ user.username }}</span>
                  </div>
                </div>
              </div>
              
              <!-- 互动排行列表 -->
              <div v-else-if="activeRankingTab === 'interaction'" class="ranking-list-horizontal">
                <div v-for="(user, index) in interactionRanking" :key="user.id" class="ranking-item-horizontal">
                  <div class="user-info">
                    <el-avatar :size="32" :src="user.avatar" :class="getRankingClass(index)"></el-avatar>
                    <span class="username" :class="getRankingClass(index)">{{ user.username }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 稿件管理内容 -->
        <div v-else-if="currentActive === 'content-articles'" class="content-section">
          <!-- 移除标题 -->
          <div class="content-body">
            <!-- 主要选择栏 -->
            <div class="content-tabs">
              <el-tabs v-model="mainTab" type="card">
                <el-tab-pane label="视频管理" name="video"></el-tab-pane>
                <el-tab-pane label="合集管理" name="collection"></el-tab-pane>
              </el-tabs>
            </div>
            
            <!-- 视频管理次级选择栏 -->
            <div v-if="mainTab === 'video'" class="video-filters">
              <!-- 第一行：全部稿件 -->
              <div class="filter-row">
                <el-radio-group v-model="statusFilter" size="small">
                  <el-radio-button value="processing">进行中</el-radio-button>
                  <el-radio-button value="published">已通过 ({{ approvedCount }})</el-radio-button>
                  <el-radio-button value="rejected">未通过 ({{ rejectedCount }})</el-radio-button>
                  <el-radio-button value="unpublished">已下架</el-radio-button>
                </el-radio-group>
              </div>
            </div>
            
            <!-- 视频稿件列表 -->
            <div v-if="mainTab === 'video'" class="article-list" v-loading="articlesLoading">
              <el-table :data="articles" stripe style="width: 100%">
                <el-table-column prop="id" label="稿件ID" width="120"></el-table-column>
                <el-table-column prop="title" label="标题" min-width="300"></el-table-column>
                <el-table-column prop="status" label="状态" width="180">
                  <template #default="scope">
                    <!-- 进行中状态显示进度条 -->
                    <div v-if="scope.row.status === 0 || scope.row.status === 1" class="progress-cell">
                      <el-progress 
                        :percentage="articleProgress[scope.row.id] || 0" 
                        :status="articleProgressStatus[scope.row.id] === 'completed' ? 'success' : ''"
                        :stroke-width="6"
                        :show-text="true"
                      />
                      <span class="progress-text">
                        {{ articleProgressStatus[scope.row.id] === 'completed' ? '已完成' : (articleProgress[scope.row.id] >= 90 ? '即将完成' : '处理中') }}
                      </span>
                    </div>
                    <!-- 其他状态显示标签 -->
                    <el-tag 
                      v-else
                      :type="getArticleStatusType(scope.row.status)" 
                      size="small"
                    >
                      {{ getArticleStatusText(scope.row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="viewCount" label="播放量" width="100"></el-table-column>
                <el-table-column prop="commentCount" label="评论数" width="100"></el-table-column>
                <el-table-column prop="createdAt" label="创建时间" width="180">
                  <template #default="scope">
                    {{ formatDate(scope.row.uploadTime) }}
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="200" fixed="right">
                  <template #default="scope">
                    <!-- 进行中状态 - 显示进度完成后的查看按钮 -->
                    <template v-if="scope.row.status === 0 || scope.row.status === 1">
                      <el-button 
                        v-if="articleProgressStatus[scope.row.id] === 'completed'" 
                        type="success" 
                        size="small"
                        @click="router.push(`/manuscript/${scope.row.id}`)"
                      >
                        查看
                      </el-button>
                      <el-button 
                        v-else 
                        type="info" 
                        size="small"
                        disabled
                      >
                        处理中...
                      </el-button>
                    </template>
                    <!-- 已发布状态 -->
                    <el-button v-if="scope.row.status === 3" type="warning" size="small" @click="unpublishArticle(scope.row.id)">下架</el-button>
                    <!-- 已下架状态 -->
                    <el-button v-if="scope.row.status === -1" type="success" size="small" @click="publishArticle(scope.row.id)">上架</el-button>
                    <!-- 删除按钮 -->
                    <el-button type="danger" size="small" @click="deleteArticle(scope.row.id)">删除</el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
            
            <!-- 分页导航栏 -->
            <div v-if="mainTab === 'video'" class="pagination">
              <el-pagination
                v-model:current-page="currentPage"
                v-model:page-size="pageSize"
                :page-sizes="[10, 20, 50, 100]"
                layout="total, sizes, prev, pager, next, jumper"
                :total="totalArticles"
                @size-change="handleSizeChange"
                @current-change="handleCurrentChange"
              ></el-pagination>
            </div>
            
            <!-- 合集管理内容 -->
            <div v-if="mainTab === 'collection'" class="collection-management">
              <!-- 合集详情视图 -->
              <div v-if="collectionDetail.visible" class="collection-detail-view">
                <!-- 返回按钮和标题 -->
                <div class="collection-detail-header">
                  <div class="header-left">
                    <el-button text @click="backToCollectionsList">
                      <el-icon><ArrowRight style="transform: rotate(180deg)" /></el-icon>
                      返回
                    </el-button>
                    <span class="header-title">我的合集和视频列表</span>
                    <el-icon><ArrowRight /></el-icon>
                    <span class="collection-name">{{ collectionDetail.collection?.title || '加载中...' }}</span>
                  </div>
                  <div class="header-right">
                    <el-button text @click="openEditCollectionDialog">
                      <el-icon><Setting /></el-icon>
                      编辑
                    </el-button>
                  </div>
                </div>

                <!-- 合集信息区 -->
                <div class="collection-info-section" v-if="collectionDetail.collection">
                  <div class="info-container">
                    <!-- 封面 -->
                    <div class="collection-cover-large">
                      <img
                        :src="collectionDetail.collection.coverUrl || getDefaultCover()"
                        :alt="collectionDetail.collection.title"
                      />
                      <div class="cover-overlay" @click="playCollectionAll">
                        <el-button type="primary" size="large" :icon="VideoPlay">
                          播放全部
                        </el-button>
                      </div>
                      <div class="video-count-badge">
                        <el-icon><VideoPlay /></el-icon>
                        <span>{{ collectionDetail.pagination.total }} 个视频</span>
                      </div>
                    </div>

                    <!-- 信息 -->
                    <div class="collection-meta-detail">
                      <h1 class="collection-title-large">{{ collectionDetail.collection.title }}</h1>
                      <p class="collection-desc-large">{{ collectionDetail.collection.description || '暂无描述' }}</p>

                      <!-- 统计信息 -->
                      <div class="stats-info">
                        <span class="stat-item">
                          <el-icon><Monitor /></el-icon>
                          {{ collectionDetail.collection.viewCount || 0 }} 次观看
                        </span>
                        <span class="stat-item">
                          更新于 {{ formatDate(collectionDetail.collection.updatedAt) }}
                        </span>
                        <span v-if="!collectionDetail.collection.isPublic" class="private-tag">私密合集</span>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- 视频网格区 -->
                <div class="collection-videos-section">
                  <!-- 头部操作栏 -->
                  <div class="videos-header">
                    <div class="sort-options">
                      <span 
                        class="sort-item" 
                        :class="{ active: collectionDetail.sortBy === 'default' }"
                        @click="handleCollectionDetailSortChange('default')"
                      >
                        默认排序
                      </span>
                      <span 
                        class="sort-item" 
                        :class="{ active: collectionDetail.sortBy === 'newest' }"
                        @click="handleCollectionDetailSortChange('newest')"
                      >
                        升序排序
                      </span>
                    </div>
                    <div class="header-actions">
                      <el-button type="primary" :icon="Setting" @click="openEditCollectionDialog">
                        编辑
                      </el-button>
                      <el-dropdown trigger="click" @command="handleCollectionCommand">
                        <el-button :icon="MoreFilled" circle />
                        <template #dropdown>
                          <el-dropdown-menu>
                            <el-dropdown-item command="delete">删除视频列表</el-dropdown-item>
                          </el-dropdown-menu>
                        </template>
                      </el-dropdown>
                    </div>
                  </div>

                  <!-- 加载状态 -->
                  <div v-if="collectionDetail.loading" class="loading-state">
                    <el-skeleton :rows="3" animated />
                  </div>

                  <!-- 空状态 -->
                  <div v-else-if="collectionDetail.manuscripts.length === 0" class="empty-state">
                    <el-empty description="暂无视频" />
                  </div>

                  <!-- 视频网格 -->
                  <div v-else class="videos-grid">
                    <!-- 添加视频卡片 -->
                    <div class="video-card add-video-card" @click="openAddVideoToCollectionDialog">
                      <div class="add-video-content">
                        <el-icon :size="40"><Plus /></el-icon>
                        <span>添加稿件</span>
                      </div>
                    </div>

                    <!-- 视频卡片 -->
                    <div
                      v-for="(manuscript, index) in collectionDetail.manuscripts"
                      :key="manuscript.id"
                      class="collection-video-card"
                    >
                      <!-- 封面区域 -->
                      <div class="collection-video-cover" @click="playManuscript(manuscript)">
                        <img
                          :src="manuscript.coverUrl || getDefaultCover()"
                          :alt="manuscript.title"
                          class="collection-video-cover-img"
                        />
                        <!-- 序号 -->
                        <div class="collection-video-index">{{ index + 1 }}</div>
                        <!-- 时长 -->
                        <div class="collection-video-duration">{{ manuscript.duration || '00:00' }}</div>
                        <!-- 更多操作 -->
                        <div class="collection-video-actions" @click.stop>
                          <el-dropdown trigger="click" @command="(cmd) => handleVideoCommand(cmd, manuscript)">
                            <el-button :icon="MoreFilled" circle size="small" />
                            <template #dropdown>
                              <el-dropdown-menu>
                                <el-dropdown-item command="remove">从列表中移除</el-dropdown-item>
                              </el-dropdown-menu>
                            </template>
                          </el-dropdown>
                        </div>
                      </div>

                      <!-- 信息区域 -->
                      <div class="collection-video-info">
                        <h3 class="collection-video-title" :title="manuscript.title">{{ manuscript.title }}</h3>
                        <div class="collection-video-meta">
                          <span class="collection-meta-item">
                            <el-icon><VideoPlay /></el-icon>
                            {{ formatNumber(manuscript.viewCount) }}
                          </span>
                          <span class="collection-meta-item">{{ formatDate(manuscript.uploadTime) }}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 合集列表视图 -->
              <template v-else>
                <!-- 顶部标题和新建按钮 -->
                <div class="collections-header">
                  <div class="collections-header-left">
                    <h2 class="collections-title">我的合集和视频列表</h2>
                    <el-button
                      type="primary"
                      :icon="Plus"
                      class="new-collection-btn"
                      @click="openCreateCollectionDialog"
                    >
                      新建
                    </el-button>
                  </div>
                  <!-- 视图切换按钮 -->
                  <div class="view-toggle">
                    <button
                      class="view-toggle-btn grid-view-btn"
                      :class="{ active: myCollections.viewType === 'grid' }"
                      @click="myCollections.viewType = 'grid'"
                    >
                      <el-icon><Menu /></el-icon>
                    </button>
                    <button
                      class="view-toggle-btn list-view-btn"
                      :class="{ active: myCollections.viewType === 'horizontal' }"
                      @click="myCollections.viewType = 'horizontal'"
                    >
                      <el-icon><Document /></el-icon>
                    </button>
                  </div>
                </div>

                <!-- 加载状态 -->
                <div v-if="myCollections.loading" class="loading-state">
                  <el-skeleton :rows="3" animated />
                </div>

                <!-- 空状态 -->
                <div v-else-if="myCollections.items.length === 0" class="empty-state">
                  <el-empty description="暂无合集">
                    <el-button
                      type="primary"
                      @click="openCreateCollectionDialog"
                    >
                      创建合集
                    </el-button>
                  </el-empty>
                </div>

                <!-- 宫格视图 -->
                <div v-else-if="myCollections.viewType === 'grid'" class="collections-grid">
                  <!-- 新建合集卡片 -->
                  <div
                    class="collection-item new-collection"
                    @click="openCreateCollectionDialog"
                  >
                    <div class="new-collection-content">
                      <el-icon :size="32"><Plus /></el-icon>
                      <span>新建合集</span>
                    </div>
                  </div>

                  <!-- 合集项 -->
                  <div
                    v-for="collection in myCollections.items"
                    :key="collection.id"
                    class="collection-item"
                    @click="goToCollectionDetail(collection.id)"
                  >
                    <div class="collection-cover">
                      <img :src="collection.coverUrl || getDefaultCover()" :alt="collection.title" class="collection-cover-img">
                      <div class="collection-video-count">
                        <el-icon><VideoPlay /></el-icon>
                        <span>{{ collection.manuscriptCount || 0 }}</span>
                      </div>
                    </div>
                    <div class="collection-info">
                      <div class="collection-title">{{ collection.title }}</div>
                      <div class="collection-date">{{ collection.manuscriptCount || 0 }}个视频</div>
                    </div>
                  </div>
                </div>

                <!-- 水平列表视图 -->
                <div v-else-if="myCollections.viewType === 'horizontal'" class="collections-horizontal">
                  <!-- 合集项 -->
                  <div
                    v-for="collection in myCollections.items"
                    :key="collection.id"
                    class="collection-horizontal-item"
                  >
                    <div class="collection-horizontal-header">
                      <h3 class="collection-horizontal-title">
                        {{ collection.title }}
                        <span class="collection-video-count-badge">{{ collection.manuscriptCount || 0 }}</span>
                      </h3>
                      <div class="collection-horizontal-actions">
                        <el-button
                          v-if="collection.videos && collection.videos.length > 0"
                          class="action-btn play-all-btn"
                          :icon="VideoPlay"
                          @click="playCollectionAllFromList(collection)"
                        >
                          播放全部
                        </el-button>
                        <el-button
                          class="action-btn more-btn"
                          @click="goToCollectionDetail(collection.id)"
                        >
                          更多
                          <el-icon><ArrowRight /></el-icon>
                        </el-button>
                      </div>
                    </div>

                    <!-- 水平视频列表 -->
                    <div class="collection-videos-horizontal">
                      <!-- 添加视频按钮 -->
                      <div
                        class="add-video-card"
                        @click="openAddVideoDialog(collection.id)"
                      >
                        <div class="add-video-content">
                          <el-icon :size="24"><Plus /></el-icon>
                          <span>添加稿件</span>
                        </div>
                      </div>

                      <!-- 视频项 -->
                      <div
                        v-for="video in collection.videos"
                        :key="video.id"
                        class="video-horizontal-item"
                        @click="router.push(`/manuscript/${video.id}`)"
                      >
                        <div class="video-horizontal-cover">
                          <img :src="video.coverUrl || getDefaultCover()" :alt="video.title" class="video-cover-img">
                          <div class="video-duration">{{ video.duration }}</div>
                        </div>
                        <div class="video-horizontal-info">
                          <div class="video-title" :title="video.title">{{ video.title }}</div>
                          <div class="video-horizontal-meta">
                            <span class="video-views">
                              <el-icon><Monitor /></el-icon>
                              {{ video.viewCount || 0 }}
                            </span>
                            <span class="video-date">{{ video.date }}</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </div>

                <!-- 数据中心内容 -->
        <div v-else-if="currentActive === 'data'" class="content-section">
          <DataCenterView />
        </div>

        <!-- 评论管理内容 -->
        <div v-else-if="currentActive === 'interaction-comment'" class="content-section">
          <!-- 移除标题 -->
          <div class="content-body">
            <div class="comment-management">
              <!-- 主标签页和搜索框 -->
              <div class="main-tabs-with-search">
                <!-- 视频评论蓝色字样 -->
                <div class="video-comment-label">视频评论</div>

                <!-- 搜索框 -->
                <div class="main-search">
                  <el-input
                    v-model="commentSearchText"
                    placeholder="搜索视频评论"
                    size="small"
                    style="width: 200px;"
                    clearable
                  >
                    <template #append>
                      <el-button size="small" @click="searchComments"><el-icon><Search /></el-icon></el-button>
                    </template>
                  </el-input>
                </div>
              </div>
              
              <!-- 搜索和过滤区域 -->
              <div class="comment-filter-bar">
                <div class="left-section">
                  <!-- 子标签页：视频评论、专栏评论、音频评论 -->
                  <div class="sub-tabs">
                    <el-radio-group v-model="activeCommentSubTab" size="small">
                      <el-radio-button value="video">视频评论</el-radio-button>

                      <el-radio-button value="audio">音频评论</el-radio-button>
                    </el-radio-group>
                  </div>
                </div>
                
                <div class="right-section">
                  <!-- 评论类型和视频筛选 -->
                  <div class="filter-dropdowns">
                    <el-select v-model="commentTypeFilter" placeholder="全部评论" size="small" style="min-width: 120px; margin-right: 10px;">
                      <el-option label="全部评论" value="all"></el-option>
                      <el-option label="评论" value="comment"></el-option>
                      <el-option label="回复" value="reply"></el-option>
                    </el-select>
                    
                    <el-select v-model="videoFilter" placeholder="全部视频" size="small" style="min-width: 120px;">
                      <el-option label="全部视频" value="all"></el-option>
                      <el-option
                        v-for="video in videoList"
                        :key="video.id"
                        :label="video.title"
                        :value="video.id"
                      ></el-option>
                    </el-select>
                  </div>
                </div>
              </div>
              
              <!-- 操作栏 -->
              <div class="comment-actions">
                <div class="action-buttons">
                  <el-button size="small" plain @click="handleSelectAll(true)">全选</el-button>
                  <el-button size="small" plain @click="handleBatchDelete">删除</el-button>
                </div>
                
                <!-- 排序选项 -->
                <div class="sort-options">
                  <el-radio-group v-model="commentSortBy" size="small">
                    <el-radio-button value="latest">最近发布</el-radio-button>
                    <el-radio-button value="likes">点赞最多</el-radio-button>
                    <el-radio-button value="replies">回复最多</el-radio-button>
                  </el-radio-group>
                </div>
              </div>
              
              <!-- 评论列表 -->
              <div class="comment-list">
                <div 
                  v-for="comment in comments" 
                  :key="comment.id" 
                  class="comment-item"
                >
                  <!-- 复选框 -->
                  <div class="comment-checkbox">
                    <el-checkbox v-model="comment.selected"></el-checkbox>
                  </div>
                  
                  <!-- 评论主体：头像、用户名、内容、操作 -->
                  <div class="comment-main">
                    <!-- 头像和用户名 -->
                    <div class="comment-header">
                      <el-avatar :size="40" :src="comment.avatar"></el-avatar>
                      <span class="username">{{ comment.username }}</span>
                    </div>
                    
                    <!-- 评论内容 -->
                    <div class="comment-content">
                      {{ comment.content }}
                    </div>
                    
                    <!-- 评论时间和操作 -->
                    <div class="comment-meta">
                      <span class="comment-time">{{ comment.time }}</span>
                      <el-button size="small" plain @click="openReplyDialog(comment)">
                        <el-icon><ChatDotRound /></el-icon>回复
                      </el-button>
                      
                      <!-- 删除按钮（鼠标悬停显示） -->
                      <div class="comment-actions-hover">
                        <el-button size="small" plain type="danger">
                          <el-icon><Delete /></el-icon>删除
                        </el-button>
                      </div>
                    </div>
                  </div>
                  
                  <!-- 视频缩略图 -->
                  <div class="comment-right">
                    <div class="video-thumbnail" v-if="comment.videoThumbnail">
                      <img :src="comment.videoThumbnail" alt="视频缩略图">
                      <div class="video-title">{{ comment.videoTitle }}</div>
                    </div>
                  </div>
                </div>
              </div>
              
              <!-- 分页 -->
              <div class="comment-pagination">
                <div class="custom-pagination">
                  <el-button 
                    v-for="(page, index) in visiblePages" 
                    :key="index"
                    :type="page === commentCurrentPage ? 'primary' : 'default'"
                    :plain="page !== commentCurrentPage"
                    :disabled="page === '...'"
                    @click="typeof page === 'number' && (commentCurrentPage = page)"
                    size="small"
                  >
                    {{ page }}
                  </el-button>
                  
                  <el-button 
                    v-if="commentCurrentPage < totalPages" 
                    @click="commentCurrentPage++"
                    size="small"
                  >
                    下一页
                  </el-button>
                  
                  <div class="pagination-info">
                    共{{ totalPages }}页 / {{ totalComments }}个
                  </div>
                </div>
              </div>
            </div>

            <el-dialog v-model="replyDialogVisible" title="回复评论" width="500px">
              <el-input
                v-model="replyContent"
                type="textarea"
                :rows="4"
                placeholder="请输入回复内容"
                maxlength="500"
                show-word-limit
              />
              <template #footer>
                <el-button @click="replyDialogVisible = false">取消</el-button>
                <el-button type="primary" @click="handleReplyComment">发送</el-button>
              </template>
            </el-dialog>
          </div>
        </div>

        <div v-else-if="currentActive === 'interaction-danmu'" class="content-section">
          <!-- 移除标题 -->
          <div class="content-body">
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
                        <el-button size="small" type="text" class="text-danger">删除弹幕</el-button>
                        <el-button size="small" type="text">忽略</el-button>
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
        </div>





                

                

                

                

                


        <!-- 投稿内容 -->
        <div v-else-if="currentActive === 'upload'" class="content-section">
          <UploadView />
        </div>
      </main>
    </div>

    <!-- 新建合集对话框 -->
    <el-dialog
      v-model="createCollectionDialogVisible"
      title="新建合集"
      width="600px"
    >
      <el-form :model="createCollectionForm" label-width="80px">
        <el-form-item label="合集名称" required>
          <el-input v-model="createCollectionForm.name" placeholder="请输入合集名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="合集描述">
          <el-input
            v-model="createCollectionForm.description"
            type="textarea"
            placeholder="请输入合集描述"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="合集封面">
          <el-upload
            action="#"
            :on-change="handleCreateCollectionCoverChange"
            :auto-upload="false"
            accept="image/*"
            :show-file-list="false"
          >
            <div class="cover-preview-small" v-if="createCollectionForm.coverUrl">
              <img :src="createCollectionForm.coverUrl" alt="封面">
            </div>
            <div class="cover-placeholder-small" v-else>
              <el-icon><Picture /></el-icon>
              <span>上传封面</span>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="公开">
          <el-switch v-model="createCollectionForm.isPublic" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createCollectionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateCollection" :loading="creatingCollection">创建</el-button>
      </template>
    </el-dialog>

    <!-- 编辑合集对话框 -->
    <el-dialog
      v-model="editCollectionDialogVisible"
      title="编辑合集"
      width="600px"
    >
      <el-form :model="editCollectionForm" label-width="80px">
        <el-form-item label="合集名称" required>
          <el-input v-model="editCollectionForm.name" placeholder="请输入合集名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="合集描述">
          <el-input
            v-model="editCollectionForm.description"
            type="textarea"
            placeholder="请输入合集描述"
            :rows="3"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        <el-form-item label="合集封面">
          <el-upload
            action="#"
            :on-change="handleEditCollectionCoverChange"
            :auto-upload="false"
            accept="image/*"
            :show-file-list="false"
          >
            <div class="cover-preview-small" v-if="editCollectionForm.coverUrl">
              <img :src="editCollectionForm.coverUrl" alt="封面">
            </div>
            <div class="cover-placeholder-small" v-else>
              <el-icon><Picture /></el-icon>
              <span>上传封面</span>
            </div>
          </el-upload>
        </el-form-item>
        <el-form-item label="公开">
          <el-switch v-model="editCollectionForm.isPublic" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="editCollectionDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleUpdateCollection" :loading="updatingCollection">保存</el-button>
      </template>
    </el-dialog>

    <!-- 添加视频对话框 -->
    <el-dialog
      v-model="addVideoDialogVisible"
      title="管理合集稿件"
      width="700px"
      destroy-on-close
    >
      <div class="video-selection-header">
        <span class="selection-tip">勾选稿件添加到合集，取消勾选从合集移除</span>
      </div>
      
      <div v-if="addingVideo" class="dialog-loading">
        <el-skeleton :rows="5" animated />
      </div>
      
      <div v-else-if="availableVideos.length === 0" class="dialog-empty">
        <el-empty description="暂无可添加的稿件" />
      </div>
      
      <div v-else class="video-select-list">
        <el-checkbox-group v-model="selectedVideos">
          <div v-for="video in availableVideos" :key="video.id" class="video-select-item">
            <el-checkbox :label="video.id">
              <div class="video-select-content">
                <img :src="video.coverUrl || getDefaultCover()" class="video-select-cover" />
                <div class="video-select-info">
                  <div class="video-select-title">{{ video.title }}</div>
                  <div class="video-select-meta">{{ formatNumber(video.viewCount) }} 播放</div>
                </div>
              </div>
            </el-checkbox>
            <span v-if="selectedVideos.includes(video.id)" class="in-collection-badge">已加入</span>
          </div>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="addVideoDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddVideoToCollection" :loading="addingVideo">更新</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { useRouter, useRoute } from 'vue-router'
import { ref, reactive, watch, computed, onMounted, onUnmounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { creatorApi, manuscriptApi, collectionApi, followApi, statsApi } from '@/api/creator'
import { videoProcessApi } from '@/api/videoProcess'
import { useUserStore } from '@/stores/user'
import * as echarts from 'echarts'
import {
  VideoPlay,
  Upload,
  House,
  Document,
  DataAnalysis,
  UserFilled,
  ChatDotRound,
  Coin,
  Setting,
  Menu,
  Message,
  Comment,
  Monitor,
  UploadFilled,
  Picture,
  Plus,
  MoreFilled,
  InfoFilled,
  WarningFilled,
  CircleCheckFilled,
  ArrowDown,
  ArrowUp,
  StarFilled,
  ArrowRight,
  Medal,
  Search,
  Delete,
  Refresh,
  Timer,
  Download
} from '@element-plus/icons-vue'
import UploadView from './UploadView.vue'
import DataCenterView from './DataCenterView.vue'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

// 当前登录用户信息 - 从localStorage获取以确保有值
const currentUser = computed(() => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    try {
      return JSON.parse(userStr)
    } catch (e) {
      return userStore.userInfo
    }
  }
  return userStore.userInfo
})

// 获取当前用户ID
const getCurrentUserId = () => {
  const userStr = localStorage.getItem('user')
  if (userStr) {
    try {
      const user = JSON.parse(userStr)
      return user?.id
    } catch (e) {
      return userStore.userInfo.id
    }
  }
  return userStore.userInfo.id
}

// 首页统计数据
const homeLoading = ref(false)
const homeError = ref(null)
const statsData = ref({
  followerCount: 0,
  totalViewCount: 0,
  totalCommentCount: 0,
  totalDanmuCount: 0,
  totalLikeCount: 0,
  totalShareCount: 0,
  totalFavoriteCount: 0,
  totalCoinCount: 0
})

const loadHomeData = async () => {
  homeLoading.value = true
  homeError.value = null
  try {
    const [overviewRes, commentsRes, viewRankingRes, interactionRankingRes] = await Promise.all([
      statsApi.getOverview(),
      statsApi.getLatestComments(5),
      statsApi.getFansRanking('view', 5),
      statsApi.getFansRanking('interaction', 5)
    ])
    
    if (overviewRes.code === 200 && overviewRes.data) {
      statsData.value = {
        followerCount: overviewRes.data.totalFollowers || 0,
        totalViewCount: overviewRes.data.totalViews || 0,
        totalCommentCount: overviewRes.data.totalComments || 0,
        totalDanmuCount: overviewRes.data.totalDanmaku || 0,
        totalLikeCount: overviewRes.data.totalLikes || 0,
        totalShareCount: overviewRes.data.totalShares || 0,
        totalFavoriteCount: overviewRes.data.totalCollections || 0,
        totalCoinCount: overviewRes.data.totalCoins || 0
      }
    }
    
    if (commentsRes.code === 200 && commentsRes.data) {
      latestComments.value = commentsRes.data
    }
    
    if (viewRankingRes.code === 200 && viewRankingRes.data) {
      viewRanking.value = viewRankingRes.data
    }
    
    if (interactionRankingRes.code === 200 && interactionRankingRes.data) {
      interactionRanking.value = interactionRankingRes.data
    }
  } catch (error) {
    console.error('加载主页数据失败:', error)
    homeError.value = '加载数据失败，请稍后重试'
  } finally {
    homeLoading.value = false
  }
}

// 侧边栏菜单引用
const menuRef = ref(null)

// 返回主站
const goToMainSite = () => {
  router.push('/')
}

// 跳转到个人主页
const goToUserProfile = () => {
  if (currentUser.value && currentUser.value.id) {
    router.push(`/profile/${currentUser.value.id}/home`)
  }
}

// 当前激活的菜单索引
const currentActive = ref('home')

// 菜单选择事件处理
const handleMenuSelect = (index, indexPath) => {
  // 根据索引导航到对应的路由
  const routeMap = {
    'home': '/create-center/home',
    'upload': '/create-center/upload',
    'content': '/create-center/content',
    'content-articles': '/create-center/content-articles',

    'data': '/create-center/data',
    'fans': '/create-center/fans',
    'interaction': '/create-center/interaction',
    'interaction-comment': '/create-center/interaction-comment',
    'interaction-danmu': '/create-center/interaction-danmu',


  }
  
  if (routeMap[index]) {
    router.push(routeMap[index])
  }
  
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 粉丝管理相关状态
const fansLoading = ref(false)
const totalFans = ref(0)
const fansCurrentPage = ref(1)
const fansPageSize = ref(10)
const fansFilter = ref('all')
const fans = ref([])
const fansStats = ref({
  totalFans: 0,
  newFansToday: 0,
  newFansThisWeek: 0,
  newFansThisMonth: 0
})

// 粉丝初始化标志
let fansInitialized = false

// 粉丝统计数据获取函数
const fetchFansStats = async () => {
  try {
    const res = await creatorApi.getMyFollowers()
    console.log('粉丝统计 - API响应:', res)
    if (res.code === 200) {
      totalFans.value = Array.isArray(res.data) ? res.data.length : 0
    }
  } catch (error) {
    console.error('获取粉丝统计失败:', error)
  }
}

// 粉丝列表获取函数
const fetchFansList = async () => {
  fansLoading.value = true
  try {
    let fansRes
    if (fansFilter.value === 'mutual') {
      fansRes = await creatorApi.getMyFollowing()
    } else {
      fansRes = await creatorApi.getMyFollowers()
    }
    
    console.log('粉丝列表 - API响应:', fansRes)
    if (fansRes.code === 200) {
      fans.value = fansRes.data || []
      totalFans.value = Array.isArray(fansRes.data) ? fansRes.data.length : 0
      
      const followingRes = await creatorApi.getMyFollowing()
      if (followingRes.code === 200) {
        const followingIds = new Set((followingRes.data || []).map(u => u.id))
        fans.value = fans.value.map(fan => ({
          ...fan,
          isFollowing: followingIds.has(fan.id)
        }))
        console.log('更新后的粉丝列表:', fans.value)
      }
    }
  } catch (error) {
    console.error('获取粉丝列表失败:', error)
    ElMessage.error('获取粉丝列表失败')
  } finally {
    fansLoading.value = false
  }
}

// 粉丝管理初始化函数
const initFansData = () => {
  if (!fansInitialized) {
    fansInitialized = true
    fetchFansStats()
    fetchFansList()
  }
}

// 监听路由变化，同步当前激活的菜单
watch(
  () => route.path,
  (newPath) => {
    // 根据当前路径设置activeIndex
    const pathMap = {
      '/create-center/home': 'home',
      '/create-center/upload': 'upload',
      '/create-center/content': 'content',
      '/create-center/content-articles': 'content-articles',

      '/create-center/data': 'data',
      '/create-center/fans': 'fans',
      '/create-center/interaction': 'interaction',
      '/create-center/interaction-comment': 'interaction-comment',
      '/create-center/interaction-danmu': 'interaction-danmu',


    }
    
    if (pathMap[newPath]) {
      currentActive.value = pathMap[newPath]
      if (menuRef.value) {
        menuRef.value.activeIndex = pathMap[newPath]
      }

      if (newPath === '/create-center/fans') {
        initFansData()
      }
    }
  },
  { immediate: true }
)

// 评论/弹幕切换标签
const activeCommentTab = ref('comment')

// 监听评论/弹幕切换标签
watch(activeCommentTab, (newVal) => {
  if (newVal === 'comment') {
    fetchComments()
    fetchVideoList()
  } else if (newVal === 'danmu') {
    fetchDanmakuList()
  }
})

// 排行切换标签
const activeRankingTab = ref('view')



// 最新评论数据
const latestComments = ref([])

// 最新弹幕数据（开发中）
const latestDanmus = ref([])
const danmuDeveloping = ref(true)

// 观看排行数据
const viewRanking = ref([])

const interactionRanking = ref([])

// 获取排名样式类
const getRankingClass = (index) => {
  if (index === 0) {
    return 'ranking-gold'
  } else if (index === 1) {
    return 'ranking-silver'
  } else if (index === 2) {
    return 'ranking-bronze'
  }
  return ''
}

// 稿件管理相关状态
// 主要选择栏
const mainTab = ref('video')

// 状态筛选
const statusFilter = ref('processing') // processing: 进行中, published: 已通过, rejected: 未通过

// 稿件列表数据
const articles = ref([])
const articlesLoading = ref(false)
const totalArticles = ref(0)

// 稿件统计数据
const approvedCount = ref(0)
const rejectedCount = ref(0)
const processingCount = ref(0)

// 进度模拟相关状态
const articleProgress = ref({})  // 存储每个稿件的进度 { [id]: progress }
const articleProgressStatus = ref({})  // 存储每个稿件的进度状态 { [id]: 'running' | 'completed' }

// 分页相关状态
const currentPage = ref(1)
const pageSize = ref(10)

// 获取稿件列表
const fetchArticles = async () => {
  articlesLoading.value = true
  try {
    const params = {
      page: currentPage.value,
      size: pageSize.value
    }
    if (statusFilter.value) {
      params.status = statusFilter.value
    }
    
    console.log('获取稿件参数:', params)
    const response = await manuscriptApi.getMyManuscripts(params)
    console.log('获取稿件响应:', response)
    
    if (response.code === 200) {
      articles.value = response.data.list || []
      totalArticles.value = response.data.total || 0
      console.log('稿件列表:', articles.value)
      
      // 为进行中的稿件启动进度模拟
      startProgressSimulation()
    }
  } catch (error) {
    console.error('获取稿件列表失败:', error)
    ElMessage.error('获取稿件列表失败')
  } finally {
    articlesLoading.value = false
  }
}

// 启动进度模拟
const startProgressSimulation = () => {
  articles.value.forEach(article => {
    // 只为进行中状态的稿件启动进度模拟 (status: 0 或 1)
    if ((article.status === 0 || article.status === 1) && !videoProcessApi.getStatus(article.id)) {
      // 根据稿件类型选择任务类型
      const taskType = article.hasSubtitle ? 'transcode' : 'transcode'
      
      videoProcessApi.startProcess(article.id, taskType, {
        onProgress: (progress, info) => {
          articleProgress.value[article.id] = progress
          articleProgressStatus.value[article.id] = info.status
        },
        onComplete: () => {
          articleProgressStatus.value[article.id] = 'completed'
          // 刷新稿件列表
          fetchManuscriptStats()
        }
      })
    }
  })
}

// 停止所有进度模拟
const stopAllProgressSimulation = () => {
  articles.value.forEach(article => {
    if (videoProcessApi.getStatus(article.id)) {
      videoProcessApi.removeProcess(article.id)
    }
  })
}

// 获取稿件统计
const fetchManuscriptStats = async () => {
  try {
    const response = await manuscriptApi.getMyStats()
    
    if (response.code === 200) {
      const stats = response.data
      processingCount.value = stats.processing || 0
      approvedCount.value = stats.published || 0
      rejectedCount.value = stats.rejected || 0
    }
  } catch (error) {
    console.error('获取稿件统计失败:', error)
  }
}

// 监听筛选条件变化，重新获取数据
watch(statusFilter, () => {
  currentPage.value = 1
  fetchArticles()
})

// 监听分页变化
watch([currentPage, pageSize], () => {
  fetchArticles()
})

// 监听当前激活菜单变化，加载稿件数据
watch(currentActive, (newVal) => {
  if (newVal === 'content-articles') {
    fetchArticles()
    fetchManuscriptStats()
    if (mainTab.value === 'collection') {
      loadUserCollections()
    }
  }
  if (newVal === 'interaction-comment') {
    fetchComments()
    fetchVideoList()
  }
})

// 监听 mainTab 变化，加载合集数据
watch(mainTab, (newVal) => {
  if (newVal === 'collection') {
    loadUserCollections()
  }
})

// 获取稿件状态类型
const getArticleStatusType = (status) => {
  const statusTypeMap = {
    0: 'info',      // 待审核
    1: 'warning',   // 进行中
    3: 'success',   // 已发布
    4: 'danger',    // 已拒绝
    '-1': 'warning' // 已下架
  }
  return statusTypeMap[status] || 'info'
}

// 获取稿件状态文本
const getArticleStatusText = (status) => {
  const statusTextMap = {
    0: '进行中',   // 待审核
    1: '处理中',
    2: '待发布',
    3: '已通过',   // 已发布
    4: '未通过',   // 已拒绝
    '-1': '已下架' // 已下架
  }
  return statusTextMap[status] || '未知'
}

// 编辑稿件
const editArticle = (id) => {
  router.push(`/create-center/edit/${id}`)
}

// 删除稿件
const deleteArticle = async (id) => {
  try {
    await ElMessageBox.confirm('确定要删除这个稿件吗？删除后无法恢复。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const response = await manuscriptApi.deleteManuscript(id)
    
    if (response.code === 200) {
      ElMessage.success('删除成功')
      fetchArticles()
      fetchManuscriptStats()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除稿件失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 下架稿件
const unpublishArticle = async (id) => {
  try {
    await ElMessageBox.confirm('确定要下架这个稿件吗？下架后观众将无法观看。', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    console.log('下架稿件ID:', id)
    const response = await manuscriptApi.unpublishManuscript(id)
    console.log('下架响应:', response)
    
    if (response.code === 200) {
      ElMessage.success('下架成功')
      fetchArticles()
      fetchManuscriptStats()
    } else {
      ElMessage.error(response.message || '下架失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('下架稿件失败:', error)
      ElMessage.error('下架失败')
    }
  }
}

// 上架稿件
const publishArticle = async (id) => {
  try {
    const response = await manuscriptApi.publishManuscript(id)
    
    if (response.code === 200) {
      ElMessage.success('上架成功')
      fetchArticles()
      fetchManuscriptStats()
    }
  } catch (error) {
    console.error('上架稿件失败:', error)
    ElMessage.error('上架失败')
  }
}

// 分页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
}

// 当前页变化
const handleCurrentChange = (current) => {
  currentPage.value = current
}

// 数据中心相关状态
const activeDataTab = ref('overview') // 活跃的数据中心标签页
const timeRange = ref('7d') // 时间范围

// 数据概览相关状态
const dataOverviewLoading = ref(false)
const dataOverview = ref({
  totalViews: 0,
  totalLikes: 0,
  totalCoins: 0,
  totalCollections: 0,
  totalShares: 0,
  totalComments: 0,
  totalDanmaku: 0,
  totalFollowers: 0,
  totalManuscripts: 0,
  viewsIncrease: 0,
  likesIncrease: 0,
  commentsIncrease: 0,
  danmakuIncrease: 0,
  sharesIncrease: 0,
  collectionsIncrease: 0,
  coinsIncrease: 0,
  updateTime: ''
})

// 趋势数据
const trendData = ref({
  dates: [],
  views: [],
  likes: [],
  comments: []
})

// 稿件排行数据
const manuscriptRanking = ref([])
const rankingSortBy = ref('views')

// ECharts 图表实例
const playTrendChartRef = ref(null)
const playTrendChart = ref(null)
const fansTrendChartRef = ref(null)
const fansTrendChart = ref(null)
const manuscriptRankingChartRef = ref(null)
const manuscriptRankingChart = ref(null)

// 初始化核心数据趋势图表
const initPlayTrendChart = () => {
  if (!playTrendChartRef.value) return
  
  if (!trendData.value || !trendData.value.dates || trendData.value.dates.length === 0) return
  
  if (playTrendChart.value) {
    playTrendChart.value.dispose()
  }
  
  playTrendChart.value = echarts.init(playTrendChartRef.value)
  
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      }
    },
    legend: {
      data: ['播放量', '点赞数', '评论数'],
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '10%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: trendData.value.dates || [],
      axisLabel: {
        color: '#606266'
      }
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLabel: {
        color: '#606266'
      }
    },
    series: [
      {
        name: '播放量',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: {
          width: 3,
          color: '#00a1d6'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(0, 161, 214, 0.25)' },
            { offset: 1, color: 'rgba(0, 161, 214, 0.02)' }
          ])
        },
        itemStyle: {
          color: '#00a1d6'
        },
        data: trendData.value.views || []
      },
      {
        name: '点赞数',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: {
          width: 3,
          color: '#fb7299'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(251, 114, 153, 0.25)' },
            { offset: 1, color: 'rgba(251, 114, 153, 0.04)' }
          ])
        },
        itemStyle: {
          color: '#fb7299'
        },
        data: trendData.value.likes || []
      },
      {
        name: '评论数',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: 8,
        lineStyle: {
          width: 3,
          color: '#f3a832'
        },
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(243, 168, 50, 0.25)' },
            { offset: 1, color: 'rgba(243, 168, 50, 0.04)' }
          ])
        },
        itemStyle: {
          color: '#f3a832'
        },
        data: trendData.value.comments || []
      }
    ]
  }
  
  playTrendChart.value.setOption(option)
}

// 初始化粉丝增长图表
const initFansTrendChart = () => {
  const chartDom = fansTrendChartRef.value
  if (!chartDom) {
    console.log('粉丝图表 DOM 不存在')
    return
  }

  if (fansTrendChart.value) {
    fansTrendChart.value.dispose()
    fansTrendChart.value = null
  }

  try {
    fansTrendChart.value = echarts.init(chartDom)

    const followersData = trendData.value?.followers || []
    let dates = trendData.value?.dates || []

    console.log('粉丝图表数据:', { followersData, dates })

    let finalFollowers = []
    if (followersData && followersData.length > 0) {
      finalFollowers = followersData
    } else {
      finalFollowers = [10, 15, 8, 20, 12, 18, 25]
    }

    if (!dates || dates.length === 0) {
      const today = new Date()
      dates = []
      for (let i = 6; i >= 0; i--) {
        const d = new Date(today)
        d.setDate(d.getDate() - i)
        dates.push(`${d.getMonth() + 1}/${d.getDate()}`)
      }
    }

    console.log('最终粉丝数据:', finalFollowers, '日期:', dates)

    const option = {
      tooltip: { trigger: 'axis' },
      xAxis: { type: 'category', data: dates },
      yAxis: { type: 'value' },
      series: [{ name: '新增粉丝', type: 'bar', data: finalFollowers }]
    }

    fansTrendChart.value.setOption(option)
    console.log('粉丝图表初始化成功')
  } catch (e) {
    console.error('粉丝图表初始化失败:', e)
  }
}

// 下载图表
const downloadChart = (chartRef, filename) => {
  if (!chartRef.value) {
    ElMessage.warning('图表未加载')
    return
  }
  
  const url = chartRef.value.getDataURL({
    type: 'png',
    pixelRatio: 2,
    backgroundColor: '#fff'
  })
  
  const link = document.createElement('a')
  link.download = `${filename}_${new Date().toLocaleDateString()}.png`
  link.href = url
  link.click()
  
  ElMessage.success('图表下载成功')
}

// 初始化稿件表现排行图表
const initManuscriptRankingChart = () => {
  if (!manuscriptRankingChartRef.value) return
  
  if (manuscriptRanking.value && manuscriptRanking.value.length === 0) return
  
  if (manuscriptRankingChart.value) {
    manuscriptRankingChart.value.dispose()
  }
  
  manuscriptRankingChart.value = echarts.init(manuscriptRankingChartRef.value)

  const top5 = (manuscriptRanking.value || []).slice(0, 5)

  if (top5.length === 0) return

  const option = {
    xAxis: { type: 'category', data: top5.map(item => item.title || '未知稿件') },
    yAxis: { type: 'value' },
    series: [{ name: '播放量', type: 'bar', data: top5.map(item => item.viewCount || 0) }]
  }
  
  manuscriptRankingChart.value.setOption(option)
}

// 更新图表数据
const updateCharts = () => {
  nextTick(() => {
    if (playTrendChart.value) {
      playTrendChart.value.setOption({
        xAxis: { data: trendData.value.dates || [] },
        series: [
          { data: trendData.value.views || [] },
          { data: trendData.value.likes || [] },
          { data: trendData.value.comments || [] }
        ]
      })
    }
    if (fansTrendChart.value) {
      fansTrendChart.value.setOption({
        xAxis: { data: trendData.value.dates || [] },
        series: [{ data: trendData.value.followers || [] }]
      })
    }
    if (manuscriptRankingChart.value) {
      initManuscriptRankingChart()
    }
  })
}

// 窗口大小变化时重绘图表
const handleResize = () => {
  playTrendChart.value?.resize()
  fansTrendChart.value?.resize()
  manuscriptRankingChart.value?.resize()
}

// 加载数据概览
const loadDataOverview = async () => {
  dataOverviewLoading.value = true
  try {
    const res = await statsApi.getOverview()
    if (res.code === 200 && res.data) {
      dataOverview.value = res.data
    }
  } catch (error) {
    console.error('加载数据概览失败:', error)
  } finally {
    dataOverviewLoading.value = false
  }
}

// 加载趋势数据
const loadTrendData = async (days = 7) => {
  console.log('开始加载趋势数据, days:', days)
  try {
    const res = await statsApi.getTrend(days)
    console.log('趋势数据响应:', res)
    if (res.code === 200 && res.data) {
      trendData.value = res.data
      console.log('trendData 已更新:', trendData.value)
    }
  } catch (error) {
    console.error('加载趋势数据失败:', error)
  }
}

// 加载稿件排行
const loadManuscriptRanking = async (sortBy = 'views', limit = 10) => {
  try {
    const res = await statsApi.getRanking(sortBy, limit)
    if (res.code === 200 && res.data) {
      manuscriptRanking.value = res.data.list || []
    }
  } catch (error) {
    console.error('加载稿件排行失败:', error)
  }
}

// 监听时间范围变化
watch(timeRange, (newVal) => {
  const daysMap = {
    '7d': 7,
    '30d': 30,
    '90d': 90,
    '1y': 365
  }
  const days = daysMap[newVal] || 7
  loadTrendData(days)
})

// 监听数据中心标签页变化
watch(activeDataTab, async (newVal) => {
  if (newVal === 'overview') {
    await Promise.all([
      loadDataOverview(),
      loadTrendData(),
      loadManuscriptRanking('views', 5)
    ])
    nextTick(() => {
      initPlayTrendChart()
      initManuscriptRankingChart()
    })
  } else if (newVal === 'article') {
    await loadManuscriptRanking(rankingSortBy.value)
  } else if (newVal === 'fans') {
    await loadTrendData()
    setTimeout(() => {
      console.log('延迟初始化粉丝图表 - watch')
      initFansTrendChart()
    }, 100)
  }
  updateCharts()
})

// 监听趋势数据变化，更新图表
watch(trendData, () => {
  updateCharts()
  if (activeDataTab.value === 'fans') {
    setTimeout(() => {
      initFansTrendChart()
    }, 100)
  }
}, { deep: true })

// 监听稿件排行变化，更新排行图表
watch(manuscriptRanking, (newVal) => {
  if (newVal && newVal.length > 0 && activeDataTab.value === 'overview') {
    nextTick(() => {
      initManuscriptRankingChart()
    })
  }
}, { deep: true })

// 监听窗口大小变化
onMounted(async () => {
  console.log('onMounted 执行, activeDataTab:', activeDataTab.value)
  window.addEventListener('resize', handleResize)

  if (activeDataTab.value === 'overview') {
    console.log('加载 overview 数据')
    await Promise.all([
      loadDataOverview(),
      loadTrendData(),
      loadManuscriptRanking('views', 5)
    ])
    nextTick(() => {
      initPlayTrendChart()
      initManuscriptRankingChart()
    })
  } else if (activeDataTab.value === 'fans') {
    console.log('加载 fans 数据')
    await loadTrendData()
    setTimeout(() => {
      console.log('延迟初始化粉丝图表 - onMounted')
      initFansTrendChart()
    }, 100)
  }
})

// 稿件分析相关状态
const articleDataTab = ref('data-overview') // 稿件数据标签页
const playTrendTimeRange = ref('30d') // 播放趋势图时间范围

// 粉丝分析相关状态
const fansTimeRange = ref('7d') // 粉丝数据时间范围

// 监听粉丝分析时间范围变化
watch(fansTimeRange, async (newVal) => {
  if (activeDataTab.value !== 'fans') return

  const daysMap = {
    '7d': 7,
    '30d': 30,
    '90d': 90,
    '1y': 365
  }
  const days = daysMap[newVal] || 7
  await loadTrendData(days)
  nextTick(() => {
    initFansTrendChart()
  })
}, { immediate: false })

// 粉丝排行数据
const playTimeRanking = ref([
  { id: 1, username: 'bili_1220763...', avatar: 'https://picsum.photos/id/1001/40/40' },
  { id: 2, username: '不顶不是人', avatar: 'https://picsum.photos/id/1002/40/40' },
  { id: 3, username: '乘風', avatar: 'https://picsum.photos/id/1003/40/40' }
])

const videoInteractionRanking = ref([
  { id: 1, username: '不顶不是人', avatar: 'https://picsum.photos/id/1002/40/40' },
  { id: 2, username: 'bili_1220763...', avatar: 'https://picsum.photos/id/1001/40/40' },
  { id: 3, username: '61-boy', avatar: 'https://picsum.photos/id/1004/40/40' }
])

const dynamicInteractionRanking = ref([
  { id: 1, username: '伶琬莹·可爱', avatar: 'https://picsum.photos/id/1005/40/40' }
])

const handleFollowFan = async (fan) => {
  try {
    if (fan.isFollowing) {
      await followApi.unfollow(fan.id)
      fan.isFollowing = false
      ElMessage.success('已取消关注')
    } else {
      await followApi.follow(fan.id)
      fan.isFollowing = true
      ElMessage.success('关注成功')
    }
  } catch (error) {
    console.error('关注操作失败:', error)
    ElMessage.error('操作失败，请重试')
  }
}

watch(fansFilter, () => {
  fansCurrentPage.value = 1
})

watch([fansCurrentPage, fansPageSize, fansFilter], () => {
  if (fansInitialized) {
    fetchFansList()
  }
})

// 合集管理相关状态
const myCollections = ref({
  viewType: 'horizontal',
  items: [],
  loading: false
})

const collectionDetail = ref({
  visible: false,
  collectionId: null,
  collection: null,
  manuscripts: [],
  loading: false,
  sortBy: 'default',
  pagination: {
    page: 1,
    size: 20,
    total: 0
  }
})

const createCollectionDialogVisible = ref(false)
const createCollectionForm = ref({
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const creatingCollection = ref(false)

const editCollectionDialogVisible = ref(false)
const editCollectionForm = ref({
  id: null,
  name: '',
  description: '',
  cover: null,
  coverUrl: '',
  isPublic: true
})
const updatingCollection = ref(false)

const addVideoDialogVisible = ref(false)
const addVideoCollectionId = ref(null)
const availableVideos = ref([])
const selectedVideos = ref([])
const addingVideo = ref(false)

const getDefaultCover = () => {
  return 'https://picsum.photos/id/1025/400/225'
}

const formatDate = (dateStr) => {
  if (!dateStr) return ''
  const date = new Date(dateStr)
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

const formatNumber = (num) => {
  if (!num) return '0'
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + '万'
  }
  return num.toString()
}

const loadUserCollections = async () => {
  console.log('【调试】loadUserCollections 被调用')
  const userId = getCurrentUserId()
  console.log('【调试】currentUser.value:', currentUser.value)
  console.log('【调试】getCurrentUserId():', userId)
  if (!userId) {
    console.log('【调试】用户ID为空，直接返回')
    return
  }

  myCollections.value.loading = true
  try {
    console.log('【调试】开始获取合集列表，用户ID:', userId)
    const response = await collectionApi.getUserCollections(userId, 1, 100)
    console.log('【调试】合集列表响应:', response)
    if (response.code === 200) {
      const list = response.data || []
      for (const collection of list) {
        try {
          const videoResponse = await collectionApi.getCollectionManuscripts(collection.id, 1, 10)
          if (videoResponse.code === 200) {
            collection.videos = (videoResponse.data || []).map(video => ({
              ...video,
              date: formatDate(video.uploadTime)
            }))
          }
        } catch (e) {
          console.error('获取合集稿件失败:', e)
          collection.videos = []
        }
      }
      myCollections.value.items = list
    }
  } catch (error) {
    console.error('获取合集列表失败:', error)
  } finally {
    myCollections.value.loading = false
  }
}

const goToCollectionDetail = async (collectionId) => {
  collectionDetail.value.visible = true
  collectionDetail.value.collectionId = collectionId
  collectionDetail.value.loading = true
  
  try {
    const response = await collectionApi.getCollectionById(collectionId)
    if (response.code === 200) {
      collectionDetail.value.collection = response.data
    }
    
    const videoResponse = await collectionApi.getCollectionManuscripts(collectionId, 1, 20)
    if (videoResponse.code === 200) {
      collectionDetail.value.manuscripts = videoResponse.data || []
      collectionDetail.value.pagination.total = videoResponse.data?.length || 0
    }
  } catch (error) {
    console.error('获取合集详情失败:', error)
  } finally {
    collectionDetail.value.loading = false
  }
}

const backToCollectionsList = () => {
  collectionDetail.value.visible = false
  collectionDetail.value.collectionId = null
  collectionDetail.value.collection = null
  collectionDetail.value.manuscripts = []
}

const openCreateCollectionDialog = () => {
  createCollectionForm.value = {
    name: '',
    description: '',
    cover: null,
    coverUrl: '',
    isPublic: true
  }
  createCollectionDialogVisible.value = true
}

const handleCreateCollectionCoverChange = (file) => {
  const isImage = file.raw.type.startsWith('image/')
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('封面图片只能是图片格式!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('封面图片大小不能超过 2MB!')
    return false
  }

  createCollectionForm.value.cover = file.raw
  createCollectionForm.value.coverUrl = URL.createObjectURL(file.raw)
  return false
}

const handleCreateCollection = async () => {
  if (!createCollectionForm.value.name.trim()) {
    ElMessage.warning('请输入合集名称')
    return
  }

  creatingCollection.value = true
  try {
    const response = await collectionApi.createCollection({
      name: createCollectionForm.value.name,
      description: createCollectionForm.value.description,
      cover: createCollectionForm.value.cover,
      isPublic: createCollectionForm.value.isPublic
    })
    if (response.code === 200) {
      ElMessage.success('创建成功')
      createCollectionDialogVisible.value = false
      await loadUserCollections()
    } else {
      ElMessage.error(response.message || '创建失败')
    }
  } catch (error) {
    console.error('创建合集失败:', error)
    ElMessage.error('创建合集失败')
  } finally {
    creatingCollection.value = false
  }
}

const openEditCollectionDialog = () => {
  if (!collectionDetail.value.collection) return
  
  editCollectionForm.value = {
    id: collectionDetail.value.collectionId,
    name: collectionDetail.value.collection.title || '',
    description: collectionDetail.value.collection.description || '',
    cover: null,
    coverUrl: collectionDetail.value.collection.coverUrl || '',
    isPublic: collectionDetail.value.collection.status === 1
  }
  editCollectionDialogVisible.value = true
}

const handleEditCollectionCoverChange = (file) => {
  const isImage = file.raw.type.startsWith('image/')
  const isLt2M = file.raw.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('封面图片只能是图片格式!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('封面图片大小不能超过 2MB!')
    return false
  }

  editCollectionForm.value.cover = file.raw
  editCollectionForm.value.coverUrl = URL.createObjectURL(file.raw)
  return false
}

const handleUpdateCollection = async () => {
  if (!editCollectionForm.value.name.trim()) {
    ElMessage.warning('请输入合集名称')
    return
  }

  updatingCollection.value = true
  try {
    const response = await collectionApi.updateCollection(editCollectionForm.value.id, {
      name: editCollectionForm.value.name,
      description: editCollectionForm.value.description,
      cover: editCollectionForm.value.cover,
      isPublic: editCollectionForm.value.isPublic
    })
    if (response.code === 200) {
      ElMessage.success('更新成功')
      editCollectionDialogVisible.value = false
      await goToCollectionDetail(editCollectionForm.value.id)
      await loadUserCollections()
    } else {
      ElMessage.error(response.message || '更新失败')
    }
  } catch (error) {
    console.error('更新合集失败:', error)
    ElMessage.error('更新合集失败')
  } finally {
    updatingCollection.value = false
  }
}

const handleCollectionCommand = async (command) => {
  if (command === 'delete') {
    try {
      await ElMessageBox.confirm(
        '确定要删除这个合集吗？删除后无法恢复。',
        '删除确认',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
      
      const response = await collectionApi.deleteCollection(collectionDetail.value.collectionId)
      if (response.code === 200) {
        ElMessage.success('删除成功')
        backToCollectionsList()
        await loadUserCollections()
      } else {
        ElMessage.error(response.message || '删除失败')
      }
    } catch (error) {
      if (error !== 'cancel') {
        console.error('删除合集失败:', error)
        ElMessage.error('删除合集失败')
      }
    }
  }
}

const handleCollectionDetailSortChange = (sort) => {
  collectionDetail.value.sortBy = sort
}

const openAddVideoToCollectionDialog = async () => {
  await openAddVideoDialog(collectionDetail.value.collectionId)
}

const openAddVideoDialog = async (collectionId) => {
  addVideoCollectionId.value = collectionId
  selectedVideos.value = []
  addingVideo.value = true
  
  try {
    console.log('开始获取数据，合集ID:', collectionId)
    
    const [manuscriptsRes, videosRes] = await Promise.all([
      collectionApi.getCollectionManuscripts(collectionId, 1, 100),
      manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    ])
    
    console.log('合集稿件响应:', manuscriptsRes)
    console.log('用户稿件响应:', videosRes)
    
    if (videosRes.code === 200) {
      const data = videosRes.data || {}
      console.log('data对象的keys:', Object.keys(data))
      console.log('data对象的完整结构:', data)
      
      const list = data.list || data.records || data.items || data.data || []
      console.log('处理前的稿件列表:', list)
      
      availableVideos.value = list.map(v => ({
        ...v,
        id: v.manuscriptId || v.id,
        coverUrl: v.coverUrl || v.cover_url,
        viewCount: v.viewCount || v.view_count || 0,
        title: v.title || '无标题'
      }))
      
      console.log('处理后的稿件列表:', availableVideos.value)
    } else {
      console.error('获取用户稿件失败:', videosRes.message)
      availableVideos.value = []
    }
    
    if (manuscriptsRes.code === 200) {
      const manuscripts = manuscriptsRes.data || []
      selectedVideos.value = manuscripts.map(m => m.manuscriptId || m.id)
      console.log('当前合集中的稿件IDs:', selectedVideos.value)
    }
  } catch (error) {
    console.error('获取数据失败:', error)
    availableVideos.value = []
  } finally {
    addingVideo.value = false
  }
  
  addVideoDialogVisible.value = true
}

const handleAddVideoToCollection = async () => {
  if (selectedVideos.value.length === 0) {
    ElMessage.warning('请选择要添加的视频')
    return
  }

  addingVideo.value = true
  try {
    for (const videoId of selectedVideos.value) {
      await collectionApi.addManuscriptToCollection(addVideoCollectionId.value, videoId, 0)
    }
    ElMessage.success('添加成功')
    addVideoDialogVisible.value = false
    
    if (collectionDetail.value.visible) {
      await goToCollectionDetail(collectionDetail.value.collectionId)
    }
    await loadUserCollections()
  } catch (error) {
    console.error('添加视频失败:', error)
    ElMessage.error('添加视频失败')
  } finally {
    addingVideo.value = false
  }
}

const handleVideoCommand = async (command, manuscript) => {
  if (command === 'remove') {
    try {
      await ElMessageBox.confirm(
        '确定要从合集中移除这个视频吗？',
        '移除确认',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
      
      const response = await collectionApi.removeManuscriptFromCollection(
        collectionDetail.value.collectionId,
        manuscript.id
      )
      if (response.code === 200) {
        ElMessage.success('移除成功')
        await goToCollectionDetail(collectionDetail.value.collectionId)
        await loadUserCollections()
      } else {
        ElMessage.error(response.message || '移除失败')
      }
    } catch (error) {
      if (error !== 'cancel') {
        console.error('移除视频失败:', error)
        ElMessage.error('移除视频失败')
      }
    }
  }
}

const playManuscript = (manuscript) => {
  router.push(`/manuscript/${manuscript.id}`)
}

const playCollectionAll = () => {
  if (collectionDetail.value.manuscripts.length > 0) {
    router.push(`/manuscript/${collectionDetail.value.manuscripts[0].id}`)
  }
}

const playCollectionAllFromList = (collection) => {
  if (collection.videos && collection.videos.length > 0) {
    router.push(`/manuscript/${collection.videos[0].id}`)
  }
}

// 回到创作中心首页
const goToCreateCenterHome = () => {
  router.push('/create-center/home')
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 跳转到投稿页面（修改为在创作中心内部显示）
const goToUpload = () => {
  router.push('/create-center/upload')
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 投稿相关变量和函数
// 简化上传表单，只保留视频文件
const uploadForm = reactive({
  videoFile: null
})

// 表单验证规则，只验证视频文件
const uploadRules = {
  videoFile: [
    { required: true, message: '请上传视频文件', trigger: 'change' }
  ]
}

// 表单引用
const uploadFormRef = ref()

// 上传组件引用
const uploadRef = ref(null)

// 上传进度
const uploadProgress = ref(0)
const isUploading = ref(false)

// 处理视频上传
const handleVideoUpload = (file) => {
  uploadForm.videoFile = file.raw
  return false // 阻止自动上传
}

// 上传视频按钮点击事件
const handleUploadClick = () => {
  if (!uploadForm.videoFile) {
    // 没有选择文件，触发文件选择对话框
    if (uploadRef.value) {
      uploadRef.value.open()
    }
  } else {
    // 已经选择了文件，执行上传操作
    handleSubmit()
  }
}

// 提交上传
const handleSubmit = () => {
  uploadFormRef.value.validate((valid) => {
    if (valid) {
      // 模拟上传进度
      isUploading.value = true
      uploadProgress.value = 0
      
      // 模拟上传速度变化
      const speeds = ['1.2 MB/s', '1.5 MB/s', '1.8 MB/s', '2.0 MB/s', '1.7 MB/s']
      let speedIndex = 0
      
      const interval = setInterval(() => {
        uploadProgress.value += 10
        uploadSpeed.value = speeds[speedIndex]
        speedIndex = (speedIndex + 1) % speeds.length
        
        if (uploadProgress.value >= 100) {
          clearInterval(interval)
          setTimeout(() => {
            isUploading.value = false
            console.log('视频上传完成:', uploadForm)
            // 上传成功后进入发布页面
            uploadStatus.value = UploadStatus.PUBLISHING
          }, 500)
        }
      }, 300)
    } else {
      return false
    }
  })
}

// 上传状态枚举
const UploadStatus = {
  UPLOADING: 'uploading',
  PUBLISHING: 'publishing',
  COMPLETED: 'completed'
}

// 上传状态
const uploadStatus = ref(UploadStatus.UPLOADING)

// 上传速度
const uploadSpeed = ref('0 KB/s')

// 发布视频表单
const publishForm = reactive({
  coverFile: null,
  coverPreview: '',
  title: '',
  videoType: 'original', // original: 自制, repost: 转载
  category: '',
  tags: [],
  tagInput: '',
  description: '',
  collection: ''
})

// 分类列表
const categories = ref([
  { value: 1, label: '动画' },
  { value: 2, label: '音乐' },
  { value: 3, label: '舞蹈' },
  { value: 4, label: '游戏' },
  { value: 5, label: '知识' },
  { value: 6, label: '资讯' },
  { value: 7, label: '美食' },
  { value: 8, label: '生活' },
  { value: 9, label: '鬼畜' },
  { value: 10, label: '时尚' },
  { value: 11, label: '娱乐' },
  { value: 12, label: '影视' }
])




// 合集列表
const collections = ref([
  { value: '', label: '不加入合集' },
  { value: 1, label: '我的合集1' },
  { value: 2, label: '我的合集2' },
  { value: 3, label: '我的合集3' }
])

// 处理封面上传
const handleCoverUpload = (file) => {
  publishForm.coverFile = file.raw
  // 生成封面预览
  const reader = new FileReader()
  reader.onload = (e) => {
    publishForm.coverPreview = e.target.result
  }
  reader.readAsDataURL(file.raw)
  return false // 阻止自动上传
}

// 添加标签
const addTag = () => {
  if (publishForm.tagInput.trim() && !publishForm.tags.includes(publishForm.tagInput.trim())) {
    publishForm.tags.push(publishForm.tagInput.trim())
    publishForm.tagInput = ''
  }
}

// 删除标签
const removeTag = (tag) => {
  const index = publishForm.tags.indexOf(tag)
  if (index > -1) {
    publishForm.tags.splice(index, 1)
  }
}

// 处理标签输入回车事件
const handleTagKeydown = (e) => {
  if (e.key === 'Enter') {
    e.preventDefault()
    addTag()
  }
}

// 存草稿
const saveDraft = () => {
  console.log('保存草稿:', publishForm)
  ElMessage.success('草稿保存成功')
}

// 立即投稿
const submitPublish = () => {
  console.log('立即投稿:', publishForm)
  // 模拟投稿进度
  isUploading.value = true
  uploadProgress.value = 0
  
  const interval = setInterval(() => {
    uploadProgress.value += 10
    if (uploadProgress.value >= 100) {
      clearInterval(interval)
      setTimeout(() => {
        isUploading.value = false
        uploadStatus.value = UploadStatus.COMPLETED
        ElMessage.success('投稿成功！')
      }, 500)
    }
  }, 300)
}

// 添加分P
const addPart = () => {
  console.log('添加分P')
  ElMessage.info('添加分P功能待实现')
}

// 返回上传页面
const backToUpload = () => {
  // 重置上传状态
  uploadStatus.value = UploadStatus.UPLOADING
  uploadForm.videoFile = null
  uploadProgress.value = 0
  uploadSpeed.value = '0 KB/s'
  // 导航到上传页面
  router.push('/create-center/upload')
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 取消上传
const cancelUpload = () => {
  // 回到首页
  router.push('/create-center/home')
  // 滚动到顶部
  window.scrollTo(0, 0)
}

// 评论管理相关数据
const activeCommentMainTab = ref('visible')
const activeCommentSubTab = ref('video')
const commentTypeFilter = ref('all')
const videoFilter = ref('all')
const commentSearchText = ref('')
const commentSortBy = ref('latest')
const commentCurrentPage = ref(1)
const commentPageSize = ref(10)

const comments = ref([])
const commentLoading = ref(false)
const totalComments = ref(0)
const totalPages = ref(0)
const videoList = ref([])
const replyDialogVisible = ref(false)
const replyCommentId = ref(null)
const replyContent = ref('')
const replyToUserId = ref(null)

const danmakuList = ref([])
const danmakuLoading = ref(false)
const danmakuTotal = ref(0)
const danmakuCurrentPage = ref(1)
const danmakuPageSize = ref(10)

const fetchDanmakuList = async () => {
  danmakuLoading.value = true
  try {
    const res = await creatorApi.getDanmakuList({
      page: danmakuCurrentPage.value,
      size: danmakuPageSize.value
    })
    console.log('弹幕API返回:', res)
    if (res.code === 200 && res.data) {
      danmakuList.value = res.data.list || []
      danmakuTotal.value = res.data.total || 0
      console.log('弹幕列表:', danmakuList.value, '总数:', danmakuTotal.value)
    }
  } catch (error) {
    console.error('获取弹幕列表失败:', error)
  } finally {
    danmakuLoading.value = false
  }
}

const deleteDanmaku = async (danmakuId) => {
  try {
    await ElMessageBox.confirm('确定要删除这条弹幕吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await creatorApi.deleteDanmaku(danmakuId)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      fetchDanmakuList()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除弹幕失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const handleDeleteComment = async (commentId) => {
  try {
    await ElMessageBox.confirm('确定要删除这条评论吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    const res = await creatorApi.deleteComment(commentId)
    if (res.code === 200) {
      ElMessage.success('删除成功')
      loadHomeData()
    } else {
      ElMessage.error(res.message || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除评论失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const getDanmakuModeText = (mode) => {
  const modeMap = {
    1: '滚动',
    4: '底部',
    5: '顶部',
    6: '逆向',
    7: '定位'
  }
  return modeMap[mode] || '滚动'
}

const getDanmakuModeType = (mode) => {
  const typeMap = {
    1: 'primary',
    4: 'warning',
    5: 'success',
    6: 'danger',
    7: 'info'
  }
  return typeMap[mode] || 'primary'
}

const formatTime = (seconds) => {
  if (!seconds && seconds !== 0) return '00:00'
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const goToVideo = (manuscriptId, time, videoOrder) => {
  if (!manuscriptId) return
  const pParam = videoOrder ? `&p=${videoOrder}` : '&p=1'
  const timeParam = time ? `&t=${Math.floor(time)}` : ''
  window.open(`/manuscript/${manuscriptId}?${pParam.substring(1)}${timeParam}`, '_blank')
}

const fetchComments = async () => {
  commentLoading.value = true
  try {
    const params = {
      page: commentCurrentPage.value,
      size: commentPageSize.value,
      manuscriptId: videoFilter.value === 'all' ? undefined : videoFilter.value
    }
    const res = await creatorApi.getComments(params)
    if (res.code === 200 && res.data) {
      comments.value = res.data.list.map(item => ({
        id: item.id,
        selected: false,
        username: item.userName || '未知用户',
        avatar: item.userAvatar || '',
        content: item.content,
        time: item.createTime,
        videoThumbnail: item.manuscriptCover || '',
        videoTitle: item.manuscriptTitle || '',
        likeCount: item.likeCount || 0,
        replyCount: item.replyCount || 0,
        liked: item.liked || false,
        userId: item.userId,
        manuscriptId: item.manuscriptId
      }))
      totalComments.value = res.data.total || 0
      totalPages.value = Math.ceil((res.data.total || 0) / commentPageSize.value)
    }
  } catch (error) {
    console.error('获取评论列表失败:', error)
  } finally {
    commentLoading.value = false
  }
}

const fetchVideoList = async () => {
  try {
    const res = await manuscriptApi.getMyManuscripts({ page: 1, size: 100 })
    if (res.code === 200 && res.data) {
      videoList.value = res.data.list || []
    }
  } catch (error) {
    console.error('获取视频列表失败:', error)
  }
}

const openReplyDialog = (comment) => {
  replyCommentId.value = comment.id
  replyToUserId.value = comment.userId
  replyContent.value = ''
  replyDialogVisible.value = true
}

const handleReplyComment = async () => {
  if (!replyContent.value.trim()) {
    ElMessage.warning('请输入回复内容')
    return
  }
  try {
    const res = await creatorApi.replyComment(replyCommentId.value, replyContent.value, replyToUserId.value)
    if (res.code === 200) {
      ElMessage.success('回复成功')
      replyDialogVisible.value = false
      fetchComments()
    }
  } catch (error) {
    console.error('回复评论失败:', error)
  }
}

const handleSelectAll = (select) => {
  comments.value.forEach(comment => {
    comment.selected = select
  })
}

const handleBatchDelete = async () => {
  const selectedComments = comments.value.filter(c => c.selected)
  if (selectedComments.length === 0) {
    ElMessage.warning('请选择要删除的评论')
    return
  }
  try {
    await Promise.all(selectedComments.map(c => creatorApi.deleteComment(c.id)))
    ElMessage.success('批量删除成功')
    fetchComments()
  } catch (error) {
    console.error('批量删除失败:', error)
  }
}

watch([commentCurrentPage, commentTypeFilter, videoFilter, commentSortBy], () => {
  fetchComments()
})

watch(commentSearchText, () => {
  if (commentCurrentPage.value !== 1) {
    commentCurrentPage.value = 1
  } else {
    fetchComments()
  }
})

// 搜索评论
const searchComments = () => {
  fetchComments()
}

onMounted(() => {
  if (currentActive.value === 'home') {
    loadHomeData()
  }
  if (currentActive.value === 'interaction-comment') {
    fetchComments()
    fetchVideoList()
  }
  if (currentActive.value === 'content-articles') {
    fetchArticles()
    fetchManuscriptStats()
  }
  if (currentActive.value === 'fans') {
    initFansData()
  }
  if (currentActive.value === 'data') {
    loadDataOverview()
    loadTrendData()
  }
})

// 组件卸载时清理进度模拟器
onUnmounted(() => {
  stopAllProgressSimulation()
  videoProcessApi.clear()
})

watch(currentActive, (newVal) => {
  if (newVal === 'home') {
    loadHomeData()
  }
  if (newVal === 'fans') {
    initFansData()
  }
  if (newVal === 'data') {
    loadDataOverview()
    loadTrendData()
  }
  if (newVal === 'interaction-comment') {
    fetchComments()
    fetchVideoList()
  }
})

// 可见页码列表
const visiblePages = computed(() => {
  const pages = []
  const current = commentCurrentPage.value
  const total = totalPages.value
  
  // 总是显示第一页
  pages.push(1)
  
  // 如果当前页码大于3，显示省略号
  if (current > 3) {
    pages.push('...')
  }
  
  // 显示当前页码附近的页码
  const start = Math.max(2, current - 1)
  const end = Math.min(total - 1, current + 1)
  
  for (let i = start; i <= end; i++) {
    pages.push(i)
  }
  
  // 如果当前页码小于总页码-2，显示省略号
  if (current < total - 2) {
    pages.push('...')
  }
  
  // 如果总页码大于1，显示最后一页
  if (total > 1) {
    pages.push(total)
  }
  
  return pages
})

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
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\taskhost.exe'
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
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\taskhost.exe'
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
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\taskhost.exe'
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
    videoTitle: '在byrut下载游戏的可能会中挖矿病毒audiog.exe\taskhost.exe'
  }
])

// 搜索弹幕
const searchDanmu = () => {
  // 实际项目中这里会调用API进行搜索
  console.log('搜索弹幕:', danmuSearchText.value)
}
</script>

<style scoped>
.create-center-container {
  width: 100%;
  height: 100vh;
  background-color: #f5f7fa;
  display: flex;
  flex-direction: column;
}

/* 顶部导航栏样式 */
.create-center-header {
  height: 60px;
  background-color: #fff;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.center-title-btn {
  font-size: 20px;
  font-weight: 600;
  color: #1890ff;
  background-color: transparent;
  border: none;
  padding: 0;
  margin: 0;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  height: 40px;
  line-height: 40px;
}

.center-title-btn:hover {
  background-color: rgba(24, 144, 255, 0.1);
  color: #409eff;
}

.main-site-btn {
  display: flex;
  align-items: center;
  gap: 5px;
  background-color: #fff;
  border: none;
  color: #606266;
}

.main-site-btn:hover {
  background-color: #f0f2f5;
  color: #1890ff;
  border: none;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 15px;
}

.user-avatar {
  cursor: pointer;
}

.up-day-box {
  background-color: #f0f9ff;
  border: 1px solid #91d5ff;
  border-radius: 4px;
  padding: 6px 12px;
  font-size: 14px;
  color: #1890ff;
  display: flex;
  align-items: center;
}

/* 主体内容区域样式 */
.create-center-main {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* 侧边栏样式 */
.sidebar {
  width: 200px;
  background-color: #fff;
  border-right: 1px solid #e0e0e0;
  padding: 20px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.upload-btn-large {
  width: 100%;
  height: 40px;
  font-size: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  background-color: #fb7299;
  border: none;
  color: #fff;
}

.upload-btn-large:hover {
  background-color: #f75982;
  color: #fff;
}

.sidebar-menu {
  border-right: none;
}

.sidebar-menu .el-menu-item {
  height: 48px;
  line-height: 48px;
  font-size: 15px;
}

.sidebar-menu .el-menu-item.is-active {
  color: #1890ff;
  background-color: #ecf5ff;
}

/* 主内容区域样式 */
.content-area {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background-color: #f5f7fa;
}

.content-header {
  margin-bottom: 20px;
}

.content-header h2 {
  font-size: 22px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.content-body {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

.loading-container {
  min-height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.developing-tip {
  padding: 40px 20px;
  text-align: center;
  color: #909399;
}

/* 仪表盘统计卡片样式 */
.dashboard-stats {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  margin-top: 20px;
}

.stat-card {
  background-color: #fafafa;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
}

.stat-number {
  font-size: 28px;
  font-weight: 600;
  color: #1890ff;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  color: #606266;
}

/* 评论/弹幕选择栏样式 */
.comment-danmu-section {
  margin-top: 30px;
  padding: 0;
  background-color: transparent;
  box-shadow: none;
}

/* 排行选择栏样式 */
.ranking-section {
  margin-top: 30px;
  background-color: transparent;
  padding: 0;
  border-radius: 0;
  box-shadow: none;
}

.section-header {
  display: flex;
  justify-content: flex-start;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e0e0e0;
}

.section-header h3 {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.tab-buttons {
  display: flex;
  gap: 10px;
}

/* 互动列表样式 - 垂直布局 */
.interaction-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.interaction-item {
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
  position: relative;
}

.interaction-item:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.interaction-actions {
  position: absolute;
  right: 15px;
  top: 50%;
  transform: translateY(-50%);
}

/* 互动列表样式 - 水平布局 */
.interaction-list-horizontal {
  display: flex;
  gap: 20px;
  margin-top: 10px;
}

.interaction-item-horizontal {
  flex: 1;
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
  min-height: 120px;
  display: flex;
  flex-direction: column;
}

.interaction-item-horizontal:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 用户信息样式 */
.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.username {
  font-weight: 500;
  color: #303133;
}

/* 内容样式 */
.interaction-content {
  color: #606266;
  margin-bottom: auto;
  line-height: 1.5;
  flex: 1;
}

/* 时间样式 */
.interaction-time {
  font-size: 12px;
  color: #909399;
  text-align: right;
  margin-top: 8px;
}

/* 评论弹幕按钮样式 */
.tab-buttons .el-button {
  border: none;
  box-shadow: none;
}

.tab-buttons .el-button--primary:not(.is-plain) {
  background-color: #1890ff;
  border-color: transparent;
}

.tab-buttons .el-button--primary.is-plain {
  background-color: transparent;
  border-color: transparent;
  color: #1890ff;
}

/* 排行列表样式 */
.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 15px;
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.ranking-item:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.ranking-index {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background-color: #1890ff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
}

.ranking-count {
  margin-left: auto;
  font-weight: 600;
  color: #f77825;
}

/* 排行列表样式 - 水平布局 */
.ranking-list-horizontal {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 15px;
  margin-top: 10px;
}

.ranking-item-horizontal {
  padding: 15px;
  background-color: #fafafa;
  border-radius: 6px;
  transition: all 0.3s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  min-height: 120px;
}

.ranking-item-horizontal:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 水平排行项中的排名索引 */
.ranking-item-horizontal .ranking-index {
  margin-bottom: 10px;
}

/* 水平排行项中的用户信息 */
.ranking-item-horizontal .user-info {
  flex-direction: column;
  gap: 5px;
  margin-bottom: 10px;
}

/* 水平排行项中的用户名称 */
.ranking-item-horizontal .username {
  font-size: 14px;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 排名样式 - 金色（第一名） */
.ranking-gold {
  border-color: #f7c13b;
  color: #f7c13b;
}

.ranking-gold + .username {
  color: #f7c13b;
  font-weight: bold;
}

/* 排名样式 - 银色（第二名） */
.ranking-silver {
  border-color: #c0c4cc;
  color: #c0c4cc;
}

.ranking-silver + .username {
  color: #c0c4cc;
  font-weight: bold;
}

/* 排名样式 - 铜色（第三名） */
.ranking-bronze {
  border-color: #e8a055;
  color: #e8a055;
}

.ranking-bronze + .username {
  color: #e8a055;
  font-weight: bold;
}

/* 头像边框样式 */
.el-avatar.ranking-gold {
  border: 2px solid #f7c13b;
  box-shadow: 0 0 8px rgba(247, 193, 59, 0.4);
}

.el-avatar.ranking-silver {
  border: 2px solid #c0c4cc;
  box-shadow: 0 0 8px rgba(192, 196, 204, 0.4);
}

.el-avatar.ranking-bronze {
  border: 2px solid #e8a055;
  box-shadow: 0 0 8px rgba(232, 160, 85, 0.4);
}

/* 水平排行项中的计数 */
.ranking-item-horizontal .ranking-count {
  margin: auto 0 0 0;
}

/* 评论管理样式 */
.comment-management {
  width: 100%;
}

/* 主标签页样式 */
.main-tabs {
  margin-bottom: 20px;
}

.main-tabs .el-radio-group {
  font-size: 16px;
}

/* 主标签页和搜索框组合样式 */
.main-tabs-with-search {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

/* 主搜索框样式 */
.main-search {
  display: flex;
  align-items: center;
}

/* 评论过滤栏样式 */
.comment-filter-bar {
  display: flex;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #e0e0e0;
  margin-bottom: 15px;
}

.left-section {
  display: flex;
  align-items: center;
  margin-left: 20px;
}

.right-section {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-left: auto;
}

/* 子标签页样式 */
.sub-tabs {
  margin-right: 20px;
}

/* 筛选下拉框样式 */
.filter-dropdowns {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 视频评论蓝色字样样式 */
.video-comment-label {
  color: #1890ff;
  font-weight: 500;
  font-size: 14px;
}

/* 评论操作栏样式 */
.comment-actions {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

/* 操作按钮组样式 */
.action-buttons {
  display: flex;
  align-items: center;
  gap: 5px; /* 紧凑排布 */
}

.action-buttons .el-button {
  margin-right: 0; /* 移除默认右边距，使用gap控制间距 */
}

/* 排序选项样式 */
.sort-options {
  display: flex;
  align-items: center;
  margin-left: auto; /* 靠右对齐 */
}

/* 评论列表样式 */
.comment-list {
  margin-bottom: 20px;
}

/* 评论项样式 */
.comment-item {
  display: flex;
  align-items: flex-start;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
  position: relative;
  gap: 15px;
}

.comment-item:hover .comment-actions-hover {
  display: flex;
}

/* 复选框样式 */
.comment-checkbox {
  margin-top: 5px;
  flex-shrink: 0;
}

/* 评论主体样式 */
.comment-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-self: flex-start;
}

/* 头像和用户名样式 */
.comment-header {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-bottom: 5px;
}

.comment-header .el-avatar {
  margin-right: 10px;
  flex-shrink: 0;
}

/* 用户名样式 */
.username {
  font-weight: 500;
  color: #303133;
  font-size: 14px;
  line-height: 40px;
}

/* 评论内容样式 */
.comment-content {
  line-height: 1.5;
  color: #303133;
  font-size: 14px;
  margin-left: 50px;
  margin-top: 0;
  margin-bottom: 0;
}

/* 评论元信息样式 */
.comment-meta {
  display: flex;
  align-items: center;
  gap: 15px;
  margin-top: 5px;
  margin-left: 50px;
}

.comment-time {
  font-size: 12px;
  color: #909399;
}

.comment-meta .el-button {
  min-width: auto;
  padding: 0 8px;
  height: 24px;
  line-height: 24px;
  font-size: 12px;
}

/* 举报和删除按钮样式（鼠标悬停显示） */
.comment-actions-hover {
  display: none;
  align-items: center;
  gap: 5px;
  margin-left: 15px;
}

.comment-actions-hover .el-button {
  min-width: auto;
  padding: 0 8px;
  height: 24px;
  line-height: 24px;
  font-size: 12px;
}

.comment-right {
  margin-left: 20px;
  align-self: flex-start;
}

/* 视频缩略图样式 */
.video-thumbnail {
  width: 120px;
  text-align: center;
  margin-top: 0;
}

.video-thumbnail img {
  width: 100%;
  height: 68px;
  object-fit: cover;
  border-radius: 4px;
  display: block;
}

.video-title {
  font-size: 12px;
  color: #606266;
  margin-top: 5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 弹幕管理样式 */
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

/* 评论分页样式 */
.comment-pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #e0e0e0;
}

.comment-total {
  font-size: 12px;
  color: #909399;
}

.pagination-control {
  display: flex;
  justify-content: flex-end;
  align-items: center;
}

/* 自定义分页样式 */
.custom-pagination {
  display: flex;
  align-items: center;
  gap: 5px;
}

.custom-pagination .el-button {
  min-width: 32px;
  height: 32px;
  padding: 0 8px;
  border-radius: 4px;
  font-size: 14px;
  line-height: 32px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.custom-pagination .el-button--primary {
  background-color: #1890ff;
  border-color: #1890ff;
  color: #fff;
}

.custom-pagination .el-button--primary.is-plain {
  background-color: #fff;
  border-color: #d9d9d9;
  color: #1890ff;
}

/* 省略号样式 */
.custom-pagination .ellipsis {
  min-width: 32px;
  height: 32px;
  line-height: 32px;
  text-align: center;
  font-size: 14px;
  color: #606266;
}

/* 分页信息样式 */
.pagination-info {
  margin-left: 15px;
  font-size: 14px;
  color: #606266;
}

/* 稿件管理页面样式 */
/* 主要选择栏 */
.content-tabs {
  margin-bottom: 15px;
}

/* 视频管理选择栏 */
.video-filters {
  margin-bottom: 20px;
}

/* 过滤行样式 */
.filter-row {
  margin-bottom: 10px;
}

/* 过滤行之间的间距 */
.filter-row:not(:last-child) {
  margin-bottom: 15px;
}

/* 单个过滤组样式 */
.filter-row .el-radio-group {
  display: flex;
  align-items: center;
  gap: 10px;
}

/* 隐藏次级选择栏和状态选择栏 */
.sub-tabs,
.status-tabs {
  display: none;
}

/* 视频稿件列表 */
.article-list {
  margin-bottom: 20px;
}

/* 进度条单元格样式 */
.progress-cell {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 140px;
}

.progress-cell .el-progress {
  width: 100%;
}

.progress-text {
  font-size: 12px;
  color: #909399;
}

/* 表格样式调整 */
.article-list .el-table {
  border-radius: 6px;
  overflow: hidden;
}

/* 分页导航栏 */
.pagination {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  margin-top: 20px;
  padding: 15px 0;
  border-top: 1px solid #e0e0e0;
}

/* 合集管理样式 */
.collection-management {
  margin-top: 20px;
}

.collection-detail-view {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
}

.collection-detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e0e0e0;
}

.collection-detail-header .header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.collection-detail-header .header-title {
  font-size: 14px;
  color: #606266;
}

.collection-detail-header .collection-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.collection-info-section {
  margin-bottom: 24px;
}

.collection-info-section .info-container {
  display: flex;
  gap: 24px;
}

.collection-cover-large {
  width: 240px;
  height: 135px;
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  flex-shrink: 0;
}

.collection-cover-large img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-cover-large .cover-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.3s;
  cursor: pointer;
}

.collection-cover-large:hover .cover-overlay {
  opacity: 1;
}

.collection-cover-large .video-count-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-meta-detail {
  flex: 1;
}

.collection-title-large {
  font-size: 20px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 8px 0;
}

.collection-desc-large {
  font-size: 14px;
  color: #606266;
  margin: 0 0 16px 0;
  line-height: 1.6;
}

.collection-meta-detail .stats-info {
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  color: #909399;
}

.collection-meta-detail .stats-info .stat-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-meta-detail .private-tag {
  background: #f4f4f5;
  color: #909399;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.collection-videos-section {
  margin-top: 24px;
}

.collection-videos-section .videos-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-videos-section .sort-options {
  display: flex;
  gap: 16px;
}

.collection-videos-section .sort-item {
  font-size: 14px;
  color: #909399;
  cursor: pointer;
  padding: 4px 0;
  border-bottom: 2px solid transparent;
  transition: all 0.2s;
}

.collection-videos-section .sort-item.active {
  color: #00a1d6;
  border-bottom-color: #00a1d6;
}

.collection-videos-section .sort-item:hover {
  color: #00a1d6;
}

.collection-videos-section .header-actions {
  display: flex;
  gap: 8px;
}

.collection-videos-section .videos-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.collection-video-card {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  transition: transform 0.2s, box-shadow 0.2s;
}

.collection-video-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.collection-video-cover {
  position: relative;
  aspect-ratio: 16/9;
  cursor: pointer;
}

.collection-video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-index {
  position: absolute;
  top: 8px;
  left: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
}

.collection-video-duration {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.collection-video-actions {
  position: absolute;
  top: 8px;
  right: 8px;
}

.collection-video-info {
  padding: 12px;
}

.collection-video-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin: 0 0 8px 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-meta {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  color: #909399;
}

.collection-meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.add-video-card {
  aspect-ratio: 16/9;
  border: 2px dashed #dcdfe6;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s;
}

.add-video-card:hover {
  border-color: #00a1d6;
  background: #f0f9ff;
}

.add-video-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.add-video-card:hover .add-video-content {
  color: #00a1d6;
}

.collections-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.collections-header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.collections-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.view-toggle {
  display: flex;
  gap: 4px;
}

.view-toggle-btn {
  width: 32px;
  height: 32px;
  border: 1px solid #dcdfe6;
  background: #fff;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}

.view-toggle-btn.active {
  background: #00a1d6;
  border-color: #00a1d6;
  color: #fff;
}

.collections-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 280px));
  gap: 16px;
  justify-content: start;
}

.collection-item {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  max-width: 280px;
  width: 100%;
}

.collection-item:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.collection-item.new-collection {
  border: 2px dashed #dcdfe6;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 200px;
}

.collection-item.new-collection:hover {
  border-color: #00a1d6;
  background: #f0f9ff;
}

.new-collection-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  color: #909399;
}

.collection-item.new-collection:hover .new-collection-content {
  color: #00a1d6;
}

.collection-cover {
  position: relative;
  aspect-ratio: 16/9;
  width: 100%;
  overflow: hidden;
}

.collection-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.collection-video-count {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.collection-info {
  padding: 12px;
}

.collection-info .collection-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-info .collection-date {
  font-size: 12px;
  color: #909399;
}

.collections-horizontal {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.collection-horizontal-item {
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
}

.collection-horizontal-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
}

.collection-horizontal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.collection-horizontal-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.collection-video-count-badge {
  background: #f4f4f5;
  color: #909399;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: normal;
  flex-shrink: 0;
}

.collection-horizontal-actions {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.collection-horizontal-actions .action-btn {
  padding: 6px 12px;
  font-size: 13px;
}

.collection-videos-horizontal {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
  scrollbar-width: thin;
  scrollbar-color: #e0e0e0 #f5f7fa;
}

.collection-videos-horizontal::-webkit-scrollbar {
  height: 6px;
}

.collection-videos-horizontal::-webkit-scrollbar-track {
  background: #f5f7fa;
  border-radius: 3px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb {
  background: #e0e0e0;
  border-radius: 3px;
}

.collection-videos-horizontal::-webkit-scrollbar-thumb:hover {
  background: #d0d0d0;
}

.collection-videos-horizontal .add-video-card {
  min-width: 160px;
  height: 100px;
  flex-shrink: 0;
  max-width: 160px;
}

.video-horizontal-item {
  min-width: 160px;
  max-width: 160px;
  flex-shrink: 0;
  cursor: pointer;
  transition: transform 0.3s ease;
}

.video-horizontal-item:hover {
  transform: translateY(-2px);
}

.video-horizontal-cover {
  position: relative;
  aspect-ratio: 16/9;
  border-radius: 6px;
  overflow: hidden;
  background: #f0f2f5;
}

.video-cover-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  transition: transform 0.3s ease;
}

.video-horizontal-item:hover .video-cover-img {
  transform: scale(1.05);
}

.video-horizontal-cover .video-duration {
  position: absolute;
  bottom: 4px;
  right: 4px;
  background: rgba(0, 0, 0, 0.7);
  color: #fff;
  padding: 1px 4px;
  border-radius: 2px;
  font-size: 11px;
  z-index: 1;
}

.video-horizontal-info {
  padding: 8px 0;
}

.video-horizontal-info .video-title {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4px;
  line-height: 1.3;
}

.video-horizontal-meta {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #909399;
  align-items: center;
}

.video-horizontal-meta .video-views {
  display: flex;
  align-items: center;
  gap: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.video-horizontal-meta .video-date {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex-shrink: 0;
  width: 80px;
  text-align: right;
}

.cover-preview-small {
  width: 120px;
  height: 68px;
  border-radius: 4px;
  overflow: hidden;
  cursor: pointer;
}

.cover-preview-small img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-placeholder-small {
  width: 120px;
  height: 68px;
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: #909399;
  font-size: 12px;
  gap: 4px;
}

.cover-placeholder-small:hover {
  border-color: #00a1d6;
  color: #00a1d6;
}

.video-select-list {
  max-height: 400px;
  overflow-y: auto;
}

.video-selection-header {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 8px;
  margin-bottom: 16px;
}

.selection-tip {
  font-size: 13px;
  color: #606266;
}

.dialog-loading,
.dialog-empty {
  padding: 40px 0;
  text-align: center;
}

.video-select-item {
  padding: 8px;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.video-select-item:last-child {
  border-bottom: none;
}

.video-select-item .el-checkbox {
  flex: 1;
}

.video-select-content {
  display: flex;
  align-items: center;
  gap: 12px;
}

.video-select-cover {
  width: 80px;
  height: 45px;
  object-fit: cover;
  border-radius: 4px;
}

.in-collection-badge {
  background: linear-gradient(135deg, #005980, #1890ff);
  color: #ffffff;
  padding: 6px 14px;
  border-radius: 4px;
  font-size: 12px;
  flex-shrink: 0;
  margin-left: 8px;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
}

.video-select-info {
  flex: 1;
}

.video-select-title {
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
}

.video-select-meta {
  font-size: 12px;
  color: #909399;
}

.loading-state {
  padding: 40px;
  text-align: center;
  color: #909399;
}

.empty-state {
  padding: 40px;
  text-align: center;
}

/* 数据中心样式 */
/* 标签页样式 */
.data-center-tabs {
  margin-bottom: 20px;
}

/* 数据概览区域 */
.data-overview {
  background-color: #fff;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.05);
}

/* 概览头部 */
.overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.overview-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.overview-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.time-select {
  width: 120px;
}

.export-btn {
  display: flex;
  align-items: center;
  gap: 5px;
}

/* 核心数据卡片 */
.core-data {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 15px;
  margin-bottom: 30px;
}

.data-card {
  background-color: #fafafa;
  padding: 15px;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  transition: all 0.3s ease;
}

.data-card:hover {
  background-color: #f5f7fa;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

/* 播放量卡片特殊样式 */
.data-card.play-count {
  background-color: #ff5050;
  color: #fff;
}

.data-card.play-count:hover {
  background-color: #ff7875;
}

.data-title {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
}

.data-card.play-count .data-title {
  color: rgba(255, 255, 255, 0.8);
}

.data-value {
  font-size: 24px;
  font-weight: 600;
  margin-bottom: 8px;
}

.data-card.play-count .data-value {
  color: #fff;
}

.data-change {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
}

.increase {
  color: #67c23a;
}

.decrease {
  color: #f56c6c;
}

.data-card.play-count .increase,
.data-card.play-count .decrease {
  color: rgba(255, 255, 255, 0.9);
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 16px;
}

.chart-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 12px;
}

.chart-container {
  height: 260px;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* ECharts 图表容器 */
.echarts-container {
  width: 100%;
  height: 240px;
}

.echarts-container.pie-chart {
  height: 280px;
}

/* 图表标题 */
.chart-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

/* 图表描述 */
.chart-desc {
  font-size: 13px;
  color: #909399;
  margin-bottom: 12px;
}

/* 稿件排行区域 */
.manuscript-ranking-section {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #ebeef5;
}

/* 粉丝增长趋势图 */
.fans-trend-chart {
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid #ebeef5;
}

.fans-trend-chart .chart-container {
  height: 280px;
}

.fans-trend-chart .echarts-container {
  width: 100%;
  height: 280px;
}

/* 排行榜图表样式 */
.echarts-container.ranking-chart {
  height: 260px;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .overview-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 15px;
  }
  
  .core-data {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 10px;
  }
  
  .data-value {
    font-size: 20px;
  }
  
  .chart-container {
    height: 250px;
  }
  
  .echarts-container {
    min-height: 220px;
  }
}

/* 内容区域通用样式 */
.content-section {
  height: 100%;
}

/* 投稿表单样式 */
.upload-form {
  max-width: 800px;
  margin: 30px auto 0;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* 提示框样式 */
.upload-tips {
  display: flex;
  gap: 20px;
  margin-bottom: 30px;
  flex-wrap: wrap;
  justify-content: center;
  width: 100%;
  max-width: 800px;
}

.tip-box {
  flex: 1;
  min-width: 250px;
  background-color: #ecf5ff;
  border: 1px solid #d9ecff;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.tip-title {
  font-size: 16px;
  font-weight: 600;
  color: #1890ff;
  margin: 0 0 8px 0;
}

.tip-content {
  font-size: 14px;
  color: #606266;
  margin: 0;
  line-height: 1.5;
}

/* 移除el-form-item__content的内边距 */
.no-content-padding .el-form-item__content {
  padding: 0;
  margin: 0;
}

/* 全屏上传区域样式 */
.full-width-upload {
  width: 100%;
  max-width: 600px;
}

/* 让el-upload-dragger填满两边白边 */
.full-width-upload .el-upload-dragger {
  width: 100%;
  height: 100%;
  margin-bottom: 10px;
  border: 1px dashed #dcdfe6;
  border-radius: 8px;
  padding: 40px 20px;
  text-align: center;
  transition: all 0.3s ease;
}

.full-width-upload .el-upload-dragger:hover {
  border-color: #409eff;
  background-color: #f0f9ff;
}

/* 表单按钮样式 */
.form-actions {
  display: flex;
  justify-content: center;
  margin-top: 10px;
  width: 100%;
}

.form-actions .el-button--primary {
  background-color: #1890ff;
  border-color: #1890ff;
  font-size: 16px;
  padding: 12px 24px;
  height: auto;
}

.form-actions .el-button--primary:hover {
  background-color: #409eff;
  border-color: #409eff;
}

/* 上传进度和速度显示 */
.upload-progress-bar {
  margin-bottom: 20px;
}

.upload-speed {
  margin-top: 10px;
  text-align: right;
  font-size: 14px;
  color: #606266;
}

/* 添加分P按钮 */
.add-part-btn {
  margin-bottom: 30px;
}

/* 发布视频表单样式 */
.publish-form {
  max-width: 800px;
  margin: 0 auto;
}

/* 封面上传样式 */
.cover-upload-section {
  display: flex;
  align-items: flex-start;
  gap: 20px;
}

.cover-uploader {
  width: 320px;
  height: 180px;
}

.cover-preview {
  width: 100%;
  height: 100%;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-preview img {
  max-width: 100%;
  max-height: 100%;
  object-fit: cover;
}

.cover-placeholder {
  width: 100%;
  height: 100%;
  border: 1px dashed #dcdfe6;
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.3s ease;
}

.cover-placeholder:hover {
  border-color: #409eff;
  color: #409eff;
}

.cover-tip {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
  color: #909399;
  font-size: 14px;
}

/* 标签输入样式 */
.tag-input-section {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 10px;
}

/* 发布按钮样式 */
.publish-actions {
  display: flex;
  gap: 12px;
  margin-top: 30px;
  justify-content: flex-end;
}

/* 上传完成页面样式 */
.upload-completed {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 50px 0;
}

/* 发布表单项样式 */
.publish-form .el-form-item {
  margin-bottom: 20px;
}

/* 粉丝管理样式 */
.fans-container {
  padding: 20px;
}

.fans-header {
  margin-bottom: 15px;
}

.fans-header h3 {
  font-size: 18px;
  font-weight: bold;
  margin: 0;
}

.fans-count-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.fans-count {
  font-size: 16px;
  color: #333;
  margin: 0;
}

.fans-filter {
  display: flex;
  align-items: center;
}

.fans-list {
  margin-bottom: 20px;
}

.fan-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 15px 0;
  border-bottom: 1px solid #f0f0f0;
}

.fan-item:last-child {
  border-bottom: none;
}

.fan-info {
  display: flex;
  align-items: center;
}

.fan-info .el-avatar {
  margin-right: 12px;
}

/* 创作设置样式 */
.settings-container {
  padding: 20px;
}

.settings-header {
  margin-bottom: 20px;
}

.settings-header h3 {
  font-size: 18px;
  font-weight: bold;
  margin: 0;
}

.settings-form {
  max-width: 600px;
}

.form-item-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.fan-username {
  font-size: 14px;
  color: #333;
}

.fan-actions {
  display: flex;
  gap: 8px;
}

.fans-pagination {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.publish-form .el-form-item__label {
  font-weight: 600;
}

/* 账号诊断样式 */
.account-diagnosis {
  padding: 20px;
}

.diagnosis-summary {
  margin-bottom: 24px;
}

.diagnosis-summary .section-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 16px 0;
  display: flex;
  align-items: center;
  gap: 8px;
}

.summary-tip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fff7e6;
  border-radius: 8px;
  color: #fa8c16;
}

.tip-icon {
  font-size: 18px;
}

.diagnosis-content {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
}

.diagnosis-metrics {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  margin-bottom: 24px;
}

.metric-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
  transition: all 0.3s ease;
}

.metric-card:hover {
  background: #e6f7ff;
}

.metric-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: linear-gradient(135deg, #1890ff, #40a9ff);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.metric-info {
  flex: 1;
}

.metric-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 4px;
}

.metric-value {
  font-size: 24px;
  font-weight: 600;
  color: #303133;
}

.diagnosis-suggestions {
  border-top: 1px solid #f0f0f0;
  padding-top: 20px;
}

.suggestions-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin: 0 0 16px 0;
}

.suggestion-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fff7e6;
  border-radius: 8px;
  margin-bottom: 12px;
  color: #fa8c16;
}

.suggestion-item:last-child {
  margin-bottom: 0;
}

.suggestion-icon {
  font-size: 18px;
}

.suggestion-icon.success {
  color: #52c41a;
}

.suggestion-item:has(.suggestion-icon.success) {
  background: #f6ffed;
  color: #52c41a;
}

/* 稿件分析样式 */
.article-analysis {
  padding: 20px;
}

.ranking-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.ranking-header .section-title {
  font-size: 18px;
  font-weight: 600;
  color: #303133;
  margin: 0;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 16px;
  background: #fff;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.3s ease;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.ranking-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.ranking-index {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #f0f2f5;
  color: #909399;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 14px;
  flex-shrink: 0;
}

.ranking-index.top-3 {
  background: linear-gradient(135deg, #ff6b6b, #ffa500);
  color: #fff;
}

.ranking-cover {
  width: 120px;
  height: 68px;
  border-radius: 6px;
  overflow: hidden;
  flex-shrink: 0;
}

.ranking-cover img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.ranking-info {
  flex: 1;
  min-width: 0;
}

.ranking-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 8px;
}

.ranking-meta {
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #909399;
}

.ranking-meta .meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.ranking-stats {
  display: flex;
  gap: 12px;
  flex-shrink: 0;
}

.stat-badge {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 8px 16px;
  border-radius: 8px;
  background: #f5f7fa;
}

.stat-badge.views {
  background: #e6f7ff;
}

.stat-badge.interaction {
  background: #fff7e6;
}

.stat-badge .stat-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.stat-badge .stat-value {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

/* 视频信息卡片 */
.video-info-card {
  display: flex;
  align-items: center;
  gap: 20px;
  background-color: #fafafa;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
}

.video-thumbnail {
  position: relative;
  width: 180px;
  height: 100px;
  border-radius: 6px;
  overflow: hidden;
}

.video-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.video-duration {
  position: absolute;
  bottom: 5px;
  right: 5px;
  background-color: rgba(0, 0, 0, 0.7);
  color: #fff;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
}

.video-details {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.video-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  line-height: 1.5;
}

.video-stats {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #606266;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 5px;
}

.video-date {
  font-size: 14px;
  color: #909399;
}

.video-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: flex-end;
}

.switch-video {
  font-size: 14px;
  color: #1890ff;
  text-decoration: none;
  display: flex;
  align-items: center;
  gap: 5px;
}

.switch-video:hover {
  text-decoration: underline;
}

.compare-btn {
  background-color: #ff4d4f;
  border-color: #ff4d4f;
}

.compare-btn:hover {
  background-color: #ff7875;
  border-color: #ff7875;
}

/* 数据标签页 */
.data-tabs {
  margin-bottom: 20px;
}

.data-tab-btn {
  font-size: 14px;
  padding: 8px 16px;
}

.data-tab-btn.active {
  background-color: #1890ff;
  color: #fff;
}

/* 核心大屏数据 */
.core-big-screen {
  background-color: #fafafa;
  border-radius: 8px;
  padding: 20px;
  margin-bottom: 20px;
  position: relative;
}

.screen-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.export-data {
  position: absolute;
  top: 20px;
  right: 20px;
  font-size: 14px;
  color: #1890ff;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
}

.export-data:hover {
  text-decoration: underline;
}

.big-screen-data {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.data-row {
  display: flex;
  gap: 15px;
  flex-wrap: wrap;
}

.data-card {
  flex: 1;
  min-width: 120px;
  background-color: #fff;
  border-radius: 6px;
  padding: 16px;
  text-align: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.data-card.highlight {
  background-color: #ff4d4f;
  color: #fff;
}

.data-label {
  font-size: 14px;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 5px;
}

.data-value {
  font-size: 24px;
  font-weight: 600;
}

/* 播放量趋势 */
.play-trend-section {
  background-color: #fafafa;
  border-radius: 8px;
  padding: 20px;
}

.trend-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 10px;
}

.trend-title {
  font-size: 16px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.trend-update {
  font-size: 14px;
  color: #909399;
}

.time-selector {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  color: #606266;
}

/* 播放趋势图 */
.trend-chart-container {
  height: 300px;
  background-color: #fff;
  border-radius: 6px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.chart-axes {
  position: absolute;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.chart-y-axis {
  position: absolute;
  left: 0;
  top: 0;
  height: 100%;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding-right: 10px;
  text-align: right;
  font-size: 12px;
  color: #909399;
}

.chart-x-axis {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  display: flex;
  justify-content: space-between;
  padding-top: 10px;
  font-size: 12px;
  color: #909399;
}

.chart-line-container {
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
}

.chart-line {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(to right, #ff4d4f, #ff4d4f);
  background-size: 100% 100%;
  background-position: 0 100%;
  background-repeat: no-repeat;
  clip-path: polygon(0% 85%, 20% 85%, 30% 80%, 40% 30%, 50% 60%, 60% 65%, 70% 63%, 80% 65%, 90% 62%, 100% 63%, 100% 100%, 0% 100%);
  mask-image: linear-gradient(to right, rgba(0, 0, 0, 1), rgba(0, 0, 0, 1));
}

.chart-area {
  position: absolute;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background: linear-gradient(to top, rgba(255, 77, 79, 0.1), rgba(255, 77, 79, 0.3));
  clip-path: polygon(0% 85%, 20% 85%, 30% 80%, 40% 30%, 50% 60%, 60% 65%, 70% 63%, 80% 65%, 90% 62%, 100% 63%, 100% 100%, 0% 100%);
}

/* 粉丝分析样式 */
.fans-analysis {
  padding: 20px;
}

/* 数据概览 */
.fans-overview {
  margin-bottom: 30px;
}

.overview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  flex-wrap: wrap;
  gap: 10px;
}

.overview-title {
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;
}

.overview-update {
  font-size: 14px;
  color: #909399;
}

/* 粉丝数据卡片网格 */
.fans-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 15px;
}

.stat-card {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat-card.highlight {
  background-color: #ff4d4f;
  color: #fff;
  overflow: hidden;
}

.stat-title {
  font-size: 14px;
  margin-bottom: 8px;
  color: #606266;
}

.stat-card.highlight .stat-title {
  color: rgba(255, 255, 255, 0.9);
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.stat-card.highlight .stat-value {
  color: #fff;
}

.stat-icon {
  position: absolute;
  top: -20px;
  right: -20px;
  font-size: 80px;
  opacity: 0.2;
  color: #fff;
}

.stat-change {
  display: flex;
  align-items: center;
  gap: 5px;
  font-size: 14px;
  margin-top: 8px;
}

.stat-change.negative {
  color: #f56c6c;
}

/* 粉丝趋势 */
.fans-trend-section {
  margin-bottom: 30px;
  background-color: #fafafa;
  border-radius: 8px;
  padding: 20px;
}

.trend-actions {
  display: flex;
  gap: 10px;
}

.export-data-btn {
  color: #1890ff;
}

.export-data-btn:hover {
  color: #409eff;
}

/* 粉丝排行 */
.fans-ranking-section {
  background-color: #fafafa;
  border-radius: 8px;
  padding: 20px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 20px;
}

.ranking-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.ranking-card {
  background-color: #fff;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

.ranking-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #f0f0f0;
}

.ranking-icon {
  font-size: 18px;
  color: #1890ff;
}

.ranking-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.ranking-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.ranking-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ranking-index {
  width: 20px;
  font-size: 14px;
  color: #909399;
  text-align: center;
}

.ranking-username {
  font-size: 14px;
  color: #303133;
}

.ranking-list.empty {
  justify-content: center;
  align-items: center;
  height: 120px;
  color: #909399;
  font-size: 14px;
}

.view-all-ranking {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

/* 表现总结 */
.diagnosis-summary {
  margin-bottom: 30px;
}

.section-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.info-icon {
  font-size: 16px;
  color: #909399;
  cursor: help;
}

.summary-tip {
  background-color: #fef0f0;
  border: 1px solid #fbc4ab;
  border-radius: 6px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #f56c6c;
}

.tip-icon {
  font-size: 16px;
}

/* 诊断内容：雷达图和指标分析 */
.diagnosis-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
  margin-bottom: 30px;
}

/* 雷达图 */
.radar-chart-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 20px;
  padding: 20px;
  background-color: #fafafa;
  border-radius: 8px;
}

.radar-chart {
  width: 300px;
  height: 300px;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.radar-axes {
  position: absolute;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.radar-axis {
  position: absolute;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.radar-axis:nth-child(1) { transform: rotate(45deg); }
.radar-axis:nth-child(2) { transform: rotate(135deg); }
.radar-axis:nth-child(3) { transform: rotate(225deg); }
.radar-axis:nth-child(4) { transform: rotate(315deg); }

.axis-label {
  position: absolute;
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.radar-axis:nth-child(1) .axis-label { top: 10px; }
.radar-axis:nth-child(2) .axis-label { right: 10px; }
.radar-axis:nth-child(3) .axis-label { bottom: 10px; }
.radar-axis:nth-child(4) .axis-label { left: 10px; }

.radar-shape {
  position: absolute;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.radar-base {
  width: 200px;
  height: 200px;
  border: 2px dashed #d9d9d9;
  border-radius: 50%;
  position: relative;
}

.radar-base::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  width: 100%;
  height: 2px;
  background-color: #d9d9d9;
  transform: translateY(-50%);
}

.radar-base::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  width: 2px;
  height: 100%;
  background-color: #d9d9d9;
  transform: translateX(-50%);
}

.radar-data {
  position: absolute;
  width: 200px;
  height: 200px;
  background-color: rgba(255, 77, 79, 0.2);
  clip-path: polygon(50% 20%, 80% 50%, 50% 80%, 20% 50%);
  border: 2px solid #ff4d4f;
}

.radar-legend {
  display: flex;
  gap: 20px;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.legend-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.legend-dot.my-data {
  background-color: #ff4d4f;
}

.legend-dot.peer-data {
  background-color: #d9d9d9;
}

/* 指标分析 */
.metrics-analysis {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
  background-color: #fafafa;
  border-radius: 8px;
}

.metric-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.metric-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.metric-content {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: #606266;
}

.metric-icon {
  font-size: 16px;
}

.metric-item:nth-child(1) .metric-icon,
.metric-item:nth-child(3) .metric-icon,
.metric-item:nth-child(4) .metric-icon {
  color: #f56c6c;
}

.metric-item:nth-child(2) .metric-icon {
  color: #67c23a;
}

.metric-link {
  color: #1890ff;
  text-decoration: none;
  margin-left: auto;
}

.metric-link:hover {
  text-decoration: underline;
}

/* 播放分布 */
.play-distribution {
  margin-top: 40px;
}

.distribution-tip {
  background-color: #fef0f0;
  border: 1px solid #fbc4ab;
  border-radius: 6px;
  padding: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #f56c6c;
  margin-bottom: 20px;
}

.highlight {
  color: #1890ff;
  font-weight: 500;
}

/* 分布内容 */
.distribution-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 30px;
}

/* 人群分布和观看途径 */
.audience-stats {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.stat-card {
  background-color: #fafafa;
  border-radius: 8px;
  padding: 20px;
}

.stat-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.stat-period {
  font-size: 14px;
  font-weight: 400;
  color: #909399;
}

.stat-sort {
  font-size: 14px;
  font-weight: 400;
  color: #909399;
  display: flex;
  align-items: center;
  gap: 4px;
  margin-left: auto;
}

/* 进度条样式 */
.progress-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.progress-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.progress-label {
  font-size: 14px;
  color: #606266;
  display: flex;
  justify-content: space-between;
}

/* 来源稿件占比 */
.source-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.source-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.source-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.source-title {
  font-size: 14px;
  color: #606266;
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-right: 10px;
}

.source-percent {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  min-width: 50px;
  text-align: right;
}

/* 合集下拉框样式 */
.publish-form .el-select {
  width: 100%;
}

/* 弹幕卡片列表样式 */
.danmaku-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 10px 0;
}

/* 弹幕卡片样式 */
.danmaku-card {
  background: #ffffff;
  border-radius: 8px;
  padding: 12px 16px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.05);
  border: 1px solid #e8e8e8;
  transition: all 0.2s ease;
}

.danmaku-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  border-color: #d0d0d0;
}

/* 弹幕内容区域 */
.danmaku-body {
  background: #f5f5f5;
  border-radius: 6px;
  padding: 10px 14px;
  margin-bottom: 10px;
}

.danmaku-content {
  font-size: 14px;
  line-height: 1.5;
  font-weight: 500;
}

/* 弹幕底部信息 */
.danmaku-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.danmaku-info {
  display: flex;
  gap: 6px;
  align-items: center;
  flex-wrap: wrap;
}

.danmaku-info .el-tag {
  display: inline-flex;
  align-items: center;
}

.danmaku-date {
  font-size: 12px;
  color: #909399;
  white-space: nowrap;
}

.danmaku-actions {
  flex-shrink: 0;
}

/* 弹幕链接样式 */
.danmaku-link {
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
}

.danmaku-link:hover {
  text-decoration: underline;
}

.danmaku-divider {
  color: #c0c4cc;
  margin: 0 4px;
}

/* 弹幕空状态 */
.developing-tip {
  padding: 60px 0;
  text-align: center;
}

/* 弹幕分页样式 */
.danmaku-list + .el-pagination {
  margin-top: 20px;
  justify-content: center;
}

</style>
