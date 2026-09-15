class DanmakuData {
  final String text;
  final int timeMs;
  final int type;
  final int color;
  final int fontSize;

  const DanmakuData({
    required this.text,
    required this.timeMs,
    this.type = 0,
    this.color = 0xFFFFFFFF,
    this.fontSize = 25,
  });

  factory DanmakuData.fromJson(Map<String, dynamic> json) {
    return DanmakuData(
      text: json['text'] as String? ?? '',
      timeMs: (json['time'] as num?)?.toInt() ?? 0,
      type: (json['type'] as num?)?.toInt() ?? 0,
      color: (json['color'] as num?)?.toInt() ?? 0xFFFFFFFF,
      fontSize: (json['font_size'] as num?)?.toInt() ?? 25,
    );
  }
}