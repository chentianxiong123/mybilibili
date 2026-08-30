package com.mybilibili.common.vo;

import lombok.Data;

/**
 * 粉丝信息VO
 */
@Data
public class FanVO {
    /**
     * 粉丝用户ID
     */
    private Integer id;

    /**
     * 粉丝用户名
     */
    private String username;

    /**
     * 粉丝昵称
     */
    private String nickname;

    /**
     * 粉丝头像
     */
    private String avatar;

    /**
     * 粉丝等级
     */
    private Integer level;

    /**
     * 粉丝签名
     */
    private String signature;

    /**
     * 是否互关
     */
    private Boolean isMutual;

    /**
     * 关注时间
     */
    private String followedAt;
}
