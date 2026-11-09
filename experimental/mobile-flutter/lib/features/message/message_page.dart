import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/message_api.dart';
import '../../core/theme/theme.dart';

final conversationsProvider = FutureProvider.autoDispose<List<Conversation>>((ref) {
  return ref.read(messageApiProvider).getConversations();
});

class MessagePage extends ConsumerWidget {
  const MessagePage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final conversations = ref.watch(conversationsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('消息')),
      body: conversations.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无消息', style: TextStyle(color: Colors.grey)))
            : ListView.separated(
                padding: const EdgeInsets.all(8),
                itemCount: list.length,
                separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
                itemBuilder: (context, index) {
                  final c = list[index];
                  return ListTile(
                    leading: CircleAvatar(
                      backgroundImage: c.peerAvatar.isNotEmpty ? NetworkImage(c.peerAvatar) : null,
                      child: c.peerAvatar.isEmpty ? const Icon(Icons.person) : null,
                    ),
                    title: Text(c.peerName, style: const TextStyle(fontSize: 14)),
                    subtitle: Row(
                      children: [
                        Expanded(child: Text(c.lastMessage, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 12, color: Colors.grey))),
                        if (c.unreadCount > 0)
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                            decoration: BoxDecoration(color: AppTheme.primaryPink, borderRadius: BorderRadius.circular(10)),
                            child: Text('${c.unreadCount}', style: const TextStyle(color: Colors.white, fontSize: 10)),
                          ),
                      ],
                    ),
                    onTap: () => Navigator.of(context).push(MaterialPageRoute(
                      builder: (_) => ChatPage(conversationId: c.id, peerName: c.peerName, peerId: c.peerId),
                    )),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }
}

class ChatPage extends ConsumerStatefulWidget {
  final int conversationId;
  final String peerName;
  final int peerId;
  const ChatPage({super.key, required this.conversationId, required this.peerName, required this.peerId});
  @override
  ConsumerState<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends ConsumerState<ChatPage> {
  final _msgController = TextEditingController();
  final _scrollController = ScrollController();
  List<Message> _messages = [];
  bool _loading = true;

  @override
  void initState() {
    super.initState();
    _loadMessages();
  }

  Future<void> _loadMessages() async {
    try {
      final msgs = await ref.read(messageApiProvider).getMessages(widget.conversationId);
      if (mounted) setState(() { _messages = msgs.reversed.toList(); _loading = false; });
      _scrollToBottom();
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _send() async {
    if (_msgController.text.trim().isEmpty) return;
    final text = _msgController.text.trim();
    _msgController.clear();
    setState(() => _messages.add(Message(senderId: 0, content: text, createdAt: DateTime.now().toIso8601String())));
    _scrollToBottom();
    try {
      await ref.read(messageApiProvider).sendMessage(widget.peerId, text);
    } catch (_) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('发送失败')));
    }
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) _scrollController.jumpTo(_scrollController.position.maxScrollExtent);
    });
  }

  @override
  void dispose() {
    _msgController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text(widget.peerName)),
      body: Column(
        children: [
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : ListView.builder(
                    controller: _scrollController,
                    padding: const EdgeInsets.all(12),
                    itemCount: _messages.length,
                    itemBuilder: (context, index) {
                      final m = _messages[index];
                      final isMe = m.senderId == 0;
                      return _buildMessageBubble(m, isMe);
                    },
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
                      controller: _msgController,
                      style: const TextStyle(fontSize: 14),
                      decoration: InputDecoration(
                        hintText: '发送消息...',
                        hintStyle: const TextStyle(fontSize: 14, color: Colors.grey),
                        filled: true, fillColor: const Color(0xFF2A2A2A),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 0),
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(18), borderSide: BorderSide.none),
                      ),
                      onSubmitted: (_) => _send(),
                    ),
                  ),
                ),
                IconButton(icon: const Icon(Icons.send, color: AppTheme.primaryPink), onPressed: _send),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMessageBubble(Message m, bool isMe) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: isMe ? MainAxisAlignment.end : MainAxisAlignment.start,
        children: [
          Flexible(
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              decoration: BoxDecoration(
                color: isMe ? AppTheme.primaryPink : const Color(0xFF2A2A2A),
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(12),
                  topRight: const Radius.circular(12),
                  bottomLeft: isMe ? const Radius.circular(12) : Radius.zero,
                  bottomRight: isMe ? Radius.zero : const Radius.circular(12),
                ),
              ),
              child: Text(m.content, style: TextStyle(fontSize: 14, color: isMe ? Colors.white : Colors.white)),
            ),
          ),
        ],
      ),
    );
  }
}