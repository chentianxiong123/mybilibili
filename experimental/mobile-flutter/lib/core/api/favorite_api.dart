import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../../shared/models/manuscript.dart';

final favoriteApiProvider = Provider<FavoriteApi>((ref) => FavoriteApi(ref.read(dioProvider)));

class FavoriteFolder {
  final int id;
  final String name;
  final String coverUrl;
  final int manuscriptCount;
  final int viewCount;
  final String createdAt;

  const FavoriteFolder({this.id = 0, this.name = '', this.coverUrl = '', this.manuscriptCount = 0, this.viewCount = 0, this.createdAt = ''});

  factory FavoriteFolder.fromJson(Map<String, dynamic> json) => FavoriteFolder(
    id: (json['id'] as num?)?.toInt() ?? 0,
    name: json['name'] as String? ?? json['title'] as String? ?? '',
    coverUrl: json['cover_url'] as String? ?? json['cover'] as String? ?? '',
    manuscriptCount: (json['manuscript_count'] as num?)?.toInt() ?? 0,
    viewCount: (json['view_count'] as num?)?.toInt() ?? 0,
    createdAt: json['created_at'] as String? ?? '',
  );
}

class FavoriteApi {
  final Dio _dio;
  FavoriteApi(this._dio);

  Future<List<FavoriteFolder>> getFolders() async {
    final res = await _dio.get('/favorites');
    return _extractList(res.data).map((e) => FavoriteFolder.fromJson(e)).toList();
  }

  Future<List<ManuscriptInfo>> getFolderManuscripts(int folderId, {int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/favorites/$folderId/videos', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => ManuscriptInfo.fromJson(e)).toList();
  }

  Future<void> createFolder(String name) async {
    await _dio.post('/favorites', data: {'name': name});
  }

  Future<void> deleteFolder(int id) async {
    await _dio.delete('/favorites/$id');
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