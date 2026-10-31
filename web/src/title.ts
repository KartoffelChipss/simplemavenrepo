import { useServerName } from '@/api/settings.ts';
import { useEffect, useSyncExternalStore } from 'react';

interface Override {
    title: string;
}

let current: Override | undefined;
const listeners = new Set<() => void>();

function setOverride(override: Override | undefined) {
    current = override;
    listeners.forEach((l) => l());
}

function subscribe(listener: () => void) {
    listeners.add(listener);
    return () => listeners.delete(listener);
}

export function usePageTitle(title: string | undefined) {
    useEffect(() => {
        if (title === undefined) return;
        const override = { title };
        setOverride(override);
        return () => {
            if (current === override) setOverride(undefined);
        };
    }, [title]);
}

export function usePageTitleOverride(): string | undefined {
    return useSyncExternalStore(subscribe, () => current?.title);
}

export function useDocumentTitle(title: string | undefined) {
    const serverName = useServerName();
    useEffect(() => {
        document.title = title ? `${title} · ${serverName}` : serverName;
    }, [title, serverName]);
}
