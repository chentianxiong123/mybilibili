import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/favorite_api.dart';
import '../../core/theme/theme.dart';
import '../video/screens/video_detail_screen.dart';

final foldersProvider = FutureProvider.autoDispose<List<FavoriteFolder>>((ref) {
  return ref.read(favoriteApiProvider).getFolders();
});

class FavoritesPage extends ConsumerWidget {
  const FavoritesPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final folders = ref.watch(foldersProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('收藏夹')),
      body: folders.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无收藏夹', style: TextStyle(color: Colors.grey)))
            : ListView.builder(
                padding: const EdgeInsets.all(8),
                itemCount: list.length,
                itemBuilder: (context, index) {
                  final f = list[index];
                  return Card(
                    color: AppTheme.cardDark,
                    child: ListTile(
                      leading: f.coverUrl.isNotEmpty
                          ? ClipRRect(borderRadius: BorderRadius.circular(4), child: SizedBox(width: 60, height: 60, child: Image.network(f.coverUrl, fit: BoxFit.cover, errorBuilder: (_, _, _) => Container(color: Colors.grey[800]))))
                          : Container(width: 60, height: 60, color: Colors.grey[800], child: const Icon(Icons.folder, color: Colors.grey)),
                      title: Text(f.name, style: const TextStyle(fontSize: 14)),
                      subtitle: Text('${f.manuscriptCount}个内容', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                      trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                      onTap: () => Navigator.of(context).push(MaterialPageRoute(
                        builder: (_) => _FolderDetailPage(folderId: f.id, folderName: f.name),
                      )),
                    ),
                  );
                },
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }
}

class _FolderDetailPage extends ConsumerWidget {
  final int folderId;
  final String folderName;
  const _FolderDetailPage({required this.folderId, required this.folderName});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final manuscripts = ref.watch(FutureProvider.autoDispose((ref) {
      return ref.read(favoriteApiProvider).getFolderManuscripts(folderId);
    }));
    return Scaffold(
      appBar: AppBar(title: Text(folderName)),
      body: manuscripts.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('收藏夹为空', style: TextStyle(color: Colors.grey)))
            : GridView.builder(
                padding: const EdgeInsets.all(8),
                gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: 2, childAspectRatio: 0.7, crossAxisSpacing: 8, mainAxisSpacing: 8,
                ),
                itemCount: list.length,
                itemBuilder: (context, index) => GestureDetector(
                  onTap: () => Navigator.of(context).push(MaterialPageRoute(
                    builder: (_) => VideoDetailScreen(manuscriptId: list[index].id),
                  )),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: ClipRRect(
                          borderRadius: BorderRadius.circular(8),
                          child: Image.network(list[index].coverUrl, fit: BoxFit.cover,
                            errorBuilder: (_, _, _) => Container(color: Colors.grey[800], child: const Icon(Icons.movie, color: Colors.grey))),
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(list[index].title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13)),
                    ],
                  ),
                ),
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (_, __) => const SizedBox.shrink(),
      ),
    );
  }
}