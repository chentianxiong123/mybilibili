import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final creatorApiProvider = Provider<CreatorApi>((ref) => CreatorApi(ref.read(dioProvider)));

class CreatorOverview {
  final int manuscriptCount;
  final int viewCount;
  final int likeCount;
  final int coinCount;
  final int collectCount;
  final int followerCount;

  const CreatorOverview({
    this.manuscriptCount = 0, this.viewCount = 0, this.likeCount = 0,
    this.coinCount = 0, this.collectCount = 0, this.followerCount = 0,
  });

  factory CreatorOverview.fromJson(Map<String, dynamic> json) => CreatorOverview(
    manuscriptCount: (json['manuscript_count'] as num?)?.toInt() ?? 0,
    viewCount: (json['view_count'] as num?)?.toInt() ?? 0,
    likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
    coinCount: (json['coin_count'] as num?)?.toInt() ?? 0,
    collectCount: (json['collect_count'] as num?)?.toInt() ?? 0,
    followerCount: (json['follower_count'] as num?)?.toInt() ?? 0,
  );
}

class CreatorComment {
  final int id;
  final int manuscriptId;
  final String manuscriptTitle;
  final int userId;
  final String username;
  final String content;
  final int likeCount;
  final String createdAt;

  const CreatorComment({
    this.id = 0, this.manuscriptId = 0, this.manuscriptTitle = '',
    this.userId = 0, this.username = '', this.content = '',
    this.likeCount = 0, this.createdAt = '',
  });

  factory CreatorComment.fromJson(Map<String, dynamic> json) => CreatorComment(
    id: (json['id'] as num?)?.toInt() ?? 0,
    manuscriptId: (json['manuscript_id'] as num?)?.toInt() ?? 0,
    manuscriptTitle: json['manuscript_title'] as String? ?? json['title'] as String? ?? '',
    userId: (json['user_id'] as num?)?.toInt() ?? 0,
    username: json['username'] as String? ?? json['nickname'] as String? ?? '',
    content: json['content'] as String? ?? '',
    likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
    createdAt: json['created_at'] as String? ?? '',
  );
}

class CreatorApi {
  final Dio _dio;
  CreatorApi(this._dio);

  Future<CreatorOverview> getOverview() async {
    final res = await _dio.get('/creator/stats/overview');
    final data = res.data is Map ? (res.data['data'] ?? res.data) : res.data;
    return CreatorOverview.fromJson(data as Map<String, dynamic>);
  }

  Future<List<Map<String, dynamic>>> getTrend({int days = 7}) async {
    final res = await _dio.get('/creator/stats/trend', queryParameters: {'days': days});
    return _extractList(res.data);
  }

  Future<List<Map<String, dynamic>>> getRanking({String sortBy = 'view_count', int limit = 10}) async {
    final res = await _dio.get('/creator/stats/ranking', queryParameters: {'sortBy': sortBy, 'limit': limit});
    return _extractList(res.data);
  }

  Future<List<CreatorComment>> getComments({int? manuscriptId, int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/creator/comments', queryParameters: {
      if (manuscriptId != null) 'manuscriptId': manuscriptId,
      'page': page, 'pageSize': pageSize,
    });
    return _extractList(res.data).map((e) => CreatorComment.fromJson(e)).toList();
  }

  Future<void> deleteComment(int commentId) async {
    await _dio.delete('/creator/comments/$commentId');
  }

  Future<void> replyComment(int commentId, String content) async {
    await _dio.post('/creator/comments/$commentId/reply', data: {'content': content});
  }

  List<Map<String, dynamic>> _extractList(dynamic data) {
    if (data is Map) {
      final d = data['data'];
      if (d is List) return d.cast<Map<String, dynamic>>();
      final list = data['list'] ?? data['items'] ?? data['records'];
      if (list is List) return list.cast<Map<String, dynamic>>();
    }
    if (data is List) return data.cast<Map<String, dynamic>>();
    return [];
  }
}