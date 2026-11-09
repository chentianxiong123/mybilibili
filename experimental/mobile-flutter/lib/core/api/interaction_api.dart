import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';
import '../../shared/models/manuscript.dart';

final interactionApiProvider = Provider<InteractionApi>((ref) => InteractionApi(ref.read(dioProvider)));

class InteractionStatus {
  final bool liked;
  final bool coined;
  final bool collected;
  final int likeCount;
  final int coinCount;
  final int collectCount;

  const InteractionStatus({this.liked = false, this.coined = false, this.collected = false, this.likeCount = 0, this.coinCount = 0, this.collectCount = 0});

  factory InteractionStatus.fromJson(Map<String, dynamic> json) => InteractionStatus(
    liked: json['liked'] == true || json['isLike'] == true,
    coined: json['coined'] == true || json['isCoin'] == true,
    collected: json['collected'] == true || json['isCollect'] == true,
    likeCount: (json['likeCount'] as num?)?.toInt() ?? 0,
    coinCount: (json['coinCount'] as num?)?.toInt() ?? 0,
    collectCount: (json['collectCount'] as num?)?.toInt() ?? 0,
  );
}

class InteractionApi {
  final Dio _dio;
  InteractionApi(this._dio);

  Future<InteractionStatus> getStatus(int manuscriptId) async {
    final res = await _dio.get('/manuscript/$manuscriptId/status');
    final data = res.data is Map ? (res.data['data'] is Map ? res.data['data'] as Map<String, dynamic> : <String, dynamic>{}) : <String, dynamic>{};
    return InteractionStatus.fromJson(data);
  }

  Future<void> like(int manuscriptId) async => await _dio.post('/manuscript/$manuscriptId/like');
  Future<void> unlike(int manuscriptId) async => await _dio.delete('/manuscript/$manuscriptId/like');
  Future<void> coin(int manuscriptId, {int count = 2}) async => await _dio.post('/manuscript/$manuscriptId/coin', data: {'count': count});
  Future<void> collect(int manuscriptId, {int? folderId}) async => await _dio.post('/manuscript/$manuscriptId/collect', data: {'folderId': folderId});
  Future<void> uncollect(int manuscriptId) async => await _dio.delete('/manuscript/$manuscriptId/collect');
  Future<void> share(int manuscriptId) async => await _dio.post('/manuscript/$manuscriptId/share');
  Future<void> follow(int userId) async => await _dio.post('/follow/$userId');
  Future<void> unfollow(int userId) async => await _dio.delete('/follow/$userId');
}