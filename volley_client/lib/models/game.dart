import 'package:intl/intl.dart';
import 'user.dart';

enum GameCategory {
  soccer,
  basketball,
  pickleball,
  flagFootball,
  volleyball,
  ultimateFrisbee,
  tennis,
  other;

  String get displayName {
    switch (this) {
      case GameCategory.soccer:
        return 'Soccer';
      case GameCategory.basketball:
        return 'Basketball';
      case GameCategory.pickleball:
        return 'Pickleball';
      case GameCategory.flagFootball:
        return 'Football';
      case GameCategory.volleyball:
        return 'Volleyball';
      case GameCategory.ultimateFrisbee:
        return 'Ultimate Frisbee';
      case GameCategory.tennis:
        return 'Tennis';
      case GameCategory.other:
        return 'Other';
    }
  }

  static GameCategory fromString(String value) {
    switch (value) {
      case 'soccer':
        return GameCategory.soccer;
      case 'basketball':
        return GameCategory.basketball;
      case 'pickleball':
        return GameCategory.pickleball;
      case 'flag_football':
        return GameCategory.flagFootball;
      case 'volleyball':
        return GameCategory.volleyball;
      case 'ultimate_frisbee':
        return GameCategory.ultimateFrisbee;
      case 'tennis':
        return GameCategory.tennis;
      default:
        return GameCategory.other;
    }
  }

  String toApiString() {
    switch (this) {
      case GameCategory.soccer:
        return 'soccer';
      case GameCategory.basketball:
        return 'basketball';
      case GameCategory.pickleball:
        return 'pickleball';
      case GameCategory.flagFootball:
        return 'flag_football';
      case GameCategory.volleyball:
        return 'volleyball';
      case GameCategory.ultimateFrisbee:
        return 'ultimate_frisbee';
      case GameCategory.tennis:
        return 'tennis';
      case GameCategory.other:
        return 'other';
    }
  }
}

enum GameStatus {
  open,
  full,
  closed,
  inProgress,
  completed,
  cancelled;

  static GameStatus fromString(String value) {
    switch (value) {
      case 'open':
        return GameStatus.open;
      case 'full':
        return GameStatus.full;
      case 'closed':
        return GameStatus.closed;
      case 'in_progress':
        return GameStatus.inProgress;
      case 'completed':
        return GameStatus.completed;
      case 'cancelled':
        return GameStatus.cancelled;
      default:
        return GameStatus.open;
    }
  }
}

enum PricingType {
  free,
  total,
  perPerson;

  static PricingType fromString(String value) {
    switch (value) {
      case 'free':
        return PricingType.free;
      case 'total':
        return PricingType.total;
      case 'per_person':
        return PricingType.perPerson;
      default:
        return PricingType.free;
    }
  }
}

enum SkillLevel {
  beginner,
  intermediate,
  advanced,
  all;

  String get displayName {
    switch (this) {
      case SkillLevel.beginner:
        return 'Beginner';
      case SkillLevel.intermediate:
        return 'Intermediate';
      case SkillLevel.advanced:
        return 'Advanced';
      case SkillLevel.all:
        return 'All Levels';
    }
  }

  static SkillLevel fromString(String value) {
    switch (value) {
      case 'beginner':
        return SkillLevel.beginner;
      case 'intermediate':
        return SkillLevel.intermediate;
      case 'advanced':
        return SkillLevel.advanced;
      default:
        return SkillLevel.all;
    }
  }
}

class Location {
  final String name;
  final String? address;
  final double? latitude;
  final double? longitude;
  final String? notes;

  Location({
    required this.name,
    this.address,
    this.latitude,
    this.longitude,
    this.notes,
  });

  factory Location.fromJson(Map<String, dynamic> json) {
    return Location(
      name: json['name'] as String,
      address: json['address'] as String?,
      latitude: json['latitude'] != null
          ? (json['latitude'] as num).toDouble()
          : null,
      longitude: json['longitude'] != null
          ? (json['longitude'] as num).toDouble()
          : null,
      notes: json['notes'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      if (address != null) 'address': address,
      if (latitude != null) 'latitude': latitude,
      if (longitude != null) 'longitude': longitude,
      if (notes != null) 'notes': notes,
    };
  }
}

class Pricing {
  final PricingType type;
  final int? amountCents;
  final String? currency;

  Pricing({
    required this.type,
    this.amountCents,
    this.currency,
  });

