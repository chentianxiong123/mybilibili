import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/manuscript_api.dart';
import '../../shared/models/manuscript.dart';
import '../video/screens/video_detail_screen.dart';
import 'widgets/video_card.dart';

final hotDataProvider = FutureProvider.autoDispose<List<ManuscriptInfo>>((ref) async {
  return ref.read(manuscriptApiProvider).getHot();
});

class HotPage extends ConsumerWidget {
  const HotPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final data = ref.watch(hotDataProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('热门')),
      body: data.when(
        data: (manuscripts) => RefreshIndicator(
          onRefresh: () async => ref.invalidate(hotDataProvider),
          child: manuscripts.isEmpty
              ? const Center(child: Text('暂无内容', style: TextStyle(color: Colors.grey)))
              : GridView.builder(
                  padding: const EdgeInsets.all(8),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2, childAspectRatio: 0.7, crossAxisSpacing: 8, mainAxisSpacing: 8,
                  ),
                  itemCount: manuscripts.length,
                  itemBuilder: (context, index) => VideoCard(
                    manuscript: manuscripts[index],
                    onTap: () => Navigator.of(context).push(MaterialPageRoute(
                      builder: (_) => VideoDetailScreen(manuscriptId: manuscripts[index].id),
                    )),
                  ),
                ),
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.cloud_off, size: 48, color: Colors.grey),
              const SizedBox(height: 16),
              Text('$e', style: const TextStyle(color: Colors.grey)),
              const SizedBox(height: 16),
              ElevatedButton(onPressed: () => ref.invalidate(hotDataProvider), child: const Text('重试')),
            ],
          ),
        ),
      ),
    );
  }
}