import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/creator_api.dart';
import '../../core/theme/theme.dart';

final trendProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.read(creatorApiProvider).getTrend(days: 7);
});

final rankingProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) {
  return ref.read(creatorApiProvider).getRanking(sortBy: 'view_count', limit: 10);
});

class CreatorStatsPage extends ConsumerWidget {
  const CreatorStatsPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final trend = ref.watch(trendProvider);
    final ranking = ref.watch(rankingProvider);
    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: const Text('数据中心'),
          bottom: const TabBar(
            tabs: [Tab(text: '趋势'), Tab(text: '排行')],
            labelColor: AppTheme.primaryPink,
            unselectedLabelColor: Colors.grey,
            indicatorColor: AppTheme.primaryPink,
          ),
        ),
        body: TabBarView(
          children: [
            trend.when(
              data: (list) => list.isEmpty
                  ? const Center(child: Text('暂无数据', style: TextStyle(color: Colors.grey)))
                  : ListView.builder(
                      padding: const EdgeInsets.all(16),
                      itemCount: list.length,
                      itemBuilder: (context, index) {
                        final item = list[index];
                        return Card(
                          color: AppTheme.cardDark,
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(item['date']?.toString() ?? '', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
                                const SizedBox(height: 8),
                                Wrap(
                                  spacing: 12, runSpacing: 4,
                                  children: [
                                    _metric('播放', item['view_count']),
                                    _metric('点赞', item['like_count']),
                                    _metric('投币', item['coin_count']),
                                    _metric('收藏', item['collect_count']),
                                    _metric('评论', item['comment_count']),
                                    _metric('弹幕', item['danmaku_count']),
                                  ],
                                ),
                              ],
                            ),
                          ),
                        );
                      },
                    ),
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
            ),
            ranking.when(
              data: (list) => list.isEmpty
                  ? const Center(child: Text('暂无数据', style: TextStyle(color: Colors.grey)))
                  : ListView.builder(
                      padding: const EdgeInsets.all(8),
                      itemCount: list.length,
                      itemBuilder: (context, index) {
                        final item = list[index];
                        return ListTile(
                          leading: CircleAvatar(
                            backgroundColor: index < 3 ? AppTheme.primaryPink : const Color(0xFF2A2A2A),
                            child: Text('${index + 1}', style: TextStyle(color: index < 3 ? Colors.white : Colors.grey, fontSize: 14)),
                          ),
                          title: Text(item['title']?.toString() ?? '', maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 14)),
                          subtitle: Text('${item['view_count'] ?? 0}播放', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                        );
                      },
                    ),
              loading: () => const Center(child: CircularProgressIndicator()),
              error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
            ),
          ],
        ),
      ),
    );
  }

  Widget _metric(String label, dynamic value) {
    return Text('$label: ${value ?? 0}', style: const TextStyle(fontSize: 12, color: Colors.grey));
  }
}