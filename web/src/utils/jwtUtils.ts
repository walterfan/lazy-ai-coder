/**
 * JWT Utility Functions
 * Provides utilities for decoding and validating JWT tokens
 */

export interface JWTPayload {
  user_id: string;
  realm_id: string;
  username: string;
  email: string;
  roles?: string[];
  iss: string;
  sub: string;
  aud: string[];
  exp: number;
  nbf: number;
  iat: number;
}

/**
 * Decode a JWT token without verification
 * @param token JWT token string
 * @returns Decoded payload or null if invalid
 */
export function decodeJWT(token: string): JWTPayload | null {
  try {
    // JWT has 3 parts: header.payload.signature
    const parts = token.split('.');
    if (parts.length !== 3) {
      console.error('Invalid JWT format');
      return null;
    }

    // Decode the payload (second part)
    const payload = parts[1];

    // JWT uses base64url encoding, need to convert to regular base64
    const base64 = payload.replace(/-/g, '+').replace(/_/g, '/');

    // Decode base64
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );

    return JSON.parse(jsonPayload) as JWTPayload;
  } catch (error) {
    console.error('Failed to decode JWT:', error);
    return null;
  }
}

/**
 * Get token expiration time in milliseconds
 * @param token JWT token string
 * @returns Expiration timestamp in milliseconds, or null if invalid
 */
export function getTokenExpiration(token: string): number | null {
  const payload = decodeJWT(token);
  if (!payload || !payload.exp) {
    return null;
  }

  // exp is in seconds, convert to milliseconds
  return payload.exp * 1000;
}

/**
 * Check if a token is expired
 * @param token JWT token string
 * @returns true if expired, false if still valid
 */
export function isTokenExpired(token: string): boolean {
  const expiration = getTokenExpiration(token);
  if (!expiration) {
    return true; // Treat invalid tokens as expired
  }

  return Date.now() >= expiration;
}

/**
 * Check if a token will expire soon (within specified minutes)
 * @param token JWT token string
 * @param minutesThreshold Minutes before expiration to consider "expiring soon"
 * @returns true if token expires within the threshold
 */
export function isTokenExpiringSoon(token: string, minutesThreshold: number = 5): boolean {
  const expiration = getTokenExpiration(token);
  if (!expiration) {
    return true; // Treat invalid tokens as expiring
  }

  const thresholdMs = minutesThreshold * 60 * 1000;
  const timeUntilExpiration = expiration - Date.now();

  return timeUntilExpiration <= thresholdMs && timeUntilExpiration > 0;
}

/**
 * Get time remaining until token expiration in milliseconds
 * @param token JWT token string
 * @returns Time remaining in milliseconds, or 0 if expired/invalid
 */
export function getTimeUntilExpiration(token: string): number {
  const expiration = getTokenExpiration(token);
  if (!expiration) {
    return 0;
  }

  const remaining = expiration - Date.now();
  return Math.max(0, remaining);
}

/**
 * Format time remaining as human-readable string
 * @param milliseconds Time in milliseconds
 * @returns Formatted string (e.g., "2h 30m", "45m", "30s")
 */
export function formatTimeRemaining(milliseconds: number): string {
  if (milliseconds <= 0) {
    return 'expired';
  }

  const seconds = Math.floor(milliseconds / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) {
    return `${days}d ${hours % 24}h`;
  } else if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  } else if (minutes > 0) {
    return `${minutes}m`;
  } else {
    return `${seconds}s`;
  }
}
