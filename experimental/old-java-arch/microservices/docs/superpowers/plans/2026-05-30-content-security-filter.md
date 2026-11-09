# 内容安全过滤增强实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在评论/回复发布前加入 Redis 缓存违禁词 + 频率限制（防 spam），解决每次查 DB 和无法防刷的问题

**Architecture:** 在 `mybilibili-comment` 模块新建违禁词缓存服务（Redis）和频率限制服务，定时刷新词表到 Redis，发布评论时先过缓存和频率检查，再决定是否入库

**Tech Stack:** Spring Data Redis (`StringRedisTemplate`)，已有模式复用

---

## 文件改动总览

### 后端 (mybilibili-comment)
| 操作 | 文件 |
|------|------|
| 新建 | `service/ProhibitedWordCacheService.java` — 违禁词缓存接口 |
| 新建 | `service/impl/ProhibitedWordCacheServiceImpl.java` — 违禁词 Redis 缓存实现 |
| 新建 | `service/SpamPreventionService.java` — 防刷接口 |
| 新建 | `service/impl/SpamPreventionServiceImpl.java` — 频率限制实现 |
| 修改 | `service/impl/CommentServiceImpl.java` — 调用缓存和防刷服务 |

### 后端 (mybilibili-common)
| 操作 | 文件 |
|------|------|
| 新建 | `service/ContentSecurityService.java` — 统一内容安全接口 |
| 新建 | `service/impl/ContentSecurityServiceImpl.java` — 统一内容安全实现（组合违禁词+防刷） |

---

### Task 1: 违禁词 Redis 缓存服务

**Files:**
- Create: `mybilibili-comment/src/main/java/com/mybilibili/comment/service/ProhibitedWordCacheService.java`
- Create: `mybilibili-comment/src/main/java/com/mybilibili/comment/service/impl/ProhibitedWordCacheServiceImpl.java`

- [ ] **Step 1: 创建 ProhibitedWordCacheService 接口**

```java
package com.mybilibili.comment.service;

import java.util.List;

public interface ProhibitedWordCacheService {

    /**
     * 检查内容是否包含违禁词
     * @param content 待检查内容
     * @return 违禁词列表（空表示通过）
     */
    List<String> check(String content);

    /**
     * 刷新缓存（从数据库重新加载违禁词到Redis）
     */
    void refreshCache();

    /**
     * 获取缓存中的违禁词数量
     */
    long getCacheSize();
}
```

- [ ] **Step 2: 创建 ProhibitedWordCacheServiceImpl 实现**

```java
package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.mapper.ProhibitedWordMapper;
import com.mybilibili.comment.service.ProhibitedWordCacheService;
import com.mybilibili.common.entity.ProhibitedWord;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;

import jakarta.annotation.PostConstruct;
import java.util.ArrayList;
import java.util.List;
import java.util.Set;

@Service
public class ProhibitedWordCacheServiceImpl implements ProhibitedWordCacheService {

    private static final String REDIS_KEY = "prohibited_words:cache";
    private static final String REDIS_SET_KEY = "prohibited_words:words";

    @Autowired
    private ProhibitedWordMapper prohibitedWordMapper;

    @Autowired
    private StringRedisTemplate redisTemplate;

    /** 启动时加载一次缓存 */
    @PostConstruct
    public void init() {
        refreshCache();
    }

    /** 每5分钟刷新一次缓存 */
    @Scheduled(fixedRate = 300000)
    public void scheduledRefresh() {
        refreshCache();
    }

    @Override
    public List<String> check(String content) {
        if (content == null || content.isEmpty()) {
            return new ArrayList<>();
        }

        List<String> found = new ArrayList<>();
        // 遍历缓存中的违禁词进行匹配
        Set<String> cachedWords = redisTemplate.opsForSet().members(REDIS_SET_KEY);
        if (cachedWords != null) {
            for (String word : cachedWords) {
                if (content.contains(word)) {
                    found.add(word);
                }
            }
        }
        return found;
    }

    @Override
    public void refreshCache() {
        List<ProhibitedWord> words = prohibitedWordMapper.selectAllEnabled();
        if (words == null || words.isEmpty()) {
            redisTemplate.delete(REDIS_SET_KEY);
            return;
        }

        // 清除旧缓存
        redisTemplate.delete(REDIS_SET_KEY);

        // 写入 Redis Set（自动去重）
        for (ProhibitedWord pw : words) {
            if (pw.getWord() != null && !pw.getWord().isEmpty()) {
                redisTemplate.opsForSet().add(REDIS_SET_KEY, pw.getWord());
            }
        }
    }

    @Override
    public long getCacheSize() {
        Long size = redisTemplate.opsForSet().size(REDIS_SET_KEY);
        return size != null ? size : 0;
    }
}
```

- [ ] **Step 3: 在 CommentServiceImpl 中注入并使用缓存服务**

在 `CommentServiceImpl.java` 中添加：

```java
@Autowired
private ProhibitedWordCacheService prohibitedWordCacheService;

// 替换 detectProhibitedWords 方法中的数据库查询逻辑
private List<String> detectProhibitedWords(String content) {
    return prohibitedWordCacheService.check(content);
}
```

