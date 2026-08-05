package com.mybilibili.video.service;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.dto.ManuscriptUpdateDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.ManuscriptVO;

import java.util.List;
import java.util.Map;

public interface ManuscriptService {
    ManuscriptVO uploadManuscript(ManuscriptUploadDTO dto, Integer userId) throws Exception;
    ManuscriptVO getManuscriptById(Integer id);
    ManuscriptVO getManuscriptWithVideos(Integer id, Integer currentUserId);
    List<ManuscriptVO> getManuscriptsByUserId(Integer userId);
    List<ManuscriptVO> getManuscriptsByUserIdWithPaging(Integer userId, Integer status, Integer page, Integer size);
    Integer countManuscriptsByUserIdAndStatus(Integer userId, Integer status);
    Map<String, Integer> getManuscriptStatsByUserId(Integer userId);
    List<ManuscriptVO> searchUserManuscripts(Integer userId, String keyword, String sort);
    List<ManuscriptVO> getRecommendedManuscripts(Integer userId);
    Map<String, Object> getManuscriptsByCategoryId(Integer categoryId, Integer page, Integer size);
    List<ManuscriptVO> getHotManuscripts(Integer userId);
    ManuscriptVO updateManuscript(Integer id, Manuscript manuscript);
    ManuscriptVO updateManuscriptByOwner(Integer id, Integer userId, ManuscriptUpdateDTO dto) throws Exception;
    void deleteManuscript(Integer id, Integer userId);
    int fixAllManuscriptDurations();

    List<ManuscriptVO> getPendingManuscripts();
    List<ManuscriptVO> getProcessingManuscripts();
    List<ManuscriptVO> getAllManuscripts();
    Map<String, Object> getManuscriptDetail(Integer manuscriptId);
    boolean approveManuscript(Integer manuscriptId, Integer reviewerId, String reason);
    boolean approveManuscriptWithProcess(Integer manuscriptId, Integer reviewerId, String reason, boolean autoProcess);
    boolean rejectManuscript(Integer manuscriptId, Integer reviewerId, String reason);
    boolean publishManuscript(Integer manuscriptId);
    boolean unpublishManuscript(Integer manuscriptId);

    /**
     * 违规下架(举报处理后,管理员人工下架)
     * 区别于普通 unpublishManuscript:还更新 reviewStatus=REJECTED + reviewReason
     * 并发 ES DELETE 事件
     */
    boolean takeDownViolatingManuscript(Integer manuscriptId, String reason);
    boolean publishManuscriptByOwner(Integer manuscriptId, Integer userId);
    boolean unpublishManuscriptByOwner(Integer manuscriptId, Integer userId);
    List<ManuscriptVO.VideoItemVO> getManuscriptVideos(Integer manuscriptId);
    Map<String, Object> getManuscriptStatistics();
    boolean retryManuscript(Integer manuscriptId);

    boolean manualProcessAll(Integer videoId);
    boolean resetVideoStatus(Integer videoId);
    Map<String, Object> getVideoSourceUrl(Integer videoId);
    Video getVideoById(Integer videoId);
    void incrementViewCount(Integer manuscriptId);
    void incrementViewCount(Integer manuscriptId, String viewerKey);

    /**
     * 处理完成时自动上架:检查稿件下所有 video 是否全部 COMPLETED,
     * 若是则改 status=PUBLISHED + 发 ES 索引事件。
     * 由 ManuscriptAutoPublishConsumer 调用。
     *
     * @return true 表示已上架,false 表示还有 video 未完成
     */
    boolean autoPublishIfAllVideosCompleted(Integer manuscriptId);
    void updateCommentCount(Integer manuscriptId, Integer count);
    void incrementCommentCount(Integer manuscriptId);
    void decrementCommentCount(Integer manuscriptId);
}
