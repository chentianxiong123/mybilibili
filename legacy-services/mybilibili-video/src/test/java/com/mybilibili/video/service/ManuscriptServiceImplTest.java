package com.mybilibili.video.service;

import com.mybilibili.common.dto.ManuscriptUpdateDTO;
import com.mybilibili.common.entity.Manuscript;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.storage.StorageService;
import com.mybilibili.mq.VideoMQProducer;
import com.mybilibili.common.storage.StorageKeys;
import com.mybilibili.video.feign.MessageClient;
import com.mybilibili.video.feign.UserClient;
import com.mybilibili.video.entity.ManuscriptEditVersion;
import com.mybilibili.video.mapper.CategoryMapper;
import com.mybilibili.video.mapper.ManuscriptEditVersionMapper;
import com.mybilibili.video.mapper.ManuscriptMapper;
import com.mybilibili.video.mapper.TagMapper;
import com.mybilibili.video.mapper.VideoMapper;
import com.mybilibili.video.service.impl.ManuscriptServiceImpl;
import org.junit.jupiter.api.Test;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.test.util.ReflectionTestUtils;
import org.springframework.mock.web.MockMultipartFile;

import java.util.Date;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.anyLong;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.never;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

class ManuscriptServiceImplTest {

    @Test
    void ownerUpdateResetsReviewAndPreservesStatistics() throws Exception {
        ManuscriptMapper manuscriptMapper = mock(ManuscriptMapper.class);
        VideoMapper videoMapper = mock(VideoMapper.class);
        Manuscript existing = manuscript(10, 42);
        when(manuscriptMapper.selectById(10)).thenReturn(existing);
        when(videoMapper.selectByManuscriptId(10)).thenReturn(List.of());

        ManuscriptServiceImpl service = newService(manuscriptMapper, videoMapper);
        ManuscriptUpdateDTO dto = new ManuscriptUpdateDTO();
        dto.setTitle(" 新标题 ");
        dto.setDescription("新简介");
        dto.setCategoryId(7);
        dto.setTags(List.of("tag-a", "tag-b"));

        service.updateManuscriptByOwner(10, 42, dto);

        assertEquals("新标题", existing.getTitle());
        assertEquals("新简介", existing.getDescription());
        assertEquals(7, existing.getCategoryId());
        assertEquals(Manuscript.STATUS_PENDING_REVIEW, existing.getStatus());
        assertEquals(Manuscript.REVIEW_STATUS_PENDING, existing.getReviewStatus());
        assertEquals(123, existing.getViewCount());
        verify(manuscriptMapper).updateById(existing);
        verify(getVersionMapper(service)).insert(any(ManuscriptEditVersion.class));
    }

    @Test
    void ownerUpdateCanReplaceVideoAndReorderParts() throws Exception {
        ManuscriptMapper manuscriptMapper = mock(ManuscriptMapper.class);
        VideoMapper videoMapper = mock(VideoMapper.class);
        StorageService storageService = mock(StorageService.class);
        Manuscript existing = manuscript(10, 42);
        when(manuscriptMapper.selectById(10)).thenReturn(existing);

        com.mybilibili.common.entity.Video video1 = new com.mybilibili.common.entity.Video();
        video1.setId(100);
        video1.setManuscriptId(10);
        video1.setVideoOrder(0);
        video1.setTitle("P1");
        video1.setSourceVideoUrl("/old/1.mp4");
        video1.setPlayUrlHd("/old/1.mp4");

        com.mybilibili.common.entity.Video video2 = new com.mybilibili.common.entity.Video();
        video2.setId(101);
        video2.setManuscriptId(10);
        video2.setVideoOrder(1);
        video2.setTitle("P2");
        video2.setSourceVideoUrl("http://127.0.0.1:9000/mybilibili/manuscripts/10/videos/101/source/video.mov");
        video2.setPlayUrlHd("/old/2.mp4");

        when(videoMapper.selectByManuscriptId(10)).thenReturn(List.of(video1, video2));
        when(storageService.upload(eq(StorageKeys.videoSource(10, 101, ".mp4")), any(), anyLong(), any())).thenReturn("/new/2.mp4");

        ManuscriptServiceImpl service = newService(manuscriptMapper, videoMapper, storageService);
        ManuscriptUpdateDTO dto = new ManuscriptUpdateDTO();
        dto.setTitle("新标题");
        dto.setCategoryId(7);
        dto.setVideos(List.of(videoUpdate(101, "P2-改", 0, new MockMultipartFile("file", "video.mp4", "video/mp4", "data".getBytes()))));

        service.updateManuscriptByOwner(10, 42, dto);

        assertEquals("P2-改", video2.getTitle());
        assertEquals(0, video2.getVideoOrder());
        assertEquals("/new/2.mp4", video2.getSourceVideoUrl());
        assertEquals(0, video2.getProcessProgress());
        assertEquals(com.mybilibili.common.entity.Video.PROCESS_STATUS_PENDING, video2.getProcessStatus());
        verify(storageService).upload(eq(StorageKeys.videoSource(10, 101, ".mp4")), any(), anyLong(), any());
        verify(storageService).delete("manuscripts/10/videos/101/source/video.mov");
        verify(storageService).deletePrefix("manuscripts/10/videos/101/transcoded/");
        verify(storageService).deletePrefix("manuscripts/10/videos/101/audio/");
        verify(storageService).deletePrefix("manuscripts/10/videos/101/subtitles/");
        verify(storageService).deletePrefix("manuscripts/10/videos/101/summary/");
        verify(videoMapper).updateById(video2);
        verify(getVersionMapper(service)).insert(any(ManuscriptEditVersion.class));
    }

