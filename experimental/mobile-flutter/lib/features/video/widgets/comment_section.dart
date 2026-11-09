import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/comment_api.dart';
import '../../../core/theme/theme.dart';

class CommentSection extends ConsumerStatefulWidget {
  final int manuscriptId;
  const CommentSection({super.key, required this.manuscriptId});
  @override
  ConsumerState<CommentSection> createState() => _CommentSectionState();
}

class _CommentSectionState extends ConsumerState<CommentSection> {
  List<CommentItem> _comments = [];
  bool _loading = true;
  final _commentController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadComments();
  }

  Future<void> _loadComments() async {
    try {
      final comments = await ref.read(commentApiProvider).getComments(widget.manuscriptId);
      if (mounted) setState(() { _comments = comments; _loading = false; });
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _postComment() async {
    if (_commentController.text.trim().isEmpty) return;
    try {
      await ref.read(commentApiProvider).addComment(widget.manuscriptId, _commentController.text.trim());
      _commentController.clear();
      _loadComments();
    } catch (_) {}
  }

  @override
  void dispose() {
    _commentController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          child: Row(
            children: [
              const Text('评论', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
              const SizedBox(width: 8),
              Text('${_comments.length}', style: const TextStyle(fontSize: 13, color: Colors.grey)),
            ],
          ),
        ),
        if (_loading)
          const Center(child: Padding(padding: EdgeInsets.all(24), child: CircularProgressIndicator()))
        else if (_comments.isEmpty)
          const Center(child: Padding(padding: EdgeInsets.all(24), child: Text('暂无评论', style: TextStyle(color: Colors.grey))))
        else
          ...List.generate(_comments.length, (i) => _buildCommentCard(_comments[i])),
        const Divider(height: 1),
        Container(
          padding: const EdgeInsets.fromLTRB(16, 8, 8, 8),
          child: Row(
            children: [
              Expanded(
                child: SizedBox(
                  height: 36,
                  child: TextField(
                    controller: _commentController,
                    style: const TextStyle(fontSize: 14),
                    decoration: InputDecoration(
                      hintText: '发一条评论...',
                      hintStyle: const TextStyle(fontSize: 14, color: Colors.grey),
                      filled: true, fillColor: const Color(0xFF2A2A2A),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(18), borderSide: BorderSide.none),
                    ),
                    onSubmitted: (_) => _postComment(),
                  ),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.send, color: AppTheme.primaryPink, size: 20),
                onPressed: _postComment,
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildCommentCard(CommentItem c) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          CircleAvatar(
            radius: 16,
            backgroundImage: c.userAvatar.isNotEmpty ? NetworkImage(c.userAvatar) : null,
            child: c.userAvatar.isEmpty ? const Icon(Icons.person, size: 16) : null,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(c.username, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w500, color: Color(0xFF99A2AA))),
                    if (c.isPinned) ...[
                      const SizedBox(width: 4),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                        decoration: BoxDecoration(color: AppTheme.primaryPink.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(4)),
                        child: const Text('置顶', style: TextStyle(fontSize: 10, color: AppTheme.primaryPink)),
                      ),
                    ],
                  ],
                ),
                const SizedBox(height: 4),
                Text(c.content, style: const TextStyle(fontSize: 14, height: 1.4)),
                const SizedBox(height: 4),
                Row(
                  children: [
                    Text(c.createdAt, style: const TextStyle(fontSize: 11, color: Colors.grey)),
                    const Spacer(),
                    IconButton(
                      icon: Icon(Icons.thumb_up, size: 14, color: c.liked ? AppTheme.primaryPink : Colors.grey),
                      onPressed: () async {
                        try {
                          if (c.liked) { await ref.read(commentApiProvider).unlikeComment(c.id); }
                          else { await ref.read(commentApiProvider).likeComment(c.id); }
                          _loadComments();
                        } catch (_) {}
                      },
                      padding: EdgeInsets.zero, constraints: const BoxConstraints(minWidth: 24, minHeight: 24),
                    ),
                    if (c.likeCount > 0) Text('${c.likeCount}', style: const TextStyle(fontSize: 11, color: Colors.grey)),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}