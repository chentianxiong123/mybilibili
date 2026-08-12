import 'package:flutter_test/flutter_test.dart';
import 'package:mybilibili_app_flutter/app.dart';

void main() {
  testWidgets('App launches', (WidgetTester tester) async {
    await tester.pumpWidget(const MyBiliApp());
    expect(find.text('mybilibili'), findsOneWidget);
  });
}