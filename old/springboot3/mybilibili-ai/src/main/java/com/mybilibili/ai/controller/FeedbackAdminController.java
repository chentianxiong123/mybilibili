package com.mybilibili.ai.controller;

import com.mybilibili.ai.entity.AiFeedback;
import com.mybilibili.ai.service.AiFeedbackService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/ai/admin/feedback")
public class FeedbackAdminController {

    @Autowired
    private AiFeedbackService aiFeedbackService;

    @GetMapping
    public List<AiFeedback> list(@RequestParam(required = false) String status) {
        if (status != null && !status.isEmpty()) {
            return aiFeedbackService.listByStatus(status);
        }
        return aiFeedbackService.listAll();
    }

    @GetMapping("/{id}")
    public AiFeedback get(@PathVariable Long id) {
        return aiFeedbackService.getById(id);
    }

    @PutMapping("/{id}/process")
    public AiFeedback process(@PathVariable Long id, @RequestBody Map<String, String> body) {
        String adminReply = body.get("adminReply");
        aiFeedbackService.process(id, adminReply);
        return aiFeedbackService.getById(id);
    }

    @DeleteMapping("/{id}")
    public void delete(@PathVariable Long id) {
        aiFeedbackService.delete(id);
    }
}