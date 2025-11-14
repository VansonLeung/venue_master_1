import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/config/api_config.dart';
import '../../core/utils/logger.dart';
import '../../core/utils/storage_service.dart';
import '../models/booking.dart';
import '../models/facility.dart';
import '../providers/dio_provider.dart';

final bookingRepositoryProvider = Provider<BookingRepository>((ref) {
  return BookingRepository(ref.read(dioProvider));
});

class BookingRepository {
  final Dio _dio;

  BookingRepository(this._dio);

  Future<List<Facility>> getFacilities({int limit = 20}) async {
    try {
      final userId = await StorageService.getUserId();
      final response = await _dio.get(
        '${ApiConfig.bookingUrl}/facilities',
        queryParameters: {'limit': limit},
        options: Options(
          headers: {
            'X-User-ID': userId,
            'X-User-Roles': 'MEMBER',
          },
        ),
      );

      return (response.data as List)
          .map((json) => Facility.fromJson(json))
          .toList();
    } catch (e) {
      AppLogger.error('Failed to fetch facilities', e);
      throw Exception('Failed to load facilities');
    }
  }

  Future<List<Booking>> getUserBookings({int limit = 20}) async {
    try {
      final userId = await StorageService.getUserId();
      if (userId == null) throw Exception('User not authenticated');

      final response = await _dio.get(
        '${ApiConfig.bookingUrl}/bookings',
        queryParameters: {
          'userId': userId,
          'limit': limit,
        },
        options: Options(
          headers: {
            'X-User-ID': userId,
            'X-User-Roles': 'MEMBER',
          },
        ),
      );

      return (response.data as List)
          .map((json) => Booking.fromJson(json))
          .toList();
    } catch (e) {
      AppLogger.error('Failed to fetch bookings', e);
      throw Exception('Failed to load bookings');
    }
  }

  Future<Booking> createBooking({
    required String facilityId,
    required DateTime startsAt,
    required DateTime endsAt,
  }) async {
    try {
      final userId = await StorageService.getUserId();
      if (userId == null) throw Exception('User not authenticated');

      final response = await _dio.post(
        '${ApiConfig.bookingUrl}/bookings',
        data: {
          'facilityId': facilityId,
          'userId': userId,
          'startsAt': startsAt.toUtc().toIso8601String(),
          'endsAt': endsAt.toUtc().toIso8601String(),
        },
        options: Options(
          headers: {
            'X-User-ID': userId,
            'X-User-Roles': 'MEMBER',
          },
        ),
      );

      return Booking.fromJson(response.data);
    } catch (e) {
      AppLogger.error('Failed to create booking', e);
      if (e is DioException && e.response?.data != null) {
        final error = e.response!.data['error'] ?? 'Failed to create booking';
        throw Exception(error);
      }
      throw Exception('Failed to create booking');
    }
  }

  Future<Booking> cancelBooking(String bookingId) async {
    try {
      final userId = await StorageService.getUserId();
      if (userId == null) throw Exception('User not authenticated');

      final response = await _dio.patch(
        '${ApiConfig.bookingUrl}/bookings/$bookingId/cancel',
        options: Options(
          headers: {
            'X-User-ID': userId,
            'X-User-Roles': 'MEMBER',
          },
        ),
      );

      return Booking.fromJson(response.data);
    } catch (e) {
      AppLogger.error('Failed to cancel booking', e);
      throw Exception('Failed to cancel booking');
    }
  }
}
