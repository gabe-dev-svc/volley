import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import 'package:geolocator/geolocator.dart';
import 'package:go_router/go_router.dart';
import '../models/game.dart';
import '../models/place.dart';
import '../providers/auth_provider.dart';
import '../services/game_service.dart';
import '../services/location_service.dart';
import '../theme/app_theme.dart';

class CreateGameScreen extends StatefulWidget {
  const CreateGameScreen({super.key});

  @override
  State<CreateGameScreen> createState() => _CreateGameScreenState();
}

class _CreateGameScreenState extends State<CreateGameScreen> {
  final _formKey = GlobalKey<FormState>();
  final _gameService = GameService();
  final _locationService = LocationService();

  GameCategory _selectedCategory = GameCategory.flagFootball;
  final _titleController = TextEditingController();
  final _descriptionController = TextEditingController();
  final _locationNameController = TextEditingController();
  final _locationAddressController = TextEditingController();
  final _locationNotesController = TextEditingController();
  DateTime _startTime = DateTime.now().add(const Duration(days: 1));
  int _durationMinutes = 60;
  final _maxParticipantsController = TextEditingController(text: '10');
  PricingType _pricingType = PricingType.free;
  final _amountController = TextEditingController(text: '50.00');
  String _currency = 'USD';
  SkillLevel _skillLevel = SkillLevel.all;
  DateTime? _signupDeadline;
  DateTime? _dropDeadline;
  final _notesController = TextEditingController();

  bool _isLoading = false;
  double? _latitude;
  double? _longitude;

  // Autocomplete state
  Timer? _debounceTimer;
  List<PlacePrediction> _placeSuggestions = [];
  bool _isLoadingSuggestions = false;

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _titleController.dispose();
    _descriptionController.dispose();
    _locationNameController.dispose();
    _locationAddressController.dispose();
    _locationNotesController.dispose();
    _maxParticipantsController.dispose();
    _amountController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  void _onLocationSearchChanged(String value) {
    // Cancel existing timer
    _debounceTimer?.cancel();

    if (value.isEmpty) {
      setState(() => _placeSuggestions = []);
      return;
    }

    // Require minimum 3 characters
    if (value.length < 3) {
      return;
    }

    // Wait 500ms after user stops typing
    _debounceTimer = Timer(const Duration(milliseconds: 500), () {
      _fetchPlaceSuggestions(value);
    });
  }

  Future<void> _fetchPlaceSuggestions(String input) async {
    setState(() => _isLoadingSuggestions = true);

    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      final authToken = await authProvider.getToken();

      if (authToken == null) {
        if (mounted) {
          setState(() => _isLoadingSuggestions = false);
        }
        return;
      }

      // Get user's current location for bias (optional)
      Position? userLocation;
      try {
        userLocation = await Geolocator.getCurrentPosition(
          locationSettings: const LocationSettings(
            accuracy: LocationAccuracy.medium,
            timeLimit: Duration(seconds: 5),
          ),
        );
      } catch (e) {
        // Location permission denied or unavailable - continue without bias
      }

      final suggestions = await _locationService.autocomplete(
        input,
        authToken,
        latitude: userLocation?.latitude,
        longitude: userLocation?.longitude,
      );

      if (mounted) {
        setState(() {
          _placeSuggestions = suggestions;
          _isLoadingSuggestions = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoadingSuggestions = false);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error fetching suggestions: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _selectPlace(PlacePrediction prediction) async {
    try {
      final authProvider = Provider.of<AuthProvider>(context, listen: false);
      final authToken = await authProvider.getToken();

      if (authToken == null) return;

      // Fetch full details including coordinates
      final details = await _locationService.getPlaceDetails(
        prediction.placeId,
        authToken,
      );

      // Update form fields
      setState(() {
        _locationAddressController.text = details.formattedAddress;
        _locationNameController.text = details.name;
        _latitude = details.latitude;
        _longitude = details.longitude;
        _placeSuggestions = []; // Clear suggestions
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error loading place details: ${e.toString()}')),
        );
      }
    }
  }

  Future<void> _selectDateTime(BuildContext context, bool isStartTime) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: isStartTime ? _startTime : (_signupDeadline ?? _startTime),
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 365)),
    );

