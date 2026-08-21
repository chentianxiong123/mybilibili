import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final liveApiProvider = Provider<LiveApi>((ref) => LiveApi(ref.read(dioProvider)));

class LiveRoom {
  final int id;
  final String roomName;
  final String roomCode;
  final int hostId;
  final String streamKey;
  final String cover;
  final String category;
  final int maxSeats;
  final int status;
  final int viewerCount;
  final String createdAt;

  const LiveRoom({
    this.id = 0, this.roomName = '', this.roomCode = '', this.hostId = 0,
    this.streamKey = '', this.cover = '', this.category = '', this.maxSeats = 0,
    this.status = 0, this.viewerCount = 0, this.createdAt = '',
  });

  factory LiveRoom.fromJson(Map<String, dynamic> json) => LiveRoom(
    id: (json['id'] as num?)?.toInt() ?? 0,
    roomName: json['room_name'] as String? ?? json['RoomName'] as String? ?? '',
    roomCode: json['room_code'] as String? ?? json['RoomCode'] as String? ?? '',
    hostId: (json['host_id'] as num?)?.toInt() ?? (json['HostID'] as num?)?.toInt() ?? 0,
    streamKey: json['stream_key'] as String? ?? json['StreamKey'] as String? ?? '',
    cover: json['cover'] as String? ?? json['Cover'] as String? ?? '',
    category: json['category'] as String? ?? json['Category'] as String? ?? '',
    maxSeats: (json['max_seats'] as num?)?.toInt() ?? (json['MaxSeats'] as num?)?.toInt() ?? 0,
    status: (json['status'] as num?)?.toInt() ?? (json['Status'] as num?)?.toInt() ?? 0,
    viewerCount: (json['viewer_count'] as num?)?.toInt() ?? (json['ViewerCount'] as num?)?.toInt() ?? 0,
    createdAt: json['created_at'] as String? ?? json['CreatedAt'] as String? ?? '',
  );

  bool get isLive => status == 1;

  String get streamUrl => 'http://192.168.31.204:8080/live/$roomCode.flv';
}

class LiveApi {
  final Dio _dio;
  LiveApi(this._dio);

  Future<List<LiveRoom>> getRooms({int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/live/rooms', queryParameters: {'page': page, 'page_size': pageSize});
    return _extractList(res.data).map((e) => LiveRoom.fromJson(e)).toList();
  }

  Future<LiveRoom> getRoom(int roomId) async {
    final res = await _dio.get('/live/room/$roomId');
    final data = res.data is Map ? (res.data['data'] ?? res.data) : res.data;
    return LiveRoom.fromJson(data as Map<String, dynamic>);
  }

  Future<LiveRoom?> getMyRoom() async {
    try {
      final res = await _dio.get('/live/room/my');
      final data = res.data is Map ? (res.data['data'] ?? res.data) : res.data;
      return LiveRoom.fromJson(data as Map<String, dynamic>);
    } catch (_) { return null; }
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