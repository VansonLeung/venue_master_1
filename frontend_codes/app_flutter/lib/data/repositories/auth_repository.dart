import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/config/api_config.dart';
import '../../core/utils/logger.dart';
import '../../core/utils/storage_service.dart';
import '../models/auth_response.dart';
import '../models/user.dart';
import '../providers/dio_provider.dart';

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepository(ref.read(dioProvider));
});

class AuthRepository {
  final Dio _dio;

  AuthRepository(this._dio);

  Future<AuthResponse> login(String email, String password) async {
    try {
      AppLogger.info('Attempting login for: $email');

      final response = await _dio.post(
        '${ApiConfig.authUrl}/login',
        data: {
          'email': email,
          'password': password,
        },
      );

      final authResponse = AuthResponse.fromJson(response.data);

      // Save tokens
      await StorageService.saveToken(authResponse.accessToken);
      await StorageService.saveRefreshToken(authResponse.refreshToken);
      await StorageService.saveUserId(authResponse.user.id);

      AppLogger.info('Login successful for user: ${authResponse.user.id}');
      return authResponse;
    } on DioException catch (e) {
      AppLogger.error('Login failed', e.message);
      if (e.response?.statusCode == 401) {
        throw Exception('Invalid email or password');
      }
      throw Exception('Login failed: ${e.message}');
    } catch (e) {
      AppLogger.error('Unexpected login error', e);
      throw Exception('An unexpected error occurred');
    }
  }

  Future<AuthResponse> refreshToken() async {
    try {
      final refreshToken = await StorageService.getRefreshToken();
      if (refreshToken == null) {
        throw Exception('No refresh token available');
      }

      final response = await _dio.post(
        '${ApiConfig.authUrl}/refresh',
        data: {'refreshToken': refreshToken},
      );

      final authResponse = AuthResponse.fromJson(response.data);

      // Save new tokens
      await StorageService.saveToken(authResponse.accessToken);
      await StorageService.saveRefreshToken(authResponse.refreshToken);

      AppLogger.info('Token refreshed successfully');
      return authResponse;
    } catch (e) {
      AppLogger.error('Token refresh failed', e);
      await logout();
      throw Exception('Session expired. Please login again.');
    }
  }

  Future<User?> getCurrentUser() async {
    try {
      final token = await StorageService.getToken();
      if (token == null) return null;

      final userId = await StorageService.getUserId();
      if (userId == null) return null;

      final response = await _dio.get(
        '${ApiConfig.userUrl}/$userId',
        options: Options(
          headers: {
            'Authorization': 'Bearer $token',
            'X-User-ID': userId,
            'X-User-Roles': 'MEMBER', // This should come from stored user data
          },
        ),
      );

      return User.fromJson(response.data);
    } catch (e) {
      AppLogger.error('Failed to get current user', e);
      return null;
    }
  }

  Future<void> logout() async {
    try {
      AppLogger.info('Logging out user');
      await StorageService.clearAuthData();
    } catch (e) {
      AppLogger.error('Logout error', e);
    }
  }

  Future<bool> isAuthenticated() async {
    final token = await StorageService.getToken();
    return token != null;
  }
}
