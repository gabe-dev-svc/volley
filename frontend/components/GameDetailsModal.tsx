import React, { useState } from 'react';
import {
    View,
    Text,
    StyleSheet,
    Modal,
    TouchableOpacity,
    ScrollView,
    ActivityIndicator,
    Alert,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { Game, apiClient } from '../lib/api';

interface GameDetailsModalProps {
    game: Game | null;
    visible: boolean;
    onClose: () => void;
    distance?: number;
    onJoinSuccess?: () => void;
}

export function GameDetailsModal({
    game,
    visible,
    onClose,
    distance,
    onJoinSuccess,
}: GameDetailsModalProps) {
    const [isJoining, setIsJoining] = useState(false);

    if (!game) return null;

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
            hour12: true,
        });
    };

    const formatPrice = () => {
        if (game.pricing.type === 'free') {
            return 'Free';
        }
        const dollars = game.pricing.amountCents / 100;
        return `$${dollars.toFixed(2)}`;
    };

    const formatPricingType = () => {
        if (game.pricing.type === 'free') {
            return 'This game is free to join!';
        }
        return game.pricing.type === 'per_person'
            ? 'Per person'
            : 'Total (split among players)';
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

    const getStatusInfo = () => {
        switch (game.status) {
            case 'open':
                return {
                    text: 'Open for Registration',
                    color: '#10B981',
                    bgColor: '#D1FAE5',
                };
            case 'full':
                return {
                    text: 'Full (Waitlist Available)',
                    color: '#F59E0B',
                    bgColor: '#FEF3C7',
                };
            case 'closed':
                return {
                    text: 'Registration Closed',
                    color: '#EF4444',
                    bgColor: '#FEE2E2',
                };
            case 'cancelled':
                return {
                    text: 'Cancelled',
                    color: '#EF4444',
                    bgColor: '#FEE2E2',
                };
            default:
                return {
                    text: game.status,
                    color: '#6B7280',
                    bgColor: '#E5E7EB',
                };
        }
    };

    const handleJoinGame = async () => {
        try {
            setIsJoining(true);
            await apiClient.joinGame(game.id);
            Alert.alert(
                'Success!',
                'You have successfully joined this game. Check your email for details.',
                [
                    {
                        text: 'OK',
                        onPress: () => {
                            onClose();
                            onJoinSuccess?.();
                        },
                    },
                ]
            );
        } catch (error) {
            Alert.alert(
                'Error',
                error instanceof Error ? error.message : 'Failed to join game'
            );
        } finally {
            setIsJoining(false);
        }
    };

    const canJoin = game.status === 'open' || game.status === 'full';
    const statusInfo = getStatusInfo();

    return (
        <Modal
            visible={visible}
            animationType="slide"
            presentationStyle="pageSheet"
            onRequestClose={onClose}
        >
            <View style={styles.container}>
                {/* Header */}
                <View style={styles.header}>
                    <View style={styles.headerLeft}>
                        <Text style={styles.emoji}>{getCategoryEmoji()}</Text>
                        <Text style={styles.headerTitle}>Game Details</Text>
                    </View>
                    <TouchableOpacity onPress={onClose} style={styles.closeButton}>
                        <Ionicons name="close" size={28} color="#6B7280" />
                    </TouchableOpacity>
                </View>

                <ScrollView style={styles.scrollView} showsVerticalScrollIndicator={false}>
                    {/* Title */}
                    <Text style={styles.title}>
                        {game.title || game.category.replace('_', ' ')}
                    </Text>

                    {/* Status Badge */}
                    <View style={[styles.statusBadge, { backgroundColor: statusInfo.bgColor }]}>
                        <Text style={[styles.statusText, { color: statusInfo.color }]}>
                            {statusInfo.text}
                        </Text>
                    </View>

                    {/* Description */}
                    {game.description && (
                        <View style={styles.section}>
                            <Text style={styles.sectionTitle}>About</Text>
                            <Text style={styles.description}>{game.description}</Text>
                        </View>
                    )}

                    {/* Organizer */}
                    {game.owner && (
                        <View style={styles.section}>
                            <Text style={styles.sectionTitle}>Organized by</Text>
                            <View style={styles.organizerCard}>
                                <View style={styles.organizerIcon}>
                                    <Text style={styles.organizerInitial}>
                                        {game.owner.firstName.charAt(0)}{game.owner.lastName.charAt(0)}
                                    </Text>
                                </View>
                                <View style={styles.organizerInfo}>
                                    <Text style={styles.organizerName}>
                                        {game.owner.firstName} {game.owner.lastName}
                                    </Text>
                                    <Text style={styles.organizerEmail}>{game.owner.email}</Text>
                                </View>
                            </View>
                        </View>
                    )}

                    {/* Date & Time */}
                    <View style={styles.section}>
                        <Text style={styles.sectionTitle}>When</Text>
                        <View style={styles.infoRow}>
                            <Ionicons name="calendar-outline" size={20} color="#10B981" />
                            <Text style={styles.infoText}>{formatDate(game.startTime)}</Text>
                        </View>
                        <View style={styles.infoRow}>
                            <Ionicons name="time-outline" size={20} color="#10B981" />
                            <Text style={styles.infoText}>
                                {formatTime(game.startTime)} ({game.durationMinutes} minutes)
                            </Text>
                        </View>
                    </View>

                    {/* Location */}
                    <View style={styles.section}>
                        <Text style={styles.sectionTitle}>Where</Text>
                        <View style={styles.infoRow}>
                            <Ionicons name="location-outline" size={20} color="#10B981" />
                            <View style={styles.locationInfo}>
                                <Text style={styles.infoText}>{game.location.name}</Text>
                                {game.location.address && (
                                    <Text style={styles.addressText}>{game.location.address}</Text>
                                )}
                                {distance !== undefined && (
                                    <Text style={styles.distanceText}>
                                        {distance.toFixed(1)} miles away
                                    </Text>
                                )}
                            </View>
                        </View>
                        {game.location.notes && (
                            <Text style={styles.notesText}>{game.location.notes}</Text>
                        )}
                    </View>

                    {/* Pricing */}
                    <View style={styles.section}>
                        <Text style={styles.sectionTitle}>Pricing</Text>
                        <View style={styles.pricingCard}>
                            <Text style={styles.priceAmount}>{formatPrice()}</Text>
                            <Text style={styles.pricingType}>{formatPricingType()}</Text>
                        </View>
                    </View>

                    {/* Participants */}
                    <View style={styles.section}>
                        <Text style={styles.sectionTitle}>Participants</Text>
                        <View style={styles.participantsCard}>
                            <View style={styles.participantsRow}>
                                <Text style={styles.participantsLabel}>Current Players:</Text>
                                <Text style={styles.participantsValue}>
                                    {game.currentParticipants || 0} / {game.maxParticipants}
                                </Text>
                            </View>
                            {game.waitlistCount !== undefined && game.waitlistCount > 0 && (
                                <View style={styles.participantsRow}>
                                    <Text style={styles.participantsLabel}>Waitlist:</Text>
                                    <Text style={styles.participantsValue}>{game.waitlistCount}</Text>
                                </View>
                            )}
                            <View style={styles.participantsRow}>
                                <Text style={styles.participantsLabel}>Skill Level:</Text>
                                <Text style={styles.participantsValue}>{game.skillLevel}</Text>
                            </View>
                        </View>
                    </View>

                    {/* Additional Notes */}
                    {game.notes && (
                        <View style={styles.section}>
                            <Text style={styles.sectionTitle}>Additional Notes</Text>
                            <Text style={styles.notesText}>{game.notes}</Text>
                        </View>
                    )}

                    <View style={{ height: 100 }} />
                </ScrollView>

                {/* Join Button */}
                {canJoin && (
                    <View style={styles.footer}>
                        <TouchableOpacity
                            style={[
                                styles.joinButton,
                                isJoining && styles.joinButtonDisabled,
                            ]}
                            onPress={handleJoinGame}
                            disabled={isJoining}
                        >
                            {isJoining ? (
                                <ActivityIndicator color="#FFFFFF" />
                            ) : (
                                <>
                                    <Ionicons name="checkmark-circle" size={24} color="#FFFFFF" />
                                    <Text style={styles.joinButtonText}>
                                        {game.status === 'full' ? 'Join Waitlist' : 'Join Game'}
                                    </Text>
                                </>
                            )}
                        </TouchableOpacity>
                    </View>
                )}
            </View>
        </Modal>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#FFFFFF',
    },
    header: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        padding: 20,
        borderBottomWidth: 1,
        borderBottomColor: '#E5E7EB',
    },
    headerLeft: {
        flexDirection: 'row',
        alignItems: 'center',
    },
    emoji: {
        fontSize: 32,
        marginRight: 12,
    },
    headerTitle: {
        fontSize: 20,
        fontWeight: 'bold',
        color: '#1F2937',
    },
    closeButton: {
        padding: 4,
    },
    scrollView: {
        flex: 1,
        padding: 20,
    },
    title: {
        fontSize: 28,
        fontWeight: 'bold',
        color: '#1F2937',
        marginBottom: 16,
        textTransform: 'capitalize',
    },
    statusBadge: {
        alignSelf: 'flex-start',
        paddingHorizontal: 16,
        paddingVertical: 8,
        borderRadius: 20,
        marginBottom: 24,
    },
    statusText: {
        fontSize: 14,
        fontWeight: '600',
    },
    section: {
        marginBottom: 24,
    },
    sectionTitle: {
        fontSize: 18,
        fontWeight: '600',
        color: '#1F2937',
        marginBottom: 12,
    },
    description: {
        fontSize: 16,
        color: '#6B7280',
        lineHeight: 24,
    },
    organizerCard: {
        flexDirection: 'row',
        alignItems: 'center',
        backgroundColor: '#F9FAFB',
        padding: 16,
        borderRadius: 12,
    },
    organizerIcon: {
        width: 48,
        height: 48,
        borderRadius: 24,
        backgroundColor: '#10B981',
        justifyContent: 'center',
        alignItems: 'center',
        marginRight: 12,
    },
    organizerInitial: {
        fontSize: 18,
        fontWeight: 'bold',
        color: '#FFFFFF',
    },
    organizerInfo: {
        flex: 1,
    },
    organizerName: {
        fontSize: 16,
        fontWeight: '600',
        color: '#1F2937',
        marginBottom: 2,
    },
    organizerEmail: {
        fontSize: 14,
        color: '#6B7280',
    },
    infoRow: {
        flexDirection: 'row',
        alignItems: 'flex-start',
        marginBottom: 12,
    },
    infoText: {
        fontSize: 16,
        color: '#1F2937',
        flex: 1,
    },
    locationInfo: {
        flex: 1,
        marginLeft: 12,
    },
    addressText: {
        fontSize: 14,
        color: '#6B7280',
        marginTop: 4,
    },
    distanceText: {
        fontSize: 13,
        color: '#9CA3AF',
        marginTop: 4,
    },
    notesText: {
        fontSize: 14,
        color: '#6B7280',
        fontStyle: 'italic',
        marginTop: 8,
        paddingLeft: 32,
    },
    pricingCard: {
        backgroundColor: '#ECFDF5',
        padding: 20,
        borderRadius: 12,
        alignItems: 'center',
    },
    priceAmount: {
        fontSize: 32,
        fontWeight: 'bold',
        color: '#10B981',
        marginBottom: 4,
    },
    pricingType: {
        fontSize: 14,
        color: '#6B7280',
    },
    participantsCard: {
        backgroundColor: '#F9FAFB',
        padding: 16,
        borderRadius: 12,
    },
    participantsRow: {
        flexDirection: 'row',
        justifyContent: 'space-between',
        alignItems: 'center',
        marginBottom: 8,
    },
    participantsLabel: {
        fontSize: 15,
        color: '#6B7280',
    },
    participantsValue: {
        fontSize: 15,
        fontWeight: '600',
        color: '#1F2937',
    },
    footer: {
        padding: 20,
        borderTopWidth: 1,
        borderTopColor: '#E5E7EB',
        backgroundColor: '#FFFFFF',
    },
    joinButton: {
        backgroundColor: '#10B981',
        flexDirection: 'row',
        alignItems: 'center',
        justifyContent: 'center',
        paddingVertical: 16,
        borderRadius: 12,
        gap: 8,
        shadowColor: '#10B981',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.3,
        shadowRadius: 8,
        elevation: 4,
    },
    joinButtonDisabled: {
        backgroundColor: '#9CA3AF',
        shadowOpacity: 0,
    },
    joinButtonText: {
        color: '#FFFFFF',
        fontSize: 18,
        fontWeight: '600',
    },
});
