import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/theme/app_colors.dart';
import '../../../data/repositories/booking_repository.dart';
import '../../../data/models/facility.dart';
import '../../../widgets/common/loading_widget.dart';
import '../../../widgets/common/error_widget.dart';
import 'facility_details_screen.dart';

final facilitiesProvider = FutureProvider<List<Facility>>((ref) async {
  final repository = ref.read(bookingRepositoryProvider);
  return repository.getFacilities(limit: 50);
});

class FacilitiesScreen extends ConsumerWidget {
  const FacilitiesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final facilitiesAsync = ref.watch(facilitiesProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Facilities'),
        actions: [
          IconButton(
            icon: const Icon(Icons.filter_list),
            onPressed: () {
              // TODO: Implement filters
            },
          ),
        ],
      ),
      body: facilitiesAsync.when(
        data: (facilities) {
          if (facilities.isEmpty) {
            return Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.event_busy,
                    size: 64,
                    color: AppColors.textHint,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'No facilities available',
                    style: Theme.of(context).textTheme.titleLarge,
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: () async {
              ref.invalidate(facilitiesProvider);
            },
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: facilities.length,
              itemBuilder: (context, index) {
                final facility = facilities[index];
                return FacilityCard(
                  facility: facility,
                  onTap: () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (context) => FacilityDetailsScreen(
                          facility: facility,
                        ),
                      ),
                    );
                  },
                );
              },
            ),
          );
        },
        loading: () => ListView.builder(
          padding: const EdgeInsets.all(16),
          itemCount: 5,
          itemBuilder: (context, index) => const ListItemShimmer(),
        ),
        error: (error, stack) => ErrorDisplayWidget(
          message: error.toString(),
          onRetry: () => ref.invalidate(facilitiesProvider),
        ),
      ),
    );
  }
}

class FacilityCard extends StatelessWidget {
  final Facility facility;
  final VoidCallback onTap;

  const FacilityCard({
    super.key,
    required this.facility,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
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
                      facility.name,
                      style: Theme.of(context).textTheme.titleLarge?.copyWith(
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                  ),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                    decoration: BoxDecoration(
                      color: (facility.available
                              ? AppColors.facilityAvailable
                              : AppColors.facilityUnavailable)
                          .withOpacity(0.1),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(
                          facility.available ? Icons.check_circle : Icons.cancel,
                          size: 16,
                          color: facility.available
                              ? AppColors.facilityAvailable
                              : AppColors.facilityUnavailable,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          facility.available ? 'Available' : 'Unavailable',
                          style: TextStyle(
                            color: facility.available
                                ? AppColors.facilityAvailable
                                : AppColors.facilityUnavailable,
                            fontSize: 12,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Text(
                facility.description,
                style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: AppColors.textSecondary,
                    ),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 12),
              Row(
                children: [
                  Icon(
                    Icons.sports_tennis,
                    size: 16,
                    color: AppColors.textSecondary,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    facility.surface,
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                  const SizedBox(width: 16),
                  Icon(
                    Icons.schedule,
                    size: 16,
                    color: AppColors.textSecondary,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    '${facility.openAt} - ${facility.closeAt}',
                    style: Theme.of(context).textTheme.bodySmall,
                  ),
                ],
              ),
              if (facility.weekdayRateCents != null) ...[
                const SizedBox(height: 12),
                Row(
                  children: [
                    Text(
                      'From ${facility.priceDisplay}',
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            color: AppColors.primary,
                            fontWeight: FontWeight.bold,
                          ),
                    ),
                    const Text(' / hour'),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
