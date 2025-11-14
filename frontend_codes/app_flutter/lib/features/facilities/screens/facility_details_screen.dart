import 'package:flutter/material.dart';
import '../../../core/theme/app_colors.dart';
import '../../../data/models/facility.dart';

class FacilityDetailsScreen extends StatelessWidget {
  final Facility facility;

  const FacilityDetailsScreen({
    super.key,
    required this.facility,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Facility Details'),
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Facility Image Placeholder
            Container(
              height: 200,
              width: double.infinity,
              color: AppColors.primary.withOpacity(0.1),
              child: Icon(
                Icons.sports_tennis,
                size: 80,
                color: AppColors.primary.withOpacity(0.5),
              ),
            ),

            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Title and Status
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          facility.name,
                          style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 12,
                          vertical: 6,
                        ),
                        decoration: BoxDecoration(
                          color: (facility.available
                                  ? AppColors.facilityAvailable
                                  : AppColors.facilityUnavailable)
                              .withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          facility.available ? 'Available' : 'Unavailable',
                          style: TextStyle(
                            color: facility.available
                                ? AppColors.facilityAvailable
                                : AppColors.facilityUnavailable,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // Description
                  Text(
                    facility.description,
                    style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                          color: AppColors.textSecondary,
                        ),
                  ),
                  const SizedBox(height: 24),

                  // Details Card
                  Card(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Details',
                            style: Theme.of(context).textTheme.titleLarge?.copyWith(
                                  fontWeight: FontWeight.bold,
                                ),
                          ),
                          const SizedBox(height: 16),
                          _DetailRow(
                            icon: Icons.sports,
                            label: 'Surface',
                            value: facility.surface,
                          ),
                          const SizedBox(height: 12),
                          _DetailRow(
                            icon: Icons.access_time,
                            label: 'Operating Hours',
                            value: '${facility.openAt} - ${facility.closeAt}',
                          ),
                          if (facility.weekdayRateCents != null) ...[
                            const SizedBox(height: 12),
                            _DetailRow(
                              icon: Icons.attach_money,
                              label: 'Weekday Rate',
                              value: facility.priceDisplay + ' / hour',
                            ),
                          ],
                          if (facility.weekendRateCents != null) ...[
                            const SizedBox(height: 12),
                            _DetailRow(
                              icon: Icons.attach_money,
                              label: 'Weekend Rate',
                              value:
                                  '\$${(facility.weekendRateCents! / 100).toStringAsFixed(2)} / hour',
                            ),
                          ],
                        ],
                      ),
                    ),
                  ),
                  const SizedBox(height: 24),

                  // TODO: Add booking button and calendar
                  if (facility.available)
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        onPressed: () {
                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text('Booking feature coming soon!'),
                            ),
                          );
                        },
                        style: ElevatedButton.styleFrom(
                          padding: const EdgeInsets.symmetric(vertical: 16),
                        ),
                        child: const Text(
                          'Book This Facility',
                          style: TextStyle(fontSize: 16),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _DetailRow extends StatelessWidget {
  final IconData icon;
  final String label;
  final String value;

  const _DetailRow({
    required this.icon,
    required this.label,
    required this.value,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Icon(icon, size: 20, color: AppColors.textSecondary),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppColors.textSecondary,
                    ),
              ),
              const SizedBox(height: 4),
              Text(
                value,
                style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                      fontWeight: FontWeight.w500,
                    ),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
