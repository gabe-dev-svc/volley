import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Game } from '../lib/api';
import { Colors, Shadows, BorderRadius, Spacing, FontSizes, FontWeights } from '../lib/theme';

interface GameCardProps {
    game: Game;
    distance?: number; // Distance in miles
    onPress?: () => void;
}

export function GameCard({ game, distance, onPress }: GameCardProps) {
    const formatDate = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleDateString('en-US', {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
        });
    };

    const formatTime = (dateString: string) => {
        const date = new Date(dateString);
        return date.toLocaleTimeString('en-US', {
            hour: 'numeric',
            minute: '2-digit',
            hour12: true,
        });
    };

    const formatPrice = () => {
        if (game.pricing.type === 'free') {
            return 'Free';
        }
        const dollars = game.pricing.amountCents / 100;
        return `$${dollars.toFixed(2)}${game.pricing.type === 'per_person' ? '/person' : ' total'}`;
    };

    const getStatusBadgeColor = () => {
        switch (game.status) {
            case 'open':
                return styles.badgeOpen;
            case 'full':
                return styles.badgeFull;
            case 'closed':
            case 'cancelled':
                return styles.badgeClosed;
            default:
                return styles.badgeDefault;
        }
    };

    const getCategoryEmoji = () => {
        switch (game.category) {
            case 'soccer': return '⚽';
            case 'basketball': return '🏀';
            case 'pickleball': return '🏓';
            case 'flag_football': return '🏈';
            case 'volleyball': return '🏐';
            case 'ultimate_frisbee': return '🥏';
            case 'tennis': return '🎾';
            default: return '🏃';
        }
    };

    return (
        <TouchableOpacity
            style={styles.card}
            onPress={onPress}
            activeOpacity={0.7}
        >
            {/* Header with emoji and status */}
            <View style={styles.header}>
                <View style={styles.headerLeft}>
                    <Text style={styles.emoji}>{getCategoryEmoji()}</Text>
                    <View>
                        <Text style={styles.title}>
                            {game.title || game.category.replace('_', ' ')}
                        </Text>
                        {distance !== undefined && (
                            <Text style={styles.distance}>{distance.toFixed(1)} mi away</Text>
                        )}
                    </View>
                </View>
                <View style={[styles.badge, getStatusBadgeColor()]}>
                    <Text style={styles.badgeText}>{game.status}</Text>
                </View>
            </View>

            {/* Location */}
            <View style={styles.row}>
                <Text style={styles.label}>📍</Text>
                <View style={styles.locationContainer}>
                    <Text style={styles.value}>{game.location.name}</Text>
                    {game.location.address && (
                        <Text style={styles.address}>{game.location.address}</Text>
                    )}
                </View>
            </View>

            {/* Date & Time */}
            <View style={styles.row}>
                <Text style={styles.label}>🕐</Text>
                <Text style={styles.value}>
                    {formatDate(game.startTime)} at {formatTime(game.startTime)}
                </Text>
            </View>

            {/* Participants */}
            <View style={styles.row}>
                <Text style={styles.label}>👥</Text>
                <Text style={styles.value}>
                    {game.currentParticipants || 0}/{game.maxParticipants} players
                    {game.waitlistCount ? ` • ${game.waitlistCount} waitlisted` : ''}
                </Text>
            </View>

            {/* Footer with price and skill level */}
            <View style={styles.footer}>
                <View style={styles.chip}>
                    <Text style={styles.chipText}>{formatPrice()}</Text>
                </View>
                <View style={styles.chip}>
                    <Text style={styles.chipText}>{game.skillLevel}</Text>
                </View>
            </View>
        </TouchableOpacity>
    );
}

const styles = StyleSheet.create({
    card: {
        backgroundColor: Colors.surface,
        borderRadius: BorderRadius.xl,
        padding: Spacing.xl,
        marginHorizontal: Spacing.xl,
        marginBottom: Spacing.lg,
        ...Shadows.large,
    },
    header: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'flex-start',
        marginBottom: Spacing.lg,
    },
    headerLeft: {
        flexDirection: 'row',
        alignItems: 'center',
        flex: 1,
    },
    emoji: {
        fontSize: FontSizes.xxxl,
        marginRight: Spacing.md,
    },
    title: {
        fontSize: FontSizes.lg,
        fontWeight: FontWeights.bold,
        color: Colors.textPrimary,
        textTransform: 'capitalize',
    },
    distance: {
        fontSize: FontSizes.xs + 1,
        color: Colors.textTertiary,
        marginTop: 2,
    },
    badge: {
        paddingHorizontal: Spacing.md,
        paddingVertical: 6,
        borderRadius: BorderRadius.xl,
    },
    badgeOpen: {
        backgroundColor: Colors.successBg,
    },
    badgeFull: {
        backgroundColor: Colors.warningBg,
    },
    badgeClosed: {
        backgroundColor: Colors.errorBg,
    },
    badgeDefault: {
        backgroundColor: Colors.border,
    },
    badgeText: {
        fontSize: 11,
        fontWeight: FontWeights.bold,
        color: Colors.textPrimary,
        textTransform: 'uppercase',
    },
    row: {
        flexDirection: 'row',
        alignItems: 'flex-start',
        marginBottom: Spacing.md,
    },
    label: {
        fontSize: FontSizes.md,
        marginRight: Spacing.sm,
        width: 24,
        marginTop: 2,
    },
    locationContainer: {
        flex: 1,
    },
    value: {
        fontSize: FontSizes.sm,
        color: Colors.textSecondary,
        flex: 1,
    },
    address: {
        fontSize: FontSizes.xs + 1,
        color: Colors.textTertiary,
        marginTop: 4,
    },
    footer: {
        flexDirection: 'row',
        gap: Spacing.sm,
        marginTop: Spacing.sm,
    },
    chip: {
        backgroundColor: Colors.accent,
        paddingHorizontal: Spacing.md,
        paddingVertical: 6,
        borderRadius: BorderRadius.md,
    },
    chipText: {
        fontSize: FontSizes.xs + 1,
        fontWeight: FontWeights.semibold,
        color: Colors.primary,
    },
});
