import React from 'react';
import { Platform, View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { Game } from '../lib/api';

interface GameMapProps {
    games: Game[];
    userLocation: { latitude: number; longitude: number };
    onBackToList: () => void;
}

// Web fallback component
function WebMapFallback({ onBackToList }: { onBackToList: () => void }) {
    return (
        <View style={styles.centerContent}>
            <Text style={styles.emoji}>🗺️</Text>
            <Text style={styles.errorTitle}>Map View</Text>
            <Text style={styles.errorText}>
                Map view is only available on mobile devices
            </Text>
            <TouchableOpacity style={styles.button} onPress={onBackToList}>
                <Text style={styles.buttonText}>Back to List</Text>
            </TouchableOpacity>
        </View>
    );
}

// Native map component
function NativeMap({ games, userLocation }: GameMapProps) {
    if (Platform.OS === 'web') {
        return null;
    }

    const MapView = require('react-native-maps').default;
    const { Marker, PROVIDER_GOOGLE } = require('react-native-maps');

    return (
        <MapView
            style={styles.map}
            provider={PROVIDER_GOOGLE}
            initialRegion={{
                latitude: userLocation.latitude,
                longitude: userLocation.longitude,
                latitudeDelta: 0.2,
                longitudeDelta: 0.2,
            }}
            showsUserLocation
        >
            {games.map(game => (
                game.location.latitude && game.location.longitude && (
                    <Marker
                        key={game.id}
                        coordinate={{
                            latitude: game.location.latitude,
                            longitude: game.location.longitude,
                        }}
                        title={game.title || game.category}
                        description={game.location.name}
                    />
                )
            ))}
        </MapView>
    );
}

export function GameMap(props: GameMapProps) {
    if (Platform.OS === 'web') {
        return <WebMapFallback onBackToList={props.onBackToList} />;
    }
    return <NativeMap {...props} />;
}

const styles = StyleSheet.create({
    map: {
        flex: 1,
    },
    centerContent: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 24,
        backgroundColor: '#ECFDF5',
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
    },
    buttonText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
});
