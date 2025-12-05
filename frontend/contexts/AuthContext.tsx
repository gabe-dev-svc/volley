import React, { createContext, useContext, useState, useEffect } from 'react';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { apiClient, AuthResponse } from '../lib/api';

interface User {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
}

interface AuthContextType {
    user: User | null;
    token: string | null;
    isLoading: boolean;
    login: (email: string, password: string) => Promise<void>;
    register: (firstName: string, lastName: string, email: string, password: string) => Promise<void>;
    logout: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_KEY = '@volley_auth_token';
const USER_KEY = '@volley_user';

export function AuthProvider({ children }: { children: React.ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    // Load auth state on mount
    useEffect(() => {
        loadAuthState();
    }, []);

    // Update API client token when token changes
    useEffect(() => {
        apiClient.setToken(token);
    }, [token]);

    const loadAuthState = async () => {
        try {
            const [storedToken, storedUser] = await Promise.all([
                AsyncStorage.getItem(TOKEN_KEY),
                AsyncStorage.getItem(USER_KEY),
            ]);

            if (storedToken && storedUser) {
                setToken(storedToken);
                setUser(JSON.parse(storedUser));
            }
        } catch (error) {
            console.error('Failed to load auth state:', error);
        } finally {
            setIsLoading(false);
        }
    };

    const saveAuthState = async (authResponse: AuthResponse) => {
        const { token: newToken, user: newUser } = authResponse;

        if (!newToken) {
            throw new Error('No token received from server');
        }

        await Promise.all([
            AsyncStorage.setItem(TOKEN_KEY, newToken),
            AsyncStorage.setItem(USER_KEY, JSON.stringify(newUser)),
        ]);

        setToken(newToken);
        setUser(newUser);
    };

    const clearAuthState = async () => {
        await Promise.all([
            AsyncStorage.removeItem(TOKEN_KEY),
            AsyncStorage.removeItem(USER_KEY),
        ]);

        setToken(null);
        setUser(null);
    };

    const login = async (email: string, password: string) => {
        try {
            const response = await apiClient.login({ email, password });
            await saveAuthState(response);
        } catch (error) {
            console.error('Login failed:', error);
            throw error;
        }
    };

    const register = async (
        firstName: string,
        lastName: string,
        email: string,
        password: string
    ) => {
        try {
            const response = await apiClient.register({
                firstName,
                lastName,
                email,
                password,
            });
            await saveAuthState(response);
        } catch (error) {
            console.error('Registration failed:', error);
            throw error;
        }
    };

    const logout = async () => {
        await clearAuthState();
    };

    return (
        <AuthContext.Provider
            value={{
                user,
                token,
                isLoading,
                login,
                register,
                logout,
            }}
        >
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
