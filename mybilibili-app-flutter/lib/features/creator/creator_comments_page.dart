import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/creator_api.dart';
import '../../core/theme/theme.dart';

final creatorCommentsProvider = FutureProvider.autoDispose<List<CreatorComment>>((ref) {
  return ref.read(creatorApiProvider).getComments();
});

class CreatorCommentsPage extends ConsumerWidget {
  const CreatorCommentsPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final comments = ref.watch(creatorCommentsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('评论管理')),
      body: comments.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无评论', style: TextStyle(color: Colors.grey)))
            : ListView.separated(
                padding: const EdgeInsets.all(8),
                itemCount: list.length,
                separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
                itemBuilder: (context, index) {
                  final c = list[index];
                  return ListTile(
                    leading: const CircleAvatar(child: Icon(Icons.person, size: 20)),
                    title: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(c.username, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w500, color: Color(0xFF99A2AA))),
                        const SizedBox(height: 4),
                        Text(c.content, style: const TextStyle(fontSize: 14)),
                      ],
                    ),
                    subtitle: Row(
                      children: [
                        if (c.manuscriptTitle.isNotEmpty)
                          Flexible(child: Text(c.manuscriptTitle, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 11, color: Colors.grey))),
                        const SizedBox(width: 8),
                        Text(c.likeCount > 0 ? '${c.likeCount}赞' : '', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                      ],
                    ),
                    trailing: PopupMenuButton<String>(
                      icon: const Icon(Icons.more_vert, color: Colors.grey, size: 18),
                      onSelected: (value) async {
                        if (value == 'delete') {
                          await ref.read(creatorApiProvider).deleteComment(c.id);
                          ref.invalidate(creatorCommentsProvider);
                        } else if (value == 'reply') {
                          _showReplyDialog(context, ref, c.id);
                        }
                      },
                      itemBuilder: (context) => [
                        const PopupMenuItem(value: 'reply', child: Text('回复')),
                        const PopupMenuItem(value: 'delete', child: Text('删除', style: TextStyle(color: Colors.red))),
                      ],
                    ),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }

  void _showReplyDialog(BuildContext context, WidgetRef ref, int commentId) {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('回复评论'),
        content: TextField(
          controller: controller,
          autofocus: true,
          maxLines: 3,
          decoration: InputDecoration(
            filled: true, fillColor: const Color(0xFF2A2A2A),
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(dialogContext), child: const Text('取消')),
          TextButton(
            onPressed: () async {
              if (controller.text.trim().isNotEmpty) {
                await ref.read(creatorApiProvider).replyComment(commentId, controller.text.trim());
                if (dialogContext.mounted) Navigator.pop(dialogContext);
                ref.invalidate(creatorCommentsProvider);
              }
            },
            child: const Text('回复', style: TextStyle(color: AppTheme.primaryPink)),
          ),
        ],
      ),
    );
  }
}