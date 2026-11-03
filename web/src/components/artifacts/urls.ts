import { useLocation } from 'react-router';

interface ArtifactId {
    groupId: string;
    artifactId: string;
}

export const artifactSettingsUrl = ({ groupId, artifactId }: ArtifactId) =>
    `/artifacts/${encodeURIComponent(groupId)}/${encodeURIComponent(artifactId)}`;

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
