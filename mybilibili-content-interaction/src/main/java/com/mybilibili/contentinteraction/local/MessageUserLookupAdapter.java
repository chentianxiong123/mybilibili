package com.mybilibili.contentinteraction.local;

import com.mybilibili.common.entity.User;
import com.mybilibili.common.vo.Result;
import com.mybilibili.common.vo.UserVO;
import com.mybilibili.interaction.mapper.UserMapper;
import com.mybilibili.message.service.UserLookupPort;
import org.springframework.stereotype.Component;

@Component
public class MessageUserLookupAdapter implements UserLookupPort {

    private final UserMapper userMapper;

    public MessageUserLookupAdapter(UserMapper userMapper) {
        this.userMapper = userMapper;
    }

    @Override
    public Result<UserVO> getUserById(Integer id) {
        User user = userMapper.findById(id);
        if (user == null) {
            return Result.error("用户不存在");
        }
        UserVO vo = new UserVO();
        vo.setId(user.getId());
        vo.setUsername(user.getUsername());
        vo.setNickname(user.getNickname());
        vo.setAvatar(user.getAvatar());
        vo.setLevel(user.getLevel());
        vo.setStatus(user.getStatus());
        return Result.success("获取成功", vo);
    }
}
