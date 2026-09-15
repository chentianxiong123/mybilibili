class UserInfo {
  final int id;
  final String username;
  final String nickname;
  final String avatar;
  final String introduction;
  final int level;
  final int status;
  final int followerCount;
  final int followingCount;
  final int likeCount;
  final String createdAt;

  const UserInfo({
    this.id = 0,
    this.username = '',
    this.nickname = '',
    this.avatar = '',
    this.introduction = '',
    this.level = 0,
    this.status = 0,
    this.followerCount = 0,
    this.followingCount = 0,
    this.likeCount = 0,
    this.createdAt = '',
  });

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: (json['id'] as num?)?.toInt() ?? 0,
      username: json['username'] as String? ?? '',
      nickname: json['nickname'] as String? ?? '',
      avatar: json['avatar'] as String? ?? '',
      introduction: json['introduction'] as String? ?? '',
      level: (json['level'] as num?)?.toInt() ?? 0,
      status: (json['status'] as num?)?.toInt() ?? 0,
      followerCount: (json['follower_count'] as num?)?.toInt() ?? 0,
      followingCount: (json['following_count'] as num?)?.toInt() ?? 0,
      likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
      createdAt: json['created_at'] as String? ?? '',
    );
  }
}