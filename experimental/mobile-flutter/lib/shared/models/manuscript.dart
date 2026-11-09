import 'user.dart';

class VideoItem {
  final int id;
  final String title;
  final String description;
  final String playUrl;
  final String playUrlHd;
  final String playUrlSd;
  final String playUrlLd;
  final String duration;
  final int durationSeconds;
  final int videoOrder;
  final int status;
  final int processStatus;
  final int processProgress;

  const VideoItem({
    this.id = 0,
    this.title = '',
    this.description = '',
    this.playUrl = '',
    this.playUrlHd = '',
    this.playUrlSd = '',
    this.playUrlLd = '',
    this.duration = '',
    this.durationSeconds = 0,
    this.videoOrder = 0,
    this.status = 0,
    this.processStatus = 0,
    this.processProgress = 0,
  });

  factory VideoItem.fromJson(Map<String, dynamic> json) {
    return VideoItem(
      id: (json['id'] as num?)?.toInt() ?? 0,
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      playUrl: json['play_url'] as String? ?? '',
      playUrlHd: json['play_url_hd'] as String? ?? '',
      playUrlSd: json['play_url_sd'] as String? ?? '',
      playUrlLd: json['play_url_ld'] as String? ?? '',
      duration: json['duration'] as String? ?? '',
      durationSeconds: (json['duration_seconds'] as num?)?.toInt() ?? 0,
      videoOrder: (json['video_order'] as num?)?.toInt() ?? 0,
      status: (json['status'] as num?)?.toInt() ?? 0,
      processStatus: (json['process_status'] as num?)?.toInt() ?? 0,
      processProgress: (json['process_progress'] as num?)?.toInt() ?? 0,
    );
  }

  String get bestPlayUrl {
    if (playUrlHd.isNotEmpty) return playUrlHd;
    if (playUrl.isNotEmpty) return playUrl;
    if (playUrlSd.isNotEmpty) return playUrlSd;
    return playUrlLd;
  }
}

class ManuscriptInfo {
  final int id;
  final String title;
  final String description;
  final String coverUrl;
  final int userId;
  final int categoryId;
  final String categoryName;
  final int viewCount;
  final int likeCount;
  final int coinCount;
  final int collectCount;
  final int shareCount;
  final int commentCount;
  final int danmakuCount;
  final String duration;
  final int durationSeconds;
  final int status;
  final int reviewStatus;
  final String reviewReason;
  final String createdAt;
  final String updatedAt;
  final UserInfo? uploader;
  final List<String> tags;
  final int firstVideoId;
  final String firstVideoPlayUrl;
  final List<VideoItem> videos;

  const ManuscriptInfo({
    this.id = 0,
    this.title = '',
    this.description = '',
    this.coverUrl = '',
    this.userId = 0,
    this.categoryId = 0,
    this.categoryName = '',
    this.viewCount = 0,
    this.likeCount = 0,
    this.coinCount = 0,
    this.collectCount = 0,
    this.shareCount = 0,
    this.commentCount = 0,
    this.danmakuCount = 0,
    this.duration = '',
    this.durationSeconds = 0,
    this.status = 0,
    this.reviewStatus = 0,
    this.reviewReason = '',
    this.createdAt = '',
    this.updatedAt = '',
    this.uploader,
    this.tags = const [],
    this.firstVideoId = 0,
    this.firstVideoPlayUrl = '',
    this.videos = const [],
  });

  factory ManuscriptInfo.fromJson(Map<String, dynamic> json) {
    return ManuscriptInfo(
      id: (json['id'] as num?)?.toInt() ?? 0,
      title: json['title'] as String? ?? '',
      description: json['description'] as String? ?? '',
      coverUrl: json['cover_url'] as String? ?? '',
      userId: (json['user_id'] as num?)?.toInt() ?? 0,
      categoryId: (json['category_id'] as num?)?.toInt() ?? 0,
      categoryName: json['category_name'] as String? ?? '',
      viewCount: (json['view_count'] as num?)?.toInt() ?? 0,
      likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
      coinCount: (json['coin_count'] as num?)?.toInt() ?? 0,
      collectCount: (json['collect_count'] as num?)?.toInt() ?? 0,
      shareCount: (json['share_count'] as num?)?.toInt() ?? 0,
      commentCount: (json['comment_count'] as num?)?.toInt() ?? 0,
      danmakuCount: (json['danmaku_count'] as num?)?.toInt() ?? 0,
      duration: json['duration'] as String? ?? '',
      durationSeconds: (json['duration_seconds'] as num?)?.toInt() ?? 0,
      status: (json['status'] as num?)?.toInt() ?? 0,
      reviewStatus: (json['review_status'] as num?)?.toInt() ?? 0,
      reviewReason: json['review_reason'] as String? ?? '',
      createdAt: json['created_at'] as String? ?? '',
      updatedAt: json['updated_at'] as String? ?? '',
      uploader: json['uploader'] is Map
          ? UserInfo.fromJson(json['uploader'] as Map<String, dynamic>)
          : null,
      tags: (json['tags'] as List?)?.cast<String>() ?? const [],
      firstVideoId: (json['first_video_id'] as num?)?.toInt() ?? 0,
      firstVideoPlayUrl: json['first_video_play_url'] as String? ?? '',
      videos: (json['videos'] as List?)
          ?.map((e) => VideoItem.fromJson(e as Map<String, dynamic>))
          .toList() ??
          const [],
    );
  }
}