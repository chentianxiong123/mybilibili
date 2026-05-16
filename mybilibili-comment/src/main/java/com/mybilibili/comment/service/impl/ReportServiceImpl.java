package com.mybilibili.comment.service.impl;

import com.mybilibili.comment.feign.ManuscriptClient;
import com.mybilibili.comment.feign.MessageClient;
import com.mybilibili.comment.mapper.CommentMapper;
import com.mybilibili.comment.mapper.DynamicCommentMapper;
import com.mybilibili.comment.mapper.ReplyMapper;
import com.mybilibili.comment.mapper.ReportMapper;
import com.mybilibili.comment.service.ReportService;
import com.mybilibili.common.entity.Comment;
import com.mybilibili.common.entity.DynamicComment;
import com.mybilibili.common.entity.Reply;
import com.mybilibili.common.entity.Report;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
public class ReportServiceImpl implements ReportService {

    @Autowired
    private ReportMapper reportMapper;

    @Autowired
    private CommentMapper commentMapper;

    @Autowired
    private ReplyMapper replyMapper;

    @Autowired
    private DynamicCommentMapper dynamicCommentMapper;

    @Autowired
    private MessageClient messageClient;

    @Autowired
    private ManuscriptClient manuscriptClient;

    @Override
    public Result<?> submitReport(Integer reporterId, String targetType, Integer targetId,
                                   Integer manuscriptId, String reason, String description) {
        try {
            // 检查是否已举报过
            Report existing = reportMapper.selectOne(
                new com.baomidou.mybatisplus.core.conditions.query.QueryWrapper<Report>()
                    .eq("reporter_id", reporterId)
                    .eq("target_type", targetType)
                    .eq("target_id", targetId)
                    .eq("status", "PENDING"));
            if (existing != null) {
                return Result.error("您已举报过该内容，请等待处理");
            }

            Report report = new Report();
            report.setReporterId(reporterId);
            report.setTargetType(targetType);
            report.setTargetId(targetId);
            report.setManuscriptId(manuscriptId);
            report.setReason(reason);
            report.setDescription(description);
            report.setStatus("PENDING");
            reportMapper.insert(report);
            return Result.success("举报成功，我们会尽快处理", null);
        } catch (Exception e) {
            return Result.error("举报失败: " + e.getMessage());
        }
    }

