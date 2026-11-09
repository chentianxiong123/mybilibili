import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final uploadApiProvider = Provider<UploadApi>((ref) => UploadApi(ref.read(dioProvider)));

class UploadApi {
  final Dio _dio;
  UploadApi(this._dio);

  Future<String> createSession({
    required String title,
    required String description,
    required int categoryId,
    List<String> tags = const [],
    String? videoUrl,
  }) async {
    final res = await _dio.post('/manuscript/upload-session', data: {
      'title': title,
      'description': description,
      'category_id': categoryId,
      'tags': tags,
      'videos': [
        {'url': videoUrl, 'total_chunks': 0},
      ],
      'total_chunks': 0,
    });
    final data = _data(res.data);
    return data['upload_id'] as String? ?? '';
  }

  Future<Map<String, dynamic>> completeSession(String uploadId) async {
    final res = await _dio.post('/manuscript/upload-session/$uploadId/complete');
    return _data(res.data);
  }

  Map<String, dynamic> _data(dynamic response) {
    if (response is Map) {
      final d = response['data'];
      if (d is Map) return Map<String, dynamic>.from(d);
    }
    return {};
  }
}
