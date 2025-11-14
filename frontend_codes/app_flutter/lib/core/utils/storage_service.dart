import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../config/api_config.dart';

class StorageService {
  static const _storage = FlutterSecureStorage();
  static SharedPreferences? _prefs;

  // Initialize shared preferences
  static Future<void> init() async {
    _prefs = await SharedPreferences.getInstance();
  }

  // Secure Storage (for tokens)
  static Future<void> saveToken(String token) async {
    await _storage.write(key: ApiConfig.accessTokenKey, value: token);
  }

  static Future<String?> getToken() async {
    return await _storage.read(key: ApiConfig.accessTokenKey);
  }

  static Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: ApiConfig.refreshTokenKey, value: token);
  }

  static Future<String?> getRefreshToken() async {
    return await _storage.read(key: ApiConfig.refreshTokenKey);
  }

  static Future<void> saveUserId(String userId) async {
    await _storage.write(key: ApiConfig.userIdKey, value: userId);
  }

  static Future<String?> getUserId() async {
    return await _storage.read(key: ApiConfig.userIdKey);
  }

  static Future<void> clearAuthData() async {
    await _storage.delete(key: ApiConfig.accessTokenKey);
    await _storage.delete(key: ApiConfig.refreshTokenKey);
    await _storage.delete(key: ApiConfig.userIdKey);
  }

  // Shared Preferences (for app settings)
  static Future<void> setBool(String key, bool value) async {
    await _prefs?.setBool(key, value);
  }

  static bool? getBool(String key) {
    return _prefs?.getBool(key);
  }

  static Future<void> setString(String key, String value) async {
    await _prefs?.setString(key, value);
  }

  static String? getString(String key) {
    return _prefs?.getString(key);
  }

  static Future<void> remove(String key) async {
    await _prefs?.remove(key);
  }

  static Future<void> clear() async {
    await _prefs?.clear();
  }
}
