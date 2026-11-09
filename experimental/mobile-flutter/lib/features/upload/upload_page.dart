import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/category_api.dart';
import '../../core/api/upload_api.dart';
import '../video/screens/video_detail_screen.dart';

final uploadCategoriesProvider = FutureProvider.autoDispose<List<Category>>((ref) {
  return ref.read(categoryApiProvider).getCategories();
});

class UploadPage extends ConsumerStatefulWidget {
  const UploadPage({super.key});

  @override
  ConsumerState<UploadPage> createState() => _UploadPageState();
}

class _UploadPageState extends ConsumerState<UploadPage> {
  final _titleCtl = TextEditingController();
  final _descCtl = TextEditingController();
  final _urlCtl = TextEditingController();
  final _tagsCtl = TextEditingController();
  int? _categoryId;
  bool _submitting = false;
  int? _createdManuscriptId;
  int _resetCounter = 0;

  @override
  void dispose() {
    _titleCtl.dispose();
    _descCtl.dispose();
    _urlCtl.dispose();
    _tagsCtl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final cats = ref.watch(uploadCategoriesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('发布视频')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          TextField(
            controller: _titleCtl,
            decoration: const InputDecoration(labelText: '标题', hintText: '请输入视频标题'),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _descCtl,
            maxLines: 3,
            decoration: const InputDecoration(labelText: '简介', hintText: '选填'),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _tagsCtl,
            decoration: const InputDecoration(labelText: '标签', hintText: '逗号分隔，选填'),
          ),
          const SizedBox(height: 12),
          cats.when(
            data: (list) => DropdownButtonFormField<int>(
              key: ValueKey(_resetCounter),
              initialValue: _categoryId,
              isExpanded: true,
              hint: const Text('选择分类'),
              items: list
                  .map((c) => DropdownMenuItem(value: c.id, child: Text(c.name, overflow: TextOverflow.ellipsis)))
                  .toList(),
              onChanged: (v) => setState(() => _categoryId = v),
            ),
            loading: () => const LinearProgressIndicator(),
            error: (e, _) => Text('分类加载失败: $e', style: const TextStyle(color: Colors.grey, fontSize: 12)),
          ),
          const SizedBox(height: 12),
          TextField(
            controller: _urlCtl,
            decoration: const InputDecoration(labelText: '视频链接', hintText: 'mp4 直链地址'),
          ),
          const SizedBox(height: 24),
          if (_createdManuscriptId != null)
            _buildResult()
          else
            ElevatedButton(
              onPressed: _submitting ? null : _submit,
              child: _submitting
                  ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2))
                  : const Text('发布'),
            ),
        ],
      ),
    );
  }

  Widget _buildResult() {
    final id = _createdManuscriptId ?? 0;
    return Column(
      children: [
        const Icon(Icons.check_circle, color: Colors.green, size: 48),
        const SizedBox(height: 12),
        Text('发布成功，稿件 ID: $id', style: const TextStyle(fontSize: 15)),
        const SizedBox(height: 16),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            OutlinedButton(
              onPressed: () => Navigator.of(context).push(MaterialPageRoute(
                builder: (_) => VideoDetailScreen(manuscriptId: id),
              )),
              child: const Text('查看稿件'),
            ),
            OutlinedButton(
              onPressed: () => setState(() {
                _createdManuscriptId = null;
                _titleCtl.clear();
                _descCtl.clear();
                _urlCtl.clear();
                _tagsCtl.clear();
                _categoryId = null;
                _resetCounter++;
              }),
              child: const Text('继续发布'),
            ),
          ],
        ),
      ],
    );
  }

  Future<void> _submit() async {
    final title = _titleCtl.text.trim();
    final url = _urlCtl.text.trim();
    if (title.isEmpty || _categoryId == null || url.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请填写标题、分类和视频链接')),
      );
      return;
    }
    setState(() => _submitting = true);
    try {
      final tags = _tagsCtl.text
          .split(',')
          .map((e) => e.trim())
          .where((e) => e.isNotEmpty)
          .toList();
      final api = ref.read(uploadApiProvider);
      final sessionId = await api.createSession(
        title: title,
        description: _descCtl.text.trim(),
        categoryId: _categoryId!,
        tags: tags,
        videoUrl: url,
      );
      final result = await api.completeSession(sessionId);
      if (!mounted) return;
      setState(() {
        _submitting = false;
        _createdManuscriptId = (result['manuscript_id'] as num?)?.toInt() ?? 0;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() => _submitting = false);
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('发布失败: $e')));
    }
  }
}
