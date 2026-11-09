import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/live_api.dart';
import '../../core/theme/theme.dart';
import 'live_room_page.dart';

final liveRoomsProvider = FutureProvider.autoDispose<List<LiveRoom>>((ref) {
  return ref.read(liveApiProvider).getRooms();
});

class LiveListPage extends ConsumerWidget {
  const LiveListPage({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final rooms = ref.watch(liveRoomsProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('直播')),
      body: rooms.when(
        data: (list) => list.isEmpty
            ? const Center(child: Text('暂无直播', style: TextStyle(color: Colors.grey)))
            : RefreshIndicator(
                onRefresh: () async => ref.invalidate(liveRoomsProvider),
                child: GridView.builder(
                  padding: const EdgeInsets.all(8),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2, childAspectRatio: 0.75, crossAxisSpacing: 8, mainAxisSpacing: 8,
                  ),
                  itemCount: list.length,
                  itemBuilder: (context, index) => _LiveCard(room: list[index]),
                ),
              ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('加载失败: $e', style: const TextStyle(color: Colors.grey))),
      ),
    );
  }
}

class _LiveCard extends StatelessWidget {
  final LiveRoom room;
  const _LiveCard({required this.room});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () => Navigator.of(context).push(MaterialPageRoute(
        builder: (_) => LiveRoomPage(roomId: room.id),
      )),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Stack(
              fit: StackFit.expand,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: room.cover.isNotEmpty
                      ? Image.network(room.cover, fit: BoxFit.cover, errorBuilder: (_, _, _) => Container(color: Colors.grey[800]))
                      : Container(color: Colors.grey[800], child: const Icon(Icons.live_tv, color: Colors.grey, size: 32)),
                ),
                Positioned(
                  left: 4, top: 4,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(
                      color: room.isLive ? AppTheme.primaryPink : Colors.grey,
                      borderRadius: BorderRadius.circular(4),
                    ),
                    child: Text(room.isLive ? '直播中' : '未开播', style: const TextStyle(color: Colors.white, fontSize: 10)),
                  ),
                ),
                Positioned(
                  right: 4, bottom: 4,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(color: Colors.black54, borderRadius: BorderRadius.circular(4)),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.visibility, size: 12, color: Colors.white),
                        const SizedBox(width: 2),
                        Text('${room.viewerCount}', style: const TextStyle(color: Colors.white, fontSize: 10)),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 6),
          Text(room.roomName, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13)),
          if (room.category.isNotEmpty)
            Text(room.category, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 11, color: Colors.grey)),
        ],
      ),
    );
  }
}