    @Test
    void ownerUpdateRejectsOtherUser() {
        ManuscriptMapper manuscriptMapper = mock(ManuscriptMapper.class);
        Manuscript existing = manuscript(10, 42);
        when(manuscriptMapper.selectById(10)).thenReturn(existing);

        ManuscriptServiceImpl service = newService(manuscriptMapper, mock(VideoMapper.class));
        ManuscriptUpdateDTO dto = new ManuscriptUpdateDTO();
        dto.setTitle("新标题");
        dto.setCategoryId(7);

        assertThrows(BusinessException.class, () -> service.updateManuscriptByOwner(10, 99, dto));
        verify(manuscriptMapper, never()).updateById(any(Manuscript.class));
    }

    @Test
    void internalUpdateMergesPartialEntityWithoutNullingExistingFields() {
        ManuscriptMapper manuscriptMapper = mock(ManuscriptMapper.class);
        VideoMapper videoMapper = mock(VideoMapper.class);
        Manuscript existing = manuscript(10, 42);
        when(manuscriptMapper.selectById(10)).thenReturn(existing);
        when(videoMapper.selectByManuscriptId(10)).thenReturn(List.of());

        ManuscriptServiceImpl service = newService(manuscriptMapper, videoMapper);
        Manuscript patch = new Manuscript();
        patch.setStatus(Manuscript.STATUS_UNPUBLISHED);
        patch.setReviewStatus(Manuscript.REVIEW_STATUS_REJECTED);

        service.updateManuscript(10, patch);

        assertEquals("原标题", existing.getTitle());
        assertEquals("/cover.jpg", existing.getCoverUrl());
        assertEquals(Manuscript.STATUS_UNPUBLISHED, existing.getStatus());
        assertEquals(Manuscript.REVIEW_STATUS_REJECTED, existing.getReviewStatus());
        verify(manuscriptMapper).updateById(existing);
    }

    private ManuscriptServiceImpl newService(ManuscriptMapper manuscriptMapper, VideoMapper videoMapper) {
        return newService(manuscriptMapper, videoMapper, mock(StorageService.class));
    }

    private ManuscriptServiceImpl newService(ManuscriptMapper manuscriptMapper, VideoMapper videoMapper, StorageService storageService) {
        ManuscriptServiceImpl service = new ManuscriptServiceImpl();
        ReflectionTestUtils.setField(service, "manuscriptMapper", manuscriptMapper);
        ReflectionTestUtils.setField(service, "manuscriptEditVersionMapper", mock(ManuscriptEditVersionMapper.class));
        ReflectionTestUtils.setField(service, "videoMapper", videoMapper);
        ReflectionTestUtils.setField(service, "categoryMapper", mock(CategoryMapper.class));
        ReflectionTestUtils.setField(service, "tagMapper", mock(TagMapper.class));
        ReflectionTestUtils.setField(service, "userClient", mock(UserClient.class));
        ReflectionTestUtils.setField(service, "storageService", storageService);
        ReflectionTestUtils.setField(service, "videoMQProducer", mock(VideoMQProducer.class));
        ReflectionTestUtils.setField(service, "messageClient", mock(MessageClient.class));
        ReflectionTestUtils.setField(service, "videoProcessPort", mock(VideoProcessPort.class));
        ReflectionTestUtils.setField(service, "redisTemplate", mock(StringRedisTemplate.class));
        return service;
    }

    private ManuscriptEditVersionMapper getVersionMapper(ManuscriptServiceImpl service) {
        return (ManuscriptEditVersionMapper) ReflectionTestUtils.getField(service, "manuscriptEditVersionMapper");
    }

    private ManuscriptUpdateDTO.VideoUpdateDTO videoUpdate(Integer id, String title, Integer order, MockMultipartFile file) {
        ManuscriptUpdateDTO.VideoUpdateDTO dto = new ManuscriptUpdateDTO.VideoUpdateDTO();
        dto.setId(id);
        dto.setTitle(title);
        dto.setVideoOrder(order);
        dto.setDurationSeconds(10);
        dto.setFile(file);
        return dto;
    }

    private Manuscript manuscript(Integer id, Integer userId) {
        Manuscript manuscript = new Manuscript();
        manuscript.setId(id);
        manuscript.setUserId(userId);
        manuscript.setTitle("原标题");
        manuscript.setDescription("原简介");
        manuscript.setCoverUrl("/cover.jpg");
        manuscript.setCategoryId(3);
        manuscript.setViewCount(123);
        manuscript.setStatus(Manuscript.STATUS_PUBLISHED);
        manuscript.setReviewStatus(Manuscript.REVIEW_STATUS_APPROVED);
        manuscript.setUploadTime(new Date());
        return manuscript;
    }
}