  factory Pricing.fromJson(Map<String, dynamic> json) {
    return Pricing(
      type: PricingType.fromString(json['type'] as String),
      amountCents: json['amountCents'] != null
          ? (json['amountCents'] as num).toInt()
          : null,
      currency: json['currency'] as String?,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type == PricingType.free
          ? 'free'
          : type == PricingType.total
              ? 'total'
              : 'per_person',
      if (amountCents != null) 'amountCents': amountCents,
      if (currency != null) 'currency': currency,
    };
  }

  bool get isFree => type == PricingType.free || (amountCents ?? 0) == 0;

  String get displayPrice {
    if (isFree) return 'Free';
    final dollars = (amountCents ?? 0) / 100;
    return '\$${dollars.toStringAsFixed(0)}';
  }
}

/// GameSummary contains essential game details for list views
class GameSummary {
  final String id;
  final GameCategory category;
  final String? title;
  final String? description;
  final Location location;
  final DateTime startTime;
  final int durationMinutes;
  final int maxParticipants;
  final int signupCount;
  final Pricing pricing;
  final DateTime signupDeadline;
  final SkillLevel skillLevel;
  final GameStatus status;
  final String? userParticipationStatus;

  GameSummary({
    required this.id,
    required this.category,
    this.title,
    this.description,
    required this.location,
    required this.startTime,
    required this.durationMinutes,
    required this.maxParticipants,
    required this.signupCount,
    required this.pricing,
    required this.signupDeadline,
    required this.skillLevel,
    required this.status,
    this.userParticipationStatus,
  });

  factory GameSummary.fromJson(Map<String, dynamic> json) {
    return GameSummary(
      id: json['id'] as String,
      category: GameCategory.fromString(json['category'] as String),
      title: json['title'] as String?,
      description: json['description'] as String?,
      location: Location.fromJson(json['location'] as Map<String, dynamic>),
      startTime: DateTime.parse(json['startTime'] as String),
      durationMinutes: json['durationMinutes'] != null
          ? (json['durationMinutes'] as num).toInt()
          : 0,
      maxParticipants: json['maxParticipants'] != null
          ? (json['maxParticipants'] as num).toInt()
          : 0,
      signupCount: json['signupCount'] != null
          ? (json['signupCount'] as num).toInt()
          : 0,
      pricing: Pricing.fromJson(json['pricing'] as Map<String, dynamic>),
      signupDeadline: DateTime.parse(json['signupDeadline'] as String),
      skillLevel: SkillLevel.fromString(json['skillLevel'] as String),
      status: GameStatus.fromString(json['status'] as String),
      userParticipationStatus: json['userParticipationStatus'] as String?,
    );
  }

  bool get isFull => signupCount >= maxParticipants;
  bool get isUserJoined => userParticipationStatus != null &&
      (userParticipationStatus == 'confirmed' || userParticipationStatus == 'waitlist');
}

/// Game represents a full game with all details including participants
class Game {
  final String id;
  final User? owner;
  final GameCategory category;
  final String? title;
  final String? description;
  final Location location;
  final DateTime startTime;
  final int durationMinutes;
  final int maxParticipants;
  final List<dynamic> confirmedParticipants; // Confirmed participants (up to max)
  final List<dynamic> waitlist; // Waitlisted participants (beyond max)
  final Pricing pricing;
  final DateTime signupDeadline;
  final DateTime? dropDeadline;
  final SkillLevel skillLevel;
  final String? notes;
  final GameStatus status;
  final DateTime? cancelledAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  Game({
    required this.id,
    this.owner,
    required this.category,
    this.title,
    this.description,
    required this.location,
    required this.startTime,
    required this.durationMinutes,
    required this.maxParticipants,
    required this.confirmedParticipants,
    required this.waitlist,
    required this.pricing,
    required this.signupDeadline,
    this.dropDeadline,
    required this.skillLevel,
    this.notes,
    required this.status,
    this.cancelledAt,
    required this.createdAt,
    required this.updatedAt,
  });

  // Computed properties
  int get currentParticipants => confirmedParticipants.length;
  int get waitlistCount => waitlist.length;

