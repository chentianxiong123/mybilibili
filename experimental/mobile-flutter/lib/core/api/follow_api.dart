import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../../shared/models/user.dart';

final followApiProvider = Provider<FollowApi>((ref) => FollowApi(ref.read(dioProvider)));

class FollowApi {
  final Dio _dio;
  FollowApi(this._dio);

  Future<List<UserInfo>> getFollowers({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/follow/me/followers', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => UserInfo.fromJson(e)).toList();
  }

  Future<List<UserInfo>> getFollowing({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/follow/me/following', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => UserInfo.fromJson(e)).toList();
  }

  Future<bool> checkFollow(int userId) async {
    final res = await _dio.get('/follow/check/$userId');
    final data = res.data is Map ? (res.data['data'] as Map? ?? {}) : {};
    return data['followed'] == true || data['isFollow'] == true;
  }

  Future<void> follow(int userId) async {
    await _dio.post('/follow/$userId');
  }

  Future<void> unfollow(int userId) async {
    await _dio.delete('/follow/$userId');
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