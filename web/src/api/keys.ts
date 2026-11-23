import type { ArtifactRef, GrantFilter } from '@/api/types.ts';

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
    detail: ({ repository, groupId, artifactId }: ArtifactRef) =>
        [...artifactKeys.all, 'detail', repository, groupId, artifactId] as const,
    deploys: (id: ArtifactRef, page: number) => [...artifactKeys.detail(id), 'deploys', page] as const,
    downloads: (id: ArtifactRef, days: number) => [...artifactKeys.detail(id), 'downloads', days] as const,
};

export const repoKeys = {
    all: ['repositories'] as const,
    list: () => [...repoKeys.all, 'list'] as const,
    detail: (name: string) => [...repoKeys.all, 'detail', name] as const,
};

export const browseKeys = {
    all: ['browse'] as const,
    list: () => [...browseKeys.all, 'list'] as const,
    path: (path: string) => [...browseKeys.all, 'path', path] as const,
    pom: (versionPath: string) => [...browseKeys.all, 'pom', versionPath] as const,
    vulnerabilities: (versionPath: string) => [...browseKeys.all, 'vulnerabilities', versionPath] as const,
};

export const javadocKeys = {
    all: ['javadoc'] as const,
    list: () => [...javadocKeys.all, 'list'] as const,
    path: (path: string) => [...javadocKeys.all, path] as const,
};
