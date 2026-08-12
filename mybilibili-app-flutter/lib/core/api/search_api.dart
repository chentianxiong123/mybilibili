import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../../shared/models/manuscript.dart';

final searchApiProvider = Provider<SearchApi>((ref) => SearchApi(ref.read(dioProvider)));

class SearchApi {
  final Dio _dio;
  SearchApi(this._dio);

  Future<List<ManuscriptInfo>> search(String keyword, {int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/search/videos', queryParameters: {'keyword': keyword, 'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<List<String>> getSuggestions(String keyword) async {
    final res = await _dio.get('/search/suggest', queryParameters: {'keyword': keyword});
    final data = _extractList(res.data);
    return data.map((e) => e['name'] as String? ?? e['value'] as String? ?? '').where((s) => s.isNotEmpty).toList();
  }

  Future<List<String>> getHotKeywords() async {
    final res = await _dio.get('/search/hot');
    final data = _extractList(res.data);
    return data.map((e) => e['name'] as String? ?? e['keyword'] as String? ?? '').where((s) => s.isNotEmpty).toList();
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