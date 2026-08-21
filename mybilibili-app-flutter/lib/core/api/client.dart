import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../utils/token_storage.dart';

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(
    baseUrl: 'http://192.168.31.204:8080/api/v1',
    connectTimeout: const Duration(seconds: 10),
    receiveTimeout: const Duration(seconds: 30),
    headers: {'Content-Type': 'application/json'},
  ));

  dio.interceptors.add(InterceptorsWrapper(
    onRequest: (options, handler) async {
      final token = await TokenStorage.getAccessToken();
      if (token != null) {
        options.headers['Authorization'] = 'Bearer $token';
      }
      final userId = await TokenStorage.getUserId();
      if (userId != null) {
        options.headers['X-User-Id'] = '$userId';
      }
      handler.next(options);
    },
    onError: (error, handler) async {
      if (error.response?.statusCode == 401) {
        final refreshToken = await TokenStorage.getRefreshToken();
        if (refreshToken != null) {
          try {
            final response = await Dio().post(
              '${dio.options.baseUrl}/user/token/refresh',
              data: {'refreshToken': refreshToken},
            );
            final newToken = response.data['data']?['accessToken'] ??
                response.data['accessToken'];
            if (newToken != null) {
              await TokenStorage.saveAccessToken(newToken.toString());
              error.requestOptions.headers['Authorization'] = 'Bearer $newToken';
              final retryResponse = await dio.fetch(error.requestOptions);
              handler.resolve(retryResponse);
              return;
            }
          } catch (_) {}
        }
        await TokenStorage.clearTokens();
      }
      handler.next(error);
    },
  ));

  return dio;
});