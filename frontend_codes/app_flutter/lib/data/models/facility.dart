import 'package:json_annotation/json_annotation.dart';

part 'facility.g.dart';

@JsonSerializable()
class Facility {
  final String id;
  final String venueId;
  final String name;
  final String description;
  final String surface;
  final String openAt;
  final String closeAt;
  final bool available;
  final int? weekdayRateCents;
  final int? weekendRateCents;
  final String currency;

  Facility({
    required this.id,
    required this.venueId,
    required this.name,
    required this.description,
    required this.surface,
    required this.openAt,
    required this.closeAt,
    required this.available,
    this.weekdayRateCents,
    this.weekendRateCents,
    required this.currency,
  });

  String get priceDisplay {
    if (weekdayRateCents == null) return 'N/A';
    return '\$${(weekdayRateCents! / 100).toStringAsFixed(2)}';
  }

  factory Facility.fromJson(Map<String, dynamic> json) =>
      _$FacilityFromJson(json);
  Map<String, dynamic> toJson() => _$FacilityToJson(this);
}
