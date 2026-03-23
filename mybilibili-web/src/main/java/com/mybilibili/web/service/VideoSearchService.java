package com.mybilibili.web.service;

import com.mybilibili.common.vo.VideoSearchVO;
import org.springframework.data.domain.Page;

import java.util.List;

/**
 * 视频搜索服务接口
 */
public interface VideoSearchService {

    /**
     * 搜索视频
     *
     * @param keyword    搜索关键词（匹配title和description）
     * @param categoryId 分类ID（可选）
     * @param tag        标签（可选）
     * @param userId     UP主ID（可选）
     * @param sort       排序方式：relevance（默认）/ time / hot
     * @param page       页码（从0开始）
     * @param size       每页大小
     * @return 分页搜索结果，包含高亮
     */
    Page<VideoSearchVO> search(String keyword, Integer categoryId, String tag, Integer userId,
                               String sort, int page, int size);

    /**
     * 获取搜索建议
     *
     * @param keyword 搜索关键词
     * @return 搜索建议列表
     */
    List<String> suggest(String keyword);
}
