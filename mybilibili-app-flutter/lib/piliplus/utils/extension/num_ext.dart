extension NumExt on double {
  double toPrecision(int n) => double.parse(toStringAsFixed(n));
}

extension IntExt on int {
  Widget cacheSize(BuildContext context) => const SizedBox.shrink();
}

import 'package:flutter/material.dart';