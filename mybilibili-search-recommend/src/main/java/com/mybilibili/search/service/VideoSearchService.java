package com.mybilibili.search.service;

import com.mybilibili.common.vo.VideoSearchVO;
import org.springframework.data.domain.Page;

import java.util.List;

public interface VideoSearchService {

    Page<VideoSearchVO> search(String keyword, Integer categoryId, String tag, Integer userId,
                               String sort, int page, int size);

    List<String> suggest(String keyword);
}