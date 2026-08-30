package com.mybilibili.user.feign;

import com.mybilibili.common.vo.ManuscriptVO;
import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;

@FeignClient(name = "mybilibili-video", path = "/manuscript")
public interface ManuscriptClient {

    @GetMapping("/{id}")
    Result<ManuscriptVO> getManuscriptById(@PathVariable("id") Integer id);
}
