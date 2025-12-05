import React, { useEffect, useState } from 'react';
import {
    View,
    Text,
    StyleSheet,
    ScrollView,
    TouchableOpacity,
    ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter, useLocalSearchParams } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { apiClient, Game, ApiRequestError } from '../../lib/api';

const CATEGORY_EMOJIS: Record<string, string> = {
    soccer: '⚽',
    basketball: '🏀',
    volleyball: '🏐',
    pickleball: '🏓',
    flag_football: '🏈',
    ultimate_frisbee: '🥏',
    tennis: '🎾',
    other: '🏃',
};

const STATUS_LABELS: Record<string, { label: string; color: string; bgColor: string }> = {
    open: { label: 'Open', color: '#10B981', bgColor: '#D1FAE5' },
    full: { label: 'Full', color: '#F59E0B', bgColor: '#FEF3C7' },
    closed: { label: 'Closed', color: '#6B7280', bgColor: '#F3F4F6' },
    in_progress: { label: 'In Progress', color: '#3B82F6', bgColor: '#DBEAFE' },
    completed: { label: 'Completed', color: '#8B5CF6', bgColor: '#EDE9FE' },
    cancelled: { label: 'Cancelled', color: '#EF4444', bgColor: '#FEE2E2' },
};

export default function GameDetailsScreen() {
    const router = useRouter();
    const params = useLocalSearchParams();
    const gameId = params.id as string;

    const [game, setGame] = useState<Game | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        // For now, we'll just show the game ID since we need to add a getGame endpoint
        // This is a placeholder implementation
        setIsLoading(false);
    }, [gameId]);

    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            weekday: 'long',
            month: 'long',
            day: 'numeric',
            year: 'numeric',
        });
    };

    const formatTime = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleTimeString('en-US', {
            hour: 'numeric',
            minute: '2-digit',
        });
    };

    const formatPrice = (amountCents: number, currency: string) => {
        return `$${(amountCents / 100).toFixed(2)}`;
    };

    if (isLoading) {
        return (
            <SafeAreaView style={styles.container}>
                <View style={styles.loadingContainer}>
                    <ActivityIndicator size="large" color="#10B981" />
                    <Text style={styles.loadingText}>Loading game details...</Text>
                </View>
            </SafeAreaView>
        );
    }

    if (error) {
        return (
            <SafeAreaView style={styles.container}>
                <View style={styles.header}>
                    <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
                        <Ionicons name="arrow-back" size={24} color="#1F2937" />
                    </TouchableOpacity>
                    <Text style={styles.headerTitle}>Game Details</Text>
                    <View style={{ width: 24 }} />
                </View>
                <View style={styles.errorContainer}>
                    <Ionicons name="alert-circle" size={48} color="#EF4444" />
                    <Text style={styles.errorText}>{error}</Text>
                    <TouchableOpacity
                        style={styles.retryButton}
                        onPress={() => {
                            setError(null);
                            setIsLoading(true);
                        }}
                    >
                        <Text style={styles.retryButtonText}>Try Again</Text>
                    </TouchableOpacity>
                </View>
            </SafeAreaView>
        );
    }

    return (
        <SafeAreaView style={styles.container}>
            <View style={styles.header}>
                <TouchableOpacity onPress={() => router.back()} style={styles.backButton}>
                    <Ionicons name="arrow-back" size={24} color="#1F2937" />
                </TouchableOpacity>
                <Text style={styles.headerTitle}>Game Details</Text>
                <View style={{ width: 24 }} />
            </View>

            <ScrollView style={styles.content}>
                <View style={styles.successCard}>
                    <View style={styles.successIcon}>
                        <Ionicons name="checkmark-circle" size={64} color="#10B981" />
                    </View>
                    <Text style={styles.successTitle}>Game Created!</Text>
                    <Text style={styles.successSubtitle}>
                        Your game has been successfully created and is now visible to other players.
                    </Text>
                    <Text style={styles.gameIdText}>Game ID: {gameId}</Text>
                </View>

                <View style={styles.infoCard}>
                    <Text style={styles.infoTitle}>What's Next?</Text>
                    <View style={styles.infoItem}>
                        <Ionicons name="people-outline" size={24} color="#10B981" />
                        <Text style={styles.infoText}>
                            Players can now find and join your game
                        </Text>
                    </View>
                    <View style={styles.infoItem}>
                        <Ionicons name="notifications-outline" size={24} color="#10B981" />
                        <Text style={styles.infoText}>
                            You'll be notified when players join
                        </Text>
                    </View>
                    <View style={styles.infoItem}>
                        <Ionicons name="calendar-outline" size={24} color="#10B981" />
                        <Text style={styles.infoText}>
                            Check "My Games" to manage your game
                        </Text>
                    </View>
                </View>

                <TouchableOpacity
                    style={styles.primaryButton}
                    onPress={() => router.push('/(tabs)')}
                >
                    <Text style={styles.primaryButtonText}>Find More Games</Text>
                </TouchableOpacity>

                <TouchableOpacity
                    style={styles.secondaryButton}
                    onPress={() => router.push('/(tabs)/my-games')}
                >
                    <Text style={styles.secondaryButtonText}>View My Games</Text>
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
        padding: 24,
    },
    loadingContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        gap: 16,
    },
    loadingText: {
        fontSize: 16,
        color: '#6B7280',
    },
    errorContainer: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 24,
        gap: 16,
    },
    errorText: {
        fontSize: 16,
        color: '#EF4444',
        textAlign: 'center',
    },
    retryButton: {
        backgroundColor: '#10B981',
        paddingVertical: 12,
        paddingHorizontal: 24,
        borderRadius: 12,
        marginTop: 8,
    },
    retryButtonText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
    successCard: {
        backgroundColor: '#FFFFFF',
        borderRadius: 16,
        padding: 32,
        alignItems: 'center',
        marginBottom: 24,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 8,
        elevation: 3,
    },
    successIcon: {
        marginBottom: 16,
    },
    successTitle: {
        fontSize: 28,
        fontWeight: 'bold',
        color: '#1F2937',
        marginBottom: 8,
    },
    successSubtitle: {
        fontSize: 16,
        color: '#6B7280',
        textAlign: 'center',
        marginBottom: 16,
    },
    gameIdText: {
        fontSize: 14,
        color: '#9CA3AF',
        fontFamily: 'monospace',
    },
    infoCard: {
        backgroundColor: '#FFFFFF',
        borderRadius: 16,
        padding: 24,
        marginBottom: 24,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 2 },
        shadowOpacity: 0.1,
        shadowRadius: 8,
        elevation: 3,
    },
    infoTitle: {
        fontSize: 20,
        fontWeight: '600',
        color: '#1F2937',
        marginBottom: 16,
    },
    infoItem: {
        flexDirection: 'row',
        alignItems: 'center',
        gap: 12,
        marginBottom: 16,
    },
    infoText: {
        flex: 1,
        fontSize: 16,
        color: '#374151',
    },
    primaryButton: {
        backgroundColor: '#10B981',
        paddingVertical: 16,
        borderRadius: 12,
        alignItems: 'center',
        marginBottom: 12,
        shadowColor: '#10B981',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.3,
        shadowRadius: 8,
        elevation: 4,
    },
    primaryButtonText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
    secondaryButton: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#10B981',
        paddingVertical: 16,
        borderRadius: 12,
        alignItems: 'center',
    },
    secondaryButtonText: {
        color: '#10B981',
        fontSize: 16,
        fontWeight: '600',
    },
});
