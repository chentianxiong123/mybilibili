import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/user_api.dart';
import '../../shared/models/manuscript.dart';
import '../home/widgets/video_card.dart';
import '../video/screens/video_detail_screen.dart';

final myManuscriptsProvider = FutureProvider.autoDispose<List<ManuscriptInfo>>((ref) {
  return ref.read(userApiProvider).getMyManuscripts();
});

class ManuscriptsPage extends ConsumerWidget {
  const ManuscriptsPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final manuscripts = ref.watch(myManuscriptsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('我的投稿')),
      body: manuscripts.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无投稿', style: TextStyle(color: Colors.grey)))
            : GridView.builder(
                padding: const EdgeInsets.all(8),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2, childAspectRatio: 0.7, crossAxisSpacing: 8, mainAxisSpacing: 8,
                ),
                itemCount: list.length,
                itemBuilder: (context, index) => VideoCard(
                  manuscript: list[index],
                  onTap: () => Navigator.of(context).push(MaterialPageRoute(
                    builder: (_) => VideoDetailScreen(manuscriptId: list[index].id),
                  )),
                ),
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }
}