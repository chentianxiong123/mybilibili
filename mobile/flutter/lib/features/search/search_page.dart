import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/search_api.dart';
import '../../shared/models/manuscript.dart';
import '../home/widgets/video_card.dart';
import '../video/screens/video_detail_screen.dart';

class SearchPage extends ConsumerStatefulWidget {
  const SearchPage({super.key});
  @override
  ConsumerState<SearchPage> createState() => _SearchPageState();
}

class _SearchPageState extends ConsumerState<SearchPage> {
  final _controller = TextEditingController();
  List<ManuscriptInfo> _results = [];
  List<String> _hotKeywords = [];
  List<String> _suggestions = [];
  bool _showSuggestions = false;

  @override
  void initState() {
    super.initState();
    _loadHotKeywords();
  }

  Future<void> _loadHotKeywords() async {
    try {
      final keywords = await ref.read(searchApiProvider).getHotKeywords();
      if (mounted) setState(() => _hotKeywords = keywords);
    } catch (_) {}
  }

  Future<void> _doSearch(String keyword) async {
    if (keyword.trim().isEmpty) return;
    setState(() { _results = []; _showSuggestions = false; });
    FocusScope.of(context).unfocus();
    try {
      final results = await ref.read(searchApiProvider).search(keyword.trim());
      if (mounted) setState(() { _results = results; });
    } catch (_) {
      if (mounted) setState(() {});
    }
  }

  Future<void> _loadSuggestions(String keyword) async {
    if (keyword.length < 2) { setState(() => _suggestions = []); return; }
    try {
      final suggestions = await ref.read(searchApiProvider).getSuggestions(keyword);
      if (mounted) setState(() => _suggestions = suggestions);
    } catch (_) {}
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: SizedBox(
          height: 36,
          child: TextField(
            controller: _controller,
            autofocus: true,
            style: const TextStyle(fontSize: 14),
            decoration: InputDecoration(
              hintText: '搜索视频...',
              hintStyle: const TextStyle(color: Colors.grey, fontSize: 14),
              filled: true, fillColor: const Color(0xFF2A2A2A),
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 0),
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(18), borderSide: BorderSide.none),
              prefixIcon: const Icon(Icons.search, size: 20, color: Colors.grey),
              suffixIcon: _controller.text.isNotEmpty
                  ? IconButton(icon: const Icon(Icons.clear, size: 18), onPressed: () {
                      _controller.clear();
                      setState(() { _results = []; _suggestions = []; _showSuggestions = false; });
                    })
                  : null,
            ),
            onChanged: (v) {
              setState(() => _showSuggestions = v.isNotEmpty);
              _loadSuggestions(v);
            },
            onSubmitted: _doSearch,
          ),
        ),
        actions: [TextButton(onPressed: () => _doSearch(_controller.text), child: const Text('搜索'))],
      ),
      body: _results.isNotEmpty
          ? _buildResults()
          : _showSuggestions && _suggestions.isNotEmpty
              ? _buildSuggestions()
              : _buildHotKeywords(),
    );
  }

  Widget _buildHotKeywords() {
    if (_hotKeywords.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('热搜', style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
          const SizedBox(height: 12),
          Wrap(
            spacing: 8, runSpacing: 8,
            children: _hotKeywords.asMap().entries.map((e) {
              return GestureDetector(
                onTap: () {
                  _controller.text = e.value;
                  _doSearch(e.value);
                },
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: const Color(0xFF2A2A2A),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  child: Text('${e.key + 1}. ${e.value}', style: const TextStyle(fontSize: 13)),
                ),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildSuggestions() {
    return ListView.separated(
      itemCount: _suggestions.length,
      separatorBuilder: (_, __) => const Divider(height: 1, color: Color(0xFF2A2A2A)),
      itemBuilder: (context, index) {
        return ListTile(
          leading: const Icon(Icons.search, size: 18, color: Colors.grey),
          title: Text(_suggestions[index], style: const TextStyle(fontSize: 14)),
          onTap: () {
            _controller.text = _suggestions[index];
            _doSearch(_suggestions[index]);
          },
        );
      },
    );
  }

  Widget _buildResults() {
    return GridView.builder(
      padding: const EdgeInsets.all(8),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2, childAspectRatio: 0.7, crossAxisSpacing: 8, mainAxisSpacing: 8,
      ),
      itemCount: _results.length,
      itemBuilder: (context, index) => VideoCard(
        manuscript: _results[index],
        onTap: () => Navigator.of(context).push(MaterialPageRoute(
          builder: (_) => VideoDetailScreen(manuscriptId: _results[index].id),
        )),
      ),
    );
  }
}