    @Override
    public Result<?> getReportList(String status, String targetType, Integer page, Integer size) {
        try {
            int offset = (page - 1) * size;
            List<Report> reports = reportMapper.selectReportList(status, targetType, offset, size);
            int total = reportMapper.countReportList(status, targetType);

            // 补充举报目标的内容信息
            List<Map<String, Object>> list = new ArrayList<>();
            for (Report report : reports) {
                Map<String, Object> item = new HashMap<>();
                item.put("id", report.getId());
                item.put("reporterId", report.getReporterId());
                item.put("targetType", report.getTargetType());
                item.put("targetId", report.getTargetId());
                item.put("manuscriptId", report.getManuscriptId());
                item.put("reason", report.getReason());
                item.put("description", report.getDescription());
                item.put("status", report.getStatus());
                item.put("adminRemark", report.getAdminRemark());
                item.put("createdAt", report.getCreatedAt());
                item.put("processedAt", report.getProcessedAt());

                // 获取被举报内容
                String targetContent = getTargetContent(report.getTargetType(), report.getTargetId());
                item.put("targetContent", targetContent);

                list.add(item);
            }

            Map<String, Object> data = new HashMap<>();
            data.put("list", list);
            data.put("total", total);
            return Result.success(data);
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    @Override
    public Result<?> processReport(Integer reportId, String action, String adminRemark, Integer adminId) {
        try {
            Report report = reportMapper.selectById(reportId);
            if (report == null) {
                return Result.error("举报记录不存在");
            }
            if (!"PENDING".equals(report.getStatus())) {
                return Result.error("该举报已处理");
            }

            // 查找被举报内容的作者
            Integer contentAuthorId = getContentAuthorId(report.getTargetType(), report.getTargetId());

            if ("resolve".equals(action)) {
                // 处理举报：下架内容
                report.setStatus("RESOLVED");
                takeDownContent(report.getTargetType(), report.getTargetId());

                // 发送系统通知给举报人
                try {
                    messageClient.sendSystemNotification(
                        report.getReporterId(),
                        "举报处理通知",
                        "您举报的内容已被处理，内容已下架。感谢您为维护社区环境做出的贡献！");
                } catch (Exception ignored) {}

                // 发送系统通知给被举报人
                if (contentAuthorId != null) {
                    try {
                        String remark = (adminRemark != null && !adminRemark.isEmpty())
                            ? "，管理员备注：" + adminRemark : "";
                        messageClient.sendSystemNotification(
                            contentAuthorId,
                            "内容下架通知",
                            "您的" + getTypeName(report.getTargetType()) + "因违规已被下架，原因：" + report.getReason() + remark);
                    } catch (Exception ignored) {}
                }
            } else if ("reject".equals(action)) {
                // 驳回举报
                report.setStatus("REJECTED");

                // 发送系统通知给举报人
                try {
                    messageClient.sendSystemNotification(
                        report.getReporterId(),
                        "举报处理通知",
                        "您举报的内容经审核未发现违规，举报已被驳回。" +
                        (adminRemark != null && !adminRemark.isEmpty() ? "备注：" + adminRemark : ""));
                } catch (Exception ignored) {}
            } else {
                return Result.error("无效的操作");
            }

            report.setAdminRemark(adminRemark);
            report.setProcessedAt(new Date());
            reportMapper.updateById(report);

            return Result.success("处理成功");
        } catch (Exception e) {
            return Result.error(e.getMessage());
        }
    }

    private String getTargetContent(String targetType, Integer targetId) {
        if ("MANUSCRIPT".equals(targetType)) {
            try {
                Result<com.mybilibili.common.vo.ManuscriptVO> res = manuscriptClient.getManuscriptById(targetId);
                if (res != null && res.getData() != null) {
                    return "[稿件] " + res.getData().getTitle();
                }
            } catch (Exception ignored) {}
            return "[稿件] ID: " + targetId;
        } else if ("COMMENT".equals(targetType)) {
            Comment comment = commentMapper.selectById(targetId);
            return comment != null ? comment.getContent() : null;
        } else if ("REPLY".equals(targetType)) {
            Reply reply = replyMapper.selectById(targetId);
            return reply != null ? reply.getContent() : null;
        } else if ("DYNAMIC_COMMENT".equals(targetType)) {
            DynamicComment dc = dynamicCommentMapper.selectById(targetId);
            return dc != null ? dc.getContent() : null;
        }
        return null;
    }

    private Integer getContentAuthorId(String targetType, Integer targetId) {
        if ("MANUSCRIPT".equals(targetType)) {
            try {
                Result<com.mybilibili.common.vo.ManuscriptVO> res = manuscriptClient.getManuscriptById(targetId);
                if (res != null && res.getData() != null) {
                    return res.getData().getUserId();
                }
            } catch (Exception ignored) {}
            return null;
        } else if ("COMMENT".equals(targetType)) {
            Comment comment = commentMapper.selectById(targetId);
            return comment != null ? comment.getUserId() : null;
        } else if ("REPLY".equals(targetType)) {
            Reply reply = replyMapper.selectById(targetId);
            return reply != null ? reply.getUserId() : null;
        } else if ("DYNAMIC_COMMENT".equals(targetType)) {
            DynamicComment dc = dynamicCommentMapper.selectById(targetId);
            return dc != null ? dc.getUserId() : null;
        }
        return null;
    }

    private void takeDownContent(String targetType, Integer targetId) {
        if ("MANUSCRIPT".equals(targetType)) {
            try {
                manuscriptClient.takeDownManuscript(targetId);
            } catch (Exception ignored) {}
        } else if ("COMMENT".equals(targetType)) {
            Comment comment = commentMapper.selectById(targetId);
            if (comment != null) {
                comment.setStatus(1);
                commentMapper.updateById(comment);
            }
        } else if ("REPLY".equals(targetType)) {
            Reply reply = replyMapper.selectById(targetId);
            if (reply != null) {
                reply.setStatus("REMOVED");
                replyMapper.updateById(reply);
            }
        } else if ("DYNAMIC_COMMENT".equals(targetType)) {
            DynamicComment dc = dynamicCommentMapper.selectById(targetId);
            if (dc != null) {
                dc.setStatus(1);
                dynamicCommentMapper.updateById(dc);
            }
        }
    }

    private String getTypeName(String targetType) {
        switch (targetType) {
            case "COMMENT": return "评论";
            case "REPLY": return "回复";
            case "DYNAMIC_COMMENT": return "动态评论";
            default: return "内容";
        }
    }
}
