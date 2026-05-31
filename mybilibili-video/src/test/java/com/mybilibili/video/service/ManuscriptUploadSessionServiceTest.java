package com.mybilibili.video.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.mybilibili.common.dto.ManuscriptUploadDTO;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.common.utils.UploadFilePathUtils;
import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.video.dto.CreateUploadSessionRequest;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.springframework.mock.web.MockMultipartFile;

import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.concurrent.atomic.AtomicReference;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.ArgumentMatchers.eq;
import static org.mockito.Mockito.mock;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

class ManuscriptUploadSessionServiceTest {

    @TempDir
    Path tempDir;

    @Test
    void assemblesChunksAndSubmitsThroughManuscriptService() throws Exception {
        ManuscriptService manuscriptService = mock(ManuscriptService.class);
        AtomicReference<ManuscriptUploadDTO> capturedDto = new AtomicReference<>();
        AtomicReference<byte[]> capturedVideoBytes = new AtomicReference<>();
        when(manuscriptService.uploadManuscript(any(ManuscriptUploadDTO.class), eq(7))).thenAnswer(invocation -> {
            ManuscriptUploadDTO dto = invocation.getArgument(0);
            capturedDto.set(dto);
            capturedVideoBytes.set(dto.getVideos().get(0).getVideo().getBytes());
            return new ManuscriptVO();
        });
        ManuscriptUploadSessionService service = newService(manuscriptService);

        ManuscriptUploadSessionService.UploadSessionInfo session = service.createSession(7, uploadRequest());
        String uploadId = session.uploadId();

        service.uploadChunk(7, uploadId, 0, 0, 2, chunk("hello "));
        ManuscriptUploadSessionService.UploadChunkStatus finalChunk =
                service.uploadChunk(7, uploadId, 0, 1, 2, chunk("world"));

        assertTrue(finalChunk.complete());
        ManuscriptUploadSessionService.UploadSessionStatus status = service.getSessionStatus(7, uploadId);
        assertEquals(1, status.uploadedParts());
        assertEquals(List.of(0, 1), status.parts().get(0).uploadedChunks());

        service.completeSession(7, uploadId, new MockMultipartFile(
                "cover",
                "cover.jpg",
                "image/jpeg",
                "cover".getBytes(StandardCharsets.UTF_8)
        ));

        verify(manuscriptService).uploadManuscript(any(ManuscriptUploadDTO.class), eq(7));
        ManuscriptUploadDTO dto = capturedDto.get();
        assertNotNull(dto);
        assertEquals("分片稿件", dto.getTitle());
        assertEquals(List.of("tag-a"), dto.getTags());
        assertEquals(1, dto.getVideos().size());
        assertEquals("video.mp4", dto.getVideos().get(0).getVideo().getOriginalFilename());
        assertArrayEquals("hello world".getBytes(StandardCharsets.UTF_8), capturedVideoBytes.get());
        assertFalse(Files.exists(tempDir.resolve("manuscript-upload-sessions").resolve(uploadId)));
    }

    @Test
    void rejectsInvalidOwnerAndUnsafeUploadId() {
        ManuscriptUploadSessionService service = newService(mock(ManuscriptService.class));
        String uploadId = service.createSession(1, uploadRequest()).uploadId();

        assertThrows(BusinessException.class, () -> service.uploadChunk(2, uploadId, 0, 0, 2, chunk("x")));
        assertThrows(BusinessException.class, () -> service.getSessionStatus(1, "../outside"));
    }

    @Test
    void rejectsOutOfRangeChunkIndex() {
        ManuscriptUploadSessionService service = newService(mock(ManuscriptService.class));
        String uploadId = service.createSession(1, uploadRequest()).uploadId();

        assertThrows(BusinessException.class, () -> service.uploadChunk(1, uploadId, 0, 2, 2, chunk("x")));
    }

    private ManuscriptUploadSessionService newService(ManuscriptService manuscriptService) {
        return new ManuscriptUploadSessionService(
                manuscriptService,
                new UploadFilePathUtils(tempDir.toString()),
                new ObjectMapper()
        );
    }

    private CreateUploadSessionRequest uploadRequest() {
        CreateUploadSessionRequest request = new CreateUploadSessionRequest();
        request.setTitle("分片稿件");
        request.setDescription("desc");
        request.setCategoryId(3);
        request.setTags(List.of("tag-a"));

        CreateUploadSessionRequest.VideoPart part = new CreateUploadSessionRequest.VideoPart();
        part.setTitle("P1");
        part.setFileName("video.mp4");
        part.setSize(11L);
        part.setVideoOrder(0);
        part.setDurationSeconds(12);
        part.setTotalChunks(2);
        request.setVideos(List.of(part));
        return request;
    }

    private MockMultipartFile chunk(String content) {
        return new MockMultipartFile(
                "file",
                "chunk",
                "application/octet-stream",
                content.getBytes(StandardCharsets.UTF_8)
        );
    }
}
