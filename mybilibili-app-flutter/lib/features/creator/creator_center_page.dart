import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/creator_api.dart';
import '../../core/theme/theme.dart';
import 'creator_manuscripts_page.dart';
import 'creator_comments_page.dart';
import 'creator_stats_page.dart';

final overviewProvider = FutureProvider.autoDispose<CreatorOverview>((ref) {
  return ref.read(creatorApiProvider).getOverview();
});

class CreatorCenterPage extends ConsumerWidget {
  const CreatorCenterPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final overview = ref.watch(overviewProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('创作中心')),
      body: ListView(
        children: [
          overview.when(
            data: (o) => Container(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('数据概览', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  const SizedBox(height: 12),
                  GridView.count(
                    shrinkWrap: true,
                    crossAxisCount: 3,
                    childAspectRatio: 1.2,
                    crossAxisSpacing: 8,
                    mainAxisSpacing: 8,
                    children: [
                      _statCard('稿件', o.manuscriptCount, Icons.video_library),
                      _statCard('播放', o.viewCount, Icons.visibility),
                      _statCard('点赞', o.likeCount, Icons.thumb_up),
                      _statCard('投币', o.coinCount, Icons.monetization_on),
                      _statCard('收藏', o.collectCount, Icons.star),
                      _statCard('粉丝', o.followerCount, Icons.people),
                    ],
                  ),
                ],
              ),
            ),
            loading: () => const Padding(padding: EdgeInsets.all(24), child: Center(child: CircularProgressIndicator())),
            error: (e, _) => Padding(padding: const EdgeInsets.all(24), child: Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey)))),
          ),
          const Divider(height: 1, color: Color(0xFF2A2A2A)),
          _menuItem(Icons.video_library, '稿件管理', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const CreatorManuscriptsPage()))),
          _menuItem(Icons.comment, '评论管理', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const CreatorCommentsPage()))),
          _menuItem(Icons.bar_chart, '数据中心', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const CreatorStatsPage()))),
        ],
      ),
    );
  }

  Widget _statCard(String label, int count, IconData icon) {
    String display;
    if (count >= 10000) {
      display = '${(count / 10000).toStringAsFixed(1)}万';
    } else {
      display = '$count';
    }
    return Container(
      decoration: BoxDecoration(color: AppTheme.cardDark, borderRadius: BorderRadius.circular(8)),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(icon, size: 20, color: AppTheme.primaryPink),
          const SizedBox(height: 4),
          Text(display, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 11, color: Colors.grey)),
        ],
      ),
    );
  }

  Widget _menuItem(IconData icon, String label, VoidCallback onTap) {
    return ListTile(
      leading: Icon(icon, color: Colors.grey, size: 22),
      title: Text(label, style: const TextStyle(fontSize: 14)),
      trailing: const Icon(Icons.chevron_right, color: Colors.grey, size: 18),
      onTap: onTap,
    );
  }
}