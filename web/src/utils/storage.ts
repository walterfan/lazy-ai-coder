/**
 * Safe wrappers around Web Storage APIs.
 *
 * Motivation: localStorage/sessionStorage access can throw (privacy mode, blocked storage),
 * and stale/corrupt values (e.g. "undefined") can crash JSON.parse().
 *
 * These helpers never throw; they return null / no-op instead.
 */

export interface StorageLike {
  getItem(key: string): string | null;
  setItem(key: string, value: string): void;
  removeItem(key: string): void;
  clear(): void;
}

function getLocalStorage(): StorageLike | null {
  try {
    if (typeof window === 'undefined') return null;
    return window.localStorage;
  } catch {
    return null;
  }
}

export const storage = {
  getItem(key: string): string | null {
    try {
      const ls = getLocalStorage();
      return ls ? ls.getItem(key) : null;
    } catch {
      return null;
    }
  },

  setItem(key: string, value: string): void {
    try {
      const ls = getLocalStorage();
      if (!ls) return;
      ls.setItem(key, value);
    } catch {
      // noop
    }
  },

  removeItem(key: string): void {
    try {
      const ls = getLocalStorage();
      if (!ls) return;
      ls.removeItem(key);
    } catch {
      // noop
    }
  },

  clear(): void {
    try {
      const ls = getLocalStorage();
      if (!ls) return;
      ls.clear();
    } catch {
      // noop
    }
  },
};

export function safeJsonParse<T>(value: string | null): T | null {
  if (!value) return null;
  try {
    return JSON.parse(value) as T;
  } catch {
    return null;
  }
}


