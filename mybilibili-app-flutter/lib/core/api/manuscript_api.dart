import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../shared/models/manuscript.dart';
import 'client.dart';

final manuscriptApiProvider = Provider<ManuscriptApi>((ref) {
  return ManuscriptApi(ref.read(dioProvider));
});

class ManuscriptApi {
  final Dio _dio;
  ManuscriptApi(this._dio);

  Future<List<ManuscriptInfo>> getRecommended({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/manuscript/recommended', queryParameters: {
      'page': page,
      'pageSize': pageSize,
    });
    final data = _extractList(response.data);
    return data.map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<List<ManuscriptInfo>> getHot({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/manuscript/hot', queryParameters: {
      'page': page,
      'pageSize': pageSize,
    });
    final data = _extractList(response.data);
    return data.map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<ManuscriptInfo> getManuscript(int id) async {
    final response = await _dio.get('/manuscript/$id');
    final data = _extractData(response.data);
    return ManuscriptInfo.fromJson(data);
  }

  Future<List<ManuscriptInfo>> getByCategory(int categoryId, {int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/manuscript/category/$categoryId', queryParameters: {
      'page': page,
      'pageSize': pageSize,
    });
    final data = _extractList(response.data);
    return data.map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<List<ManuscriptInfo>> search(String keyword, {int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/search/videos', queryParameters: {
      'keyword': keyword,
      'page': page,
      'pageSize': pageSize,
    });
    final data = _extractList(response.data);
    return data.map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  List<Map<String, dynamic>> _extractList(dynamic response) {
    if (response is Map<String, dynamic>) {
      final data = response['data'];
      if (data is List) return data.cast<Map<String, dynamic>>();
      if (data is Map<String, dynamic>) {
        final list = data['list'] ?? data['items'] ?? data['records'];
        if (list is List) return list.cast<Map<String, dynamic>>();
      }
      final list = response['list'] ?? response['items'];
      if (list is List) return list.cast<Map<String, dynamic>>();
    }
    if (response is List) return response.cast<Map<String, dynamic>>();
    return [];
  }

  Map<String, dynamic> _extractData(dynamic response) {
    if (response is Map<String, dynamic>) {
      final data = response['data'];
      if (data is Map<String, dynamic>) return data;
      return response;
    }
    return {};
  }
}