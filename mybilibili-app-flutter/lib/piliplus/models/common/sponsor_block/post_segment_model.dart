import 'action_type.dart';

class PostSegmentModel {
  final int segment;
  final int category;
  final ActionType actionType;

  PostSegmentModel({
    required this.segment,
    required this.category,
    required this.actionType,
  });
}