import { useSyncExternalStore } from 'react';

export type Theme = 'system' | 'light' | 'dark';

const STORAGE_KEY = 'smr.theme';

const systemDark = matchMedia('(prefers-color-scheme: dark)');
const listeners = new Set<() => void>();

function readTheme(): Theme {
    try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored === 'light' || stored === 'dark') return stored;
    } catch {
        // Storage unavailable
    }
    return 'system';
}

let current = readTheme();

function apply() {
    const dark = current === 'dark' || (current === 'system' && systemDark.matches);
    document.documentElement.classList.toggle('dark', dark);
}

systemDark.addEventListener('change', apply);
apply();

export function setTheme(theme: Theme) {
    current = theme;
    try {
        if (theme === 'system') localStorage.removeItem(STORAGE_KEY);
        else localStorage.setItem(STORAGE_KEY, theme);
    } catch {
        // Storage unavailable
    }
    apply();
    listeners.forEach((l) => l());
}

function subscribe(listener: () => void) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

export function useTheme(): Theme {
    return useSyncExternalStore(subscribe, () => current);
}
