import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final watchHistoryApiProvider = Provider<WatchHistoryApi>((ref) => WatchHistoryApi(ref.read(dioProvider)));

class WatchHistoryItem {
  final int id;
  final int manuscriptId;
  final String title;
  final String coverUrl;
  final String duration;
  final int durationSeconds;
  final int progressSeconds;
  final String uploaderName;
  final String watchedAt;

  const WatchHistoryItem({this.id = 0, this.manuscriptId = 0, this.title = '', this.coverUrl = '', this.duration = '', this.durationSeconds = 0, this.progressSeconds = 0, this.uploaderName = '', this.watchedAt = ''});

  factory WatchHistoryItem.fromJson(Map<String, dynamic> json) => WatchHistoryItem(
    id: (json['id'] as num?)?.toInt() ?? 0,
    manuscriptId: (json['manuscript_id'] as num?)?.toInt() ?? (json['manuscriptId'] as num?)?.toInt() ?? 0,
    title: json['title'] as String? ?? '',
    coverUrl: json['cover_url'] as String? ?? json['cover'] as String? ?? '',
    duration: json['duration'] as String? ?? '',
    durationSeconds: (json['duration_seconds'] as num?)?.toInt() ?? 0,
    progressSeconds: (json['progress_seconds'] as num?)?.toInt() ?? 0,
    uploaderName: json['uploader_name'] as String? ?? json['uploader'] as String? ?? '',
    watchedAt: json['watched_at'] as String? ?? json['created_at'] as String? ?? '',
  );
}

class WatchHistoryApi {
  final Dio _dio;
  WatchHistoryApi(this._dio);

  Future<List<WatchHistoryItem>> getHistory({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/watch-history', queryParameters: {'page': page, 'pageSize': pageSize});
    return _extractList(res.data).map((e) => WatchHistoryItem.fromJson(e)).toList();
  }

  Future<void> deleteHistory(int id) async {
    await _dio.delete('/watch-history/$id');
  }

  Future<void> clearHistory() async {
    await _dio.delete('/watch-history');
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