import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/user_api.dart';
import '../../shared/models/manuscript.dart';
import '../../core/theme/theme.dart';
import '../home/widgets/video_card.dart';
import '../video/screens/video_detail_screen.dart';

final creatorManuscriptsProvider = FutureProvider.autoDispose<List<ManuscriptInfo>>((ref) {
  return ref.read(userApiProvider).getMyManuscripts();
});

class CreatorManuscriptsPage extends ConsumerWidget {
  const CreatorManuscriptsPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final manuscripts = ref.watch(creatorManuscriptsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('稿件管理')),
      body: manuscripts.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无投稿', style: TextStyle(color: Colors.grey)))
            : ListView.separated(
                padding: const EdgeInsets.all(8),
                itemCount: list.length,
                separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
                itemBuilder: (context, index) {
                  final m = list[index];
                  return _ManuscriptTile(manuscript: m, onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => VideoDetailScreen(manuscriptId: m.id))));
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }
}

class _ManuscriptTile extends StatelessWidget {
  final ManuscriptInfo manuscript;
  final VoidCallback onTap;
  const _ManuscriptTile({required this.manuscript, required this.onTap});

  String get _statusText {
    switch (manuscript.status) {
      case 0: return '审核中';
      case 1: return '已发布';
      case 2: return '未通过';
      default: return '未知';
    }
  }

  Color get _statusColor {
    switch (manuscript.status) {
      case 0: return Colors.orange;
      case 1: return AppTheme.primaryPink;
      case 2: return Colors.red;
      default: return Colors.grey;
    }
  }

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: ClipRRect(
        borderRadius: BorderRadius.circular(4),
        child: SizedBox(
          width: 100, height: 56,
          child: Image.network(manuscript.coverUrl, fit: BoxFit.cover, errorBuilder: (_, _, _) => Container(color: Colors.grey[800])),
        ),
      ),
      title: Text(manuscript.title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 14)),
      subtitle: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
            decoration: BoxDecoration(color: _statusColor.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(4)),
            child: Text(_statusText, style: TextStyle(fontSize: 10, color: _statusColor)),
          ),
          const SizedBox(width: 8),
          Text('${manuscript.viewCount}播放', style: const TextStyle(fontSize: 11, color: Colors.grey)),
          const SizedBox(width: 8),
          Text('${manuscript.likeCount}赞', style: const TextStyle(fontSize: 11, color: Colors.grey)),
        ],
      ),
      trailing: const Icon(Icons.chevron_right, color: Colors.grey),
      onTap: onTap,
    );
  }
}