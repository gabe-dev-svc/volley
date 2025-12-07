import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../models/game.dart';
import '../providers/game_provider.dart';
import '../providers/auth_provider.dart';
import '../theme/app_theme.dart';
import '../widgets/app_bottom_nav.dart';

class GamesNearYouScreen extends StatefulWidget {
  const GamesNearYouScreen({super.key});

  @override
  State<GamesNearYouScreen> createState() => _GamesNearYouScreenState();
}

class _GamesNearYouScreenState extends State<GamesNearYouScreen> {
  final TextEditingController _locationController = TextEditingController();
  bool _isSearching = false;
  bool _usingCurrentLocation = false;

  @override
  void initState() {
    super.initState();
    _initializeLocation();
  }

  Future<void> _initializeLocation() async {
    final gameProvider = Provider.of<GameProvider>(context, listen: false);

    // Try to get current location
    final success = await gameProvider.getCurrentLocation();

    if (success && mounted) {
      // If we got GPS location, show "Current Location" and set flag
      setState(() {
        _usingCurrentLocation = true;
        _locationController.text = 'Current Location';
      });
      await gameProvider.loadGames();
    } else if (mounted) {
      // If GPS failed, use default zip code
      setState(() {
        _usingCurrentLocation = false;
        _locationController.text = '90210';
      });
      final geocoded = await gameProvider.geocodeLocation('90210');
      if (geocoded) {
        await gameProvider.loadGames();
      }
    }
  }

