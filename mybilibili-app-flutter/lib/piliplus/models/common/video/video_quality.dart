class VideoQuality {
  final int? code;
  final String? desc;
  final String? shortDesc;

  VideoQuality({this.code, this.desc, this.shortDesc});

  static VideoQuality fromCode(int code) {
    return VideoQuality(code: code, desc: '自动', shortDesc: '自动');
  }
}