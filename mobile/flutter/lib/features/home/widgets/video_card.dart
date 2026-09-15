import 'package:flutter/material.dart';
import '../../../shared/models/manuscript.dart';

String formatCount(int count) {
  if (count >= 100000000) return '${(count / 100000000).toStringAsFixed(1)}亿';
  if (count >= 10000) return '${(count / 10000).toStringAsFixed(1)}万';
  return '$count';
}

class VideoCard extends StatelessWidget {
  final ManuscriptInfo manuscript;
  final VoidCallback? onTap;
  final bool showStatistics;

  const VideoCard({super.key, required this.manuscript, this.onTap, this.showStatistics = true});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          ClipRRect(
            borderRadius: BorderRadius.circular(8),
            child: AspectRatio(
              aspectRatio: 16 / 9,
              child: Stack(
                fit: StackFit.expand,
                children: [
                  Image.network(manuscript.coverUrl, fit: BoxFit.cover,
                    errorBuilder: (_, _, _) => Container(color: Colors.grey[800], child: const Icon(Icons.movie, color: Colors.grey))),
                  if (manuscript.duration.isNotEmpty)
                    Positioned(
                      right: 4, bottom: 4,
                      child: Container(
                        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                        decoration: BoxDecoration(color: Colors.black54, borderRadius: BorderRadius.circular(4)),
                        child: Text(manuscript.duration, style: const TextStyle(color: Colors.white, fontSize: 11)),
                      ),
                    ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 6),
          Text(manuscript.title, maxLines: 2, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 13, height: 1.3)),
          if (showStatistics) ...[
            const SizedBox(height: 4),
            Row(
              children: [
                if (manuscript.uploader?.nickname.isNotEmpty ?? false)
                  Expanded(
                    child: Text(manuscript.uploader!.nickname, maxLines: 1, overflow: TextOverflow.ellipsis, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                  ),
                Text('${formatCount(manuscript.viewCount)}观看', style: const TextStyle(fontSize: 11, color: Colors.grey)),
              ],
            ),
          ],
        ],
      ),
    );
  }
}