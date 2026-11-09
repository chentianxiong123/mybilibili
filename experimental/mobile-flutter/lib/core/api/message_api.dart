import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final messageApiProvider = Provider<MessageApi>((ref) => MessageApi(ref.read(dioProvider)));

class Conversation {
  final int id;
  final int peerId;
  final String peerName;
  final String peerAvatar;
  final String lastMessage;
  final int unreadCount;
  final String updatedAt;

  const Conversation({
    this.id = 0, this.peerId = 0, this.peerName = '', this.peerAvatar = '',
    this.lastMessage = '', this.unreadCount = 0, this.updatedAt = '',
  });

  factory Conversation.fromJson(Map<String, dynamic> json) => Conversation(
    id: (json['id'] as num?)?.toInt() ?? 0,
    peerId: (json['peer_id'] as num?)?.toInt() ?? (json['receiver_id'] as num?)?.toInt() ?? 0,
    peerName: json['peer_name'] as String? ?? json['receiver_name'] as String? ?? '',
    peerAvatar: json['peer_avatar'] as String? ?? '',
    lastMessage: json['last_message'] as String? ?? json['lastMessage'] as String? ?? '',
    unreadCount: (json['unread_count'] as num?)?.toInt() ?? 0,
    updatedAt: json['updated_at'] as String? ?? json['updatedAt'] as String? ?? '',
  );
}

class Message {
  final int id;
  final int senderId;
  final int receiverId;
  final String content;
  final int messageType;
  final String createdAt;

  const Message({
    this.id = 0, this.senderId = 0, this.receiverId = 0,
    this.content = '', this.messageType = 1, this.createdAt = '',
  });

  factory Message.fromJson(Map<String, dynamic> json) => Message(
    id: (json['id'] as num?)?.toInt() ?? 0,
    senderId: (json['sender_id'] as num?)?.toInt() ?? (json['userId'] as num?)?.toInt() ?? 0,
    receiverId: (json['receiver_id'] as num?)?.toInt() ?? 0,
    content: json['content'] as String? ?? '',
    messageType: (json['message_type'] as num?)?.toInt() ?? 1,
    createdAt: json['created_at'] as String? ?? '',
  );
}

class MessageApi {
  final Dio _dio;
  MessageApi(this._dio);

  Future<List<Conversation>> getConversations() async {
    final res = await _dio.get('/message/conversations');
    return _extractList(res.data).map((e) => Conversation.fromJson(e)).toList();
  }

  Future<List<Message>> getMessages(int conversationId, {int page = 1, int pageSize = 20}) async {
    final res = await _dio.get('/message/conversations/$conversationId/messages', queryParameters: {'page': page, 'page_size': pageSize});
    return _extractList(res.data).map((e) => Message.fromJson(e)).toList();
  }

  Future<void> sendMessage(int receiverId, String content, {int messageType = 1}) async {
    await _dio.post('/message/send', data: {'receiver_id': receiverId, 'content': content, 'message_type': messageType});
  }

  Future<void> markRead(int messageId) async {
    await _dio.put('/message/$messageId/read');
  }

  Future<void> batchRead(List<int> messageIds) async {
    await _dio.put('/message/batch/read', data: {'messageIds': messageIds});
  }

  Future<int> getUnreadCount() async {
    final res = await _dio.get('/message/unread/counts');
    final data = res.data is Map ? res.data : {};
    return (data['unread_count'] as num?)?.toInt() ?? 0;
  }

  Future<void> deleteConversation(int id) async {
    await _dio.delete('/message/conversations/$id');
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