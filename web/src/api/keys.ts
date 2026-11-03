import type { GrantFilter } from '@/api/types.ts';

export const aboutKeys = {
    all: ['about'] as const,
};

export const statsKeys = {
    all: ['stats'] as const,
    storage: () => [...statsKeys.all, 'storage'] as const,
    downloads: (days: number) => [...statsKeys.all, 'downloads', days] as const,
};

export const settingsKeys = {
    all: ['settings'] as const,
};

export const meKeys = {
    all: ['me'] as const,
    tokens: () => [...meKeys.all, 'tokens'] as const,
};

export const userKeys = {
    all: ['users'] as const,
    list: () => [...userKeys.all, 'list'] as const,
    detail: (username: string) => [...userKeys.all, 'detail', username] as const,
    tokens: (username: string) => [...userKeys.detail(username), 'tokens'] as const,
};

export const teamKeys = {
    all: ['teams'] as const,
    list: () => [...teamKeys.all, 'list'] as const,
    detail: (name: string) => [...teamKeys.all, 'detail', name] as const,
};

export const grantKeys = {
    all: ['grants'] as const,
    list: (filter: GrantFilter = {}) => [...grantKeys.all, 'list', filter] as const,
};

export const artifactKeys = {
    all: ['artifacts'] as const,
    list: () => [...artifactKeys.all, 'list'] as const,
    detail: (groupId: string, artifactId: string) => [...artifactKeys.all, 'detail', groupId, artifactId] as const,
    deploys: (groupId: string, artifactId: string, page: number) =>
        [...artifactKeys.detail(groupId, artifactId), 'deploys', page] as const,
    downloads: (groupId: string, artifactId: string, days: number) =>
        [...artifactKeys.detail(groupId, artifactId), 'downloads', days] as const,
};

export const browseKeys = {
    all: ['browse'] as const,
    list: () => [...browseKeys.all, 'list'] as const,
    path: (path: string) => [...browseKeys.all, 'path', path] as const,
    pom: (versionPath: string) => [...browseKeys.all, 'pom', versionPath] as const,
};

export const javadocKeys = {
    all: ['javadoc'] as const,
    list: () => [...javadocKeys.all, 'list'] as const,
    path: (path: string) => [...javadocKeys.all, path] as const,
};
