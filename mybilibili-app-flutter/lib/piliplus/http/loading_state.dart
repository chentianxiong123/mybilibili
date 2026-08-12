sealed class LoadingState<T> {
  const LoadingState();
}

class Success<T> extends LoadingState<T> {
  final T response;
  final int? total;
  const Success(this.response, {this.total});
}

class Loading<T> extends LoadingState<T> {
  const Loading();
}

class Error<T> extends LoadingState<T> {
  final dynamic message;
  final dynamic extra;
  const Error(this.message, {this.extra});
}

extension LoadingStateExt<T> on LoadingState<T> {
  T? get dataOrNull => switch (this) {
    Success<T>(:final response) => response,
    _ => null,
  };
}