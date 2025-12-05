import { useState, useEffect } from 'react';
import * as Location from 'expo-location';

export interface LocationCoords {
    latitude: number;
    longitude: number;
}

export interface UseLocationResult {
    location: LocationCoords | null;
    error: string | null;
    isLoading: boolean;
    permissionStatus: Location.PermissionStatus | null;
    requestPermission: () => Promise<void>;
    setManualLocation: (coords: LocationCoords) => void;
    refreshLocation: () => Promise<void>;
}

export function useLocation(): UseLocationResult {
    const [location, setLocation] = useState<LocationCoords | null>(null);
    const [error, setError] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [permissionStatus, setPermissionStatus] = useState<Location.PermissionStatus | null>(null);

    const requestPermission = async () => {
        try {
            const { status } = await Location.requestForegroundPermissionsAsync();
            setPermissionStatus(status);

            if (status === Location.PermissionStatus.GRANTED) {
                await getCurrentLocation();
            } else {
                setError('Location permission denied');
                setIsLoading(false);
            }
        } catch (err) {
            setError('Failed to request location permission');
            setIsLoading(false);
        }
    };

    const getCurrentLocation = async () => {
        try {
            setIsLoading(true);
            setError(null);

            // Add timeout to prevent hanging
            const timeoutPromise = new Promise<never>((_, reject) => {
                setTimeout(() => reject(new Error('Location request timed out. Please try manually entering your location.')), 5000);
            });

            const locationPromise = Location.getCurrentPositionAsync({
                accuracy: Location.Accuracy.Balanced,
            });

            const position = await Promise.race([locationPromise, timeoutPromise]);

            setLocation({
                latitude: position.coords.latitude,
                longitude: position.coords.longitude,
            });
        } catch (err) {
            const errorMessage = err instanceof Error ? err.message : 'Failed to get current location';
            setError(errorMessage);
            console.error('Location error:', err);
        } finally {
            setIsLoading(false);
        }
    };

    const setManualLocation = (coords: LocationCoords) => {
        setLocation(coords);
        setError(null);
        setIsLoading(false);
    };

    const refreshLocation = async () => {
        if (permissionStatus === Location.PermissionStatus.GRANTED) {
            await getCurrentLocation();
        }
    };

    useEffect(() => {
        let mounted = true;
        let timeoutId: NodeJS.Timeout;

        const initLocation = async () => {
            try {
                // Set a safety timeout to ensure we never hang forever
                timeoutId = setTimeout(() => {
                    if (mounted) {
                        console.warn('Location initialization timed out');
                        setError('Location initialization timed out');
                        setIsLoading(false);
                    }
                }, 8000);

                // Check existing permission status
                const { status } = await Location.getForegroundPermissionsAsync();

                if (!mounted) return;

                setPermissionStatus(status);

                if (status === Location.PermissionStatus.GRANTED) {
                    await getCurrentLocation();
                } else {
                    if (mounted) {
                        setIsLoading(false);
                    }
                }

                // Clear timeout if we completed successfully
                clearTimeout(timeoutId);
            } catch (err) {
                console.error('Permission check error:', err);
                if (mounted) {
                    setError('Failed to check location permissions');
                    setIsLoading(false);
                }
            }
        };

        initLocation();

        return () => {
            mounted = false;
            if (timeoutId) clearTimeout(timeoutId);
        };
    }, []);

    return {
        location,
        error,
        isLoading,
        permissionStatus,
        requestPermission,
        setManualLocation,
        refreshLocation,
    };
}
