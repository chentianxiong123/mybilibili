import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/dynamic_api.dart';
import '../../core/theme/theme.dart';
import '../../shared/models/manuscript.dart';
import '../video/screens/video_detail_screen.dart';

final dynamicsProvider = FutureProvider.autoDispose<List<DynamicItem>>((ref) {
  return ref.read(dynamicApiProvider).getFollowingDynamics();
});

class DynamicPage extends ConsumerWidget {
  const DynamicPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final dynamics = ref.watch(dynamicsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('动态')),
      body: dynamics.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无动态', style: TextStyle(color: Colors.grey)))
            : RefreshIndicator(
                onRefresh: () async => ref.invalidate(dynamicsProvider),
                child: ListView.separated(
                  padding: const EdgeInsets.all(8),
                  itemCount: list.length,
                  separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
                  itemBuilder: (context, index) => _DynamicCard(item: list[index], ref: ref),
                ),
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }
}

class _DynamicCard extends ConsumerStatefulWidget {
  final DynamicItem item;
  final WidgetRef ref;
  const _DynamicCard({required this.item, required this.ref});
  @override
  ConsumerState<_DynamicCard> createState() => _DynamicCardState();
}

class _DynamicCardState extends ConsumerState<_DynamicCard> {
  late bool _liked = widget.item.liked;
  late int _likeCount = widget.item.likeCount;

  @override
  Widget build(BuildContext context) {
    final item = widget.item;
    return Card(
      color: AppTheme.cardDark,
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                CircleAvatar(
                  radius: 18,
                  backgroundImage: item.userAvatar.isNotEmpty ? NetworkImage(item.userAvatar) : null,
                  child: item.userAvatar.isEmpty ? const Icon(Icons.person, size: 18) : null,
                ),
                const SizedBox(width: 8),
                Expanded(child: Text(item.username, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500))),
                Text(_formatTime(item.createdAt), style: const TextStyle(fontSize: 11, color: Colors.grey)),
              ],
            ),
            const SizedBox(height: 8),
            if (item.content.isNotEmpty)
              Text(item.content, style: const TextStyle(fontSize: 14, height: 1.4)),
            if (item.refManuscriptId > 0) ...[
              const SizedBox(height: 8),
              GestureDetector(
                onTap: () => Navigator.of(context).push(MaterialPageRoute(
                  builder: (_) => VideoDetailScreen(manuscriptId: item.refManuscriptId),
                )),
                child: Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(color: const Color(0xFF181818), borderRadius: BorderRadius.circular(8)),
                  child: Row(
                    children: [
                      if (item.refManuscriptCover.isNotEmpty)
                        ClipRRect(
                          borderRadius: BorderRadius.circular(4),
                          child: SizedBox(width: 80, height: 50, child: Image.network(item.refManuscriptCover, fit: BoxFit.cover, errorBuilder: (_, _, _) => Container(color: Colors.grey[800]))),
                        ),
                      const SizedBox(width: 8),
                      Expanded(child: Text(item.refManuscriptTitle, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13, color: Colors.grey))),
                    ],
                  ),
                ),
              ),
            ],
            const SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _actionBtn(Icons.thumb_up, _liked, _likeCount, () async {
                  try {
                    if (_liked) { await ref.read(dynamicApiProvider).unlike(item.id); setState(() { _liked = false; _likeCount--; }); }
                    else { await ref.read(dynamicApiProvider).like(item.id); setState(() { _liked = true; _likeCount++; }); }
                  } catch (_) {}
                }),
                _actionBtn(Icons.comment, false, item.commentCount, () {}),
                _actionBtn(Icons.share, false, item.shareCount, () {}),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _actionBtn(IconData icon, bool active, int count, VoidCallback onTap) {
    return GestureDetector(
      onTap: onTap,
      child: Row(
        children: [
          Icon(icon, size: 18, color: active ? AppTheme.primaryPink : Colors.grey),
          const SizedBox(width: 4),
          Text('$count', style: TextStyle(fontSize: 12, color: active ? AppTheme.primaryPink : Colors.grey)),
        ],
      ),
    );
  }

  String _formatTime(String t) {
    try { return t.substring(5, 16).replaceFirst('T', ' '); } catch (_) { return t; }
  }
}