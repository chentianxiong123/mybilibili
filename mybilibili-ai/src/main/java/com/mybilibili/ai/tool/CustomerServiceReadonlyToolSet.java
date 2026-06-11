package com.mybilibili.ai.tool;

import com.mybilibili.ai.entity.AiSession;
import com.mybilibili.ai.mapper.AiSessionMapper;
import com.mybilibili.ai.mapper.ManuscriptMapper;
import com.mybilibili.ai.mapper.VideoMapper;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Video;
import org.springframework.ai.tool.annotation.Tool;

import java.util.LinkedHashMap;
import java.util.Map;

public class CustomerServiceReadonlyToolSet {

    private static final String TYPE_CUSTOMER_SERVICE = "CUSTOMER_SERVICE";

    private final Integer userId;
    private final VideoMapper videoMapper;
    private final ManuscriptMapper manuscriptMapper;
    private final AiSessionMapper aiSessionMapper;

    public CustomerServiceReadonlyToolSet(Long userId,
                                          VideoMapper videoMapper,
                                          ManuscriptMapper manuscriptMapper,
                                          AiSessionMapper aiSessionMapper) {
        this.userId = userId == null ? null : userId.intValue();
        this.videoMapper = videoMapper;
        this.manuscriptMapper = manuscriptMapper;
        this.aiSessionMapper = aiSessionMapper;
    }

    @Tool(name = "queryMyManuscriptStatus", description = "只读查询当前登录用户自己的稿件状态。只能用于解释稿件是否待审核、处理中、待发布、已发布、驳回、处理失败或已下架。")
    public Map<String, Object> queryMyManuscriptStatus(Integer manuscriptId) {
        Map<String, Object> result = baseResult("queryMyManuscriptStatus");
        if (manuscriptId == null) {
            return error(result, "请提供稿件ID");
        }

        Manuscript manuscript = manuscriptMapper.selectById(manuscriptId);
        if (manuscript == null) {
            result.put("found", false);
            result.put("message", "未找到该稿件");
            return result;
        }
        if (!isOwner(manuscript.getUserId())) {
            return forbidden(result);
        }

        result.put("found", true);
        result.put("manuscriptId", manuscript.getId());
        result.put("title", manuscript.getTitle());
        result.put("status", manuscript.getStatus());
        result.put("statusText", manuscriptStatusText(manuscript.getStatus()));
        result.put("reviewStatus", manuscript.getReviewStatus());
        result.put("reviewStatusText", reviewStatusText(manuscript.getReviewStatus()));
        result.put("uploadTime", manuscript.getUploadTime());
        result.put("updatedAt", manuscript.getUpdatedAt());
        if (Integer.valueOf(Manuscript.STATUS_REJECTED).equals(manuscript.getStatus())) {
            result.put("reviewReason", manuscript.getReviewReason());
        }
        return result;
    }

    @Tool(name = "queryMyVideoProcessStatus", description = "只读查询当前登录用户自己的视频处理状态。只能解释转码、音频提取、字幕生成、AI摘要生成、完成或失败状态；不能触发重试或处理。")
    public Map<String, Object> queryMyVideoProcessStatus(Integer videoId) {
        Map<String, Object> result = baseResult("queryMyVideoProcessStatus");
        Video video = loadOwnedVideo(videoId, result);
        if (video == null) {
            return result;
        }

        result.put("found", true);
        result.put("videoId", video.getId());
        result.put("manuscriptId", video.getManuscriptId());
        result.put("title", video.getTitle());
        result.put("processStatus", video.getProcessStatus());
        result.put("processStatusText", processStatusText(video.getProcessStatus()));
        result.put("processProgress", video.getProcessProgress() == null ? 0 : video.getProcessProgress());
        result.put("processStage", video.getProcessStage());
        result.put("failed", isFailedStatus(video.getProcessStatus()));
        if (isFailedStatus(video.getProcessStatus())) {
            result.put("message", "视频处理失败，客服不能直接重试处理任务；如需处理请转人工。");
        }
        return result;
    }

    @Tool(name = "queryVideoAiAssets", description = "只读查询当前登录用户自己的视频是否已有字幕和AI摘要。只返回是否生成，不返回字幕或摘要正文。")
    public Map<String, Object> queryVideoAiAssets(Integer videoId) {
        Map<String, Object> result = baseResult("queryVideoAiAssets");
        Video video = loadOwnedVideo(videoId, result);
        if (video == null) {
            return result;
        }

        result.put("found", true);
        result.put("videoId", video.getId());
        result.put("manuscriptId", video.getManuscriptId());
        result.put("title", video.getTitle());
        result.put("hasSubtitle", Integer.valueOf(1).equals(video.getHasSubtitle()));
        result.put("hasSummary", Integer.valueOf(1).equals(video.getHasSummary()));
        result.put("processStatus", video.getProcessStatus());
        result.put("processStatusText", processStatusText(video.getProcessStatus()));
        return result;
    }

