import { request, seg } from '@/api/client.ts';
import { artifactKeys, browseKeys } from '@/api/keys.ts';
import type { Artifact, CreateArtifactRequest, DeployPage, UpdateArtifactRequest } from '@/api/types.ts';
import { keepPreviousData, queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

const artifactPath = (groupId: string, artifactId: string) => `/artifacts/${seg(groupId)}/${seg(artifactId)}`;

export const listArtifacts = (signal?: AbortSignal) => request<Artifact[]>('GET', '/artifacts', undefined, signal);
export const getArtifact = (groupId: string, artifactId: string, signal?: AbortSignal) =>
    request<Artifact>('GET', artifactPath(groupId, artifactId), undefined, signal);
export const createArtifact = (body: CreateArtifactRequest) => request<Artifact>('POST', '/artifacts', body);
export const updateArtifact = (groupId: string, artifactId: string, body: UpdateArtifactRequest) =>
    request<Artifact>('PATCH', artifactPath(groupId, artifactId), body);
export const deleteArtifact = (groupId: string, artifactId: string) =>
    request<void>('DELETE', artifactPath(groupId, artifactId));
export const listDeploys = (groupId: string, artifactId: string, page: number, signal?: AbortSignal) =>
    request<DeployPage>('GET', `${artifactPath(groupId, artifactId)}/deploys?page=${page}`, undefined, signal);

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

/** One page of the deploy log, newest first. page starts at 1. */
export const deploysQuery = ({ groupId, artifactId }: ArtifactId, page: number) =>
    queryOptions({
        queryKey: artifactKeys.deploys(groupId, artifactId, page),
        queryFn: ({ signal }) => listDeploys(groupId, artifactId, page, signal),
        // Keep showing the current page while the next one loads so the table doesn't jump
        placeholderData: keepPreviousData,
    });

export function useArtifacts() {
    return useQuery(artifactsQuery());
}

export function useArtifact(id: ArtifactId) {
    return useQuery(artifactQuery(id));
}

export function useDeploys(id: ArtifactId, page: number) {
    return useQuery(deploysQuery(id, page));
}

export function useCreateArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createArtifact,
        onSuccess: () =>
            Promise.all([
                queryClient.invalidateQueries({ queryKey: artifactKeys.list() }),
                queryClient.invalidateQueries({ queryKey: browseKeys.all }),
            ]),
    });
}

export function useUpdateArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ groupId, artifactId, ...body }: UpdateArtifactRequest & ArtifactId) =>
            updateArtifact(groupId, artifactId, body),
        onSuccess: (artifact) => {
            queryClient.setQueryData(artifactKeys.detail(artifact.groupId, artifact.artifactId), artifact);
            return Promise.all([
                queryClient.invalidateQueries({ queryKey: artifactKeys.list() }),
                queryClient.invalidateQueries({ queryKey: browseKeys.all }),
            ]);
        },
    });
}

export function useDeleteArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ groupId, artifactId }: ArtifactId) => deleteArtifact(groupId, artifactId),
        onSuccess: (_, { groupId, artifactId }) => {
            queryClient.removeQueries({ queryKey: artifactKeys.detail(groupId, artifactId) });
            return Promise.all([
                queryClient.invalidateQueries({ queryKey: artifactKeys.list() }),
                queryClient.invalidateQueries({ queryKey: browseKeys.all }),
            ]);
        },
    });
}
