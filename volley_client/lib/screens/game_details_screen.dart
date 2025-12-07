import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:go_router/go_router.dart';
import '../models/game.dart';
import '../models/user.dart';
import '../services/game_service.dart';
import '../providers/auth_provider.dart';
import '../theme/app_theme.dart';

class GameDetailsScreen extends StatefulWidget {
  final String gameId;

  const GameDetailsScreen({super.key, required this.gameId});

  @override
  State<GameDetailsScreen> createState() => _GameDetailsScreenState();
}

class _GameDetailsScreenState extends State<GameDetailsScreen> {
  final GameService _gameService = GameService();
  bool _isLoading = true;
  Game? _game;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadGameDetails();
  }

  Future<void> _loadGameDetails() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    try {
      // Get auth token if available
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      final authToken = await authProvider.getToken();

      final game = await _gameService.getGame(widget.gameId, authToken: authToken);
      setState(() {
        _game = game;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _errorMessage = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            context.go('/games');
          },
        ),
        title: const Text('Game Details'),
        actions: [
          IconButton(
            icon: const Icon(Icons.share),
            onPressed: () {
              // TODO: Implement share functionality
            },
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _errorMessage != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text('Error: $_errorMessage'),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadGameDetails,
                        child: const Text('Retry'),
                      ),
                    ],
                  ),
                )
              : _game == null
                  ? const Center(child: Text('Game not found'))
                  : _buildGameDetails(),
      bottomNavigationBar: _game != null ? _buildJoinButton() : null,
    );
  }

  Widget _buildGameDetails() {
    final game = _game!;

    return SingleChildScrollView(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Map placeholder
          Container(
            height: 200,
            width: double.infinity,
            color: Colors.grey[300],
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.map, size: 48, color: Colors.grey),
                  const SizedBox(height: 8),
                  Text(
                    '${game.location.latitude?.toStringAsFixed(4)}, ${game.location.longitude?.toStringAsFixed(4)}',
                    style: const TextStyle(color: Colors.grey),
                  ),
                ],
              ),
            ),
          ),

          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Title
                Text(
                  game.title ?? '${game.category.displayName} Game',
                  style: const TextStyle(
                    fontSize: 28,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 12),

                // Category and Status chips
                Row(
                  children: [
                    _buildChip(
                      game.category.displayName,
                      AppColors.success.withOpacity(0.2),
                      AppColors.success,
                    ),
                    const SizedBox(width: 8),
                    _buildChip(
                      game.status == GameStatus.open ? 'Open' : game.status.toString().split('.').last,
                      AppColors.success.withOpacity(0.2),
                      AppColors.success,
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                // Info grid
                Row(
                  children: [
                    Expanded(
                      child: _buildInfoCard(
                        Icons.calendar_today,
                        'Time',
                        DateFormat('EEE, MMM d, h:mm a').format(game.startTime.toLocal()),
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _buildInfoCard(
                        Icons.timer,
                        'Duration',
                        '${game.durationMinutes} mins',
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: _buildInfoCard(
                        Icons.bar_chart,
                        'Skill Level',
                        game.skillLevel.displayName,
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: _buildInfoCard(
                        Icons.attach_money,
                        'Price',
                        game.pricing.displayPrice,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                // Location
                _buildLocationCard(game.location),
                const SizedBox(height: 16),

                // Location Notes
                if (game.location.notes != null) ...[
                  _buildLocationNotesCard(game.location.notes!),
                  const SizedBox(height: 24),
                ],

                // About this game
                if (game.description != null) ...[
                  const Text(
                    'About this game',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    game.description!,
                    style: const TextStyle(
                      fontSize: 14,
                      color: AppColors.textSecondary,
                      height: 1.5,
                    ),
                  ),
                  const SizedBox(height: 24),
                ],

                // Game Notes
                if (game.notes != null) ...[
                  const Text(
                    'Game Notes',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    game.notes!,
                    style: const TextStyle(
                      fontSize: 14,
                      color: AppColors.textSecondary,
                      height: 1.5,
                    ),
                  ),
                  const SizedBox(height: 24),
                ],

                // Participants
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      'Participants (${game.currentParticipants}/${game.maxParticipants})',
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    TextButton(
                      onPressed: () {
                        // TODO: Show all participants
                      },
                      child: const Text('View All'),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                _buildParticipantsPreview(game),
                const SizedBox(height: 24),

                // Host
                if (game.owner != null) ...[
                  const Text(
                    'Host',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  _buildHostCard(game.owner!),
                  const SizedBox(height: 100), // Space for bottom button
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildChip(String label, Color backgroundColor, Color textColor) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: backgroundColor,
        borderRadius: BorderRadius.circular(20),
      ),
      child: Text(
        label,
        style: TextStyle(
          color: textColor,
          fontWeight: FontWeight.w600,
          fontSize: 14,
        ),
      ),
    );
  }

  Widget _buildInfoCard(IconData icon, String title, String value) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.greyLight),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: AppColors.success, size: 24),
          const SizedBox(height: 8),
          Text(
            title,
            style: const TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            value,
            style: const TextStyle(
              fontSize: 13,
              color: AppColors.textSecondary,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLocationCard(Location location) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.greyLight),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: AppColors.success.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(
              Icons.location_on,
              color: AppColors.success,
              size: 24,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  location.name,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                if (location.address != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    location.address!,
                    style: const TextStyle(
                      fontSize: 13,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLocationNotesCard(String notes) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.primaryLight,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.greyLight),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: AppColors.success.withOpacity(0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(
              Icons.notes,
              color: AppColors.success,
              size: 20,
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Location Notes',
                  style: TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  notes,
                  style: const TextStyle(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildParticipantsPreview(Game game) {
    final participants = game.confirmedParticipants.take(5).toList();

    return SizedBox(
      height: 80,
      child: Stack(
        children: [
          ...participants.asMap().entries.map((entry) {
            final index = entry.key;
            final participant = entry.value;

            // Get participant name and last name initial
            String displayName = 'Player ${index + 1}';
            if (participant is Map) {
              final firstName = participant['firstName'] as String?;
              final lastName = participant['lastName'] as String?;

              if (firstName != null && firstName.isNotEmpty) {
                displayName = firstName;
                if (lastName != null && lastName.isNotEmpty) {
                  displayName += ' ${lastName[0]}.';
                }
              }
            }

            return Positioned(
              left: index * 50.0,
              child: Column(
                children: [
                  CircleAvatar(
                    radius: 25,
                    backgroundColor: AppColors.primary,
                    child: Text(
                      displayName.isNotEmpty ? displayName[0].toUpperCase() : '?',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(height: 4),
                  SizedBox(
                    width: 50,
                    child: Text(
                      displayName,
                      style: const TextStyle(
                        fontSize: 10,
                        color: AppColors.textSecondary,
                      ),
                      textAlign: TextAlign.center,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            );
          }),
          if (game.currentParticipants > 5)
            Positioned(
              left: 5 * 50.0,
              child: Column(
                children: [
                  CircleAvatar(
                    radius: 25,
                    backgroundColor: AppColors.success,
                    child: Text(
                      '+${game.currentParticipants - 5}',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  const SizedBox(height: 4),
                  const SizedBox(
                    width: 50,
                    child: Text(
                      'more',
                      style: TextStyle(
                        fontSize: 10,
                        color: AppColors.textSecondary,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildHostCard(User owner) {
    // Display full name for host
    final firstName = owner.firstName;
    final lastName = owner.lastName;
    final email = owner.email;

    String hostName = 'Host';
    String initials = 'H';

    // Build full name
    if (firstName.isNotEmpty && lastName.isNotEmpty) {
      hostName = '$firstName $lastName';
      initials = '${firstName[0]}${lastName[0]}'.toUpperCase();
    } else if (firstName.isNotEmpty) {
      hostName = firstName;
      initials = firstName[0].toUpperCase();
    } else if (lastName.isNotEmpty) {
      hostName = lastName;
      initials = lastName[0].toUpperCase();
    } else if (email.isNotEmpty) {
      hostName = email;
      initials = email[0].toUpperCase();
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppColors.greyLight),
      ),
      child: Row(
        children: [
          CircleAvatar(
            radius: 24,
            backgroundColor: AppColors.primary,
            child: Text(
              initials,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 16,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(
              hostName,
              style: const TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          IconButton(
            icon: const Icon(Icons.message, color: AppColors.success),
            onPressed: () {
              // TODO: Implement messaging
            },
          ),
        ],
      ),
    );
  }

  Widget _buildJoinButton() {
    final game = _game!;
    final authProvider = Provider.of<AuthProvider>(context, listen: false);

    // Check if user is a participant
    final currentUserId = authProvider.user?.id;
    final isUserJoined = currentUserId != null &&
        (game.confirmedParticipants.any((p) => p['id'] == currentUserId) ||
         game.waitlist.any((p) => p['id'] == currentUserId));

    String buttonText;
    Color buttonColor;
    VoidCallback? onPressed;

    if (isUserJoined) {
      buttonText = 'Drop From Game';
      buttonColor = Colors.red;
      onPressed = () => _showDropConfirmationDialog();
    } else {
      buttonText = game.pricing.isFree
          ? 'Join Game'
          : 'Join Game - ${game.pricing.displayPrice}';
      buttonColor = AppColors.success;
      onPressed = () => _handleJoinGame();
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 8,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: SafeArea(
        child: ElevatedButton(
          onPressed: onPressed,
          style: ElevatedButton.styleFrom(
            backgroundColor: buttonColor,
            padding: const EdgeInsets.symmetric(vertical: 16),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
          ),
          child: Text(
            buttonText,
            style: const TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
        ),
      ),
    );
  }

  Future<void> _handleJoinGame() async {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);

    if (!authProvider.isAuthenticated) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Please login to join a game'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    // Get auth token
    final token = await authProvider.getToken();
    if (token == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Authentication error. Please login again.'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    try {
      await _gameService.joinGame(widget.gameId, token);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Successfully joined game!'),
            backgroundColor: Colors.green,
          ),
        );
        // Reload game details to reflect the change
        _loadGameDetails();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to join game: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  Future<void> _showDropConfirmationDialog() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Drop From Game'),
        content: const Text(
          'Are you sure you want to drop from this game? You will lose your position entirely and cannot rejoin if the game is full.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(true),
            style: TextButton.styleFrom(
              foregroundColor: Colors.red,
            ),
            child: const Text('Drop'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _handleDropGame();
    }
  }

  Future<void> _handleDropGame() async {
    final authProvider = Provider.of<AuthProvider>(context, listen: false);

    if (!authProvider.isAuthenticated) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Please login to drop from a game'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    // Get auth token
    final token = await authProvider.getToken();
    if (token == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Authentication error. Please login again.'),
            backgroundColor: Colors.red,
          ),
        );
      }
      return;
    }

    try {
      await _gameService.leaveGame(widget.gameId, token);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Successfully dropped from game'),
            backgroundColor: Colors.green,
          ),
        );
        // Reload game details to reflect the change
        _loadGameDetails();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Failed to drop from game: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }
}