  Future<void> _searchLocation() async {
    final location = _locationController.text.trim();
    if (location.isEmpty) return;

    // Don't search if they haven't changed from "Current Location"
    if (location == 'Current Location') return;

    setState(() {
      _isSearching = true;
      _usingCurrentLocation = false; // Clear the flag when manually searching
    });

    final gameProvider = Provider.of<GameProvider>(context, listen: false);
    final success = await gameProvider.geocodeLocation(location);

    if (success) {
      await gameProvider.loadGames();
    } else if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(gameProvider.errorMessage ?? 'Failed to find location'),
          backgroundColor: Colors.red,
        ),
      );
      gameProvider.clearError();
    }

    setState(() {
      _isSearching = false;
    });
  }

  @override
  void dispose() {
    _locationController.dispose();
    super.dispose();
  }

  IconData _getCategoryIcon(GameCategory category) {
    switch (category) {
      case GameCategory.basketball:
        return Icons.sports_basketball;
      case GameCategory.soccer:
        return Icons.sports_soccer;
      case GameCategory.volleyball:
        return Icons.sports_volleyball;
      case GameCategory.pickleball:
        return Icons.sports_tennis;
      default:
        return Icons.sports;
    }
  }

  String _formatDateTime(DateTime dateTime) {
    final now = DateTime.now();
    final today = DateTime(now.year, now.month, now.day);
    final tomorrow = today.add(const Duration(days: 1));
    final gameDate = DateTime(dateTime.year, dateTime.month, dateTime.day);

    String dayString;
    if (gameDate == today) {
      dayString = 'Today';
    } else if (gameDate == tomorrow) {
      dayString = 'Tomorrow';
    } else {
      dayString = DateFormat('E, MMM d').format(dateTime);
    }

    final timeString = DateFormat('@ h:mm a').format(dateTime);
    return '$dayString $timeString';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      bottomNavigationBar: const AppBottomNav(currentIndex: 1),
      appBar: AppBar(
        backgroundColor: Colors.white,
        elevation: 0,
        title: const Text(
          'Games Near You',
          style: TextStyle(
            color: Colors.black,
            fontSize: 24,
            fontWeight: FontWeight.bold,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.tune, color: Colors.black),
            onPressed: () {
              // TODO: Open filter sheet
            },
          ),
        ],
      ),
      body: Consumer<GameProvider>(
        builder: (context, gameProvider, child) {
          return Column(
            children: [
              // Location and Search Radius
              Padding(
                padding: const EdgeInsets.all(20.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Location',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _locationController,
                      textInputAction: TextInputAction.search,
                      onSubmitted: (_) => _searchLocation(),
                      decoration: InputDecoration(
                        filled: true,
                        fillColor: AppColors.primaryLight,
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                          borderSide: BorderSide.none,
                        ),
                        hintText: 'Enter zip code or address',
                        suffixIcon: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            if (_isSearching)
                              const Padding(
                                padding: EdgeInsets.all(12.0),
                                child: SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: AppColors.primary,
                                  ),
                                ),
                              )
                            else
                              IconButton(
                                icon: const Icon(Icons.search, color: AppColors.primary),
                                onPressed: _searchLocation,
                              ),
                            IconButton(
                              icon: const Icon(Icons.my_location, color: AppColors.primary),
                              onPressed: _initializeLocation,
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 20),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Search Radius',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Text(
                          '${gameProvider.radiusInMiles.toInt()} miles',
                          style: const TextStyle(
                            fontSize: 16,
                            color: AppColors.primary,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    SliderTheme(
                      data: SliderTheme.of(context).copyWith(
                        activeTrackColor: AppColors.primary,
                        inactiveTrackColor: AppColors.primaryLight,
                        thumbColor: AppColors.primary,
                        overlayColor: AppColors.primary.withValues(alpha: 0.2),
                        trackHeight: 6,
                      ),
                      child: Slider(
                        value: gameProvider.radiusInMiles,
                        min: 1,
                        max: 50,
                        onChanged: (value) {
                          gameProvider.setRadiusInMiles(value);
                        },
                        onChangeEnd: (value) {
                          gameProvider.loadGames();
                        },
                      ),
                    ),
                  ],
                ),
              ),

              // Category Pills
              SizedBox(
                height: 50,
                child: ListView(
                  scrollDirection: Axis.horizontal,
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  children: [
                    _CategoryPill(
                      label: 'Football',
                      isSelected: gameProvider.selectedCategory == GameCategory.flagFootball,
                      onTap: () {
                        gameProvider.setCategory(
                          gameProvider.selectedCategory == GameCategory.flagFootball
                              ? null
                              : GameCategory.flagFootball,
                        );
                      },
                    ),
                    const SizedBox(width: 12),
                    _CategoryPill(
                      label: 'Soccer',
                      isSelected: gameProvider.selectedCategory == GameCategory.soccer,
                      onTap: () {
                        gameProvider.setCategory(
                          gameProvider.selectedCategory == GameCategory.soccer
                              ? null
                              : GameCategory.soccer,
                        );
                      },
                    ),
                    const SizedBox(width: 12),
                    _CategoryPill(
                      label: 'Basketball',
                      isSelected: gameProvider.selectedCategory == GameCategory.basketball,
                      onTap: () {
                        gameProvider.setCategory(
                          gameProvider.selectedCategory == GameCategory.basketball
                              ? null
                              : GameCategory.basketball,
                        );
                      },
                    ),
                    const SizedBox(width: 12),
                    _CategoryPill(
                      label: 'Pickleball',
                      isSelected: gameProvider.selectedCategory == GameCategory.pickleball,
                      onTap: () {
                        gameProvider.setCategory(
                          gameProvider.selectedCategory == GameCategory.pickleball
                              ? null
                              : GameCategory.pickleball,
                        );
                      },
                    ),
                  ],
                ),
              ),

              const SizedBox(height: 20),

              // Games List
              Expanded(
                child: gameProvider.isLoading
                    ? const Center(child: CircularProgressIndicator())
                    : gameProvider.errorMessage != null
                        ? Center(
                            child: Column(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Text(
                                  gameProvider.errorMessage!,
                                  style: const TextStyle(color: Colors.red),
                                ),
                                const SizedBox(height: 16),
                                ElevatedButton(
                                  onPressed: () => gameProvider.loadGames(),
                                  child: const Text('Retry'),
                                ),
                              ],
                            ),
                          )
                        : gameProvider.games.isEmpty
                            ? const Center(
                                child: Text('No games found nearby'),
                              )
                            : ListView.builder(
                                padding: const EdgeInsets.symmetric(horizontal: 20),
                                itemCount: gameProvider.games.length,
                                itemBuilder: (context, index) {
                                  final game = gameProvider.games[index];
                                  return _GameCard(
                                    game: game,
                                    categoryIcon: _getCategoryIcon(game.category),
                                    formattedDateTime: _formatDateTime(game.startTime),
                                  );
                                },
                              ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _CategoryPill extends StatelessWidget {
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _CategoryPill({
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primary : AppColors.primaryLight,
          borderRadius: BorderRadius.circular(25),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? AppColors.white : AppColors.black,
            fontSize: 16,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }
}

class _GameCard extends StatelessWidget {
  final GameSummary game;
  final IconData categoryIcon;
  final String formattedDateTime;

  const _GameCard({
    required this.game,
    required this.categoryIcon,
    required this.formattedDateTime,
  });

  @override
  Widget build(BuildContext context) {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final gameProvider = Provider.of<GameProvider>(context, listen: false);

    return GestureDetector(
      onTap: () {
        context.go('/game/${game.id}');
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(20),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.08),
              blurRadius: 10,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Expanded(
                child: Text(
                  game.title ?? '${game.category.displayName} Game',
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              Row(
                children: [
                  if (game.userParticipationStatus != null &&
                      (game.userParticipationStatus == 'confirmed' ||
                       game.userParticipationStatus == 'waitlist')) ...[
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                      decoration: BoxDecoration(
                        color: game.userParticipationStatus == 'confirmed'
                            ? Colors.green.shade100
                            : Colors.orange.shade100,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        game.userParticipationStatus == 'confirmed'
                            ? 'Confirmed'
                            : 'Waitlist',
                        style: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: game.userParticipationStatus == 'confirmed'
                              ? Colors.green.shade700
                              : Colors.orange.shade700,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                  ],
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: AppColors.primaryLight,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(
                      categoryIcon,
                      color: AppColors.primary,
                      size: 24,
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              const Icon(Icons.calendar_today, size: 16, color: AppColors.textSecondary),
              const SizedBox(width: 8),
              Text(
                formattedDateTime,
                style: const TextStyle(
                  fontSize: 14,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              const Icon(Icons.location_on, size: 16, color: AppColors.textSecondary),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  game.location.name,
                  style: const TextStyle(
                    fontSize: 14,
                    color: AppColors.textSecondary,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              Row(
                children: [
                  const Icon(Icons.attach_money, size: 18, color: AppColors.textPrimary),
                  const SizedBox(width: 4),
                  Text(
                    game.pricing.displayPrice,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              const SizedBox(width: 24),
              Row(
                children: [
                  const Icon(Icons.bar_chart, size: 18, color: AppColors.textPrimary),
                  const SizedBox(width: 4),
                  Text(
                    game.skillLevel.displayName,
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
              const SizedBox(width: 24),
              Row(
                children: [
                  const Icon(Icons.people, size: 18, color: AppColors.textPrimary),
                  const SizedBox(width: 4),
                  Text(
                    '${game.signupCount}/${game.maxParticipants}',
                    style: const TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: game.isUserJoined
                  ? null
                  : game.isFull
                      ? null
                      : () async {
                          if (!authProvider.isAuthenticated) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              const SnackBar(
                                content: Text('Please login to join a game'),
                              ),
                            );
                            return;
                          }

                          // Get auth token
                          final token = await authProvider.getToken();
                          if (token == null) {
                            if (context.mounted) {
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('Authentication error. Please login again.'),
                                  backgroundColor: Colors.red,
                                ),
                              );
                            }
                            return;
                          }

                          final success = await gameProvider.joinGame(
                            game.id,
                            token,
                          );

                          if (context.mounted) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text(
                                  success
                                      ? 'Successfully joined game!'
                                      : 'Failed to join game',
                                ),
                                backgroundColor: success ? Colors.green : Colors.red,
                              ),
                            );
                          }
                        },
              style: ElevatedButton.styleFrom(
                backgroundColor: game.isUserJoined || game.isFull
                    ? Colors.grey.shade300
                    : AppColors.primary,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: Text(
                game.isUserJoined
                    ? 'Already Joined'
                    : game.isFull
                        ? 'Join Waitlist'
                        : 'Quick Join',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: game.isUserJoined || game.isFull
                      ? Colors.grey.shade600
                      : AppColors.black,
                ),
              ),
            ),
          ),
        ],
      ),
      ),
    );
  }
}
