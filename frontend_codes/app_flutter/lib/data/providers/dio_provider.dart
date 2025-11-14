import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/config/api_config.dart';
import '../../core/utils/logger.dart';
import '../../core/utils/storage_service.dart';

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(
    BaseOptions(
      connectTimeout: ApiConfig.connectTimeout,
      receiveTimeout: ApiConfig.receiveTimeout,
      sendTimeout: ApiConfig.sendTimeout,
      headers: {
        'Content-Type': 'application/json',
      },
    ),
  );

  // Request Interceptor
  dio.interceptors.add(
    InterceptorsWrapper(
      onRequest: (options, handler) async {
        // Add auth token to requests
        final token = await StorageService.getToken();
        if (token != null) {
          options.headers['Authorization'] = 'Bearer $token';
        }

        AppLogger.debug('Request: ${options.method} ${options.uri}');
        return handler.next(options);
      },
      onResponse: (response, handler) {
        AppLogger.debug('Response: ${response.statusCode} ${response.requestOptions.uri}');
        return handler.next(response);
      },
      onError: (error, handler) async {
        AppLogger.error(
          'Error: ${error.response?.statusCode} ${error.requestOptions.uri}',
          error.message,
        );

        // Handle 401 Unauthorized - try to refresh token
        if (error.response?.statusCode == 401) {
          final refreshToken = await StorageService.getRefreshToken();
          if (refreshToken != null) {
            try {
              // Attempt to refresh token
              final response = await dio.post(
                ApiConfig.authUrl + '/refresh',
                data: {'refreshToken': refreshToken},
              );

              final newToken = response.data['accessToken'];
              await StorageService.saveToken(newToken);

              // Retry the original request
              final options = error.requestOptions;
              options.headers['Authorization'] = 'Bearer $newToken';
              final retryResponse = await dio.fetch(options);
              return handler.resolve(retryResponse);
            } catch (e) {
              // Refresh failed - clear auth and redirect to login
              await StorageService.clearAuthData();
              return handler.reject(error);
            }
          }
        }

        return handler.next(error);
      },
    ),
  );

  return dio;
});