  factory Game.fromJson(Map<String, dynamic> json) {
    return Game(
      id: json['id'] as String,
      owner: json['owner'] != null
          ? User.fromJson(json['owner'] as Map<String, dynamic>)
          : null,
      category: GameCategory.fromString(json['category'] as String),
      title: json['title'] as String?,
      description: json['description'] as String?,
      location: Location.fromJson(json['location'] as Map<String, dynamic>),
      startTime: DateTime.parse(json['startTime'] as String),
      durationMinutes: json['durationMinutes'] != null
          ? (json['durationMinutes'] as num).toInt()
          : 0,
      maxParticipants: json['maxParticipants'] != null
          ? (json['maxParticipants'] as num).toInt()
          : 0,
      confirmedParticipants: json['confirmedParticipants'] != null
          ? (json['confirmedParticipants'] as List<dynamic>)
          : [],
      waitlist: json['waitlist'] != null
          ? (json['waitlist'] as List<dynamic>)
          : [],
      pricing: Pricing.fromJson(json['pricing'] as Map<String, dynamic>),
      signupDeadline: DateTime.parse(json['signupDeadline'] as String),
      dropDeadline: json['dropDeadline'] != null
          ? DateTime.parse(json['dropDeadline'] as String)
          : null,
      skillLevel: SkillLevel.fromString(json['skillLevel'] as String),
      notes: json['notes'] as String?,
      status: GameStatus.fromString(json['status'] as String),
      cancelledAt: json['cancelledAt'] != null
          ? DateTime.parse(json['cancelledAt'] as String)
          : null,
      createdAt: DateTime.parse(json['createdAt'] as String),
      updatedAt: DateTime.parse(json['updatedAt'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      if (owner != null) 'owner': owner!.toJson(),
      'category': category.toApiString(),
      if (title != null) 'title': title,
      if (description != null) 'description': description,
      'location': location.toJson(),
      'startTime': startTime.toIso8601String(),
      'durationMinutes': durationMinutes,
      'maxParticipants': maxParticipants,
      'confirmedParticipants': confirmedParticipants,
      'waitlist': waitlist,
      'pricing': pricing.toJson(),
      'signupDeadline': signupDeadline.toIso8601String(),
      if (dropDeadline != null) 'dropDeadline': dropDeadline!.toIso8601String(),
      'skillLevel': skillLevel.toString().split('.').last,
      if (notes != null) 'notes': notes,
      'status': status.toString().split('.').last,
      if (cancelledAt != null) 'cancelledAt': cancelledAt!.toIso8601String(),
      'createdAt': createdAt.toIso8601String(),
      'updatedAt': updatedAt.toIso8601String(),
    };
  }

  bool get isFull => currentParticipants >= maxParticipants;
  bool get hasWaitlist => waitlistCount > 0;
}

/// CreateGameRequest for creating a new game
class CreateGameRequest {
  final GameCategory category;
  final String? title;
  final String? description;
  final Location location;
  final DateTime startTime;
  final int durationMinutes;
  final int maxParticipants;
  final Pricing pricing;
  final DateTime? signupDeadline;
  final SkillLevel skillLevel;
  final String? notes;

  CreateGameRequest({
    required this.category,
    this.title,
    this.description,
    required this.location,
    required this.startTime,
    required this.durationMinutes,
    required this.maxParticipants,
    required this.pricing,
    this.signupDeadline,
    this.skillLevel = SkillLevel.all,
    this.notes,
  });

  Map<String, dynamic> toJson() {
    String formatDateTime(DateTime dt) {
      final offset = dt.timeZoneOffset;
      final offsetSign = offset.isNegative ? '-' : '+';
      final offsetHours = offset.inHours.abs().toString().padLeft(2, '0');
      final offsetMinutes = (offset.inMinutes.abs() % 60).toString().padLeft(2, '0');

      return '${dt.year.toString().padLeft(4, '0')}-${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')}'
             'T${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}:${dt.second.toString().padLeft(2, '0')}'
             '$offsetSign$offsetHours:$offsetMinutes';
    }

    return {
      'category': category.toApiString(),
      if (title != null && title!.isNotEmpty) 'title': title,
      if (description != null && description!.isNotEmpty) 'description': description,
      'location': location.toJson(),
      'startTime': formatDateTime(startTime),
      'durationMinutes': durationMinutes,
      'maxParticipants': maxParticipants,
      'pricing': pricing.toJson(),
      if (signupDeadline != null) 'signupDeadline': formatDateTime(signupDeadline!),
      'skillLevel': skillLevel.toString().split('.').last,
      if (notes != null && notes!.isNotEmpty) 'notes': notes,
    };
  }
}
