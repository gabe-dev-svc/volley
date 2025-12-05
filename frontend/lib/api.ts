import Constants from 'expo-constants';

// Get API URL from environment or default to localhost
const API_URL = Constants.expoConfig?.extra?.apiUrl || 'http://192.168.1.129:8080/v1';

export interface RegisterRequest {
    firstName: string;
    lastName: string;
    email: string;
    password: string;
}

export interface LoginRequest {
    email: string;
    password: string;
}

export interface AuthResponse {
    token?: string; // Only for mobile clients
    user: {
        id: string;
        email: string;
        firstName: string;
        lastName: string;
    };
}

export interface ApiError {
    error: string;
    code?: string;
    details?: Record<string, any>;
}

export class ApiRequestError extends Error {
    statusCode: number;
    apiError?: ApiError;

    constructor(message: string, statusCode: number, apiError?: ApiError) {
        super(message);
        this.name = 'ApiRequestError';
        this.statusCode = statusCode;
        this.apiError = apiError;
    }
}

export type GameCategory =
    | 'soccer'
    | 'basketball'
    | 'pickleball'
    | 'flag_football'
    | 'volleyball'
    | 'ultimate_frisbee'
    | 'tennis'
    | 'other';

export type GameStatus =
    | 'open'
    | 'full'
    | 'closed'
    | 'in_progress'
    | 'completed'
    | 'cancelled';

export type SkillLevel = 'beginner' | 'intermediate' | 'advanced' | 'all';

export type PricingType = 'free' | 'total' | 'per_person';

export interface GameLocation {
    name: string;
    address?: string;
    latitude?: number;
    longitude?: number;
    notes?: string;
}

export interface Pricing {
    type: PricingType;
    amountCents: number;
    currency: string;
}

export interface User {
    id: string;
    email: string;
    firstName: string;
    lastName: string;
}

export interface Game {
    id: string;
    ownerId: string;
    owner?: User;
    category: GameCategory;
    title?: string;
    description?: string;
    location: GameLocation;
    startTime: string;
    durationMinutes: number;
    maxParticipants: number;
    currentParticipants?: number;
    waitlistCount?: number;
    pricing: Pricing;
    signupDeadline: string;
    skillLevel: SkillLevel;
    notes?: string;
    status: GameStatus;
    cancelledAt?: string | null;
    createdAt: string;
    updatedAt: string;
}

export interface ListGamesParams {
    categories: GameCategory[];
    latitude: number;
    longitude: number;
    radius?: number; // meters, default 16093.4 (10 miles)
    timeFilter?: 'upcoming' | 'past' | 'all';
    status?: GameStatus;
    limit?: number;
    offset?: number;
}

export interface ListGamesResponse {
    games: Game[];
}

export interface JoinGameResponse {
    success: boolean;
    message?: string;
}

export interface CreateGameRequest {
    category: GameCategory;
    title?: string;
    description?: string;
    location: GameLocation;
    startTime: string; // ISO 8601 date string
    durationMinutes: number;
    maxParticipants: number;
    pricing: Pricing;
    signupDeadline?: string; // ISO 8601 date string, defaults to startTime
    skillLevel?: SkillLevel;
    notes?: string;
}

class ApiClient {
    private baseUrl: string;
    private token: string | null = null;

    constructor(baseUrl: string) {
        this.baseUrl = baseUrl;
    }

    setToken(token: string | null) {
        this.token = token;
    }

    private async request<T>(
        endpoint: string,
        options: RequestInit = {}
    ): Promise<T> {
        const headers: Record<string, string> = {
            'Content-Type': 'application/json',
            'X-Client-Type': 'mobile', // Important for mobile clients to get JWT in response
        };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const response = await fetch(`${this.baseUrl}${endpoint}`, {
            ...options,
            headers,
        });

        const data = await response.json();

        if (!response.ok) {
            const error: ApiError = data;
            throw new ApiRequestError(
                error.error || 'An error occurred',
                response.status,
                error
            );
        }

        return data as T;
    }

    async register(request: RegisterRequest): Promise<AuthResponse> {
        return this.request<AuthResponse>('/auth/register', {
            method: 'POST',
            body: JSON.stringify(request),
        });
    }

    async login(request: LoginRequest): Promise<AuthResponse> {
        return this.request<AuthResponse>('/auth/login', {
            method: 'POST',
            body: JSON.stringify(request),
        });
    }

    async listGames(params: ListGamesParams): Promise<ListGamesResponse> {
        const queryParams = new URLSearchParams();

        // Required parameters
        params.categories.forEach(cat => queryParams.append('categories', cat));
        queryParams.append('latitude', params.latitude.toString());
        queryParams.append('longitude', params.longitude.toString());

        // Optional parameters
        if (params.radius !== undefined) {
            queryParams.append('radius', params.radius.toString());
        }
        if (params.timeFilter) {
            queryParams.append('timeFilter', params.timeFilter);
        }
        if (params.status) {
            queryParams.append('status', params.status);
        }
        if (params.limit !== undefined) {
            queryParams.append('limit', params.limit.toString());
        }
        if (params.offset !== undefined) {
            queryParams.append('offset', params.offset.toString());
        }

        return this.request<ListGamesResponse>(`/games?${queryParams.toString()}`);
    }

    async createGame(request: CreateGameRequest): Promise<Game> {
        return this.request<Game>('/games', {
            method: 'POST',
            body: JSON.stringify(request),
        });
    }

    async joinGame(gameId: string): Promise<JoinGameResponse> {
        return this.request<JoinGameResponse>(`/games/${gameId}/join`, {
            method: 'POST',
        });
    }

    async leaveGame(gameId: string): Promise<JoinGameResponse> {
        return this.request<JoinGameResponse>(`/games/${gameId}/leave`, {
            method: 'POST',
        });
    }
}

export const apiClient = new ApiClient(API_URL);
