import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/user_api.dart';
import '../../shared/models/manuscript.dart';
import '../home/widgets/video_card.dart';
import '../video/screens/video_detail_screen.dart';

final userPageProvider = FutureProvider.family<UserProfile, int>((ref, userId) {
  return ref.read(userApiProvider).getUserProfile(userId);
});

final userManuscriptsProvider = FutureProvider.family<List<ManuscriptInfo>, int>((ref, userId) {
  return ref.read(userApiProvider).getUserManuscripts(userId);
});

class UserPage extends ConsumerWidget {
  final int userId;
  const UserPage({super.key, required this.userId});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final user = ref.watch(userPageProvider(userId));
    final manuscripts = ref.watch(userManuscriptsProvider(userId));
    return Scaffold(
      appBar: AppBar(title: const Text('UP主主页')),
      body: user.when(
        data: (u) => CustomScrollView(
          slivers: [
            SliverToBoxAdapter(child: _buildUserHeader(u)),
            SliverToBoxAdapter(child: _buildStats(u)),
            const SliverToBoxAdapter(child: SizedBox(height: 8)),
            const SliverToBoxAdapter(child: Padding(
              padding: EdgeInsets.symmetric(horizontal: 16),
              child: Text('投稿', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            )),
            manuscripts.when(
              data: (list) => SliverPadding(
                padding: const EdgeInsets.all(8),
                sliver: SliverGrid(
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2, childAspectRatio: 0.7, crossAxisSpacing: 8, mainAxisSpacing: 8,
                  ),
                  delegate: SliverChildBuilderDelegate((context, index) {
                    final m = list[index];
                    return VideoCard(
                      manuscript: m,
                      onTap: () => Navigator.of(context).push(MaterialPageRoute(
                        builder: (_) => VideoDetailScreen(manuscriptId: m.id),
                      )),
                    );
                  }, childCount: list.length),
                ),
              ),
              loading: () => const SliverFillRemaining(child: Center(child: CircularProgressIndicator())),
              error: (_, __) => const SliverFillRemaining(child: Center(child: Text('加载失败'))),
            ),
          ],
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e')),
      ),
    );
  }

  Widget _buildUserHeader(UserProfile u) {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          CircleAvatar(
            radius: 36,
            backgroundImage: u.avatar.isNotEmpty ? NetworkImage(u.avatar) : null,
            child: u.avatar.isEmpty ? const Icon(Icons.person, size: 36) : null,
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(u.nickname, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                if (u.introduction.isNotEmpty) ...[
                  const SizedBox(height: 4),
                  Text(u.introduction, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13, color: Colors.grey)),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStats(UserProfile u) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        children: [
          _statItem('${u.followerCount}', '粉丝'),
          _statItem('${u.followingCount}', '关注'),
          _statItem('${u.manuscriptCount}', '稿件'),
          _statItem('${u.likeCount}', '获赞'),
        ],
      ),
    );
  }

  Widget _statItem(String count, String label) {
    return Expanded(
      child: Column(
        children: [
          Text(count, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 11, color: Colors.grey)),
        ],
      ),
    );
  }
}