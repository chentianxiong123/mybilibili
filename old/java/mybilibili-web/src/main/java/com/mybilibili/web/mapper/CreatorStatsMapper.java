package com.mybilibili.web.mapper;

import com.mybilibili.common.vo.LatestCommentVO;
import com.mybilibili.common.vo.RankingVO;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;

import java.util.List;

/**
 * 创作者统计Mapper
 */
@Mapper
public interface CreatorStatsMapper {

    /**
     * 获取用户粉丝总数
     *
     * @param userId 用户ID
     * @return 粉丝总数
     */
    Integer countFollowers(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总播放量
     *
     * @param userId 用户ID
     * @return 总播放量
     */
    Integer sumViewCount(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总评论数
     *
     * @param userId 用户ID
     * @return 总评论数
     */
    Integer sumCommentCount(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总点赞数
     *
     * @param userId 用户ID
     * @return 总点赞数
     */
    Integer sumLikeCount(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总分享数
     *
     * @param userId 用户ID
     * @return 总分享数
     */
    Integer sumShareCount(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总收藏数
     *
     * @param userId 用户ID
     * @return 总收藏数
     */
    Integer sumCollectCount(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总投币数
     *
     * @param userId 用户ID
     * @return 总投币数
     */
    Integer sumCoinCount(@Param("userId") Integer userId);

    /**
     * 获取用户稿件总数
     *
     * @param userId 用户ID
     * @return 稿件总数
     */
    Integer countManuscripts(@Param("userId") Integer userId);

    /**
     * 获取用户稿件的最新评论
     *
     * @param userId 用户ID
     * @param limit  限制数量
     * @return 最新评论列表
     */
    List<LatestCommentVO> selectLatestComments(@Param("userId") Integer userId, @Param("limit") Integer limit);

    /**
     * 获取用户稿件观看排行榜
     *
     * @param userId 用户ID
     * @param limit  限制数量
     * @return 观看排行榜
     */
    List<RankingVO> selectViewRankings(@Param("userId") Integer userId, @Param("limit") Integer limit);

    /**
     * 获取用户稿件互动排行榜
     *
     * @param userId 用户ID
     * @param limit  限制数量
     * @return 互动排行榜
     */
    List<RankingVO> selectInteractionRankings(@Param("userId") Integer userId, @Param("limit") Integer limit);
}
