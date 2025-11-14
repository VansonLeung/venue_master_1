import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_colors.dart';
import '../../../data/repositories/booking_repository.dart';
import '../../../data/models/booking.dart';
import '../../../widgets/common/loading_widget.dart';
import '../../../widgets/common/error_widget.dart';
import 'package:intl/intl.dart';

final upcomingBookingsProvider = FutureProvider<List<Booking>>((ref) async {
  final repository = ref.read(bookingRepositoryProvider);
  return repository.getUserBookings(limit: 5);
});

class UpcomingBookingsWidget extends ConsumerWidget {
  const UpcomingBookingsWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final bookingsAsync = ref.watch(upcomingBookingsProvider);

    return bookingsAsync.when(
      data: (bookings) {
        if (bookings.isEmpty) {
          return Card(
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                children: [
                  Icon(
                    Icons.calendar_today_outlined,
                    size: 48,
                    color: AppColors.textHint,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'No upcoming bookings',
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Book a facility to get started!',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppColors.textHint,
                        ),
                  ),
                ],
              ),
            ),
          );
        }

        return Column(
          children: bookings.map((booking) {
            return BookingCard(booking: booking);
          }).toList(),
        );
      },
      loading: () => const LoadingWidget(),
      error: (error, stack) => ErrorDisplayWidget(
        message: error.toString(),
        onRetry: () => ref.refresh(upcomingBookingsProvider),
      ),
    );
  }
}

class BookingCard extends StatelessWidget {
  final Booking booking;

  const BookingCard({super.key, required this.booking});

  Color _getStatusColor() {
    if (booking.isPending) return AppColors.statusPending;
    if (booking.isConfirmed) return AppColors.statusConfirmed;
    if (booking.isCancelled) return AppColors.statusCancelled;
    return AppColors.statusCompleted;
  }

  @override
  Widget build(BuildContext context) {
    final startDate = DateTime.parse(booking.startsAt);
    final endDate = DateTime.parse(booking.endsAt);

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: Text(
                    booking.facility?.name ?? 'Facility',
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                          fontWeight: FontWeight.bold,
                        ),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  decoration: BoxDecoration(
                    color: _getStatusColor().withOpacity(0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text(
                    booking.status,
                    style: TextStyle(
                      color: _getStatusColor(),
                      fontSize: 12,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Icon(Icons.calendar_today, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 8),
                Text(
                  DateFormat('MMM dd, yyyy').format(startDate),
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.access_time, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 8),
                Text(
                  '${DateFormat('h:mm a').format(startDate)} - ${DateFormat('h:mm a').format(endDate)}',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
              ],
            ),
            const SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.attach_money, size: 16, color: AppColors.textSecondary),
                const SizedBox(width: 8),
                Text(
                  booking.priceDisplay,
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: AppColors.primary,
                      ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