    if (pickedDate != null && context.mounted) {
      final TimeOfDay? pickedTime = await showTimePicker(
        context: context,
        initialTime: TimeOfDay.fromDateTime(isStartTime ? _startTime : (_signupDeadline ?? _startTime)),
      );

      if (pickedTime != null) {
        setState(() {
          final newDateTime = DateTime(
            pickedDate.year,
            pickedDate.month,
            pickedDate.day,
            pickedTime.hour,
            pickedTime.minute,
          );
          if (isStartTime) {
            _startTime = newDateTime;
          } else {
            _signupDeadline = newDateTime;
          }
        });
      }
    }
  }

  Future<void> _selectDropDeadline(BuildContext context) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: _dropDeadline ?? _startTime,
      firstDate: DateTime.now(),
      lastDate: _startTime,
    );

    if (pickedDate != null && context.mounted) {
      final TimeOfDay? pickedTime = await showTimePicker(
        context: context,
        initialTime: TimeOfDay.fromDateTime(_dropDeadline ?? _startTime),
      );

      if (pickedTime != null) {
        setState(() {
          _dropDeadline = DateTime(
            pickedDate.year,
            pickedDate.month,
            pickedDate.day,
            pickedTime.hour,
            pickedTime.minute,
          );
        });
      }
    }
  }

  Future<void> _createGame() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    final authProvider = Provider.of<AuthProvider>(context, listen: false);
    final authToken = await authProvider.getToken();

    if (authToken == null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Please log in to create a game')),
        );
      }
      return;
    }

    setState(() => _isLoading = true);

    try {
      final location = Location(
        name: _locationNameController.text.trim().isEmpty
            ? 'Location'
            : _locationNameController.text.trim(),
        address: _locationAddressController.text.trim().isEmpty
            ? null
            : _locationAddressController.text.trim(),
        latitude: _latitude,
        longitude: _longitude,
        notes: _locationNotesController.text.trim().isEmpty
            ? null
            : _locationNotesController.text.trim(),
      );

      final pricing = Pricing(
        type: _pricingType,
        amountCents: _pricingType == PricingType.free
            ? 0
            : (double.parse(_amountController.text) * 100).toInt(),
        currency: _currency,
      );

      final request = CreateGameRequest(
        category: _selectedCategory,
        title: _titleController.text.trim().isEmpty ? null : _titleController.text.trim(),
        description: _descriptionController.text.trim().isEmpty ? null : _descriptionController.text.trim(),
        location: location,
        startTime: _startTime,
        durationMinutes: _durationMinutes,
        maxParticipants: int.parse(_maxParticipantsController.text),
        pricing: pricing,
        signupDeadline: _signupDeadline,
        skillLevel: _skillLevel,
        notes: _notesController.text.trim().isEmpty ? null : _notesController.text.trim(),
      );

      final createdGame = await _gameService.createGame(request, authToken);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Game created successfully!')),
        );
        // Navigate to the game details screen
        context.go('/game/${createdGame.id}');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.toString()}')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Create Game'),
        actions: [
          TextButton(
            onPressed: () => context.go('/games'),
            child: const Text(
              'Cancel',
              style: TextStyle(color: AppColors.primary, fontSize: 16),
            ),
          ),
        ],
      ),
      body: Form(
        key: _formKey,
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              _buildSectionTitle('Game Type'),
              const SizedBox(height: 12),
              _buildGameTypeSelector(),
              const SizedBox(height: 24),

              _buildSectionTitle('Game Details'),
              const SizedBox(height: 12),
              _buildTextField(
                controller: _titleController,
                label: 'Title',
                hint: 'e.g. Saturday Morning Hoops',
              ),
              const SizedBox(height: 16),
              _buildTextField(
                controller: _descriptionController,
                label: 'Description',
                hint: 'Casual 5v5 game, all are welcome!',
                maxLines: 3,
              ),
              const SizedBox(height: 24),

              _buildSectionTitle('Location'),
              const SizedBox(height: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Address', style: TextStyle(fontWeight: FontWeight.w600)),
                  const SizedBox(height: 8),
                  TextFormField(
                    controller: _locationAddressController,
                    onChanged: _onLocationSearchChanged,
                    decoration: InputDecoration(
                      hintText: 'Search for a park, court, or address',
                      prefixIcon: const Icon(Icons.search, color: AppColors.grey),
                      suffixIcon: _isLoadingSuggestions
                          ? const Padding(
                              padding: EdgeInsets.all(12.0),
                              child: SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              ),
                            )
                          : null,
                      filled: true,
                      fillColor: AppColors.primaryLight,
                      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: AppColors.greyLight),
                      ),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: AppColors.greyLight),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: AppColors.primary, width: 2),
                      ),
                    ),
                  ),
                  if (_placeSuggestions.isNotEmpty)
                    Container(
                      margin: const EdgeInsets.only(top: 8),
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(12),
                        boxShadow: [
                          BoxShadow(
                            color: Colors.black.withOpacity(0.1),
                            blurRadius: 8,
                            offset: const Offset(0, 2),
                          ),
                        ],
                      ),
                      child: ListView.builder(
                        shrinkWrap: true,
                        physics: const NeverScrollableScrollPhysics(),
                        itemCount: _placeSuggestions.length,
                        itemBuilder: (context, index) {
                          final suggestion = _placeSuggestions[index];
                          return ListTile(
                            leading: const Icon(Icons.location_on, color: AppColors.primary),
                            title: Text(
                              suggestion.mainText,
                              style: const TextStyle(fontWeight: FontWeight.w600),
                            ),
                            subtitle: Text(suggestion.secondaryText),
                            onTap: () => _selectPlace(suggestion),
                          );
                        },
                      ),
                    ),
                ],
              ),
              const SizedBox(height: 16),
              _buildTextField(
                controller: _locationNotesController,
                label: 'Notes',
                hint: 'e.g. Court 3, near the west entrance',
              ),
              const SizedBox(height: 24),

              _buildSectionTitle('Date & Time'),
              const SizedBox(height: 12),
              Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Start Time', style: TextStyle(fontWeight: FontWeight.w600)),
                        const SizedBox(height: 8),
                        InkWell(
                          onTap: () => _selectDateTime(context, true),
                          child: Container(
                            padding: const EdgeInsets.all(16),
                            decoration: BoxDecoration(
                              color: AppColors.primaryLight,
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: AppColors.greyLight),
                            ),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Text(
                                    DateFormat('MM/dd/yyyy, hh:mm a').format(_startTime),
                                    style: const TextStyle(fontSize: 14),
                                  ),
                                ),
                                const Icon(Icons.calendar_today, size: 20),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        const Text('Duration', style: TextStyle(fontWeight: FontWeight.w600)),
                        const SizedBox(height: 8),
                        DropdownButtonFormField<int>(
                          initialValue: _durationMinutes,
                          decoration: InputDecoration(
                            filled: true,
                            fillColor: AppColors.primaryLight,
                            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(12),
                              borderSide: const BorderSide(color: AppColors.greyLight),
                            ),
                          ),
                          items: [30, 60, 90, 120, 180].map((minutes) {
                            return DropdownMenuItem(
                              value: minutes,
                              child: Text('$minutes min'),
                            );
                          }).toList(),
                          onChanged: (value) {
                            setState(() => _durationMinutes = value!);
                          },
                        ),
                      ],
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 24),

              _buildSectionTitle('Participants & Pricing'),
              const SizedBox(height: 12),
              _buildTextField(
                controller: _maxParticipantsController,
                label: 'Max Participants',
                keyboardType: TextInputType.number,
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'Required';
                  }
                  final num = int.tryParse(value);
                  if (num == null || num < 2) {
                    return 'Must be at least 2';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 16),
              const Text('Price', style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 8),
              Row(
                children: [
                  _buildPriceTypeChip('Free', PricingType.free),
                  const SizedBox(width: 8),
                  _buildPriceTypeChip('Total', PricingType.total),
                  const SizedBox(width: 8),
                  _buildPriceTypeChip('Per Person', PricingType.perPerson),
                ],
              ),
              if (_pricingType != PricingType.free) ...[
                const SizedBox(height: 16),
                Row(
                  children: [
                    Expanded(
                      flex: 2,
                      child: _buildTextField(
                        controller: _amountController,
                        label: 'Amount',
                        keyboardType: TextInputType.number,
                        prefixText: '\$ ',
                      ),
                    ),
                    const SizedBox(width: 12),
                    Expanded(
                      child: DropdownButtonFormField<String>(
                        initialValue: _currency,
                        decoration: InputDecoration(
                          labelText: 'Currency',
                          filled: true,
                          fillColor: AppColors.primaryLight,
                          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                            borderSide: const BorderSide(color: AppColors.greyLight),
                          ),
                        ),
                        items: ['USD', 'EUR', 'GBP'].map((currency) {
                          return DropdownMenuItem(
                            value: currency,
                            child: Text(currency),
                          );
                        }).toList(),
                        onChanged: (value) {
                          setState(() => _currency = value!);
                        },
                      ),
                    ),
                  ],
                ),
              ],
              const SizedBox(height: 24),

              _buildSectionTitle('Additional Information'),
              const SizedBox(height: 12),
              const Text('Skill Level', style: TextStyle(fontWeight: FontWeight.w600)),
              const SizedBox(height: 8),
              Wrap(
                spacing: 8,
                children: [
                  _buildSkillLevelChip('All', SkillLevel.all),
                  _buildSkillLevelChip('Beginner', SkillLevel.beginner),
                  _buildSkillLevelChip('Intermediate', SkillLevel.intermediate),
                  _buildSkillLevelChip('Advanced', SkillLevel.advanced),
                ],
              ),
              const SizedBox(height: 16),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Text('Signup Deadline', style: TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(width: 4),
                      const Text('Optional', style: TextStyle(color: AppColors.textSecondary, fontSize: 12)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  InkWell(
                    onTap: () => _selectDateTime(context, false),
                    child: Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: AppColors.primaryLight,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: AppColors.greyLight),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              _signupDeadline != null
                                  ? DateFormat('MM/dd/yyyy, hh:mm a').format(_signupDeadline!)
                                  : 'Select signup deadline',
                              style: TextStyle(
                                fontSize: 14,
                                color: _signupDeadline != null ? AppColors.textPrimary : AppColors.textHint,
                              ),
                            ),
                          ),
                          const Icon(Icons.calendar_today, size: 20),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Text('Drop Deadline', style: TextStyle(fontWeight: FontWeight.w600)),
                      const SizedBox(width: 4),
                      const Text('Optional', style: TextStyle(color: AppColors.textSecondary, fontSize: 12)),
                    ],
                  ),
                  const SizedBox(height: 8),
                  InkWell(
                    onTap: () => _selectDropDeadline(context),
                    child: Container(
                      padding: const EdgeInsets.all(16),
                      decoration: BoxDecoration(
                        color: AppColors.primaryLight,
                        borderRadius: BorderRadius.circular(12),
                        border: Border.all(color: AppColors.greyLight),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              _dropDeadline != null
                                  ? DateFormat('MM/dd/yyyy, hh:mm a').format(_dropDeadline!)
                                  : 'mm/dd/yyyy, --:--',
                              style: TextStyle(
                                fontSize: 14,
                                color: _dropDeadline != null ? AppColors.textPrimary : AppColors.textHint,
                              ),
                            ),
                          ),
                          const Icon(Icons.calendar_today, size: 20),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              _buildTextField(
                controller: _notesController,
                label: 'Game Notes',
                hint: 'e.g. Bring a dark and a light shirt.',
                maxLines: 3,
              ),
              const SizedBox(height: 32),

              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _isLoading ? null : _createGame,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppColors.primary,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                  child: _isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        )
                      : const Text(
                          'Create Game',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                ),
              ),
              const SizedBox(height: 32),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: const TextStyle(
        fontSize: 18,
        fontWeight: FontWeight.bold,
      ),
    );
  }

  Widget _buildGameTypeSelector() {
    final categories = [
      GameCategory.flagFootball,
      GameCategory.soccer,
      GameCategory.basketball,
      GameCategory.pickleball,
      GameCategory.volleyball,
    ];

    return Wrap(
      spacing: 8,
      runSpacing: 8,
      children: categories.map((GameCategory category) {
        final isSelected = _selectedCategory == category;
        return InkWell(
          onTap: () => setState(() => _selectedCategory = category),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
            decoration: BoxDecoration(
              color: isSelected ? AppColors.primary : AppColors.primaryLight,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(
                color: isSelected ? AppColors.primary : AppColors.greyLight,
              ),
            ),
            child: Text(
              category.displayName,
              style: TextStyle(
                color: isSelected ? Colors.white : AppColors.textPrimary,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
        );
      }).toList(),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    String? hint,
    int maxLines = 1,
    TextInputType? keyboardType,
    IconData? prefixIcon,
    String? prefixText,
    String? Function(String?)? validator,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontWeight: FontWeight.w600)),
        const SizedBox(height: 8),
        TextFormField(
          controller: controller,
          maxLines: maxLines,
          keyboardType: keyboardType,
          validator: validator,
          decoration: InputDecoration(
            hintText: hint,
            prefixIcon: prefixIcon != null ? Icon(prefixIcon, color: AppColors.grey) : null,
            prefixText: prefixText,
            filled: true,
            fillColor: AppColors.primaryLight,
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 16),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.greyLight),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.greyLight),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: const BorderSide(color: AppColors.primary, width: 2),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildPriceTypeChip(String label, PricingType type) {
    final isSelected = _pricingType == type;
    return InkWell(
      onTap: () => setState(() => _pricingType = type),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primary : AppColors.primaryLight,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(
            color: isSelected ? AppColors.primary : AppColors.greyLight,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? Colors.white : AppColors.textPrimary,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }

  Widget _buildSkillLevelChip(String label, SkillLevel level) {
    final isSelected = _skillLevel == level;
    return InkWell(
      onTap: () => setState(() => _skillLevel = level),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
        decoration: BoxDecoration(
          color: isSelected ? AppColors.primary : AppColors.primaryLight,
          borderRadius: BorderRadius.circular(24),
          border: Border.all(
            color: isSelected ? AppColors.primary : AppColors.greyLight,
          ),
        ),
        child: Text(
          label,
          style: TextStyle(
            color: isSelected ? Colors.white : AppColors.textPrimary,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}
