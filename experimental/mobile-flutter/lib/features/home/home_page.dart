import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/api/manuscript_api.dart';
import '../../core/api/category_api.dart';
import '../../shared/models/manuscript.dart';
import '../video/screens/video_detail_screen.dart';
import 'widgets/banner_carousel.dart';
import 'widgets/video_card.dart';

final homeDataProvider = FutureProvider.autoDispose<Map<int, List<ManuscriptInfo>>>((ref) async {
  final manuscripts = await ref.read(manuscriptApiProvider).getRecommended();
  return {0: manuscripts};
});

final categoriesProvider = FutureProvider.autoDispose<List<Category>>((ref) async {
  return ref.read(categoryApiProvider).getCategories();
});

class HomePage extends ConsumerStatefulWidget {
  const HomePage({super.key});
  @override
  ConsumerState<HomePage> createState() => _HomePageState();
}

class _HomePageState extends ConsumerState<HomePage> with SingleTickerProviderStateMixin {
  int _selectedTab = 0;
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 1, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final categories = ref.watch(categoriesProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('mybilibili'),
        actions: [
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () => Navigator.of(context).pushNamed('/search'),
          ),
        ],
      ),
      body: Column(
        children: [
          categories.whenOrNull(data: (cats) {
            final tabCount = cats.length + 1;
            if (tabCount != _tabController.length) {
              WidgetsBinding.instance.addPostFrameCallback((_) {
                if (mounted) _tabController.dispose();
                _tabController = TabController(length: tabCount, vsync: this);
              });
            }
            return Container(
              height: 44,
              color: const Color(0xFF222222),
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                itemCount: cats.length + 1,
                itemBuilder: (context, index) {
                  final label = index == 0 ? '推荐' : cats[index - 1].name;
                  return GestureDetector(
                    onTap: () => setState(() => _selectedTab = index),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      alignment: Alignment.center,
                      decoration: BoxDecoration(
                        border: _selectedTab == index
                            ? const Border(bottom: BorderSide(color: Color(0xFFFB7299), width: 2))
                            : null,
                      ),
                      child: Text(label, style: TextStyle(
                        color: _selectedTab == index ? const Color(0xFFFB7299) : Colors.grey,
                        fontSize: 14, fontWeight: _selectedTab == index ? FontWeight.w600 : FontWeight.normal,
                      )),
                    ),
                  );
                },
              ),
            );
          }) ?? const SizedBox.shrink(),
          Expanded(
            child: _selectedTab == 0 ? _buildRecommended() : _buildCategoryContent(_selectedTab - 1),
          ),
        ],
      ),
    );
  }

  Widget _buildRecommended() {
    final data = ref.watch(homeDataProvider);
    return data.when(
      data: (map) {
        final manuscripts = map[0] ?? [];
        return RefreshIndicator(
          onRefresh: () async => ref.invalidate(homeDataProvider),
          child: CustomScrollView(
            slivers: [
              SliverToBoxAdapter(child: Padding(
                padding: const EdgeInsets.only(top: 8),
                child: BannerCarousel(fallbackTitle: 'mybilibili'),
              )),
              SliverPadding(
                padding: const EdgeInsets.all(8),
                sliver: SliverGrid(
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2, childAspectRatio: 0.7, crossAxisSpacing: 8, mainAxisSpacing: 8,
                  ),
                  delegate: SliverChildBuilderDelegate((context, index) {
                    final m = manuscripts[index];
                    return VideoCard(
                      manuscript: m,
                      onTap: () => Navigator.of(context).push(MaterialPageRoute(
                        builder: (_) => VideoDetailScreen(manuscriptId: m.id),
                      )),
                    );
                  }, childCount: manuscripts.length),
                ),
              ),
            ],
          ),
        );
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (e, _) => Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.cloud_off, size: 48, color: Colors.grey),
            const SizedBox(height: 16),
            Text('$e', style: const TextStyle(color: Colors.grey)),
            const SizedBox(height: 16),
            ElevatedButton(onPressed: () => ref.invalidate(homeDataProvider), child: const Text('重试')),
          ],
        ),
      ),
    );
  }

  Widget _buildCategoryContent(int index) {
    final categories = ref.watch(categoriesProvider);
    return categories.when(
      data: (cats) {
        if (index >= cats.length) return const SizedBox.shrink();
        return _buildCategoryVideos(cats[index].id);
      },
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (_, __) => const SizedBox.shrink(),
    );
  }

  Widget _buildCategoryVideos(int categoryId) {
    final provider = FutureProvider.autoDispose<List<ManuscriptInfo>>((ref) async {
      return ref.read(manuscriptApiProvider).getByCategory(categoryId);
    });
    final data = ref.watch(provider);
    return data.when(
      data: (manuscripts) => RefreshIndicator(
        onRefresh: () async => ref.invalidate(provider),
        child: GridView.builder(
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
      error: (_, __) => const SizedBox.shrink(),
    );
  }
}