    @Tool(name = "queryCustomerSessionStatus", description = "只读查询当前用户最近一条AI客服会话状态。用于解释是否AI服务中、等待人工或已处理。")
    public Map<String, Object> queryCustomerSessionStatus() {
        Map<String, Object> result = baseResult("queryCustomerSessionStatus");
        if (userId == null) {
            return error(result, "未登录，无法查询客服会话");
        }

        AiSession session = aiSessionMapper.selectLatestByUserIdAndType(userId.longValue(), TYPE_CUSTOMER_SERVICE);
        if (session == null) {
            result.put("found", false);
            result.put("message", "暂无客服会话");
            return result;
        }

        result.put("found", true);
        result.put("sessionId", session.getId());
        result.put("status", session.getStatus());
        result.put("statusText", customerSessionStatusText(session.getStatus()));
        result.put("createdAt", session.getCreatedAt());
        result.put("updatedAt", session.getUpdatedAt());
        return result;
    }

    private Video loadOwnedVideo(Integer videoId, Map<String, Object> result) {
        if (videoId == null) {
            error(result, "请提供视频ID");
            return null;
        }

        Video video = videoMapper.selectById(videoId);
        if (video == null) {
            result.put("found", false);
            result.put("message", "未找到该视频");
            return null;
        }

        Manuscript manuscript = manuscriptMapper.selectById(video.getManuscriptId());
        if (manuscript == null) {
            result.put("found", false);
            result.put("message", "未找到视频所属稿件");
            return null;
        }
        if (!isOwner(manuscript.getUserId())) {
            forbidden(result);
            return null;
        }
        return video;
    }

    private boolean isOwner(Integer ownerUserId) {
        return userId != null && ownerUserId != null && ownerUserId.equals(userId);
    }

    private Map<String, Object> baseResult(String tool) {
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("tool", tool);
        result.put("readonly", true);
        return result;
    }

    private Map<String, Object> error(Map<String, Object> result, String message) {
        result.put("error", true);
        result.put("message", message);
        return result;
    }

    private Map<String, Object> forbidden(Map<String, Object> result) {
        result.put("forbidden", true);
        result.put("message", "只能查询当前登录用户自己的数据");
        return result;
    }

    private String manuscriptStatusText(Integer status) {
        if (status == null) return "未知";
        return switch (status) {
            case Manuscript.STATUS_PENDING_REVIEW -> "待审核";
            case Manuscript.STATUS_PROCESSING -> "处理中";
            case Manuscript.STATUS_PUBLISHED -> "已发布";
            case Manuscript.STATUS_REJECTED -> "已驳回";
            case Manuscript.STATUS_PROCESS_FAILED -> "处理失败";
            case Manuscript.STATUS_UNPUBLISHED -> "已下架";
            default -> "状态: " + status;
        };
    }

    private String reviewStatusText(Integer status) {
        if (status == null) return "未知";
        return switch (status) {
            case Manuscript.REVIEW_STATUS_PENDING -> "待审核";
            case Manuscript.REVIEW_STATUS_APPROVED -> "审核通过";
            case Manuscript.REVIEW_STATUS_REJECTED -> "审核拒绝";
            default -> "审核状态: " + status;
        };
    }

    private String processStatusText(Integer status) {
        if (status == null) return "未知";
        return switch (status) {
            case Video.PROCESS_STATUS_PENDING -> "待处理";
            case Video.PROCESS_STATUS_TRANSCODING -> "视频转码中";
            case Video.PROCESS_STATUS_TRANSCODE_FAILED -> "转码失败";
            case Video.PROCESS_STATUS_TRANSCODE_SUCCESS -> "转码成功";
            case Video.PROCESS_STATUS_AUDIO_EXTRACTING -> "音频提取中";
            case Video.PROCESS_STATUS_AUDIO_FAILED -> "音频提取失败";
            case Video.PROCESS_STATUS_AUDIO_SUCCESS -> "音频提取成功";
            case Video.PROCESS_STATUS_SUBTITLE_GENERATING -> "字幕生成中";
            case Video.PROCESS_STATUS_SUBTITLE_FAILED -> "字幕生成失败";
            case Video.PROCESS_STATUS_SUBTITLE_SUCCESS -> "字幕生成成功";
            case Video.PROCESS_STATUS_AI_SUMMARIZING -> "AI摘要生成中";
            case Video.PROCESS_STATUS_AI_FAILED -> "AI摘要生成失败";
            case Video.PROCESS_STATUS_AI_SUCCESS -> "AI摘要生成成功";
            case Video.PROCESS_STATUS_COMPLETED -> "处理完成";
            default -> "处理状态: " + status;
        };
    }

    private boolean isFailedStatus(Integer status) {
        return Integer.valueOf(Video.PROCESS_STATUS_TRANSCODE_FAILED).equals(status)
                || Integer.valueOf(Video.PROCESS_STATUS_AUDIO_FAILED).equals(status)
                || Integer.valueOf(Video.PROCESS_STATUS_SUBTITLE_FAILED).equals(status)
                || Integer.valueOf(Video.PROCESS_STATUS_AI_FAILED).equals(status);
    }

    private String customerSessionStatusText(Integer status) {
        if (status == null) return "未知";
        return switch (status) {
            case 0 -> "AI服务中";
            case 1 -> "等待人工客服";
            case 2 -> "已处理";
            default -> "客服状态: " + status;
        };
    }
}
