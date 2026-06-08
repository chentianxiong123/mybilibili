package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.SupportTicket;
import com.mybilibili.ai.service.SupportTicketService;
import com.mybilibili.common.vo.Result;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/ai/tickets")
public class SupportTicketController {

    private final SupportTicketService supportTicketService;

    public SupportTicketController(SupportTicketService supportTicketService) {
        this.supportTicketService = supportTicketService;
    }

    @PostMapping
    public Result<SupportTicket> create(@RequestBody Map<String, Object> request,
                                        @RequestHeader(value = "X-User-Id", required = false) Long userId) {
        SupportTicket ticket = new SupportTicket();
        ticket.setUserId(userId);
        ticket.setSource(asString(request.get("source"), SupportTicketService.SOURCE_USER_FEEDBACK));
        ticket.setCategory(asString(request.get("category"), SupportTicketService.CATEGORY_GENERAL));
        ticket.setPriority(asString(request.get("priority"), SupportTicketService.PRIORITY_NORMAL));
        ticket.setTitle(asString(request.get("title"), null));
        ticket.setContent(asString(request.get("content"), null));
        return Result.success(supportTicketService.create(ticket));
    }

    private String asString(Object value, String defaultValue) {
        if (value == null) {
            return defaultValue;
        }
        String text = value.toString().trim();
        return text.isEmpty() ? defaultValue : text;
    }
}
