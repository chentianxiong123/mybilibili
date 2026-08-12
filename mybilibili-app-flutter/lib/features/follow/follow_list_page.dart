import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/follow_api.dart';
import '../../core/theme/theme.dart';
import '../user/user_page.dart';

class FollowListPage extends ConsumerWidget {
  final bool showFollowers;
  const FollowListPage({super.key, required this.showFollowers});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final provider = FutureProvider.autoDispose((ref) {
      return showFollowers
          ? ref.read(followApiProvider).getFollowers()
          : ref.read(followApiProvider).getFollowing();
    });
    final data = ref.watch(provider);
    return Scaffold(
      appBar: AppBar(title: Text(showFollowers ? '粉丝' : '关注')),
      body: data.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无数据', style: TextStyle(color: Colors.grey)))
            : ListView.separated(
                padding: const EdgeInsets.all(8),
                itemCount: list.length,
                separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
                itemBuilder: (context, index) {
                  final u = list[index] as dynamic;
                  return ListTile(
                    leading: CircleAvatar(
                      backgroundImage: u.avatar.isNotEmpty ? NetworkImage(u.avatar) : null,
                      child: u.avatar.isEmpty ? const Icon(Icons.person) : null,
                    ),
                    title: Text(u.nickname, style: const TextStyle(fontSize: 14)),
                    subtitle: Text(u.introduction, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 12, color: Colors.grey)),
                    trailing: _FollowButton(userId: u.id),
                    onTap: () => Navigator.of(context).push(MaterialPageRoute(
                      builder: (_) => UserPage(userId: u.id),
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

class _FollowButton extends ConsumerStatefulWidget {
  final int userId;
  const _FollowButton({required this.userId});
  @override
  ConsumerState<_FollowButton> createState() => _FollowButtonState();
}

class _FollowButtonState extends ConsumerState<_FollowButton> {
  bool _following = false;
  bool _loading = false;

  @override
  void initState() {
    super.initState();
    _checkFollow();
  }

  Future<void> _checkFollow() async {
    try {
      final following = await ref.read(followApiProvider).checkFollow(widget.userId);
      if (mounted) setState(() => _following = following);
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 32,
      child: ElevatedButton(
        onPressed: _loading ? null : () async {
          setState(() => _loading = true);
          try {
            if (_following) { await ref.read(followApiProvider).unfollow(widget.userId); setState(() => _following = false); }
            else { await ref.read(followApiProvider).follow(widget.userId); setState(() => _following = true); }
          } catch (_) {}
          if (mounted) setState(() => _loading = false);
        },
        style: ElevatedButton.styleFrom(
          backgroundColor: _following ? const Color(0xFF2A2A2A) : AppTheme.primaryPink,
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(horizontal: 16),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        ),
        child: Text(_following ? '已关注' : '+ 关注', style: const TextStyle(fontSize: 12)),
      ),
    );
  }
}