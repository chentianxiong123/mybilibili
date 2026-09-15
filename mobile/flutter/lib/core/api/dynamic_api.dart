import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final dynamicApiProvider = Provider<DynamicApi>((ref) => DynamicApi(ref.read(dioProvider)));

class DynamicItem {
  final int id;
  final int userId;
  final String username;
  final String userAvatar;
  final String content;
  final int type;
  final int refManuscriptId;
  final String refManuscriptTitle;
  final String refManuscriptCover;
  final int likeCount;
  final int commentCount;
  final int shareCount;
  final bool liked;
  final String createdAt;

  const DynamicItem({
    this.id = 0, this.userId = 0, this.username = '', this.userAvatar = '',
    this.content = '', this.type = 0, this.refManuscriptId = 0,
    this.refManuscriptTitle = '', this.refManuscriptCover = '',
    this.likeCount = 0, this.commentCount = 0, this.shareCount = 0,
    this.liked = false, this.createdAt = '',
  });

  factory DynamicItem.fromJson(Map<String, dynamic> json) => DynamicItem(
    id: (json['id'] as num?)?.toInt() ?? 0,
    userId: (json['user_id'] as num?)?.toInt() ?? (json['userId'] as num?)?.toInt() ?? 0,
    username: json['username'] as String? ?? json['nickname'] as String? ?? '',
    userAvatar: json['user_avatar'] as String? ?? json['avatar'] as String? ?? '',
    content: json['content'] as String? ?? '',
    type: (json['type'] as num?)?.toInt() ?? 0,
    refManuscriptId: (json['ref_manuscript_id'] as num?)?.toInt() ?? (json['refManuscriptId'] as num?)?.toInt() ?? 0,
    refManuscriptTitle: json['ref_manuscript_title'] as String? ?? json['title'] as String? ?? '',
    refManuscriptCover: json['ref_manuscript_cover'] as String? ?? json['cover'] as String? ?? '',
    likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
    commentCount: (json['comment_count'] as num?)?.toInt() ?? 0,
    shareCount: (json['share_count'] as num?)?.toInt() ?? 0,
    liked: json['liked'] == true || json['isLike'] == true,
    createdAt: json['created_at'] as String? ?? json['createdAt'] as String? ?? '',
  );
}

class DynamicComment {
  final int id;
  final int userId;
  final String username;
  final String userAvatar;
  final String content;
  final int likeCount;
  final String createdAt;
  final List<DynamicComment> replies;

  const DynamicComment({
    this.id = 0, this.userId = 0, this.username = '', this.userAvatar = '',
    this.content = '', this.likeCount = 0, this.createdAt = '', this.replies = const [],
  });

  factory DynamicComment.fromJson(Map<String, dynamic> json) => DynamicComment(
    id: (json['id'] as num?)?.toInt() ?? 0,
    userId: (json['user_id'] as num?)?.toInt() ?? 0,
    username: json['username'] as String? ?? '',
    userAvatar: json['user_avatar'] as String? ?? '',
    content: json['content'] as String? ?? '',
    likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
    createdAt: json['created_at'] as String? ?? '',
    replies: (json['replies'] as List?)?.map((e) => DynamicComment.fromJson(e as Map<String, dynamic>)).toList() ?? [],
  );
}

class DynamicApi {
  final Dio _dio;
  DynamicApi(this._dio);

  Future<List<DynamicItem>> getFollowingDynamics({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/dynamic/following', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => DynamicItem.fromJson(e)).toList();
  }

  Future<List<DynamicItem>> getAllDynamics({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/dynamic/list', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => DynamicItem.fromJson(e)).toList();
  }

  Future<List<DynamicItem>> getUserDynamics(int userId, {int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/dynamic/user/$userId', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => DynamicItem.fromJson(e)).toList();
  }

  Future<void> like(int dynamicId) async => await _dio.post('/dynamic/like/$dynamicId');
  Future<void> unlike(int dynamicId) async => await _dio.delete('/dynamic/like/$dynamicId');
  Future<void> delete(int dynamicId) async => await _dio.delete('/dynamic/$dynamicId');

  Future<void> publish(String content, {int type = 0, int? refManuscriptId}) async {
    await _dio.post('/dynamic/publish', queryParameters: {'content': content, 'type': type, 'ref_manuscript_id': refManuscriptId});
  }

  Future<List<DynamicComment>> getComments(int dynamicId, {int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/dynamic/comment/list', queryParameters: {'dynamicId': dynamicId, 'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => DynamicComment.fromJson(e)).toList();
  }

  Future<void> addComment(int dynamicId, String content, {int? parentId, int? replyUserId}) async {
    await _dio.post('/dynamic/comment/add', data: {
      'dynamicId': dynamicId, 'content': content, 'parentId': parentId, 'replyUserId': replyUserId,
    });
  }

  Future<void> likeComment(int commentId) async => await _dio.post('/dynamic/comment/like/$commentId');
  Future<void> deleteComment(int commentId) async => await _dio.delete('/dynamic/comment/delete/$commentId');

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