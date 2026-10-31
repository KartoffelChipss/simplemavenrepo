import { request, seg } from '@/api/client.ts';
import { artifactKeys } from '@/api/keys.ts';
import type { Artifact, CreateArtifactRequest, UpdateArtifactRequest } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

const artifactPath = (groupId: string, artifactId: string) => `/artifacts/${seg(groupId)}/${seg(artifactId)}`;

export const listArtifacts = (signal?: AbortSignal) => request<Artifact[]>('GET', '/artifacts', undefined, signal);
export const getArtifact = (groupId: string, artifactId: string, signal?: AbortSignal) =>
    request<Artifact>('GET', artifactPath(groupId, artifactId), undefined, signal);
export const createArtifact = (body: CreateArtifactRequest) => request<Artifact>('POST', '/artifacts', body);
export const updateArtifact = (groupId: string, artifactId: string, body: UpdateArtifactRequest) =>
    request<Artifact>('PATCH', artifactPath(groupId, artifactId), body);
export const deleteArtifact = (groupId: string, artifactId: string) =>
    request<void>('DELETE', artifactPath(groupId, artifactId));

interface ArtifactId {
    groupId: string;
    artifactId: string;
}

export const artifactsQuery = () =>
    queryOptions({
        queryKey: artifactKeys.list(),
        queryFn: ({ signal }) => listArtifacts(signal),
    });

export const artifactQuery = ({ groupId, artifactId }: ArtifactId) =>
    queryOptions({
        queryKey: artifactKeys.detail(groupId, artifactId),
        queryFn: ({ signal }) => getArtifact(groupId, artifactId, signal),
    });

export function useArtifacts() {
    return useQuery(artifactsQuery());
}

export function useArtifact(id: ArtifactId) {
    return useQuery(artifactQuery(id));
}

export function useCreateArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createArtifact,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: artifactKeys.list() }),
    });
}

export function useUpdateArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ groupId, artifactId, ...body }: UpdateArtifactRequest & ArtifactId) =>
            updateArtifact(groupId, artifactId, body),
        onSuccess: (artifact) => {
            queryClient.setQueryData(artifactKeys.detail(artifact.groupId, artifact.artifactId), artifact);
            return queryClient.invalidateQueries({ queryKey: artifactKeys.list() });
        },
    });
}

export function useDeleteArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ groupId, artifactId }: ArtifactId) => deleteArtifact(groupId, artifactId),
        onSuccess: (_, { groupId, artifactId }) => {
            queryClient.removeQueries({ queryKey: artifactKeys.detail(groupId, artifactId) });
            return queryClient.invalidateQueries({ queryKey: artifactKeys.list() });
        },
    });
}
