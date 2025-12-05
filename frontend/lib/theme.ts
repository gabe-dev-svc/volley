/**
 * Centralized theme configuration for consistent styling across the app
 */

export const Colors = {
  // Primary Colors (Vibrant Green)
  primary: '#10B981',        // Emerald 500 - Main brand color
  primaryLight: '#34D399',   // Emerald 400 - Lighter variant
  primaryDark: '#059669',    // Emerald 600 - Darker variant

  // Background Colors
  background: '#ECFDF5',     // Emerald 50 - Main background
  backgroundLight: '#F0FDF4', // Emerald 50 lighter
  surface: '#FFFFFF',        // White - Cards, modals
  surfaceSecondary: '#F9FAFB', // Gray 50 - Secondary surface

  // Accent Colors
  accent: '#D1FAE5',         // Emerald 100 - Light green accent
  accentLight: '#A7F3D0',    // Emerald 200 - Lighter accent

  // Text Colors
  textPrimary: '#1F2937',    // Gray 800 - Primary text
  textSecondary: '#6B7280',  // Gray 500 - Secondary text
  textTertiary: '#9CA3AF',   // Gray 400 - Tertiary text
  textInverse: '#FFFFFF',    // White - Text on colored backgrounds

  // Border Colors
  border: '#E5E7EB',         // Gray 200 - Default borders
  borderLight: '#F3F4F6',    // Gray 100 - Light borders
  borderDark: '#D1D5DB',     // Gray 300 - Dark borders

  // Status Colors
  success: '#10B981',        // Emerald 500 - Success state
  successBg: '#D1FAE5',      // Emerald 100 - Success background

  warning: '#F59E0B',        // Amber 500 - Warning state
  warningBg: '#FEF3C7',      // Amber 100 - Warning background

  error: '#EF4444',          // Red 500 - Error state
  errorBg: '#FEE2E2',        // Red 100 - Error background

  info: '#3B82F6',           // Blue 500 - Info state
  infoBg: '#DBEAFE',         // Blue 100 - Info background

  // Interactive States
  hover: '#059669',          // Emerald 600 - Hover state
  active: '#047857',         // Emerald 700 - Active/pressed state
  disabled: '#9CA3AF',       // Gray 400 - Disabled state

  // Shadow Colors
  shadow: '#10B981',         // Emerald 500 - For colored shadows
  shadowDark: '#000000',     // Black - For dark shadows
} as const;

export const Shadows = {
  small: {
    shadowColor: Colors.shadowDark,
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 2,
    elevation: 1,
  },
  medium: {
    shadowColor: Colors.shadowDark,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 2,
  },
  large: {
    shadowColor: Colors.shadowDark,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 8,
    elevation: 4,
  },
  primary: {
    shadowColor: Colors.primary,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 4,
  },
} as const;

export const Spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
  xxxl: 48,
} as const;

export const BorderRadius = {
  sm: 8,
  md: 12,
  lg: 16,
  xl: 20,
  full: 9999,
} as const;

export const FontSizes = {
  xs: 12,
  sm: 14,
  md: 16,
  lg: 18,
  xl: 20,
  xxl: 24,
  xxxl: 32,
} as const;

export const FontWeights = {
  normal: '400' as const,
  medium: '500' as const,
  semibold: '600' as const,
  bold: '700' as const,
} as const;

// Type exports for TypeScript autocomplete
export type ColorKey = keyof typeof Colors;
export type ShadowKey = keyof typeof Shadows;
export type SpacingKey = keyof typeof Spacing;
export type BorderRadiusKey = keyof typeof BorderRadius;
export type FontSizeKey = keyof typeof FontSizes;
export type FontWeightKey = keyof typeof FontWeights;
