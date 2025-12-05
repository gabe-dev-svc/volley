import React, { useState, useEffect, useCallback } from 'react';
import {
    View,
    Text,
    StyleSheet,
    FlatList,
    TouchableOpacity,
    RefreshControl,
    ActivityIndicator,
    TextInput,
    Alert,
    Modal,
    ScrollView,
    Platform,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useLocation } from '../../hooks/useLocation';
import { apiClient, Game, GameCategory } from '../../lib/api';
import { GameCard } from '../../components/GameCard';
import { GameDetailsModal } from '../../components/GameDetailsModal';

const MILES_TO_METERS = 1609.34;
const DEFAULT_RADIUS_MILES = 10;

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

export default function GamesListScreen() {
    const router = useRouter();
    const { location, isLoading: locationLoading, error: locationError, requestPermission, setManualLocation } = useLocation();

    const [games, setGames] = useState<Game[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);
    const [selectedCategory, setSelectedCategory] = useState<GameCategory>('basketball');
    const [radiusMiles, setRadiusMiles] = useState(DEFAULT_RADIUS_MILES);
    const [searchQuery, setSearchQuery] = useState('');
    const [selectedGame, setSelectedGame] = useState<Game | null>(null);
    const [showZipModal, setShowZipModal] = useState(false);
    const [zipCode, setZipCode] = useState('');
    const [isGeocodingZip, setIsGeocodingZip] = useState(false);

    const calculateDistance = (lat1: number, lon1: number, lat2: number, lon2: number): number => {
        const R = 3959; // Earth's radius in miles
        const dLat = (lat2 - lat1) * Math.PI / 180;
        const dLon = (lon2 - lon1) * Math.PI / 180;
        const a =
            Math.sin(dLat / 2) * Math.sin(dLat / 2) +
            Math.cos(lat1 * Math.PI / 180) * Math.cos(lat2 * Math.PI / 180) *
            Math.sin(dLon / 2) * Math.sin(dLon / 2);
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        return R * c;
    };

    const loadGames = useCallback(async () => {
        if (!location) return;

        try {
            setIsLoading(true);
            const response = await apiClient.listGames({
                categories: [selectedCategory],
                latitude: location.latitude,
                longitude: location.longitude,
                radius: radiusMiles * MILES_TO_METERS,
                timeFilter: 'upcoming',
                limit: 50,
            });
            setGames(response.games);
        } catch (error) {
            Alert.alert('Error', 'Failed to load games');
            console.error('Load games error:', error);
        } finally {
            setIsLoading(false);
            setRefreshing(false);
        }
    }, [location, selectedCategory, radiusMiles]);

    useEffect(() => {
        if (location) {
            loadGames();
        }
    }, [location, selectedCategory, radiusMiles]);

    const onRefresh = () => {
        setRefreshing(true);
        loadGames();
    };

    const handleCategoryChange = (category: GameCategory) => {
        setSelectedCategory(category);
    };

    const filteredGames = games.filter(game =>
        game.location.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        game.title?.toLowerCase().includes(searchQuery.toLowerCase())
    );

    const handleZipCodeSubmit = async () => {
        console.log('=== handleZipCodeSubmit START ===');
        console.log('zipCode:', zipCode);

        if (!zipCode.trim()) {
            console.log('Empty zip code');
            Alert.alert('Error', 'Please enter a zip code');
            return;
        }

        // Basic US zip code validation
        const cleanZip = zipCode.trim();
        console.log('cleanZip:', cleanZip);
        if (!/^\d{5}(-\d{4})?$/.test(cleanZip)) {
            console.log('Invalid zip code format');
            Alert.alert('Error', 'Please enter a valid US zip code (e.g., 94102)');
            return;
        }

        try {
            console.log('Setting isGeocodingZip to true');
            setIsGeocodingZip(true);

            // Use the fallback API directly for reliability
            console.log('Fetching from API...');
            const url = `https://api.zippopotam.us/us/${cleanZip}`;
            console.log('URL:', url);

            const response = await fetch(url);
            console.log('Response status:', response.status);

            if (response.ok) {
                const data = await response.json();
                console.log('API data:', data);

                if (data && data.places && data.places.length > 0) {
                    const latitude = parseFloat(data.places[0].latitude);
                    const longitude = parseFloat(data.places[0].longitude);
                    console.log('Parsed coordinates:', { latitude, longitude });

                    console.log('Calling setManualLocation...');
                    setManualLocation({ latitude, longitude });

                    console.log('Closing modal...');
                    setShowZipModal(false);
                    setZipCode('');

                    console.log('Success!');
                } else {
                    console.log('No places found in response');
                    Alert.alert('Error', 'Could not find location for this zip code');
                }
            } else {
                console.log('Response not OK');
                Alert.alert('Error', 'Could not find location for this zip code');
            }
        } catch (error) {
            console.error('Geocoding error:', error);
            Alert.alert('Error', 'Failed to geocode zip code: ' + (error instanceof Error ? error.message : 'Unknown error'));
        } finally {
            console.log('Setting isGeocodingZip to false');
            setIsGeocodingZip(false);
            console.log('=== handleZipCodeSubmit END ===');
        }
    };

    const renderZipModal = () => (
        <Modal
            visible={showZipModal}
            transparent
            animationType="slide"
            onRequestClose={() => setShowZipModal(false)}
        >
            <View style={styles.modalOverlay}>
                <View style={styles.modalContent}>
                    <View style={styles.modalHeader}>
                        <Text style={styles.modalTitle}>Enter Your Location</Text>
                        <TouchableOpacity onPress={() => setShowZipModal(false)}>
                            <Ionicons name="close" size={24} color="#6B7280" />
                        </TouchableOpacity>
                    </View>
                    <Text style={styles.zipModalDescription}>
                        Enter your zip code to find games near you
                    </Text>
                    <TextInput
                        style={styles.zipInput}
                        placeholder="Enter zip code (e.g., 94102)"
                        value={zipCode}
                        onChangeText={setZipCode}
                        keyboardType="number-pad"
                        maxLength={10}
                        placeholderTextColor="#9CA3AF"
                        autoFocus
                        returnKeyType="done"
                        onSubmitEditing={handleZipCodeSubmit}
                    />
                    <TouchableOpacity
                        style={[styles.zipSubmitButton, isGeocodingZip && styles.zipSubmitButtonDisabled]}
                        onPress={() => {
                            console.log('Button pressed!');
                            handleZipCodeSubmit();
                        }}
                        disabled={isGeocodingZip}
                        activeOpacity={0.7}
                    >
                        {isGeocodingZip ? (
                            <ActivityIndicator color="#FFFFFF" />
                        ) : (
                            <Text style={styles.zipSubmitButtonText}>Find Games</Text>
                        )}
                    </TouchableOpacity>
                </View>
            </View>
        </Modal>
    );

    if (locationLoading) {
        return (
            <SafeAreaView style={styles.container}>
                <View style={styles.contentWrapper}>
                    <View style={styles.centerContent}>
                        <ActivityIndicator size="large" color="#10B981" />
                        <Text style={styles.loadingText}>Getting your location...</Text>
                        <Text style={styles.loadingSubtext}>This may take a few seconds</Text>
                        <TouchableOpacity
                            style={[styles.button, styles.buttonSecondary]}
                            onPress={() => setShowZipModal(true)}
                        >
                            <Text style={[styles.buttonText, styles.buttonSecondaryText]}>
                                Enter Zip Code Instead
                            </Text>
                        </TouchableOpacity>
                    </View>
                    {renderZipModal()}
                </View>
            </SafeAreaView>
        );
    }

    if (locationError || !location) {
        return (
            <SafeAreaView style={styles.container}>
                <View style={styles.contentWrapper}>
                    <View style={styles.centerContent}>
                        <Text style={styles.emoji}>📍</Text>
                        <Text style={styles.errorTitle}>Location Required</Text>
                        <Text style={styles.errorText}>
                            {locationError || 'We need your location to show nearby games'}
                        </Text>
                        <TouchableOpacity style={styles.button} onPress={requestPermission}>
                            <Text style={styles.buttonText}>Enable Location</Text>
                        </TouchableOpacity>
                        <TouchableOpacity
                            style={[styles.button, styles.buttonSecondary]}
                            onPress={() => setShowZipModal(true)}
                        >
                            <Text style={[styles.buttonText, styles.buttonSecondaryText]}>
                                Enter Zip Code
                            </Text>
                        </TouchableOpacity>
                    </View>
                    {renderZipModal()}
                </View>
            </SafeAreaView>
        );
    }

    return (
        <SafeAreaView style={styles.container}>
            <View style={styles.contentWrapper}>
            {/* Header */}
            <View style={styles.header}>
                <Text style={styles.title}>Games Near You</Text>
            </View>

            {/* Category Pill Slider */}
            <View style={styles.categorySliderWrapper}>
                <ScrollView
                    horizontal
                    showsHorizontalScrollIndicator={false}
                    style={styles.categorySlider}
                    contentContainerStyle={styles.categorySliderContent}
                >
                    {SPORT_CATEGORIES.map(cat => (
                        <TouchableOpacity
                            key={cat.key}
                            style={[
                                styles.categoryPill,
                                selectedCategory === cat.key && styles.categoryPillActive
                            ]}
                            onPress={() => handleCategoryChange(cat.key)}
                        >
                            <Text style={styles.categoryPillEmoji}>{cat.emoji}</Text>
                            <Text style={[
                                styles.categoryPillText,
                                selectedCategory === cat.key && styles.categoryPillTextActive
                            ]}>
                                {cat.label}
                            </Text>
                        </TouchableOpacity>
                    ))}
                </ScrollView>
            </View>

            {/* Search & Filters */}
            <View style={styles.searchContainer}>
                <View style={styles.searchBar}>
                    <Ionicons name="search" size={20} color="#9CA3AF" />
                    <TextInput
                        style={styles.searchInput}
                        placeholder="Search location..."
                        value={searchQuery}
                        onChangeText={setSearchQuery}
                        placeholderTextColor="#9CA3AF"
                    />
                </View>
            </View>

            {/* Radius Indicator */}
            <View style={styles.radiusContainer}>
                <Text style={styles.radiusText}>
                    Showing games within {radiusMiles} miles • {filteredGames.length} games
                </Text>
            </View>

            {/* Games List */}
            <FlatList
                data={filteredGames}
                renderItem={({ item }) => {
                    const distance = calculateDistance(
                        location.latitude,
                        location.longitude,
                        item.location.latitude || location.latitude,
                        item.location.longitude || location.longitude
                    );
                    return (
                        <GameCard
                            game={item}
                            distance={distance}
                            onPress={() => setSelectedGame(item)}
                        />
                    );
                }}
                keyExtractor={item => item.id}
                refreshControl={
                    <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
                }
                ListEmptyComponent={
                    <View style={styles.emptyContainer}>
                        <Text style={styles.emptyEmoji}>🏀</Text>
                        <Text style={styles.emptyText}>No games found nearby</Text>
                        <Text style={styles.emptySubtext}>
                            Try adjusting your filters or radius
                        </Text>
                    </View>
                }
                contentContainerStyle={styles.listContent}
            />

            <GameDetailsModal
                game={selectedGame}
                visible={selectedGame !== null}
                onClose={() => setSelectedGame(null)}
                distance={
                    selectedGame
                        ? calculateDistance(
                              location.latitude,
                              location.longitude,
                              selectedGame.location.latitude || location.latitude,
                              selectedGame.location.longitude || location.longitude
                          )
                        : undefined
                }
                onJoinSuccess={loadGames}
            />

            {/* Floating Action Button */}
            <TouchableOpacity
                style={styles.fab}
                onPress={() => router.push('/create-game')}
            >
                <Ionicons name="add" size={28} color="#FFFFFF" />
            </TouchableOpacity>
            </View>
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#ECFDF5',
        alignItems: 'center',
    },
    contentWrapper: {
        flex: 1,
        width: '100%',
        maxWidth: 1200,
    },
    centerContent: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 24,
    },
    header: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: 24,
        paddingBottom: 12,
    },
    title: {
        fontSize: 32,
        fontWeight: 'bold',
        color: '#1F2937',
    },
    categorySliderWrapper: {
        height: 52,
        marginBottom: 16,
    },
    categorySlider: {
        flexGrow: 0,
        paddingLeft: 24,
    },
    categorySliderContent: {
        paddingRight: 24,
        paddingVertical: 2,
    },
    categoryPill: {
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 24,
        paddingHorizontal: 20,
        marginRight: 12,
        minWidth: 100,
        height: 48,
    },
    categoryPillActive: {
        backgroundColor: '#ECFDF5',
        borderColor: '#10B981',
    },
    categoryPillEmoji: {
        fontSize: 20,
        marginRight: 8,
    },
    categoryPillText: {
        fontSize: 16,
        fontWeight: '600',
        color: '#6B7280',
    },
    categoryPillTextActive: {
        color: '#10B981',
    },
    searchContainer: {
        paddingHorizontal: 24,
        marginBottom: 12,
    },
    searchBar: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: '#FFFFFF',
        borderRadius: 12,
        padding: 12,
        gap: 8,
    },
    searchInput: {
        flex: 1,
        fontSize: 16,
        color: '#1F2937',
    },
    radiusContainer: {
        paddingHorizontal: 24,
        paddingVertical: 12,
    },
    radiusText: {
        fontSize: 13,
        color: '#6B7280',
    },
    listContent: {
        paddingBottom: 24,
    },
    emptyContainer: {
        alignItems: 'center',
        padding: 48,
    },
    emptyEmoji: {
        fontSize: 64,
        marginBottom: 16,
    },
    emptyText: {
        fontSize: 18,
        fontWeight: '600',
        color: '#1F2937',
        marginBottom: 8,
    },
    emptySubtext: {
        fontSize: 14,
        color: '#9CA3AF',
        textAlign: 'center',
    },
    loadingText: {
        marginTop: 16,
        fontSize: 16,
        color: '#6B7280',
    },
    loadingSubtext: {
        marginTop: 8,
        marginBottom: 24,
        fontSize: 14,
        color: '#9CA3AF',
    },
    emoji: {
        fontSize: 64,
        marginBottom: 16,
    },
    errorTitle: {
        fontSize: 24,
        fontWeight: 'bold',
        color: '#1F2937',
        marginBottom: 8,
    },
    errorText: {
        fontSize: 16,
        color: '#6B7280',
        textAlign: 'center',
        marginBottom: 24,
    },
    button: {
        backgroundColor: '#10B981',
        paddingVertical: 16,
        paddingHorizontal: 32,
        borderRadius: 12,
        marginBottom: 12,
        width: '100%',
        alignItems: 'center',
    },
    buttonSecondary: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#10B981',
    },
    buttonText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
    buttonSecondaryText: {
        color: '#10B981',
    },
    modalOverlay: {
        flex: 1,
        backgroundColor: 'rgba(0, 0, 0, 0.5)',
        justifyContent: 'flex-end',
    },
    modalContent: {
        backgroundColor: '#FFFFFF',
        borderTopLeftRadius: 24,
        borderTopRightRadius: 24,
        padding: 24,
        maxHeight: '80%',
    },
    modalHeader: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 24,
    },
    modalTitle: {
        fontSize: 24,
        fontWeight: 'bold',
        color: '#1F2937',
    },
    zipModalDescription: {
        fontSize: 15,
        color: '#6B7280',
        marginBottom: 20,
        textAlign: 'center',
    },
    zipInput: {
        backgroundColor: '#F9FAFB',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        padding: 16,
        fontSize: 18,
        marginBottom: 16,
        color: '#1F2937',
        textAlign: 'center',
        fontWeight: '600',
    },
    zipSubmitButton: {
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
    zipSubmitButtonDisabled: {
        backgroundColor: '#9CA3AF',
        shadowOpacity: 0,
    },
    zipSubmitButtonText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
    fab: {
        position: 'absolute',
        bottom: 24,
        right: 24,
        width: 56,
        height: 56,
        borderRadius: 28,
        backgroundColor: '#10B981',
        justifyContent: 'center',
        alignItems: 'center',
        shadowColor: '#10B981',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.4,
        shadowRadius: 12,
        elevation: 8,
    },
});
