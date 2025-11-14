class AppConstants {
  // App Info
  static const String appName = 'Venue Master';
  static const String appVersion = '1.0.0';

  // Pagination
  static const int defaultPageSize = 20;
  static const int maxPageSize = 100;

  // Date Formats
  static const String dateFormat = 'yyyy-MM-dd';
  static const String timeFormat = 'HH:mm';
  static const String dateTimeFormat = 'yyyy-MM-dd HH:mm';
  static const String displayDateFormat = 'MMM dd, yyyy';
  static const String displayTimeFormat = 'hh:mm a';
  static const String displayDateTimeFormat = 'MMM dd, yyyy hh:mm a';

  // Booking
  static const int minBookingDurationMinutes = 30;
  static const int maxBookingDurationHours = 8;
  static const int advanceBookingDays = 90;

  // Currency
  static const String defaultCurrency = 'CAD';
  static const String currencySymbol = '\$';

  // User Roles
  static const String roleMember = 'MEMBER';
  static const String roleOperator = 'OPERATOR';
  static const String roleAdmin = 'ADMIN';
  static const String roleVenueAdmin = 'VENUE_ADMIN';

  // Booking Status
  static const String bookingStatusPending = 'PENDING_PAYMENT';
  static const String bookingStatusConfirmed = 'CONFIRMED';
  static const String bookingStatusCancelled = 'CANCELLED';
  static const String bookingStatusCompleted = 'COMPLETED';

  // Membership Types
  static const String membershipMonthly = 'MONTHLY_PREMIUM';
  static const String membershipYearly = 'YEARLY_PREMIUM';
  static const String membershipBasic = 'BASIC';

  // Animation Durations
  static const Duration shortAnimation = Duration(milliseconds: 200);
  static const Duration mediumAnimation = Duration(milliseconds: 300);
  static const Duration longAnimation = Duration(milliseconds: 500);

  // Debounce Durations
  static const Duration searchDebounce = Duration(milliseconds: 500);
  static const Duration refreshDebounce = Duration(milliseconds: 1000);
}
