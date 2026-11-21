import type { ArtifactRef } from '@/api/types.ts';
import { useLocation } from 'react-router';

export const artifactSettingsUrl = ({ repository, groupId, artifactId }: ArtifactRef) =>
    `/artifacts/${encodeURIComponent(repository)}/${encodeURIComponent(groupId)}/${encodeURIComponent(artifactId)}`;

export interface BackTo {
    to: string;
    label: string;
}

export function useBackState(label: string): { back: BackTo } {
    const { pathname, search } = useLocation();
    return { back: { to: pathname + search, label } };
}

export const artifactBrowseUrl = (artifact: { path: string }) =>
    `/browse/${artifact.path.split('/').map(encodeURIComponent).join('/')}`;
