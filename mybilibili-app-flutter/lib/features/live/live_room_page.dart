import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/live_api.dart';
import '../../core/theme/theme.dart';
import '../video/player/widgets/player_view.dart';

final liveRoomProvider = FutureProvider.family<LiveRoom, int>((ref, roomId) {
  return ref.read(liveApiProvider).getRoom(roomId);
});

class LiveRoomPage extends ConsumerStatefulWidget {
  final int roomId;
  const LiveRoomPage({super.key, required this.roomId});
  @override
  ConsumerState<LiveRoomPage> createState() => _LiveRoomPageState();
}

class _LiveRoomPageState extends ConsumerState<LiveRoomPage> {
  final _chatController = TextEditingController();
  final List<ChatMessage> _messages = [];

  @override
  void dispose() {
    _chatController.dispose();
    super.dispose();
  }

  void _sendChat() {
    if (_chatController.text.trim().isEmpty) return;
    setState(() {
      _messages.add(ChatMessage(text: _chatController.text.trim(), isMe: true));
      _chatController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    final room = ref.watch(liveRoomProvider(widget.roomId));
    return Scaffold(
      body: room.when(
        data: (r) => _buildContent(r),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e')),
      ),
    );
  }

  Widget _buildContent(LiveRoom room) {
    return Column(
      children: [
        AspectRatio(
          aspectRatio: 16 / 9,
          child: Stack(
            children: [
              room.isLive
                  ? BilibiliPlayer(
                      url: room.streamUrl,
                      title: room.roomName,
                      showBackButton: true,
                      onBack: () => Navigator.of(context).pop(),
                    )
                  : Container(
                      color: Colors.black,
                      child: const Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.live_tv, size: 48, color: Colors.grey),
                            SizedBox(height: 8),
                            Text('主播未开播', style: TextStyle(color: Colors.grey)),
                          ],
                        ),
                      ),
                    ),
              if (room.isLive)
                Positioned(
                  left: 12, top: MediaQuery.of(context).padding.top + 56,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(color: AppTheme.primaryPink, borderRadius: BorderRadius.circular(4)),
                    child: const Text('直播中', style: TextStyle(color: Colors.white, fontSize: 10)),
                  ),
                ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(room.roomName, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        if (room.category.isNotEmpty)
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                            decoration: BoxDecoration(color: const Color(0xFF2A2A2A), borderRadius: BorderRadius.circular(4)),
                            child: Text(room.category, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                          ),
                        const SizedBox(width: 8),
                        const Icon(Icons.visibility, size: 14, color: Colors.grey),
                        const SizedBox(width: 2),
                        Text('${room.viewerCount}', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                      ],
                    ),
                  ],
                ),
              ),
              IconButton(
                icon: const Icon(Icons.share, color: Colors.grey),
                onPressed: () {},
              ),
            ],
          ),
        ),
        const Divider(height: 1, color: Color(0xFF2A2A2A)),
        Expanded(
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
            itemCount: _messages.length,
            itemBuilder: (context, index) => _buildChatBubble(_messages[index]),
          ),
        ),
        Container(
          padding: const EdgeInsets.fromLTRB(12, 8, 12, 8),
          decoration: const BoxDecoration(color: Color(0xFF222222)),
          child: Row(
            children: [
              Expanded(
                child: SizedBox(
                  height: 36,
                  child: TextField(
                    controller: _chatController,
                    style: const TextStyle(fontSize: 14),
                    decoration: InputDecoration(
                      hintText: '说点什么...',
                      hintStyle: const TextStyle(fontSize: 14, color: Colors.grey),
                      filled: true, fillColor: const Color(0xFF2A2A2A),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(18), borderSide: BorderSide.none),
                    ),
                    onSubmitted: (_) => _sendChat(),
                  ),
                ),
              ),
              IconButton(icon: const Icon(Icons.send, color: AppTheme.primaryPink), onPressed: _sendChat),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildChatBubble(ChatMessage msg) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (!msg.isMe) ...[
            const Icon(Icons.person, size: 16, color: Colors.grey),
            const SizedBox(width: 4),
          ],
          Flexible(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: msg.isMe ? AppTheme.primaryPink.withValues(alpha: 0.2) : const Color(0xFF2A2A2A),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(msg.text, style: TextStyle(fontSize: 13, color: msg.isMe ? AppTheme.primaryPink : Colors.white)),
            ),
          ),
        ],
      ),
    );
  }
}

class ChatMessage {
  final String text;
  final bool isMe;
  const ChatMessage({required this.text, this.isMe = false});
}