原有 `prohibitedWordMapper` 仍保留（用于违禁词管理页面 CRUD），缓存层透明替换检测逻辑。

---

### Task 2: 防刷服务（频率限制）

**Files:**
- Create: `mybilibili-comment/src/main/java/com/mybilibili/comment/service/SpamPreventionService.java`
- Create: `mybilibili-comment/src/main/java/com/mybilibili/comment/service/impl/SpamPreventionServiceImpl.java`

- [ ] **Step 1: 创建 SpamPreventionService 接口**

```java
package com.mybilibili.comment.service;

public interface SpamPreventionService {

    /**
     * 检查用户是否超出发布频率
     * @param userId 用户ID
     * @param action 操作类型（comment / reply）
     * @return true=超出限制，false=正常
     */
    boolean isRateLimited(Integer userId, String action);

    /**
     * 记录用户发布行为（每次发布后调用）
     */
    void recordAction(Integer userId, String action);

    /**
     * 获取当前用户在该操作的剩余可发布次数
     */
    long getRemainingCount(Integer userId, String action);
}
```

- [ ] **Step 2: 创建 SpamPreventionServiceImpl 实现**

使用 Redis 滑动窗口计数：

```java
package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.service.SpamPreventionService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;

import java.util.concurrent.TimeUnit;

@Service
public class SpamPreventionServiceImpl implements SpamPreventionService {

    /** 评论：60秒内最多10条 */
    private static final int COMMENT_WINDOW_SECONDS = 60;
    private static final int COMMENT_MAX_COUNT = 10;

    /** 回复：60秒内最多20条 */
    private static final int REPLY_WINDOW_SECONDS = 60;
    private static final int REPLY_MAX_COUNT = 20;

    @Autowired
    private StringRedisTemplate redisTemplate;

    private String key(Integer userId, String action) {
        return "spam:" + action + ":" + userId;
    }

    @Override
    public boolean isRateLimited(Integer userId, String action) {
        if (userId == null) return false;

        int maxCount = "comment".equals(action) ? COMMENT_MAX_COUNT : REPLY_MAX_COUNT;
        String k = key(userId, action);
        String countStr = redisTemplate.opsForValue().get(k);
        if (countStr == null) return false;

        int count = Integer.parseInt(countStr);
        return count >= maxCount;
    }

    @Override
    public void recordAction(Integer userId, String action) {
        if (userId == null) return;

        String k = key(userId, action);
        Long count = redisTemplate.opsForValue().increment(k);
        if (count != null && count == 1) {
            // 首次写入，设置过期时间
            int window = "comment".equals(action) ? COMMENT_WINDOW_SECONDS : REPLY_WINDOW_SECONDS;
            redisTemplate.expire(k, window, TimeUnit.SECONDS);
        }
    }

    @Override
    public long getRemainingCount(Integer userId, String action) {
        if (userId == null) return 0;

        int maxCount = "comment".equals(action) ? COMMENT_MAX_COUNT : REPLY_MAX_COUNT;
        String k = key(userId, action);
        String countStr = redisTemplate.opsForValue().get(k);
        if (countStr == null) return maxCount;

        int count = Integer.parseInt(countStr);
        return Math.max(0, maxCount - count);
    }
}
```

- [ ] **Step 3: 在 CommentServiceImpl 中集成防刷**

在 `CommentServiceImpl.java` 中添加：

```java
@Autowired
private SpamPreventionService spamPreventionService;
```

修改 `addComment` 方法开头：

```java
@Override
public CommentVO addComment(Integer manuscriptId, Integer userId, String content) {
    // 频率检查
    if (spamPreventionService.isRateLimited(userId, "comment")) {
        throw new RuntimeException("发布太频繁，请稍后再试");
    }

    // 违禁词检查（已通过缓存服务替换原有方法）
    List<String> prohibitedWords = detectProhibitedWords(content);
    boolean hasProhibitedWords = !prohibitedWords.isEmpty();
    // ... 原有逻辑不变 ...

    // 发布成功后记录行为
    if (!hasProhibitedWords) {
        spamPreventionService.recordAction(userId, "comment");
    }
    // ...
}
```

同样修改 `addReply` 方法和 `addCommentByManuscriptId` 方法。

---

### Task 3: 统一内容安全服务（可选，后续扩展）

**Files:**
- Create: `mybilibili-common/src/main/java/com/mybilibili/common/service/ContentSecurityService.java`
- Create: `mybilibili-common/src/main/java/com/mybilibili/common/service/impl/ContentSecurityServiceImpl.java`

此服务作为统一入口，将来可以在这里加入：
- 正则 spam 检测（URL 重复、广告格式）
- 多层过滤组合（违禁词 → 频率 → 正则 → LLM）
- 举报审核结果反哺预过滤

---

### 测试验证

1. 启动 comment 服务，确认 Redis 连接正常
2. 连续发 15 条评论 → 第 11 条起应被频率限制拦截
3. 发送包含违禁词的内容 → 应被标记为已删除
4. 违禁词表更新后 5 分钟内自动刷新缓存

---

### 未涵盖（后续扩展）

- 正则 spam 检测（URL spam、广告格式）
- 举报审核结果反哺预过滤词表
- LLM 辅助审核（llama-guard 接入）
- 动态评论（DynamicComment）的预过滤