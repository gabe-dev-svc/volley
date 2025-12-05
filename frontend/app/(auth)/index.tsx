import React, { useState } from 'react';
import {
    View,
    Text,
    TextInput,
    TouchableOpacity,
    StyleSheet,
    KeyboardAvoidingView,
    Platform,
    ScrollView,
    Alert,
    ActivityIndicator,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useRouter } from 'expo-router';
import { AntDesign } from '@expo/vector-icons';
import { useAuth } from '../../contexts/AuthContext';
import { ApiRequestError } from '../../lib/api';

export default function LoginScreen() {
    const router = useRouter();
    const { login, register } = useAuth();
    const [isLogin, setIsLogin] = useState(true);
    const [email, setEmail] = useState('');
    const [password, setPassword] = useState('');
    const [firstName, setFirstName] = useState('');
    const [lastName, setLastName] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async () => {
        // Clear previous errors
        setError(null);

        // Validation
        if (!email || !password) {
            setError('Please fill in all required fields');
            return;
        }

        if (!isLogin && (!firstName || !lastName)) {
            setError('Please enter your first and last name');
            return;
        }

        setIsLoading(true);

        try {
            if (isLogin) {
                await login(email, password);
            } else {
                await register(firstName, lastName, email, password);
            }
            // Navigation will happen automatically via the root layout when auth state changes
        } catch (error) {
            // Handle 401 Unauthorized errors (incorrect credentials)
            if (error instanceof ApiRequestError && error.statusCode === 401) {
                setError(isLogin
                    ? 'Invalid email or password. Please try again.'
                    : 'Unable to create account. Please check your details.');
            } else if (error instanceof ApiRequestError && error.statusCode === 409) {
                // Handle 409 Conflict (duplicate account)
                setError('An account with this email already exists.');
            } else {
                // Generic error handling
                const errorMessage = error instanceof Error ? error.message : 'An error occurred';
                setError(errorMessage);
            }
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <SafeAreaView style={styles.container}>
            <KeyboardAvoidingView
                behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
                style={styles.keyboardView}
            >
                <ScrollView contentContainerStyle={styles.scrollContent}>
                    <View style={styles.contentWrapper}>
                        {/* Header */}
                        <View style={styles.header}>
                        <Text style={styles.emoji}>🏀</Text>
                        <Text style={styles.title}>
                            {isLogin ? 'Welcome Back' : 'Get Started'}
                        </Text>
                        <Text style={styles.subtitle}>
                            {isLogin ? 'Sign in to join pickup games' : 'Create your account'}
                        </Text>
                    </View>

                    {/* Card */}
                    <View style={styles.card}>
                        {/* Toggle */}
                        <View style={styles.toggle}>
                            <TouchableOpacity
                                style={[styles.toggleBtn, isLogin && styles.toggleBtnActive]}
                                onPress={() => {
                                    setIsLogin(true);
                                    setError(null);
                                }}
                            >
                                <Text style={[styles.toggleText, isLogin && styles.toggleTextActive]}>
                                    Sign In
                                </Text>
                            </TouchableOpacity>
                            <TouchableOpacity
                                style={[styles.toggleBtn, !isLogin && styles.toggleBtnActive]}
                                onPress={() => {
                                    setIsLogin(false);
                                    setError(null);
                                }}
                            >
                                <Text style={[styles.toggleText, !isLogin && styles.toggleTextActive]}>
                                    Sign Up
                                </Text>
                            </TouchableOpacity>
                        </View>

                        {/* Google Sign In - Coming Soon */}
                        <View style={styles.googleContainer}>
                            <TouchableOpacity style={styles.googleBtn} disabled>
                                <View style={{ flexDirection: 'row', alignItems: 'center' }}>
                                    <AntDesign name="google" size={20} color="#DB4437" style={{ marginRight: 8 }} />
                                    <Text style={styles.googleText}>Continue with Google</Text>
                                </View>
                            </TouchableOpacity>
                            <View style={styles.comingSoonBadge}>
                                <Text style={styles.comingSoonText}>Coming Soon</Text>
                            </View>
                        </View>

                        {/* Divider */}
                        <View style={styles.divider}>
                            <View style={styles.dividerLine} />
                            <Text style={styles.dividerText}>Or continue with email</Text>
                            <View style={styles.dividerLine} />
                        </View>

                        {/* Form */}
                        <View style={styles.form}>
                            {!isLogin && (
                                <>
                                    <TextInput
                                        style={styles.input}
                                        placeholder="First Name"
                                        value={firstName}
                                        onChangeText={setFirstName}
                                        autoCapitalize="words"
                                        placeholderTextColor="#9CA3AF"
                                    />
                                    <TextInput
                                        style={styles.input}
                                        placeholder="Last Name"
                                        value={lastName}
                                        onChangeText={setLastName}
                                        autoCapitalize="words"
                                        placeholderTextColor="#9CA3AF"
                                    />
                                </>
                            )}
                            <TextInput
                                style={styles.input}
                                placeholder="Email"
                                value={email}
                                onChangeText={(text) => {
                                    setEmail(text);
                                    setError(null);
                                }}
                                keyboardType="email-address"
                                autoCapitalize="none"
                                autoCorrect={false}
                                placeholderTextColor="#9CA3AF"
                            />
                            <TextInput
                                style={styles.input}
                                placeholder="Password"
                                value={password}
                                onChangeText={(text) => {
                                    setPassword(text);
                                    setError(null);
                                }}
                                secureTextEntry
                                placeholderTextColor="#9CA3AF"
                            />
                        </View>

                        {/* Error Message */}
                        {error && (
                            <View style={styles.errorContainer}>
                                <Text style={styles.errorText}>{error}</Text>
                            </View>
                        )}

                        {/* Submit */}
                        <TouchableOpacity
                            style={[styles.submitBtn, isLoading && styles.submitBtnDisabled]}
                            onPress={handleSubmit}
                            disabled={isLoading}
                        >
                            {isLoading ? (
                                <ActivityIndicator color="#FFFFFF" />
                            ) : (
                                <Text style={styles.submitText}>
                                    {isLogin ? 'Sign In' : 'Create Account'}
                                </Text>
                            )}
                        </TouchableOpacity>

                        {/* Switch */}
                        <TouchableOpacity onPress={() => setIsLogin(!isLogin)}>
                            <Text style={styles.switchText}>
                                {isLogin ? "Don't have an account? " : 'Already have an account? '}
                                <Text style={styles.switchLink}>
                                    {isLogin ? 'Sign up' : 'Sign in'}
                                </Text>
                            </Text>
                        </TouchableOpacity>
                    </View>
                    </View>
                </ScrollView>
            </KeyboardAvoidingView>
        </SafeAreaView>
    );
}

const styles = StyleSheet.create({
    container: {
        flex: 1,
        backgroundColor: '#ECFDF5',
    },
    keyboardView: {
        flex: 1,
    },
    scrollContent: {
        flexGrow: 1,
        justifyContent: 'center',
        padding: 24,
        alignItems: 'center',
        width: '100%',
    },
    contentWrapper: {
        width: '100%',
        maxWidth: 480,
    },
    header: {
        alignItems: 'center',
        marginBottom: 32,
    },
    emoji: {
        fontSize: 64,
        marginBottom: 16,
    },
    title: {
        fontSize: 32,
        fontWeight: 'bold',
        color: '#1F2937',
        marginBottom: 8,
    },
    subtitle: {
        fontSize: 16,
        color: '#6B7280',
        textAlign: 'center',
    },
    card: {
        backgroundColor: '#FFFFFF',
        borderRadius: 24,
        padding: 24,
        shadowColor: '#000',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.1,
        shadowRadius: 12,
        elevation: 5,
    },
    toggle: {
        flexDirection: 'row',
        backgroundColor: '#F3F4F6',
        borderRadius: 12,
        padding: 4,
        marginBottom: 24,
    },
    toggleBtn: {
        flex: 1,
        paddingVertical: 12,
        borderRadius: 8,
        alignItems: 'center',
    },
    toggleBtnActive: {
        backgroundColor: '#10B981',
    },
    toggleText: {
        fontSize: 14,
        fontWeight: '600',
        color: '#6B7280',
    },
    toggleTextActive: {
        color: '#FFFFFF',
    },
    googleContainer: {
        position: 'relative',
        marginBottom: 24,
    },
    googleBtn: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        paddingVertical: 14,
        borderRadius: 12,
        alignItems: 'center',
        opacity: 0.6,
    },
    googleText: {
        fontSize: 14,
        fontWeight: '600',
        color: '#374151',
    },
    comingSoonBadge: {
        position: 'absolute',
        top: '50%',
        left: '50%',
        transform: [{ translateX: -45 }, { translateY: -12 }],
        backgroundColor: '#10B981',
        paddingHorizontal: 12,
        paddingVertical: 6,
        borderRadius: 20,
    },
    comingSoonText: {
        color: '#FFFFFF',
        fontSize: 11,
        fontWeight: 'bold',
    },
    divider: {
        flexDirection: 'row',
        alignItems: 'center',
        marginBottom: 24,
    },
    dividerLine: {
        flex: 1,
        height: 1,
        backgroundColor: '#E5E7EB',
    },
    dividerText: {
        paddingHorizontal: 12,
        fontSize: 13,
        color: '#9CA3AF',
    },
    form: {
        marginBottom: 24,
    },
    input: {
        backgroundColor: '#FFFFFF',
        borderWidth: 2,
        borderColor: '#E5E7EB',
        borderRadius: 12,
        padding: 16,
        fontSize: 16,
        marginBottom: 12,
        color: '#1F2937',
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
    submitBtn: {
        backgroundColor: '#10B981',
        paddingVertical: 16,
        borderRadius: 12,
        alignItems: 'center',
        marginBottom: 16,
        shadowColor: '#10B981',
        shadowOffset: { width: 0, height: 4 },
        shadowOpacity: 0.3,
        shadowRadius: 8,
        elevation: 4,
    },
    submitBtnDisabled: {
        backgroundColor: '#9CA3AF',
        shadowOpacity: 0,
    },
    submitText: {
        color: '#FFFFFF',
        fontSize: 16,
        fontWeight: '600',
    },
    switchText: {
        textAlign: 'center',
        color: '#6B7280',
        fontSize: 14,
    },
    switchLink: {
        color: '#10B981',
        fontWeight: '600',
    },
});