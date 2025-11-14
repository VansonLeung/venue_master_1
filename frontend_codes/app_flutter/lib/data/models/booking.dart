import 'package:json_annotation/json_annotation.dart';
import 'facility.dart';

part 'booking.g.dart';

@JsonSerializable()
class Booking {
  final String id;
  final String facilityId;
  final String userId;
  final String startsAt;
  final String endsAt;
  final String status;
  final int amountCents;
  final String currency;
  final String? paymentIntent;
  final Facility? facility;

  Booking({
    required this.id,
    required this.facilityId,
    required this.userId,
    required this.startsAt,
    required this.endsAt,
    required this.status,
    required this.amountCents,
    required this.currency,
    this.paymentIntent,
    this.facility,
  });

  String get priceDisplay {
    return '\$${(amountCents / 100).toStringAsFixed(2)}';
  }

  bool get isPending => status == 'PENDING_PAYMENT';
  bool get isConfirmed => status == 'CONFIRMED';
  bool get isCancelled => status == 'CANCELLED';
  bool get isCompleted => status == 'COMPLETED';

  factory Booking.fromJson(Map<String, dynamic> json) =>
      _$BookingFromJson(json);
  Map<String, dynamic> toJson() => _$BookingToJson(this);
}
