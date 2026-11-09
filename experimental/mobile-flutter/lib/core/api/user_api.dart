import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../../shared/models/manuscript.dart';

final userApiProvider = Provider<UserApi>((ref) => UserApi(ref.read(dioProvider)));

class UserProfile {
  final int id;
  final String username;
  final String nickname;
  final String avatar;
  final String introduction;
  final int level;
  final int followerCount;
  final int followingCount;
  final int likeCount;
  final int manuscriptCount;
  final String createdAt;

  const UserProfile({this.id = 0, this.username = '', this.nickname = '', this.avatar = '', this.introduction = '', this.level = 0, this.followerCount = 0, this.followingCount = 0, this.likeCount = 0, this.manuscriptCount = 0, this.createdAt = ''});

  factory UserProfile.fromJson(Map<String, dynamic> json) => UserProfile(
    id: (json['id'] as num?)?.toInt() ?? 0,
    username: json['username'] as String? ?? '',
    nickname: json['nickname'] as String? ?? json['name'] as String? ?? '',
    avatar: json['avatar'] as String? ?? json['avatar_url'] as String? ?? '',
    introduction: json['introduction'] as String? ?? json['sign'] as String? ?? '',
    level: (json['level'] as num?)?.toInt() ?? 0,
    followerCount: (json['follower_count'] as num?)?.toInt() ?? 0,
    followingCount: (json['following_count'] as num?)?.toInt() ?? 0,
    likeCount: (json['like_count'] as num?)?.toInt() ?? 0,
    manuscriptCount: (json['manuscript_count'] as num?)?.toInt() ?? 0,
    createdAt: json['created_at'] as String? ?? '',
  );
}

class UserApi {
  final Dio _dio;
  UserApi(this._dio);

  Future<UserProfile> getUserProfile(int userId) async {
    final res = await _dio.get('/profile/$userId');
    return _extractData(res.data, UserProfile.fromJson);
  }

  Future<UserProfile> getCurrentUserProfile() async {
    final res = await _dio.get('/user/me');
    return _extractData(res.data, UserProfile.fromJson);
  }

  Future<List<ManuscriptInfo>> getUserManuscripts(int userId, {int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/manuscript/user/$userId', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<List<ManuscriptInfo>> getMyManuscripts({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/manuscript/me/list', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<void> updateProfile(Map<String, dynamic> data) async {
    await _dio.put('/user/me', data: data);
  }

  Future<void> updateAvatar(String avatarUrl) async {
    await _dio.post('/user/me/avatar', data: {'avatar': avatarUrl});
  }

  T _extractData<T>(dynamic res, T Function(Map<String, dynamic>) fromJson) {
    if (res is Map) {
      final data = res['data'];
      if (data is Map<String, dynamic>) return fromJson(data);
    }
    return fromJson(res as Map<String, dynamic>);
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