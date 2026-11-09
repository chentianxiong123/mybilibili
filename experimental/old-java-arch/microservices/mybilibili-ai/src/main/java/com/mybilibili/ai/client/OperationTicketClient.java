package com.mybilibili.ai.client;

import com.mybilibili.common.vo.Result;
import org.springframework.cloud.openfeign.FeignClient;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;

import java.util.Map;

@FeignClient(name = "mybilibili-search-recommend", contextId = "operationTicketClient", path = "/operation/internal/tickets")
public interface OperationTicketClient {

    @PostMapping("/customer-session")
    Result<Void> createFromCustomerSession(@RequestBody Map<String, Object> request);

    @PutMapping("/session/{sessionId}/process")
    Result<Void> processBySession(@PathVariable("sessionId") Long sessionId,
                                  @RequestBody Map<String, Object> request);
}
