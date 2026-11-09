package com.mybilibili.admin.service.impl;

import com.mybilibili.admin.mapper.CommentReviewMapper;
import com.mybilibili.admin.service.ContentReviewAdminService;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Service
public class ContentReviewAdminServiceImpl implements ContentReviewAdminService {

    @Autowired
    private CommentReviewMapper commentReviewMapper;

    @Override
    public Result<?> getPendingList(String contentType, Integer page, Integer size) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;

            int offset = (page - 1) * size;
            List<Map<String, Object>> list = new ArrayList<>();
            int total = 0;

            if (contentType == null || "COMMENT".equals(contentType)) {
                List<Map<String, Object>> comments = commentReviewMapper.selectPendingComments(offset, size);
                for (Map<String, Object> comment : comments) {
                    comment.put("contentType", "COMMENT");
                    list.add(comment);
                }
                total += commentReviewMapper.countPendingComments();
            }

            if (contentType == null || "REPLY".equals(contentType)) {
                int replyOffset = contentType == null ? offset : offset;
                int replySize = contentType == null ? size : size;
                List<Map<String, Object>> replies = commentReviewMapper.selectPendingReplies(replyOffset, replySize);
                for (Map<String, Object> reply : replies) {
                    reply.put("contentType", "REPLY");
                    list.add(reply);
                }
                total += commentReviewMapper.countPendingReplies();
            }

            // 按创建时间排序
            list.sort((a, b) -> {
                Object timeA = a.get("createdAt");
                Object timeB = b.get("createdAt");
                if (timeA == null || timeB == null) return 0;
                return timeB.toString().compareTo(timeA.toString());
            });

            Map<String, Object> data = new HashMap<>();
            data.put("list", list);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取待审核列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> getAllContent(String contentType, String status, Integer page, Integer size) {
        try {
            if (page == null || page < 1) page = 1;
            if (size == null || size < 1) size = 10;

            int offset = (page - 1) * size;
            List<Map<String, Object>> list = new ArrayList<>();
            int total = 0;

            // 将字符串状态转换为数字状态（comments表使用int类型：0-正常，1-已下架）
            Integer commentStatus = null;
            if ("NORMAL".equals(status)) {
                commentStatus = 0;
            } else if ("REMOVED".equals(status)) {
                commentStatus = 1;
            }

            if (contentType == null || "COMMENT".equals(contentType)) {
                List<Map<String, Object>> comments = commentReviewMapper.selectComments(commentStatus, offset, size);
                for (Map<String, Object> comment : comments) {
                    comment.put("contentType", "COMMENT");
                    list.add(comment);
                }
                total += commentReviewMapper.countComments(commentStatus);
            }

            if (contentType == null || "REPLY".equals(contentType)) {
                List<Map<String, Object>> replies = commentReviewMapper.selectReplies(status, offset, size);
                for (Map<String, Object> reply : replies) {
                    reply.put("contentType", "REPLY");
                    list.add(reply);
                }
                total += commentReviewMapper.countReplies(status);
            }

            // 按创建时间排序
            list.sort((a, b) -> {
                Object timeA = a.get("createdAt");
                Object timeB = b.get("createdAt");
                if (timeA == null || timeB == null) return 0;
                return timeB.toString().compareTo(timeA.toString());
            });

            Map<String, Object> data = new HashMap<>();
            data.put("list", list);
            data.put("total", total);
            data.put("page", page);
            data.put("size", size);

            return Result.success("获取内容列表成功", data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> restoreContent(String type, Integer id) {
        try {
            int result;
            if ("COMMENT".equals(type)) {
                result = commentReviewMapper.restoreComment(id);
            } else if ("REPLY".equals(type)) {
                result = commentReviewMapper.restoreReply(id);
            } else {
                return Result.error("无效的内容类型");
            }

            if (result > 0) {
                return Result.success("恢复成功", null);
            } else {
                return Result.error("恢复失败，内容不存在");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> deleteContent(String type, Integer id) {
        try {
            int result;
            if ("COMMENT".equals(type)) {
                result = commentReviewMapper.deleteComment(id);
            } else if ("REPLY".equals(type)) {
                result = commentReviewMapper.deleteReply(id);
            } else {
                return Result.error("无效的内容类型");
            }

            if (result > 0) {
                return Result.success("删除成功", null);
            } else {
                return Result.error("删除失败，内容不存在");
            }
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> batchProcess(String action, List<Map<String, Object>> items) {
        try {
            if (items == null || items.isEmpty()) {
                return Result.error("没有选择任何内容");
            }

            int successCount = 0;
            int failCount = 0;

            for (Map<String, Object> item : items) {
                String type = (String) item.get("type");
                Integer id = (Integer) item.get("id");

                if (type == null || id == null) {
                    failCount++;
                    continue;
                }

                try {
                    if ("restore".equals(action)) {
                        if ("COMMENT".equals(type)) {
                            commentReviewMapper.restoreComment(id);
                        } else if ("REPLY".equals(type)) {
                            commentReviewMapper.restoreReply(id);
                        }
                    } else if ("delete".equals(action)) {
                        if ("COMMENT".equals(type)) {
                            commentReviewMapper.deleteComment(id);
                        } else if ("REPLY".equals(type)) {
                            commentReviewMapper.deleteReply(id);
                        }
                    }
                    successCount++;
                } catch (Exception e) {
                    failCount++;
                }
            }

            Map<String, Object> result = new HashMap<>();
            result.put("successCount", successCount);
            result.put("failCount", failCount);

            return Result.success("批量处理完成", result);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }
}
