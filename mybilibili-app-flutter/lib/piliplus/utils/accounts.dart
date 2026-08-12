import 'package:mybilibili_app_flutter/piliplus/models/common/account_type.dart';

class AccountInfo {
  bool get isLogin => false;
  int get mid => 0;
}

class Accounts {
  static final AccountInfo main = AccountInfo();
  static AccountInfo get(AccountType type) => AccountInfo();
}