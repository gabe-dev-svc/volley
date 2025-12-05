class User {
  final String id;
  final String email;
  final String firstName;
  final String lastName;
  final DateTime? createdAt;

  User({
    required this.id,
    required this.email,
    required this.firstName,
    required this.lastName,
    this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String,
      email: json['email'] as String,
      firstName: json['firstName'] as String,
      lastName: json['lastName'] as String,
      createdAt: json['createdAt'] != null
          ? DateTime.parse(json['createdAt'] as String)
          : null,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'email': email,
      'firstName': firstName,
      'lastName': lastName,
      'createdAt': createdAt?.toIso8601String(),
    };
  }

  String get fullName => '$firstName $lastName';
}
