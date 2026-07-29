package com.mybilibili.operation.controller;

import com.mybilibili.common.vo.Result;
import com.mybilibili.operation.service.SupportTicketService;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/operation/internal/tickets")
public class SupportTicketInternalController {

    private final SupportTicketService supportTicketService;

    public SupportTicketInternalController(SupportTicketService supportTicketService) {
        this.supportTicketService = supportTicketService;
    }

    @PostMapping("/customer-session")
    public Result<Void> createFromCustomerSession(@RequestBody Map<String, Object> request) {
        supportTicketService.createFromCustomerSession(
                asLong(request.get("userId")),
                asLong(request.get("sessionId")),
                asString(request.get("title")),
                asString(request.get("content")),
                asString(request.get("entryReply"))
        );
        return Result.success();
    }

    @PutMapping("/session/{sessionId}/process")
    public Result<Void> processBySession(@PathVariable Long sessionId,
                                         @RequestBody Map<String, Object> request) {
        supportTicketService.processBySession(
                sessionId,
                asLong(request.get("adminId")),
                asString(request.get("adminReply"))
        );
        return Result.success();
    }

    private Long asLong(Object value) {
        if (value == null) {
            return null;
        }
        if (value instanceof Number number) {
            return number.longValue();
        }
        String text = value.toString().trim();
        return text.isEmpty() ? null : Long.parseLong(text);
    }

    private String asString(Object value) {
        if (value == null) {
            return null;
        }
        String text = value.toString().trim();
        return text.isEmpty() ? null : text;
    }
}
