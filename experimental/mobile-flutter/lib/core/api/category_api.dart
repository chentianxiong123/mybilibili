import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'client.dart';

final categoryApiProvider = Provider<CategoryApi>((ref) => CategoryApi(ref.read(dioProvider)));

class Category {
  final int id;
  final String name;
  final String icon;
  final int sortOrder;

  const Category({this.id = 0, this.name = '', this.icon = '', this.sortOrder = 0});

  factory Category.fromJson(Map<String, dynamic> json) => Category(
    id: (json['id'] as num?)?.toInt() ?? 0,
    name: json['name'] as String? ?? '',
    icon: json['icon'] as String? ?? '',
    sortOrder: (json['sort_order'] as num?)?.toInt() ?? 0,
  );
}

class CategoryApi {
  final Dio _dio;
  CategoryApi(this._dio);

  Future<List<Category>> getCategories() async {
    final res = await _dio.get('/category');
    final data = _extractList(res.data);
    return data.map((e) => Category.fromJson(e)).toList();
  }

  List<Map<String, dynamic>> _extractList(dynamic data) {
    if (data is Map) {
      final d = data['data'];
      if (d is List) return d.cast<Map<String, dynamic>>();
    }
    if (data is List) return data.cast<Map<String, dynamic>>();
    return [];
  }
}