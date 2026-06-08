package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.SupportTicket;
import com.mybilibili.ai.service.SupportTicketService;
import com.mybilibili.common.vo.Result;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/tickets")
public class SupportTicketAdminController {

    @Autowired
    private SupportTicketService supportTicketService;

    @GetMapping
    public Result<List<SupportTicket>> list(@RequestParam(required = false) String status) {
        if (status != null && !status.isEmpty()) {
            return Result.success(supportTicketService.listByStatus(status));
        }
        return Result.success(supportTicketService.listAll());
    }

    @GetMapping("/{id}")
    public Result<SupportTicket> get(@PathVariable Long id) {
        SupportTicket ticket = supportTicketService.getById(id);
        return ticket != null ? Result.success(ticket) : Result.error("工单不存在");
    }

    @PutMapping("/{id}/process")
    public Result<SupportTicket> process(@PathVariable Long id,
                                         @RequestBody Map<String, String> body,
                                         @RequestHeader(value = "X-Admin-Id", required = false) Long adminId) {
        String adminReply = body.get("adminReply");
        return Result.success(supportTicketService.process(id, adminId, adminReply));
    }

    @DeleteMapping("/{id}")
    public Result<Void> delete(@PathVariable Long id) {
        supportTicketService.delete(id);
        return Result.success();
    }
}
