class BaseEpisodeItem {
  final int cid;
  BaseEpisodeItem({required this.cid});
}

class UgcSeasonSection {
  final List<BaseEpisodeItem> episodes;
  UgcSeasonSection({required this.episodes});
}

class UgcSeason {
  final List<UgcSeasonSection>? sections;
  UgcSeason({this.sections});
}