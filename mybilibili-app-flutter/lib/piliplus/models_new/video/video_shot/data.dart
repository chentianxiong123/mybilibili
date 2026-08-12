class VideoShotData {
  final int imgXLen;
  final int imgYLen;
  final int totalPerImage;
  int imgXSize;
  int imgYSize;
  final List<String> image;

  VideoShotData({
    required this.imgXLen,
    required this.imgYLen,
    required this.totalPerImage,
    this.imgXSize = 0,
    this.imgYSize = 0,
    required this.image,
  });
}