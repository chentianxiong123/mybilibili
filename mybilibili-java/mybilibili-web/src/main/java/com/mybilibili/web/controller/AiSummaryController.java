package com.mybilibili.web.controller;

import com.mybilibili.common.entity.Video;
import com.mybilibili.common.vo.Result;
import com.mybilibili.web.mapper.VideoMapper;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;

/**
 * AI摘要流式输出控制器
 * 为前端提供视频摘要的流式获取接口
 */
@RestController
@RequestMapping("/ai-summary")
@Tag(name = "AI视频摘要", description = "视频AI摘要流式输出接口")
public class AiSummaryController {

    @Autowired
    private VideoMapper videoMapper;

    /**
     * 流式获取视频AI摘要（SSE）
     * 模拟实时生成效果，实际是从预生成的摘要文件中读取并分段发送
     */
    @GetMapping(value = "/stream/{videoId}", produces = MediaType.TEXT_EVENT_STREAM_VALUE)
    @Operation(summary = "流式获取视频AI摘要", description = "使用SSE流式输出视频AI摘要，模拟实时生成效果")
    public SseEmitter streamSummary(@PathVariable Integer videoId) {
        SseEmitter emitter = new SseEmitter(120000L); // 2分钟超时

        // 异步处理流式输出
        new Thread(() -> {
            try {
                // 1. 获取视频信息
                Video video = videoMapper.selectById(videoId);
                if (video == null) {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("视频不存在"));
                    emitter.complete();
                    return;
                }

                // 2. 检查是否有摘要
                if (video.getHasSummary() == null || video.getHasSummary() != 1) {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("该视频暂无AI摘要"));
                    emitter.complete();
                    return;
                }

                // 3. 发送开始事件
                emitter.send(SseEmitter.event()
                    .name("start")
                    .data("开始生成摘要..."));

                // 4. 读取摘要文件
                String summaryContent = readSummaryFile(videoId, video.getManuscriptId());
                if (summaryContent == null || summaryContent.isEmpty()) {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("摘要文件不存在或为空"));
                    emitter.complete();
                    return;
                }

                // 5. 模拟流式输出 - 按字符逐个发送
                // 使用随机延迟模拟AI思考过程
                int totalLength = summaryContent.length();
                int chunkSize = 5; // 每次发送5个字符
                int position = 0;

                // 发送总长度信息
                emitter.send(SseEmitter.event()
                    .name("meta")
                    .data("{\"totalLength\":" + totalLength + "}"));

                while (position < totalLength) {
                    int end = Math.min(position + chunkSize, totalLength);
                    String chunk = summaryContent.substring(position, end);

                    // 发送数据块 - 使用Base64编码避免换行符问题
                    String encodedChunk = java.util.Base64.getEncoder().encodeToString(chunk.getBytes("UTF-8"));
                    emitter.send(SseEmitter.event()
                        .name("data")
                        .data(encodedChunk));

                    position = end;

                    // 随机延迟 30-80ms，模拟打字效果
                    int delay = 25 + (int)(Math.random() * 40);
                    Thread.sleep(delay);

                    // 偶尔停顿一下，模拟AI思考
                    if (position % 60 == 0 && Math.random() > 0.6) {
                        Thread.sleep(80 + (int)(Math.random() * 100));
                    }
                }

                // 6. 发送完成事件
                emitter.send(SseEmitter.event()
                    .name("done")
                    .data("摘要生成完成"));

                emitter.complete();

            } catch (Exception e) {
                try {
                    emitter.send(SseEmitter.event()
                        .name("error")
                        .data("生成摘要失败: " + e.getMessage()));
                    emitter.complete();
                } catch (IOException ex) {
                    emitter.completeWithError(ex);
                }
            }
        }).start();

        return emitter;
    }

    /**
     * 直接获取视频AI摘要（非流式，用于快速获取完整摘要）
     */
    @GetMapping("/{videoId}")
    @Operation(summary = "获取视频AI摘要", description = "直接获取视频的完整AI摘要内容")
    public Result<String> getSummary(@PathVariable Integer videoId) {
        try {
            // 1. 获取视频信息
            Video video = videoMapper.selectById(videoId);
            if (video == null) {
                return Result.error("视频不存在");
            }

            // 2. 检查是否有摘要
            if (video.getHasSummary() == null || video.getHasSummary() != 1) {
                return Result.error("该视频暂无AI摘要");
            }

            // 3. 读取摘要文件
            String summaryContent = readSummaryFile(videoId, video.getManuscriptId());
            if (summaryContent == null || summaryContent.isEmpty()) {
                return Result.error("摘要文件不存在或为空");
            }

            return Result.success("获取成功", summaryContent);

        } catch (Exception e) {
            return Result.error("获取摘要失败: " + e.getMessage());
        }
    }

    /**
     * 检查视频是否有AI摘要
     */
    @GetMapping("/check/{videoId}")
    @Operation(summary = "检查视频是否有AI摘要", description = "快速检查视频是否已生成AI摘要")
    public Result<Boolean> checkSummary(@PathVariable Integer videoId) {
        try {
            Video video = videoMapper.selectById(videoId);
            if (video == null) {
                return Result.error("视频不存在");
            }

            boolean hasSummary = video.getHasSummary() != null && video.getHasSummary() == 1;
            return Result.success("检查成功", hasSummary);

        } catch (Exception e) {
            return Result.error("检查失败: " + e.getMessage());
        }
    }

    /**
     * 读取摘要文件内容
     */
    private String readSummaryFile(Integer videoId, Integer manuscriptId) {
        try {
            // 构建摘要文件路径 - 与UploadFilePathUtils保持一致
            // 路径: {projectRoot}/uploads/manuscripts/{manuscriptId}/videos/{videoId}/summary/ai-summary.txt
            // 注意：user.dir在web模块是mybilibili-java/mybilibili-web目录
            String basePath = System.getProperty("user.dir");
            File currentDir = new File(basePath);
            
            // 向上查找直到找到uploads目录或到达mybilibili目录
            String projectRoot = null;
            File tempDir = currentDir;
            while (tempDir != null) {
                // 检查当前目录下是否有uploads文件夹
                File uploadsDir = new File(tempDir, "uploads");
                if (uploadsDir.exists() && uploadsDir.isDirectory()) {
                    projectRoot = tempDir.getAbsolutePath();
                    break;
                }
                // 如果目录名是mybilibili，也停止
                if ("mybilibili".equals(tempDir.getName())) {
                    projectRoot = tempDir.getAbsolutePath();
                    break;
                }
                tempDir = tempDir.getParentFile();
            }
            
            // 如果没找到，使用默认路径
            if (projectRoot == null) {
                projectRoot = currentDir.getParentFile().getParent();
            }
            
            System.out.println("项目根目录: " + projectRoot);
            System.out.println("稿件ID: " + manuscriptId + ", 视频ID: " + videoId);
            
            String summaryPath = projectRoot + "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/summary/ai-summary.txt";

            File summaryFile = new File(summaryPath);
            System.out.println("尝试读取摘要文件: " + summaryFile.getAbsolutePath());
            
            if (!summaryFile.exists()) {
                // 尝试旧路径（兼容旧数据）
                summaryPath = projectRoot + "/uploads/manuscripts/" + manuscriptId + "/videos/" + videoId + "/ai_summary.txt";
                summaryFile = new File(summaryPath);
                System.out.println("尝试旧路径: " + summaryFile.getAbsolutePath());

                if (!summaryFile.exists()) {
                    // 尝试其他可能的路径
                    summaryPath = projectRoot + "/uploads/videos/" + videoId + "/ai_summary.txt";
                    summaryFile = new File(summaryPath);
                    System.out.println("尝试其他路径: " + summaryFile.getAbsolutePath());

                    if (!summaryFile.exists()) {
                        System.err.println("摘要文件不存在，最终尝试路径: " + summaryPath);
                        return null;
                    }
                }
            }

            System.out.println("成功读取摘要文件: " + summaryFile.getAbsolutePath());
            String content = new String(Files.readAllBytes(summaryFile.toPath()), "UTF-8");
            
            // 提取纯摘要内容（去除文件头）
            return extractSummaryContent(content);

        } catch (IOException e) {
            System.err.println("读取摘要文件失败: " + e.getMessage());
            return null;
        }
    }

    /**
     * 从文件内容中提取纯摘要内容（去除文件头）
     */
    private String extractSummaryContent(String fileContent) {
        if (fileContent == null || fileContent.isEmpty()) {
            return "";
        }
        
        // 查找第一个 "【视频摘要】" 或 "### 视频摘要" 或 "视频摘要" 的位置
        int startIndex = -1;
        String[] markers = {"【视频摘要】", "### 视频摘要", "视频摘要", "### 摘要"};
        
        for (String marker : markers) {
            startIndex = fileContent.indexOf(marker);
            if (startIndex != -1) {
                // 找到标记，从标记开始返回，保留原始格式
                String content = fileContent.substring(startIndex);
                System.out.println("找到标记 '" + marker + "'，提取内容长度: " + content.length());
                return content;
            }
        }
        
        // 如果没有找到标记，尝试查找第一个空行后的内容
        String[] lines = fileContent.split("\n");
        StringBuilder result = new StringBuilder();
        boolean foundEmptyLine = false;
        
        for (String line : lines) {
            // 跳过文件头（标题、生成时间等）
            if (line.startsWith("=") || line.startsWith("视频标题:") || line.startsWith("生成时间:")) {
                continue;
            }
            if (line.trim().isEmpty()) {
                foundEmptyLine = true;
                continue;
            }
            if (foundEmptyLine) {
                if (result.length() > 0) {
                    result.append("\n");
                }
                result.append(line);
            }
        }
        
        String extracted = result.toString();
        System.out.println("未找到标记，提取内容长度: " + extracted.length());
        return extracted.isEmpty() ? fileContent : extracted;
    }
}
