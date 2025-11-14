import 'package:json_annotation/json_annotation.dart';

part 'user.g.dart';

@JsonSerializable()
class User {
  final String id;
  final String email;
  final String firstName;
  final String lastName;
  final List<String> roles;

  User({
    required this.id,
    required this.email,
    required this.firstName,
    required this.lastName,
    required this.roles,
  });

  String get fullName => '$firstName $lastName';

  bool get isAdmin =>
      roles.contains('ADMIN') || roles.contains('VENUE_ADMIN');

  bool get isOperator => roles.contains('OPERATOR');

  bool get isMember => roles.contains('MEMBER');

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
  Map<String, dynamic> toJson() => _$UserToJson(this);
}
