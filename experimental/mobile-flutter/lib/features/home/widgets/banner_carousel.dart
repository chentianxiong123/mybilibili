import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/api/banner_api.dart';

class BannerCarousel extends ConsumerStatefulWidget {
  final String? fallbackTitle;
  const BannerCarousel({super.key, this.fallbackTitle});

  @override
  ConsumerState<BannerCarousel> createState() => _BannerCarouselState();
}

class _BannerCarouselState extends ConsumerState<BannerCarousel> {
  final _pageController = PageController();
  int _currentPage = 0;
  List<BannerItem> _banners = [];

  @override
  void initState() {
    super.initState();
    _loadBanners();
  }

  Future<void> _loadBanners() async {
    try {
      final banners = await ref.read(bannerApiProvider).getHomeBanners();
      if (mounted) setState(() => _banners = banners);
    } catch (_) {}
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (_banners.isEmpty) return const SizedBox.shrink();
    return SizedBox(
      height: 150,
      child: Stack(
        children: [
          PageView.builder(
            controller: _pageController,
            itemCount: _banners.length,
            onPageChanged: (i) => setState(() => _currentPage = i),
            itemBuilder: (context, index) {
              final b = _banners[index];
              return Container(
                margin: const EdgeInsets.symmetric(horizontal: 8),
                clipBehavior: Clip.antiAlias,
                decoration: BoxDecoration(borderRadius: BorderRadius.circular(8)),
                child: Stack(
                  fit: StackFit.expand,
                  children: [
                    Image.network(b.imageUrl, fit: BoxFit.cover,
                      errorBuilder: (_, _, _) => Container(color: Colors.grey[800])),
                    if (b.title.isNotEmpty)
                      Positioned(
                        left: 12, right: 12, bottom: 8,
                        child: Text(b.title, style: const TextStyle(color: Colors.white, fontSize: 14, fontWeight: FontWeight.w500), maxLines: 1, overflow: TextOverflow.ellipsis),
                      ),
                  ],
                ),
              );
            },
          ),
          if (_banners.length > 1)
            Positioned(
              right: 16, bottom: 8,
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(color: Colors.black54, borderRadius: BorderRadius.circular(10)),
                child: Text('${_currentPage + 1}/${_banners.length}', style: const TextStyle(color: Colors.white, fontSize: 11)),
              ),
            ),
        ],
      ),
    );
  }
}