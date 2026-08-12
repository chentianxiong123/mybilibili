import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/theme/theme.dart';
import '../../core/api/user_api.dart';
import 'history_page.dart';
import 'favorites_page.dart';
import 'manuscripts_page.dart';
import 'edit_profile_page.dart';
import '../user/user_page.dart';

final profileProvider = FutureProvider.autoDispose<UserProfile>((ref) {
  return ref.read(userApiProvider).getCurrentUserProfile();
});

class ProfilePage extends ConsumerWidget {
  const ProfilePage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final profile = ref.watch(profileProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('我的')),
      body: ListView(
        children: [
          profile.when(
            data: (u) => InkWell(
              onTap: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => UserPage(userId: u.id),
              )),
              child: Container(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    CircleAvatar(
                      radius: 32,
                      backgroundImage: u.avatar.isNotEmpty ? NetworkImage(u.avatar) : null,
                      child: u.avatar.isEmpty ? const Icon(Icons.person, size: 32) : null,
                    ),
                    const SizedBox(width: 12),
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(u.nickname, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                        const SizedBox(height: 4),
                        Text('${u.followerCount} 粉丝', style: const TextStyle(fontSize: 13, color: Colors.grey)),
                      ],
                    ),
                    const Spacer(),
                    const Icon(Icons.chevron_right, color: Colors.grey),
                  ],
                ),
              ),
            ),
            loading: () => const Padding(padding: EdgeInsets.all(16), child: Center(child: CircularProgressIndicator())),
            error: (_, __) => Container(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  const CircleAvatar(radius: 32, child: Icon(Icons.person, size: 32)),
                  const SizedBox(width: 12),
                  const Text('未登录', style: TextStyle(fontSize: 18)),
                  const Spacer(),
                  TextButton(
                    onPressed: () => Navigator.of(context).pushNamed('/login'),
                    child: const Text('登录', style: TextStyle(color: AppTheme.primaryPink)),
                  ),
                ],
              ),
            ),
          ),
          const Divider(height: 1, color: Color(0xFF2A2A2A)),
          _menuItem(Icons.history, '历史记录', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const HistoryPage()))),
          _menuItem(Icons.folder, '收藏夹', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const FavoritesPage()))),
          _menuItem(Icons.video_library, '我的投稿', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const ManuscriptsPage()))),
          _menuItem(Icons.edit, '编辑资料', () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const EditProfilePage()))),
          _menuItem(Icons.mail, '消息中心', () => Navigator.of(context).pushNamed('/message')),
          _menuItem(Icons.people, '关注列表', () => Navigator.of(context).pushNamed('/follow/following')),
          _menuItem(Icons.people_outline, '粉丝列表', () => Navigator.of(context).pushNamed('/follow/followers')),
          _menuItem(Icons.settings, '设置', () {}),
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