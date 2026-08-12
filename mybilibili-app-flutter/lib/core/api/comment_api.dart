import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final commentApiProvider = Provider<CommentApi>((ref) => CommentApi(ref.read(dioProvider)));

class CommentItem {
  final int id;
  final int userId;
  final String username;
  final String userAvatar;
  final int userLevel;
  final String content;
  final int likeCount;
  final bool liked;
  final bool isPinned;
  final String createdAt;
  final int replyCount;
  final List<CommentItem> replies;

  const CommentItem({this.id = 0, this.userId = 0, this.username = '', this.userAvatar = '', this.userLevel = 0, this.content = '', this.likeCount = 0, this.liked = false, this.isPinned = false, this.createdAt = '', this.replyCount = 0, this.replies = const []});

  factory CommentItem.fromJson(Map<String, dynamic> json) => CommentItem(
    id: (json['id'] as num?)?.toInt() ?? 0,
    userId: (json['userId'] as num?)?.toInt() ?? 0,
    username: json['username'] as String? ?? json['nickname'] as String? ?? '',
    userAvatar: json['userAvatar'] as String? ?? json['avatar'] as String? ?? '',
    userLevel: (json['userLevel'] as num?)?.toInt() ?? 0,
    content: json['content'] as String? ?? '',
    likeCount: (json['likeCount'] as num?)?.toInt() ?? 0,
    liked: json['liked'] == true || json['isLike'] == true,
    isPinned: json['isPinned'] == true || json['pinned'] == true,
    createdAt: json['createdAt'] as String? ?? '',
    replyCount: (json['replyCount'] as num?)?.toInt() ?? 0,
    replies: (json['replies'] as List?)?.map((e) => CommentItem.fromJson(e as Map<String, dynamic>)).toList() ?? [],
  );
}

class CommentApi {
  final Dio _dio;
  CommentApi(this._dio);

  Future<List<CommentItem>> getComments(int manuscriptId, {int page = 1, int pageSize = 20, String sort = 'time'}) async {
    final res = await _dio.get('/comment/list', queryParameters: {'manuscriptId': manuscriptId, 'page': page, 'pageSize': pageSize, 'sort': sort});
    return _extractList(res.data).map((e) => CommentItem.fromJson(e)).toList();
  }

  Future<void> addComment(int manuscriptId, String content) async => await _dio.post('/comment/add', data: {'manuscriptId': manuscriptId, 'content': content});
  Future<void> replyComment(int commentId, String content) async => await _dio.post('/comment/reply', data: {'commentId': commentId, 'content': content});
  Future<void> likeComment(int commentId) async => await _dio.post('/comment/$commentId/like');
  Future<void> unlikeComment(int commentId) async => await _dio.delete('/comment/$commentId/like');

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