import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/manuscript_api.dart';
import '../../../core/api/interaction_api.dart';
import '../../../core/api/danmaku_api.dart';
import '../../../core/theme/theme.dart';
import '../../../shared/models/manuscript.dart';
import '../player/widgets/player_view.dart';
import '../player/danmaku_model.dart';
import '../widgets/comment_section.dart';

class VideoDetailScreen extends ConsumerStatefulWidget {
  final int manuscriptId;
  final ManuscriptInfo? preloadedManuscript;
  const VideoDetailScreen({super.key, required this.manuscriptId, this.preloadedManuscript});
  @override
  ConsumerState<VideoDetailScreen> createState() => _VideoDetailScreenState();
}

class _VideoDetailScreenState extends ConsumerState<VideoDetailScreen> {
  ManuscriptInfo? _manuscript;
  bool _loading = true;
  String? _error;
  int _currentVideoIndex = 0;
  InteractionStatus? _interactionStatus;
  List<DanmakuData> _danmakuList = [];

  @override
  void initState() {
    super.initState();
    if (widget.preloadedManuscript != null) {
      _manuscript = widget.preloadedManuscript;
      _loading = false;
      _loadInteractionStatus();
      _loadDanmaku();
    } else {
      _loadManuscript();
    }
  }

  Future<void> _loadManuscript() async {
    try {
      final manuscript = await ref.read(manuscriptApiProvider).getManuscript(widget.manuscriptId);
      if (mounted) setState(() { _manuscript = manuscript; _loading = false; });
      _loadInteractionStatus();
      _loadDanmaku();
    } catch (e) {
      if (mounted) setState(() { _error = e.toString(); _loading = false; });
    }
  }

  Future<void> _loadInteractionStatus() async {
    try {
      final status = await ref.read(interactionApiProvider).getStatus(widget.manuscriptId);
      if (mounted) setState(() => _interactionStatus = status);
    } catch (_) {}
  }

  Future<void> _loadDanmaku() async {
    try {
      final danmaku = await ref.read(danmakuApiProvider).getDanmaku(_manuscript?.videos.firstOrNull?.id ?? 0);
      if (mounted) setState(() => _danmakuList = danmaku);
    } catch (_) {}
  }

  VideoItem get _currentVideo {
    final videos = _manuscript?.videos ?? [];
    return _currentVideoIndex < videos.length ? videos[_currentVideoIndex] : VideoItem();
  }

