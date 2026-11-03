import type { BrowseEntry } from '@/api/types.ts';

export const isJavadocJar = (entry: BrowseEntry) => entry.type === 'file' && entry.name.endsWith('-javadoc.jar');
export const join = (...parts: string[]) => parts.filter(Boolean).join('/');
export const browseUrl = (path: string) =>
    `/browse/${path.split('/').filter(Boolean).map(encodeURIComponent).join('/')}`;
