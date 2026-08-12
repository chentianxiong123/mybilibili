import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../utils/token_storage.dart';

final authApiProvider = Provider<AuthApi>((ref) => AuthApi(ref.read(dioProvider)));

class AuthApi {
  final Dio _dio;
  AuthApi(this._dio);

  Future<Map<String, dynamic>> login(String username, String password) async {
    final response = await _dio.post('/user/login', data: {
      'username': username,
      'password': password,
    });
    final data = response.data is Map ? (response.data['data'] ?? response.data) as Map<String, dynamic> : <String, dynamic>{};
    final token = data['token'] as String?;
    final refreshToken = data['refresh_token'] as String?;
    final userId = (data['user_id'] as num?)?.toInt();
    if (token != null) {
      await TokenStorage.saveAccessToken(token);
      if (refreshToken != null) await TokenStorage.saveRefreshToken(refreshToken);
      if (userId != null) await TokenStorage.saveUserId(userId);
    }
    return data;
  }

  Future<Map<String, dynamic>> register({
    required String username,
    required String password,
    String? nickname,
    String? email,
  }) async {
    final response = await _dio.post('/user/register', data: {
      'username': username,
      'password': password,
      if (nickname != null) 'nickname': nickname,
      if (email != null) 'email': email,
    });
    final data = response.data is Map ? (response.data['data'] ?? response.data) as Map<String, dynamic> : <String, dynamic>{};
    final token = data['token'] as String?;
    final userId = (data['user_id'] as num?)?.toInt();
    if (token != null) {
      await TokenStorage.saveAccessToken(token);
      if (userId != null) await TokenStorage.saveUserId(userId);
    }
    return data;
  }
}