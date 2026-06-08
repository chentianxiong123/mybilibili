package com.mybilibili.operation.service.impl;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.mybilibili.common.exception.BusinessException;
import com.mybilibili.operation.entity.SupportTicket;
import com.mybilibili.operation.mapper.SupportTicketMapper;
import com.mybilibili.operation.service.SupportTicketService;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.List;
import java.util.concurrent.ThreadLocalRandom;

@Service
public class SupportTicketServiceImpl implements SupportTicketService {

    private final SupportTicketMapper supportTicketMapper;

    public SupportTicketServiceImpl(SupportTicketMapper supportTicketMapper) {
        this.supportTicketMapper = supportTicketMapper;
    }

    @Override
    public List<SupportTicket> listAll() {
        return supportTicketMapper.selectList(new LambdaQueryWrapper<SupportTicket>()
                .orderByDesc(SupportTicket::getCreatedAt));
    }

    @Override
    public List<SupportTicket> listByStatus(String status) {
        return supportTicketMapper.selectList(new LambdaQueryWrapper<SupportTicket>()
                .eq(SupportTicket::getStatus, status)
                .orderByDesc(SupportTicket::getCreatedAt));
    }

    @Override
    public SupportTicket getById(Long id) {
        return supportTicketMapper.selectById(id);
    }

    @Override
    @Transactional
    public SupportTicket create(SupportTicket ticket) {
        Date now = new Date();
        if (ticket.getTicketNo() == null || ticket.getTicketNo().isBlank()) {
            ticket.setTicketNo(generateTicketNo());
        }
        if (ticket.getSource() == null || ticket.getSource().isBlank()) {
            ticket.setSource(SOURCE_USER_FEEDBACK);
        }
        if (ticket.getCategory() == null || ticket.getCategory().isBlank()) {
            ticket.setCategory(CATEGORY_GENERAL);
        }
        if (ticket.getPriority() == null || ticket.getPriority().isBlank()) {
            ticket.setPriority(PRIORITY_NORMAL);
        }
        if (ticket.getStatus() == null || ticket.getStatus().isBlank()) {
            ticket.setStatus(STATUS_PENDING);
        }
        if (ticket.getTitle() == null || ticket.getTitle().isBlank()) {
            ticket.setTitle(defaultTitle(ticket.getContent()));
        }
        if (ticket.getContent() == null || ticket.getContent().isBlank()) {
            throw new BusinessException("工单内容不能为空");
        }
        ticket.setCreatedAt(now);
        ticket.setUpdatedAt(now);
        supportTicketMapper.insert(ticket);
        return ticket;
    }

    @Override
    public SupportTicket createFromCustomerSession(Long userId, Long sessionId, String title, String content, String entryReply) {
        if (sessionId != null) {
            SupportTicket existing = supportTicketMapper.selectOne(new LambdaQueryWrapper<SupportTicket>()
                    .eq(SupportTicket::getSessionId, sessionId)
                    .ne(SupportTicket::getStatus, STATUS_PROCESSED)
                    .last("LIMIT 1"));
            if (existing != null) {
                return existing;
            }
        }
        SupportTicket ticket = new SupportTicket();
        ticket.setUserId(userId);
        ticket.setSessionId(sessionId);
        ticket.setSource(SOURCE_AI_CUSTOMER_SERVICE);
        ticket.setCategory(CATEGORY_COMPLAINT);
        ticket.setPriority(PRIORITY_NORMAL);
        ticket.setTitle(title == null || title.isBlank() ? "客服转人工工单" : title);
        ticket.setContent(content == null || content.isBlank() ? "用户请求人工客服介入" : content);
        ticket.setEntryReply(entryReply);
        return create(ticket);
    }

    @Override
    @Transactional
    public SupportTicket process(Long id, Long adminId, String adminReply) {
        if (adminReply == null || adminReply.isBlank()) {
            throw new BusinessException("处理结果不能为空");
        }
        SupportTicket ticket = supportTicketMapper.selectById(id);
        if (ticket == null) {
            throw new BusinessException("工单不存在");
        }
        ticket.setStatus(STATUS_PROCESSED);
        ticket.setAdminReply(adminReply);
        ticket.setAssigneeAdminId(adminId);
        ticket.setProcessedAt(new Date());
        ticket.setUpdatedAt(new Date());
        supportTicketMapper.updateById(ticket);
        return ticket;
    }

    @Override
    @Transactional
    public SupportTicket processBySession(Long sessionId, Long adminId, String adminReply) {
        if (sessionId == null) {
            return null;
        }
        SupportTicket ticket = supportTicketMapper.selectOne(new LambdaQueryWrapper<SupportTicket>()
                .eq(SupportTicket::getSessionId, sessionId)
                .ne(SupportTicket::getStatus, STATUS_PROCESSED)
                .orderByDesc(SupportTicket::getCreatedAt)
                .last("LIMIT 1"));
        if (ticket == null) {
            return null;
        }
        return process(ticket.getId(), adminId, adminReply);
    }

    @Override
    public void delete(Long id) {
        supportTicketMapper.deleteById(id);
    }

    private String generateTicketNo() {
        String date = new SimpleDateFormat("yyyyMMddHHmmss").format(new Date());
        int suffix = ThreadLocalRandom.current().nextInt(1000, 10000);
        return "T" + date + suffix;
    }

    private String defaultTitle(String content) {
        if (content == null || content.isBlank()) {
            return "用户反馈工单";
        }
        String normalized = content.trim().replaceAll("\\s+", " ");
        return normalized.length() > 40 ? normalized.substring(0, 40) : normalized;
    }
}
