import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../../features/video/player/danmaku_model.dart';

final danmakuApiProvider = Provider<DanmakuApi>((ref) => DanmakuApi(ref.read(dioProvider)));

class DanmakuApi {
  final Dio _dio;
  DanmakuApi(this._dio);

  Future<List<DanmakuData>> getDanmaku(int videoId) async {
    final res = await _dio.get('/danmaku/video/$videoId');
    return _extractList(res.data).map((e) => DanmakuData.fromJson(e)).toList();
  }

  Future<void> sendDanmaku(int videoId, String text, {int type = 0, int color = 0xFFFFFFFF, int fontSize = 25}) async {
    await _dio.post('/danmaku/send', data: {
      'videoId': videoId, 'text': text, 'type': type, 'color': color, 'fontSize': fontSize,
    });
  }

  List<Map<String, dynamic>> _extractList(dynamic data) {
    if (data is Map) {
      final d = data['data'];
      if (d is List) return d.cast<Map<String, dynamic>>();
    }
    if (data is List) return data.cast<Map<String, dynamic>>();
    return [];
  }
}