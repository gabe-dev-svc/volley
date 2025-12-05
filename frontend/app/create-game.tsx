import React, { useState, useCallback, useRef, useEffect } from 'react';
import {
    View,
    Text,
    StyleSheet,
    ScrollView,
    TextInput,
    TouchableOpacity,
    ActivityIndicator,
    Alert,
    FlatList,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { apiClient, GameCategory, SkillLevel, PricingType, ApiRequestError } from '../lib/api';
import { useLocation } from '../hooks/useLocation';

const GOOGLE_PLACES_API_KEY = process.env.EXPO_PUBLIC_GOOGLE_PLACES_API_KEY || '';

interface PlacePrediction {
    place_id: string;
    description: string;
    structured_formatting: {
        main_text: string;
        secondary_text: string;
    };
}

interface PlaceDetails {
    name: string;
    formatted_address: string;
    geometry: {
        location: {
            lat: number;
            lng: number;
        };
    };
}

const SPORT_CATEGORIES: { key: GameCategory; label: string; emoji: string }[] = [
    { key: 'soccer', label: 'Soccer', emoji: '⚽' },
    { key: 'basketball', label: 'Basketball', emoji: '🏀' },
    { key: 'volleyball', label: 'Volleyball', emoji: '🏐' },
    { key: 'pickleball', label: 'Pickleball', emoji: '🏓' },
    { key: 'flag_football', label: 'Flag Football', emoji: '🏈' },
    { key: 'ultimate_frisbee', label: 'Ultimate Frisbee', emoji: '🥏' },
    { key: 'tennis', label: 'Tennis', emoji: '🎾' },
    { key: 'other', label: 'Other', emoji: '🏃' },
];

const SKILL_LEVELS: { key: SkillLevel; label: string }[] = [
    { key: 'all', label: 'All Levels' },
    { key: 'beginner', label: 'Beginner' },
    { key: 'intermediate', label: 'Intermediate' },
    { key: 'advanced', label: 'Advanced' },
];

const PRICING_TYPES: { key: PricingType; label: string; description: string }[] = [
    { key: 'free', label: 'Free', description: 'No cost to join' },
    { key: 'per_person', label: 'Per Person', description: 'Fixed price per player' },
    { key: 'total', label: 'Split Total', description: 'Total cost split among players' },
];

export default function CreateGameScreen() {
    const router = useRouter();
    const { location } = useLocation();
    const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // Form state
    const [category, setCategory] = useState<GameCategory>('basketball');
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [locationSearchQuery, setLocationSearchQuery] = useState('');
    const [locationName, setLocationName] = useState('');
    const [locationAddress, setLocationAddress] = useState('');
    const [locationLatitude, setLocationLatitude] = useState<number | undefined>(undefined);
    const [locationLongitude, setLocationLongitude] = useState<number | undefined>(undefined);
    const [locationNotes, setLocationNotes] = useState('');
    const [startDate, setStartDate] = useState('');
    const [startTimeStr, setStartTimeStr] = useState('');
    const [durationMinutes, setDurationMinutes] = useState('90');
    const [maxParticipants, setMaxParticipants] = useState('10');
    const [skillLevel, setSkillLevel] = useState<SkillLevel>('all');
    const [pricingType, setPricingType] = useState<PricingType>('free');
    const [priceAmount, setPriceAmount] = useState('');
    const [notes, setNotes] = useState('');

    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    // Places autocomplete state
    const [predictions, setPredictions] = useState<PlacePrediction[]>([]);
    const [isSearching, setIsSearching] = useState(false);
    const [showPredictions, setShowPredictions] = useState(false);

    // Search for places using the new Places API
    const searchPlaces = useCallback(async (query: string) => {
        if (!query || query.length < 3) {
            setPredictions([]);
            return;
        }

        setIsSearching(true);
        try {
            // Use the new Places API (Text Search)
            const url = `https://places.googleapis.com/v1/places:searchText`;

            const response = await fetch(url, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-Goog-Api-Key': GOOGLE_PLACES_API_KEY,
                    'X-Goog-FieldMask': 'places.id,places.displayName,places.formattedAddress,places.location',
                },
                body: JSON.stringify({
                    textQuery: query,
                    maxResultCount: 5,
                }),
            });

            const data = await response.json();

            if (data.places && data.places.length > 0) {
                // Convert new API results to prediction format
                const predictions: PlacePrediction[] = data.places.map((place: any) => {
                    const addressParts = place.formattedAddress?.split(',') || [];
                    return {
                        place_id: place.id,
                        description: place.formattedAddress || place.displayName?.text || '',
                        structured_formatting: {
                            main_text: place.displayName?.text || addressParts[0] || '',
                            secondary_text: addressParts.slice(1).join(',').trim() || '',
                        },
                    };
                });
                setPredictions(predictions);
            } else {
                setPredictions([]);
            }
        } catch (err) {
            console.error('Error searching places:', err);
            console.error('Error details:', err);
            setPredictions([]);
        } finally {
            setIsSearching(false);
        }
    }, []);

    // Get place details using the new Places API
    const getPlaceDetails = async (placeId: string) => {
        try {
            const url = `https://places.googleapis.com/v1/places/${placeId}`;

            const response = await fetch(url, {
                method: 'GET',
                headers: {
                    'X-Goog-Api-Key': GOOGLE_PLACES_API_KEY,
                    'X-Goog-FieldMask': 'displayName,formattedAddress,location',
                },
            });

            const data = await response.json();

            if (data) {
                setLocationName(data.displayName?.text || data.formattedAddress?.split(',')[0] || '');
                setLocationAddress(data.formattedAddress || '');
                setLocationLatitude(data.location?.latitude);
                setLocationLongitude(data.location?.longitude);
                setLocationSearchQuery('');
                setPredictions([]);
                setShowPredictions(false);
            }
        } catch (err) {
            console.error('Error getting place details:', err);
        }
    };

    const handleLocationSearchChange = (text: string) => {
        setLocationSearchQuery(text);
        setShowPredictions(true);

        // Clear previous timeout
        if (searchTimeoutRef.current) {
            clearTimeout(searchTimeoutRef.current);
        }

        // Simple debounce
        searchTimeoutRef.current = setTimeout(() => {
            searchPlaces(text);
        }, 300);
    };

    // Cleanup timeout on unmount
    useEffect(() => {
        return () => {
            if (searchTimeoutRef.current) {
                clearTimeout(searchTimeoutRef.current);
            }
        };
    }, []);

    const clearLocation = () => {
        setLocationName('');
        setLocationAddress('');
        setLocationLatitude(undefined);
        setLocationLongitude(undefined);
        setLocationSearchQuery('');
        setPredictions([]);
        setShowPredictions(false);
    };

    const handleSubmit = async () => {
        setError(null);

        // Validation
        if (!locationName.trim()) {
            setError('Please enter a location name');
            return;
        }

        if (!locationAddress.trim()) {
            setError('Please enter a location address');
            return;
        }

        if (!startDate.trim()) {
            setError('Please enter a start date');
            return;
        }

        if (!startTimeStr.trim()) {
            setError('Please enter a start time');
            return;
        }

        // Parse date and time
        let startDateTime: Date;
        try {
            const dateTime = new Date(`${startDate}T${startTimeStr}`);
            if (isNaN(dateTime.getTime())) {
                setError('Invalid date or time format');
                return;
            }
            if (dateTime < new Date()) {
                setError('Start time must be in the future');
                return;
            }
            startDateTime = dateTime;
        } catch {
            setError('Invalid date or time format');
            return;
        }

        const duration = parseInt(durationMinutes);
        if (isNaN(duration) || duration < 15) {
            setError('Duration must be at least 15 minutes');
            return;
        }

        const maxP = parseInt(maxParticipants);
        if (isNaN(maxP) || maxP < 2) {
            setError('Must allow at least 2 participants');
            return;
        }

        // Parse price
        let amountCents = 0;
        if (pricingType !== 'free') {
            const amount = parseFloat(priceAmount);
            if (isNaN(amount) || amount < 0) {
                setError('Please enter a valid price');
                return;
            }
            amountCents = Math.round(amount * 100);
        }

        setIsLoading(true);

        try {
            const game = await apiClient.createGame({
                category,
                title: title.trim() || undefined,
                description: description.trim() || undefined,
                location: {
                    name: locationName.trim(),
                    address: locationAddress.trim(),
                    latitude: locationLatitude ?? location?.latitude,
                    longitude: locationLongitude ?? location?.longitude,
                    notes: locationNotes.trim() || undefined,
                },
                startTime: startDateTime.toISOString(),
                durationMinutes: duration,
                maxParticipants: maxP,
                pricing: {
                    type: pricingType,
                    amountCents,
                    currency: 'USD',
                },
                skillLevel,
                notes: notes.trim() || undefined,
            });

            // Navigate to game details page
            router.push(`/game/${game.id}`);
        } catch (err) {
            if (err instanceof ApiRequestError) {
                setError(err.message);
            } else {
                setError(err instanceof Error ? err.message : 'Failed to create game');
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <SafeAreaView style={styles.container}>
            <View style={styles.header}>
                <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
                    <Ionicons name="arrow-back" size={24} color="#1F2937" />
                </TouchableOpacity>
                <Text style={styles.headerTitle}>Create Game</Text>
                <View style={{ width: 24 }} />
            </View>

            <ScrollView style={styles.content} contentContainerStyle={styles.scrollContent}>
                {/* Sport Category */}
                <View style={styles.section}>
                    <Text style={styles.label}>Sport *</Text>
                    <ScrollView
                        horizontal
                        showsHorizontalScrollIndicator={false}
                        style={styles.categoryScroll}
                    >
                        {SPORT_CATEGORIES.map(cat => (
                            <TouchableOpacity
                                key={cat.key}
                                style={[
                                    styles.categoryButton,
                                    category === cat.key && styles.categoryButtonActive
                                ]}
                                onPress={() => setCategory(cat.key)}
                            >
                                <Text style={styles.categoryEmoji}>{cat.emoji}</Text>
                                <Text style={[
                                    styles.categoryText,
                                    category === cat.key && styles.categoryTextActive
                                ]}>
                                    {cat.label}
                                </Text>
                            </TouchableOpacity>
                        ))}
                    </ScrollView>
                </View>

                {/* Title */}
                <View style={styles.section}>
                    <Text style={styles.label}>Title (Optional)</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="e.g., Sunday Morning Pickup"
                        value={title}
                        onChangeText={setTitle}
                        placeholderTextColor="#9CA3AF"
                    />
                </View>

                {/* Description */}
                <View style={styles.section}>
                    <Text style={styles.label}>Description (Optional)</Text>
                    <TextInput
                        style={[styles.input, styles.textArea]}
                        placeholder="Tell players what to expect..."
                        value={description}
                        onChangeText={setDescription}
                        multiline
                        numberOfLines={3}
                        placeholderTextColor="#9CA3AF"
                    />
                </View>

                {/* Location */}
                <View style={styles.section}>
                    <Text style={styles.sectionTitle}>Location</Text>

                    {!locationAddress ? (
                        <>
                            <Text style={styles.label}>Search for a place *</Text>
                            <View style={styles.autocompleteContainer}>
                                <TextInput
                                    style={styles.input}
                                    placeholder="Search for a location..."
                                    value={locationSearchQuery}
                                    onChangeText={handleLocationSearchChange}
                                    placeholderTextColor="#9CA3AF"
                                    onFocus={() => setShowPredictions(true)}
                                />
                                {isSearching && (
                                    <View style={styles.searchingIndicator}>
                                        <ActivityIndicator size="small" color="#10B981" />
                                    </View>
                                )}
                            </View>

                            {showPredictions && predictions.length > 0 && (
                                <View style={styles.predictionsContainer}>
                                    <FlatList
                                        data={predictions}
                                        keyExtractor={(item) => item.place_id}
                                        renderItem={({ item }) => (
                                            <TouchableOpacity
                                                style={styles.predictionItem}
                                                onPress={() => getPlaceDetails(item.place_id)}
                                            >
                                                <Ionicons name="location-outline" size={20} color="#6B7280" />
                                                <View style={styles.predictionText}>
                                                    <Text style={styles.predictionMainText}>
                                                        {item.structured_formatting.main_text}
                                                    </Text>
                                                    <Text style={styles.predictionSecondaryText}>
                                                        {item.structured_formatting.secondary_text}
                                                    </Text>
                                                </View>
                                            </TouchableOpacity>
                                        )}
                                        scrollEnabled={false}
                                    />
                                </View>
                            )}
                        </>
                    ) : (
                        <View style={styles.selectedLocationContainer}>
                            <View style={styles.selectedLocationHeader}>
                                <Ionicons name="location" size={20} color="#10B981" />
                                <Text style={styles.selectedLocationName}>{locationName}</Text>
                            </View>
                            <Text style={styles.selectedLocationAddress}>{locationAddress}</Text>
                            <TouchableOpacity
                                onPress={clearLocation}
                                style={styles.clearLocationButton}
                            >
                                <Text style={styles.clearLocationText}>Change Location</Text>
                            </TouchableOpacity>
                        </View>
                    )}

                    <Text style={styles.label}>Location Notes (Optional)</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="e.g., Park in the back, meet at Field 1"
                        value={locationNotes}
                        onChangeText={setLocationNotes}
                        placeholderTextColor="#9CA3AF"
                    />
                </View>

                {/* Date & Time */}
                <View style={styles.section}>
                    <Text style={styles.sectionTitle}>When</Text>

                    <View style={styles.row}>
                        <View style={styles.halfWidth}>
                            <Text style={styles.label}>Date * (YYYY-MM-DD)</Text>
                            <TextInput
                                style={styles.input}
                                placeholder="2025-12-01"
                                value={startDate}
                                onChangeText={setStartDate}
                                placeholderTextColor="#9CA3AF"
                            />
                        </View>

                        <View style={styles.halfWidth}>
                            <Text style={styles.label}>Time * (HH:MM)</Text>
                            <TextInput
                                style={styles.input}
                                placeholder="14:00"
                                value={startTimeStr}
                                onChangeText={setStartTimeStr}
                                placeholderTextColor="#9CA3AF"
                            />
                        </View>
                    </View>

                    <Text style={styles.label}>Duration (minutes) *</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="90"
                        value={durationMinutes}
                        onChangeText={setDurationMinutes}
                        keyboardType="number-pad"
                        placeholderTextColor="#9CA3AF"
                    />
                </View>

                {/* Game Details */}
                <View style={styles.section}>
                    <Text style={styles.sectionTitle}>Game Details</Text>

                    <Text style={styles.label}>Max Players *</Text>
                    <TextInput
                        style={styles.input}
                        placeholder="10"
                        value={maxParticipants}
                        onChangeText={setMaxParticipants}
                        keyboardType="number-pad"
                        placeholderTextColor="#9CA3AF"
                    />

                    <Text style={styles.label}>Skill Level</Text>
                    <View style={styles.buttonGroup}>
                        {SKILL_LEVELS.map(level => (
                            <TouchableOpacity
                                key={level.key}
                                style={[
                                    styles.buttonGroupItem,
                                    skillLevel === level.key && styles.buttonGroupItemActive
                                ]}
                                onPress={() => setSkillLevel(level.key)}
                            >
                                <Text style={[
                                    styles.buttonGroupText,
                                    skillLevel === level.key && styles.buttonGroupTextActive
                                ]}>
                                    {level.label}
                                </Text>
                            </TouchableOpacity>
                        ))}
                    </View>
                </View>

                {/* Pricing */}
                <View style={styles.section}>
                    <Text style={styles.sectionTitle}>Pricing</Text>

                    {PRICING_TYPES.map(type => (
                        <TouchableOpacity
                            key={type.key}
                            style={styles.radioOption}
                            onPress={() => setPricingType(type.key)}
                        >
                            <View style={styles.radioLeft}>
                                <View style={[
                                    styles.radio,
                                    pricingType === type.key && styles.radioSelected
                                ]}>
                                    {pricingType === type.key && (
                                        <View style={styles.radioDot} />
                                    )}
                                </View>
                                <View>
                                    <Text style={styles.radioLabel}>{type.label}</Text>
                                    <Text style={styles.radioDescription}>{type.description}</Text>
                                </View>
                            </View>
                        </TouchableOpacity>
                    ))}

                    {pricingType !== 'free' && (
                        <>
                            <Text style={styles.label}>
                                {pricingType === 'per_person' ? 'Price Per Person' : 'Total Price'}
                            </Text>
                            <View style={styles.priceInput}>
                                <Text style={styles.priceCurrency}>$</Text>
                                <TextInput
                                    style={styles.priceTextInput}
                                    placeholder="0.00"
                                    value={priceAmount}
                                    onChangeText={setPriceAmount}
                                    keyboardType="decimal-pad"
                                    placeholderTextColor="#9CA3AF"
                                />
                            </View>
                        </>
                    )}
                </View>

                {/* Additional Notes */}
                <View style={styles.section}>
                    <Text style={styles.label}>Additional Notes (Optional)</Text>
                    <TextInput
                        style={[styles.input, styles.textArea]}
                        placeholder="Anything else players should know..."
                        value={notes}
                        onChangeText={setNotes}
                        multiline
                        numberOfLines={3}
                        placeholderTextColor="#9CA3AF"
                    />
                </View>

                {/* Error Message */}
                {error && (
                    <View style={styles.errorContainer}>
                        <Text style={styles.errorText}>{error}</Text>
                    </View>
                )}

                {/* Submit Button */}
                <TouchableOpacity
                    style={[styles.submitButton, isLoading && styles.submitButtonDisabled]}
                    onPress={handleSubmit}
                    disabled={isLoading}
                >
                    {isLoading ? (
                        <ActivityIndicator color="#FFFFFF" />
                    ) : (
                        <Text style={styles.submitButtonText}>Create Game</Text>
                    )}
                </TouchableOpacity>

                <View style={{ height: 40 }} />
            </ScrollView>
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#ECFDF5',
    },
    header: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: 16,
        backgroundColor: '#FFFFFF',
        borderBottomWidth: 1,
        borderBottomColor: '#E5E7EB',
    },
    backButton: {
        padding: 4,
    },
    headerTitle: {
        fontSize: 18,
        fontWeight: '600',
        color: '#1F2937',
    },
    content: {
        flex: 1,
    },
    scrollContent: {
        padding: 24,
    },
    section: {
        marginBottom: 24,
    },
    sectionTitle: {
        fontSize: 18,
        fontWeight: '600',
        color: '#1F2937',
        marginBottom: 16,
    },
    label: {
        fontSize: 14,
        fontWeight: '600',
        color: '#374151',
        marginBottom: 8,
        marginTop: 12,
    },
    input: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        padding: 16,
        fontSize: 16,
        color: '#1F2937',
    },
    textArea: {
        minHeight: 80,
        textAlignVertical: 'top',
    },
    categoryScroll: {
        flexGrow: 0,
    },
    categoryButton: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        paddingVertical: 12,
        paddingHorizontal: 16,
        marginRight: 12,
        alignItems: 'center',
        minWidth: 100,
    },
    categoryButtonActive: {
        borderColor: '#10B981',
        backgroundColor: '#ECFDF5',
    },
    categoryEmoji: {
        fontSize: 24,
        marginBottom: 4,
    },
    categoryText: {
        fontSize: 12,
        color: '#6B7280',
        fontWeight: '600',
    },
    categoryTextActive: {
        color: '#10B981',
    },
    row: {
        flexDirection: 'row',
        gap: 12,
    },
    halfWidth: {
        flex: 1,
    },
    buttonGroup: {
        flexDirection: 'row',
        flexWrap: 'wrap',
        gap: 8,
    },
    buttonGroupItem: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        paddingVertical: 10,
        paddingHorizontal: 16,
    },
    buttonGroupItemActive: {
        borderColor: '#10B981',
        backgroundColor: '#ECFDF5',
    },
    buttonGroupText: {
        fontSize: 14,
        color: '#6B7280',
        fontWeight: '600',
    },
    buttonGroupTextActive: {
        color: '#10B981',
    },
    radioOption: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'space-between',
        paddingVertical: 12,
        borderBottomWidth: 1,
        borderBottomColor: '#E5E7EB',
    },
    radioLeft: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 12,
    },
    radio: {
        width: 24,
        height: 24,
        borderRadius: 12,
        borderWidth: 2,
        borderColor: '#E5E7EB',
        justifyContent: 'center',
        alignItems: 'center',
    },
    radioSelected: {
        borderColor: '#10B981',
    },
    radioDot: {
        width: 12,
        height: 12,
        borderRadius: 6,
        backgroundColor: '#10B981',
    },
    radioLabel: {
        fontSize: 16,
        fontWeight: '600',
        color: '#1F2937',
    },
    radioDescription: {
        fontSize: 13,
        color: '#6B7280',
    },
    priceInput: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        paddingHorizontal: 16,
    },
    priceCurrency: {
        fontSize: 20,
        fontWeight: '600',
        color: '#1F2937',
        marginRight: 4,
    },
    priceTextInput: {
        flex: 1,
        fontSize: 16,
        color: '#1F2937',
        paddingVertical: 16,
    },
    errorContainer: {
        backgroundColor: '#FEE2E2',
        borderWidth: 1,
        borderColor: '#EF4444',
        borderRadius: 12,
        padding: 12,
        marginBottom: 16,
    },
    errorText: {
        color: '#DC2626',
        fontSize: 14,
        textAlign: 'center',
    },
    submitButton: {
        backgroundColor: '#10B981',
        paddingVertical: 16,
        borderRadius: 12,
        alignItems: 'center',
        shadowColor: '#10B981',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.3,
        shadowRadius: 8,
        elevation: 4,
    },
    submitButtonDisabled: {
        backgroundColor: '#9CA3AF',
        shadowOpacity: 0,
    },
    submitButtonText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
    autocompleteContainer: {
        position: 'relative',
        marginBottom: 12,
    },
    searchingIndicator: {
        position: 'absolute',
        right: 16,
        top: 16,
    },
    predictionsContainer: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        marginTop: 4,
        maxHeight: 300,
    },
    predictionItem: {
        flexDirection: 'row',
        alignItems: 'center',
        padding: 12,
        gap: 12,
        borderBottomWidth: 1,
        borderBottomColor: '#E5E7EB',
    },
    predictionText: {
        flex: 1,
    },
    predictionMainText: {
        fontSize: 16,
        fontWeight: '500',
        color: '#1F2937',
        marginBottom: 2,
    },
    predictionSecondaryText: {
        fontSize: 14,
        color: '#6B7280',
    },
    selectedLocationContainer: {
        backgroundColor: '#ECFDF5',
        borderWidth: 2,
        borderColor: '#10B981',
        borderRadius: 12,
        padding: 16,
        marginBottom: 12,
    },
    selectedLocationHeader: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 8,
        marginBottom: 4,
    },
    selectedLocationName: {
        fontSize: 16,
        fontWeight: '600',
        color: '#1F2937',
        flex: 1,
    },
    selectedLocationAddress: {
        fontSize: 14,
        color: '#6B7280',
        marginLeft: 28,
        marginBottom: 8,
    },
    clearLocationButton: {
        alignSelf: 'flex-start',
        paddingVertical: 4,
        paddingHorizontal: 12,
        backgroundColor: '#FFFFFF',
        borderRadius: 8,
        marginLeft: 28,
    },
    clearLocationText: {
        fontSize: 12,
        fontWeight: '600',
        color: '#10B981',
    },
});
