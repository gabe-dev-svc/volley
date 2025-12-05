import React from 'react';
import { View, Text, StyleSheet, TouchableOpacity } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useAuth } from '../../contexts/AuthContext';

export default function ProfileScreen() {
    const { user, logout } = useAuth();

    return (
        <SafeAreaView style={styles.container}>
            <View style={styles.contentWrapper}>
                <View style={styles.content}>
                    <Text style={styles.emoji}>👤</Text>
                    <Text style={styles.title}>Profile</Text>
                    {user && (
                        <View style={styles.userInfo}>
                            <Text style={styles.userName}>
                                {user.firstName} {user.lastName}
                            </Text>
                            <Text style={styles.userEmail}>{user.email}</Text>
                        </View>
                    )}
                    <TouchableOpacity style={styles.button} onPress={logout}>
                        <Text style={styles.buttonText}>Sign Out</Text>
                    </TouchableOpacity>
                </View>
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
    content: {
        flex: 1,
        justifyContent: 'center',
        alignItems: 'center',
        padding: 24,
    },
    emoji: {
        fontSize: 64,
        marginBottom: 16,
    },
    title: {
        fontSize: 32,
        fontWeight: 'bold',
        color: '#1F2937',
        marginBottom: 24,
    },
    userInfo: {
        alignItems: 'center',
        marginBottom: 32,
    },
    userName: {
        fontSize: 20,
        fontWeight: '600',
        color: '#1F2937',
        marginBottom: 4,
    },
    userEmail: {
        fontSize: 16,
        color: '#6B7280',
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
