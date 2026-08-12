enum SuperResolutionType {
  disable,
  efficiency,
  quality;

  String get label => switch (this) {
    SuperResolutionType.disable => '关闭',
    SuperResolutionType.efficiency => '高效',
    SuperResolutionType.quality => '质量',
  };
}