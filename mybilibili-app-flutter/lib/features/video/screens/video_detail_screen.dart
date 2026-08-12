import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/manuscript_api.dart';
import '../../../shared/models/manuscript.dart';
import '../player/widgets/player_view.dart';

class VideoDetailScreen extends ConsumerStatefulWidget {
  final int manuscriptId;
  const VideoDetailScreen({super.key, required this.manuscriptId});

  @override
  ConsumerState<VideoDetailScreen> createState() => _VideoDetailScreenState();
}

class _VideoDetailScreenState extends ConsumerState<VideoDetailScreen> {
  ManuscriptInfo? _manuscript;
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadManuscript();
  }

  Future<void> _loadManuscript() async {
    try {
      final manuscript = await ref.read(manuscriptApiProvider).getManuscript(widget.manuscriptId);
      if (mounted) {
        setState(() {
          _manuscript = manuscript;
          _loading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _loading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text('加载失败: $_error'))
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    if (_manuscript == null) return const SizedBox.shrink();
    final m = _manuscript!;
    final videoUrl = m.videos.isNotEmpty
        ? m.videos.first.bestPlayUrl
        : m.firstVideoPlayUrl;

    return Column(
      children: [
        AspectRatio(
          aspectRatio: 16 / 9,
          child: BilibiliPlayer(
            url: videoUrl,
            title: m.title,
            danmakuList: [],
            showBackButton: true,
            onBack: () => Navigator.of(context).pop(),
          ),
        ),
        Expanded(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(m.title, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                const SizedBox(height: 8),
                Row(
                  children: [
                    _buildStat(Icons.visibility, '${m.viewCount}'),
                    const SizedBox(width: 16),
                    _buildStat(Icons.thumb_up, '${m.likeCount}'),
                    const SizedBox(width: 16),
                    _buildStat(Icons.message, '${m.commentCount}'),
                  ],
                ),
                if (m.uploader != null) ...[
                  const SizedBox(height: 16),
                  ListTile(
                    leading: CircleAvatar(
                      backgroundImage: NetworkImage(m.uploader!.avatar),
                    ),
                    title: Text(m.uploader!.nickname),
                    subtitle: Text('${m.uploader!.followerCount} 粉丝'),
                  ),
                ],
                if (m.description.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  Text(m.description, style: const TextStyle(color: Colors.grey)),
                ],
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildStat(IconData icon, String text) {
    return Row(
      children: [
        Icon(icon, size: 16, color: Colors.grey),
        const SizedBox(width: 4),
        Text(text, style: const TextStyle(color: Colors.grey)),
      ],
    );
  }
}