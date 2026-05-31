package com.mybilibili.video.controller;

import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import com.mybilibili.video.service.ManuscriptService;
import com.mybilibili.video.service.ManuscriptUploadSessionService;
import org.junit.jupiter.api.Test;
import org.mockito.ArgumentCaptor;
import org.springframework.mock.web.MockMultipartFile;
import org.springframework.mock.web.MockMultipartHttpServletRequest;
import org.springframework.test.util.ReflectionTestUtils;

import java.nio.charset.StandardCharsets;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

class ManuscriptControllerUploadTest {

    @Test
    void legacyUploadParsesIndexedVideosDynamically() throws Exception {
        ManuscriptService manuscriptService = mock(ManuscriptService.class);
        when(manuscriptService.uploadManuscript(any(ManuscriptUploadDTO.class), eq(42))).thenReturn(new ManuscriptVO());

        ManuscriptController controller = new ManuscriptController();
        ReflectionTestUtils.setField(controller, "manuscriptService", manuscriptService);
        ReflectionTestUtils.setField(controller, "uploadSessionService", mock(ManuscriptUploadSessionService.class));

        MockMultipartHttpServletRequest request = new MockMultipartHttpServletRequest();
        request.addHeader("X-User-Id", "42");
        request.addParameter("title", "动态分P稿件");
        request.addParameter("description", "desc");
        request.addParameter("categoryId", "3");
        request.addParameter("tags", "tag-a", "tag-a", "tag-b");
        request.addFile(new MockMultipartFile("cover", "cover.jpg", "image/jpeg", "cover".getBytes(StandardCharsets.UTF_8)));

        for (int i = 0; i < 6; i++) {
            request.addFile(new MockMultipartFile(
                    "videos[" + i + "].file",
                    "part-" + i + ".mp4",
                    "video/mp4",
                    ("video-" + i).getBytes(StandardCharsets.UTF_8)
            ));
            request.addParameter("videos[" + i + "].videoOrder", String.valueOf(5 - i));
            request.addParameter("videos[" + i + "].durationSeconds", String.valueOf(10 + i));
        }
        request.addParameter("videos[0].title", "第一个分P");

        Result<ManuscriptVO> result = controller.uploadManuscript(request);

        assertEquals(200, result.getCode());
        ArgumentCaptor<ManuscriptUploadDTO> dtoCaptor = ArgumentCaptor.forClass(ManuscriptUploadDTO.class);
        verify(manuscriptService).uploadManuscript(dtoCaptor.capture(), eq(42));

        ManuscriptUploadDTO dto = dtoCaptor.getValue();
        assertEquals("动态分P稿件", dto.getTitle());
        assertEquals(3, dto.getCategoryId());
        assertEquals(2, dto.getTags().size());
        assertEquals(6, dto.getVideos().size());
        assertEquals("part-5", dto.getVideos().get(0).getTitle());
        assertEquals(0, dto.getVideos().get(0).getVideoOrder());
        assertEquals("第一个分P", dto.getVideos().get(5).getTitle());
        assertNotNull(dto.getCover());
    }
}
