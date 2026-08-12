import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../features/home/home_page.dart';
import '../../features/home/dynamic_page.dart';
import '../../features/home/hot_page.dart';
import '../../features/auth/login_page.dart';
import '../../features/profile/profile_page.dart';
import '../../features/search/search_page.dart';
import '../../features/user/user_page.dart';
import '../../features/dynamic/dynamic_page.dart' as dyn;
import '../../features/message/message_page.dart';
import '../../features/follow/follow_list_page.dart';
import '../../features/live/live_list_page.dart';
import '../../features/creator/creator_center_page.dart';
import '../../features/video/screens/video_detail_screen.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/home',
    routes: [
      ShellRoute(
        builder: (context, state, child) => AppShell(child: child),
        routes: [
          GoRoute(path: '/home', pageBuilder: (context, state) => const NoTransitionPage(child: HomePage())),
          GoRoute(path: '/hot', pageBuilder: (context, state) => const NoTransitionPage(child: LiveListPage())),
          GoRoute(path: '/dynamic', pageBuilder: (context, state) => const NoTransitionPage(child: dyn.DynamicPage())),
          GoRoute(path: '/profile', pageBuilder: (context, state) => const NoTransitionPage(child: ProfilePage())),
        ],
      ),
      GoRoute(path: '/login', builder: (context, state) => const LoginPage()),
      GoRoute(path: '/video/:id', builder: (context, state) {
        final id = int.parse(state.pathParameters['id']!);
        return VideoDetailScreen(manuscriptId: id);
      }),
      GoRoute(path: '/user/:id', builder: (context, state) {
        final id = int.parse(state.pathParameters['id']!);
        return UserPage(userId: id);
      }),
      GoRoute(
        path: '/search',
        builder: (context, state) => const SearchPage(),
        pageBuilder: (context, state) => CustomTransitionPage(
          child: const SearchPage(),
          transitionsBuilder: (context, animation, secondaryAnimation, child) =>
              SlideTransition(position: Tween<Offset>(begin: const Offset(1, 0), end: Offset.zero).animate(animation), child: child),
        ),
      ),
      GoRoute(path: '/message', builder: (context, state) => const MessagePage()),
      GoRoute(path: '/live', builder: (context, state) => const LiveListPage()),
      GoRoute(path: '/creator', builder: (context, state) => const CreatorCenterPage()),
      GoRoute(path: '/follow/followers', builder: (context, state) => const FollowListPage(showFollowers: true)),
      GoRoute(path: '/follow/following', builder: (context, state) => const FollowListPage(showFollowers: false)),
    ],
  );
});

class AppShell extends StatelessWidget {
  final Widget child;
  const AppShell({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: child,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _calculateIndex(context),
        onTap: (index) => _onTap(context, index),
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.home), label: '首页'),
          BottomNavigationBarItem(icon: Icon(Icons.live_tv), label: '直播'),
          BottomNavigationBarItem(icon: Icon(Icons.dynamic_feed), label: '动态'),
          BottomNavigationBarItem(icon: Icon(Icons.person), label: '我的'),
        ],
      ),
    );
  }

  int _calculateIndex(BuildContext context) {
    final location = GoRouterState.of(context).uri.toString();
    if (location.startsWith('/home')) return 0;
    if (location.startsWith('/hot') || location.startsWith('/live')) return 1;
    if (location.startsWith('/dynamic')) return 2;
    if (location.startsWith('/profile')) return 3;
    return 0;
  }

  void _onTap(BuildContext context, int index) {
    switch (index) {
      case 0: context.go('/home');
      case 1: context.go('/live');
      case 2: context.go('/dynamic');
      case 3: context.go('/profile');
    }
  }
}