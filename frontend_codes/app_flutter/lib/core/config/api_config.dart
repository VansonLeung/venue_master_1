/// API configuration for Venue Master backend services
class ApiConfig {
  // Base URL - Use localhost for iOS Simulator, 10.0.2.2 for Android Emulator
  static const String _baseUrlIOS = 'http://localhost';
  static const String _baseUrlAndroid = 'http://10.0.2.2';

  // Auto-detect platform (you can also make this configurable)
  static String get baseUrl => _baseUrlIOS; // Change to _baseUrlAndroid for Android

  // Service Ports
  static const int gatewayPort = 8080;
  static const int authPort = 8081;
  static const int userPort = 8082;
  static const int bookingPort = 8083;
  static const int foodPort = 8084;
  static const int parkingPort = 8085;
  static const int shopPort = 8086;
  static const int paymentPort = 8087;
  static const int notificationPort = 8088;

  // GraphQL Endpoint
  static String get graphqlUrl => '$baseUrl:$gatewayPort/graphql';

  // REST Endpoints
  static String get authUrl => '$baseUrl:$authPort/v1/auth';
  static String get userUrl => '$baseUrl:$userPort/v1/users';
  static String get bookingUrl => '$baseUrl:$bookingPort/v1';
  static String get foodUrl => '$baseUrl:$foodPort/v1';
  static String get parkingUrl => '$baseUrl:$parkingPort/v1/parking';
  static String get shopUrl => '$baseUrl:$shopPort/v1';
  static String get paymentUrl => '$baseUrl:$paymentPort/v1/payments';
  static String get notificationUrl =>
      '$baseUrl:$notificationPort/v1/notifications';

  // API Timeouts
  static const Duration connectTimeout = Duration(seconds: 30);
  static const Duration receiveTimeout = Duration(seconds: 30);
  static const Duration sendTimeout = Duration(seconds: 30);

  // Storage Keys
  static const String accessTokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userIdKey = 'user_id';
}
