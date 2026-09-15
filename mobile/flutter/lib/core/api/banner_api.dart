import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final bannerApiProvider = Provider<BannerApi>((ref) => BannerApi(ref.read(dioProvider)));

class BannerItem {
  final int id;
  final String title;
  final String imageUrl;
  final String link;
  final int type;
  final int sortOrder;

  const BannerItem({this.id = 0, this.title = '', this.imageUrl = '', this.link = '', this.type = 0, this.sortOrder = 0});

  factory BannerItem.fromJson(Map<String, dynamic> json) => BannerItem(
    id: (json['id'] as num?)?.toInt() ?? 0,
    title: json['title'] as String? ?? '',
    imageUrl: json['image_url'] as String? ?? '',
    link: json['link'] as String? ?? '',
    type: (json['type'] as num?)?.toInt() ?? 0,
    sortOrder: (json['sort_order'] as num?)?.toInt() ?? 0,
  );
}

class BannerApi {
  final Dio _dio;
  BannerApi(this._dio);

  Future<List<BannerItem>> getHomeBanners() async {
    final res = await _dio.get('/banner-images/home');
    return _extractList(res.data).map((e) => BannerItem.fromJson(e)).toList();
  }

  List<Map<String, dynamic>> _extractList(dynamic data) {
    if (data is Map) {
      final d = data['data'];
      if (d is List) return d.cast<Map<String, dynamic>>();
      final list = data['list'] as List?;
      if (list != null) return list.cast<Map<String, dynamic>>();
    }
    if (data is List) return data.cast<Map<String, dynamic>>();
    return [];
  }
}