package com.mybilibili.operation.service;

import com.mybilibili.operation.entity.SupportTicket;

import java.util.List;

public interface SupportTicketService {
    String STATUS_PENDING = "PENDING";
    String STATUS_PROCESSED = "PROCESSED";

    String SOURCE_USER_FEEDBACK = "USER_FEEDBACK";
    String SOURCE_AI_CUSTOMER_SERVICE = "AI_CUSTOMER_SERVICE";

    String CATEGORY_GENERAL = "GENERAL";
    String CATEGORY_COMPLAINT = "COMPLAINT";

    String PRIORITY_NORMAL = "NORMAL";

    List<SupportTicket> listAll();

    List<SupportTicket> listByStatus(String status);

    SupportTicket getById(Long id);

    SupportTicket create(SupportTicket ticket);

    SupportTicket createFromCustomerSession(Long userId, Long sessionId, String title, String content, String entryReply);

    SupportTicket process(Long id, Long adminId, String adminReply);

    SupportTicket processBySession(Long sessionId, Long adminId, String adminReply);

    void delete(Long id);
}
