interface ArtifactId {
    groupId: string;
    artifactId: string;
}

export const artifactSettingsUrl = ({ groupId, artifactId }: ArtifactId) =>
    `/artifacts/${encodeURIComponent(groupId)}/${encodeURIComponent(artifactId)}`;

export const artifactBrowseUrl = (artifact: { path: string }) =>
    `/browse/${artifact.path.split('/').map(encodeURIComponent).join('/')}`;