  String get _playUrl {
    final v = _currentVideo;
    return v.bestPlayUrl.isNotEmpty ? v.bestPlayUrl : (_manuscript?.firstVideoPlayUrl ?? '');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Column(mainAxisAlignment: MainAxisAlignment.center, children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.grey),
                  const SizedBox(height: 16),
                  Text('$_error', style: const TextStyle(color: Colors.grey)),
                  const SizedBox(height: 16),
                  ElevatedButton(onPressed: _loadManuscript, child: const Text('重试')),
                ]))
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    if (_manuscript == null) return const SizedBox.shrink();
    final m = _manuscript!;
    return CustomScrollView(
      slivers: [
        SliverToBoxAdapter(child: AspectRatio(
          aspectRatio: 16 / 9,
          child: BilibiliPlayer(
            url: _playUrl,
            title: m.title,
            danmakuList: _danmakuList,
            showBackButton: true,
            onBack: () => Navigator.of(context).pop(),
          ),
        )),
        SliverToBoxAdapter(child: _buildInfo(m)),
        SliverToBoxAdapter(child: _buildInteractionBar(m)),
        if (m.videos.length > 1) SliverToBoxAdapter(child: _buildPartList(m)),
        if (m.tags.isNotEmpty) SliverToBoxAdapter(child: _buildTags(m)),
        if (m.uploader != null) SliverToBoxAdapter(child: _buildUploader(m)),
        SliverToBoxAdapter(child: CommentSection(manuscriptId: m.id)),
      ],
    );
  }

  Widget _buildInfo(ManuscriptInfo m) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(m.title, style: const TextStyle(fontSize: 17, fontWeight: FontWeight.bold, height: 1.3)),
          const SizedBox(height: 8),
          Row(children: [
            Text('${m.viewCount}次观看', style: const TextStyle(fontSize: 12, color: Colors.grey)),
            const SizedBox(width: 12),
            Text(formatTime(m.createdAt), style: const TextStyle(fontSize: 12, color: Colors.grey)),
          ]),
        ],
      ),
    );
  }

  String formatTime(String t) {
    try { return t.substring(0, 10); } catch (_) { return t; }
  }

  Widget _buildInteractionBar(ManuscriptInfo m) {
    final s = _interactionStatus;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: [
          _interactBtn(Icons.thumb_up, s?.liked == true, '${s?.likeCount ?? m.likeCount}', () async {
            try {
              if (s?.liked == true) { await ref.read(interactionApiProvider).unlike(widget.manuscriptId); }
              else { await ref.read(interactionApiProvider).like(widget.manuscriptId); }
              _loadInteractionStatus();
            } catch (_) {}
          }, AppTheme.primaryPink),
          _interactBtn(Icons.monetization_on, s?.coined == true, '${s?.coinCount ?? m.coinCount}', () async {
            try { await ref.read(interactionApiProvider).coin(widget.manuscriptId); _loadInteractionStatus(); } catch (_) {}
          }, Colors.amber),
          _interactBtn(Icons.star, s?.collected == true, '${s?.collectCount ?? m.collectCount}', () async {
            try {
              if (s?.collected == true) { await ref.read(interactionApiProvider).uncollect(widget.manuscriptId); }
              else { await ref.read(interactionApiProvider).collect(widget.manuscriptId); }
              _loadInteractionStatus();
            } catch (_) {}
          }, AppTheme.primaryPink),
          _interactBtn(Icons.share, false, '${m.shareCount}', () async {
            try { await ref.read(interactionApiProvider).share(widget.manuscriptId); } catch (_) {}
          }, null),
        ],
      ),
    );
  }

  Widget _interactBtn(IconData icon, bool active, String count, VoidCallback onTap, Color? activeColor) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        children: [
          Icon(icon, size: 22, color: active ? (activeColor ?? AppTheme.primaryPink) : Colors.grey),
          const SizedBox(height: 4),
          Text(count, style: TextStyle(fontSize: 11, color: active ? (activeColor ?? AppTheme.primaryPink) : Colors.grey)),
        ],
      ),
    );
  }

  Widget _buildPartList(ManuscriptInfo m) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('分P', style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold)),
          const SizedBox(height: 8),
          SizedBox(
            height: 44,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: m.videos.length,
              separatorBuilder: (_, __) => const SizedBox(width: 8),
              itemBuilder: (context, index) {
                final v = m.videos[index];
                final active = index == _currentVideoIndex;
                return GestureDetector(
                  onTap: () => setState(() => _currentVideoIndex = index),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                    decoration: BoxDecoration(
                      color: active ? AppTheme.primaryPink.withValues(alpha: 0.2) : const Color(0xFF2A2A2A),
                      borderRadius: BorderRadius.circular(8),
                      border: active ? Border.all(color: AppTheme.primaryPink) : null,
                    ),
                    child: Text('P${v.videoOrder} ${v.title}', style: TextStyle(fontSize: 12, color: active ? AppTheme.primaryPink : Colors.grey)),
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTags(ManuscriptInfo m) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 8, 16, 0),
      child: Wrap(
        spacing: 6, runSpacing: 6,
        children: m.tags.map((t) => Container(
          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
          decoration: BoxDecoration(color: const Color(0xFF2A2A2A), borderRadius: BorderRadius.circular(4)),
          child: Text(t, style: const TextStyle(fontSize: 11, color: Colors.grey)),
        )).toList(),
      ),
    );
  }

  Widget _buildUploader(ManuscriptInfo m) {
    final u = m.uploader!;
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
      child: Row(
        children: [
          CircleAvatar(
            radius: 20,
            backgroundImage: u.avatar.isNotEmpty ? NetworkImage(u.avatar) : null,
            child: u.avatar.isEmpty ? const Icon(Icons.person, size: 20) : null,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(u.nickname, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w500)),
                Text('${u.followerCount}粉丝', style: const TextStyle(fontSize: 11, color: Colors.grey)),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
            decoration: BoxDecoration(
              color: AppTheme.primaryPink,
              borderRadius: BorderRadius.circular(16),
            ),
            child: const Text('+ 关注', style: TextStyle(color: Colors.white, fontSize: 13)),
          ),
        ],
      ),
    );
  }
}