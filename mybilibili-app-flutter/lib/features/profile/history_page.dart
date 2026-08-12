import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/watch_history_api.dart';
import '../../core/theme/theme.dart';
import '../video/screens/video_detail_screen.dart';

final historyProvider = FutureProvider.autoDispose<List<WatchHistoryItem>>((ref) {
  return ref.read(watchHistoryApiProvider).getHistory();
});

class HistoryPage extends ConsumerWidget {
  const HistoryPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final history = ref.watch(historyProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('历史记录'),
        actions: [
          IconButton(
            icon: const Icon(Icons.delete_sweep),
            onPressed: () async {
              await ref.read(watchHistoryApiProvider).clearHistory();
              ref.invalidate(historyProvider);
            },
          ),
        ],
      ),
      body: history.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无观看历史', style: TextStyle(color: Colors.grey)))
            : ListView.separated(
                padding: const EdgeInsets.all(8),
                itemCount: list.length,
                separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
                itemBuilder: (context, index) {
                  final h = list[index];
                  return ListTile(
                    leading: ClipRRect(
                      borderRadius: BorderRadius.circular(4),
                      child: SizedBox(
                        width: 100,
                        height: 56,
                        child: Image.network(h.coverUrl, fit: BoxFit.cover,
                          errorBuilder: (_, _, _) => Container(color: Colors.grey[800]),
                        ),
                      ),
                    ),
                    title: Text(h.title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 14)),
                    subtitle: Row(
                      children: [
                        Text(h.uploaderName, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                        if (h.progressSeconds > 0) ...[
                          const SizedBox(width: 8),
                          Text('${h.progressSeconds ~/ 60}分钟前', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                        ],
                      ],
                    ),
                    trailing: IconButton(
                      icon: const Icon(Icons.close, size: 18),
                      onPressed: () async {
                        await ref.read(watchHistoryApiProvider).deleteHistory(h.id);
                        ref.invalidate(historyProvider);
                      },
                    ),
                    onTap: () => Navigator.of(context).push(MaterialPageRoute(
                      builder: (_) => VideoDetailScreen(manuscriptId: h.manuscriptId),
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