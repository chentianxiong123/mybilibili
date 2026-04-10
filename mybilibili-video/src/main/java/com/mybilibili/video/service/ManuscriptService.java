package com.mybilibili.video.service;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
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
    List<ManuscriptVO> getRecommendedManuscripts(Integer userId);
    Map<String, Object> getManuscriptsByCategoryId(Integer categoryId, Integer page, Integer size);
    List<ManuscriptVO> getHotManuscripts(Integer userId);
    ManuscriptVO updateManuscript(Integer id, Manuscript manuscript);
    void deleteManuscript(Integer id, Integer userId);
    int fixAllManuscriptDurations();

    List<ManuscriptVO> getPendingManuscripts();
    List<ManuscriptVO> getProcessingManuscripts();
    List<ManuscriptVO> getReadyManuscripts();
    List<ManuscriptVO> getAllManuscripts();
    Map<String, Object> getManuscriptDetail(Integer manuscriptId);
    boolean approveManuscript(Integer manuscriptId, Integer reviewerId, String reason);
    boolean rejectManuscript(Integer manuscriptId, Integer reviewerId, String reason);
    boolean publishManuscript(Integer manuscriptId);
    boolean unpublishManuscript(Integer manuscriptId);
    List<ManuscriptVO.VideoItemVO> getManuscriptVideos(Integer manuscriptId);
    Map<String, Object> getManuscriptStatistics();
    boolean retryManuscript(Integer manuscriptId);

    boolean manualTranscode(Integer videoId);
    boolean manualExtractAudio(Integer videoId);
    boolean manualGenerateSubtitle(Integer videoId);
    boolean manualAiSummary(Integer videoId);
    boolean manualProcessAll(Integer videoId);
    boolean resetVideoStatus(Integer videoId);
    Map<String, Object> getVideoProcessStatus(Integer videoId);
    Map<String, Object> getVideoSourceUrl(Integer videoId);
    Video getVideoById(Integer videoId);
}
