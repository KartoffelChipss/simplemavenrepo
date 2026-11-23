import { request, seg } from '@/api/client.ts';
import { artifactKeys, browseKeys, statsKeys } from '@/api/keys.ts';
import type {
    Artifact,
    ArtifactRef,
    ArtifactDetails,
    ArtifactDownloads,
    CreateArtifactRequest,
    DeployPage,
    DownloadDays,
    UpdateArtifactRequest,
} from '@/api/types.ts';
import { keepPreviousData, queryOptions, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';

const artifactPath = ({ repository, groupId, artifactId }: ArtifactRef) =>
    `/artifacts/${seg(repository)}/${seg(groupId)}/${seg(artifactId)}`;

export const listArtifacts = (signal?: AbortSignal) => request<Artifact[]>('GET', '/artifacts', undefined, signal);
export const getArtifact = (id: ArtifactRef, signal?: AbortSignal) =>
    request<ArtifactDetails>('GET', artifactPath(id), undefined, signal);
export const createArtifact = (body: CreateArtifactRequest) => request<Artifact>('POST', '/artifacts', body);
export const updateArtifact = (id: ArtifactRef, body: UpdateArtifactRequest) =>
    request<ArtifactDetails>('PATCH', artifactPath(id), body);
export const deleteArtifact = (id: ArtifactRef, deleteFiles: boolean) =>
    request<void>('DELETE', `${artifactPath(id)}${deleteFiles ? '?deleteFiles=true' : ''}`);
export const listDeploys = (id: ArtifactRef, page: number, signal?: AbortSignal) =>
    request<DeployPage>('GET', `${artifactPath(id)}/deploys?page=${page}`, undefined, signal);

export const getArtifactDownloads = (id: ArtifactRef, days: DownloadDays, signal?: AbortSignal) =>
    request<ArtifactDownloads>('GET', `${artifactPath(id)}/downloads?days=${days}`, undefined, signal);

export const artifactsQuery = () =>
    queryOptions({
        queryKey: artifactKeys.list(),
        queryFn: ({ signal }) => listArtifacts(signal),
    });

export const artifactQuery = (id: ArtifactRef) =>
    queryOptions({
        queryKey: artifactKeys.detail(id),
        queryFn: ({ signal }) => getArtifact(id, signal),
    });

/** One page of the deploy log, newest first. page starts at 1. */
export const deploysQuery = (id: ArtifactRef, page: number) =>
    queryOptions({
        queryKey: artifactKeys.deploys(id, page),
        queryFn: ({ signal }) => listDeploys(id, page, signal),
        // Keep showing the current page while the next one loads so the table doesn't jump
        placeholderData: keepPreviousData,
    });

export const artifactDownloadsQuery = (id: ArtifactRef, days: DownloadDays) =>
    queryOptions({
        queryKey: artifactKeys.downloads(id, days),
        queryFn: ({ signal }) => getArtifactDownloads(id, days, signal),
        // Keep showing the current period while the next one loads so the chart doesn't jump
        placeholderData: keepPreviousData,
    });

export function useArtifacts() {
    return useQuery(artifactsQuery());
}

export function useArtifact(id: ArtifactRef) {
    return useQuery(artifactQuery(id));
}

export function useDeploys(id: ArtifactRef, page: number) {
    return useQuery(deploysQuery(id, page));
}

export function useArtifactDownloads(id: ArtifactRef, days: DownloadDays) {
    return useQuery(artifactDownloadsQuery(id, days));
}

export function useCreateArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createArtifact,
        onSuccess: () =>
            Promise.all([
                queryClient.invalidateQueries({ queryKey: artifactKeys.list() }),
                queryClient.invalidateQueries({ queryKey: browseKeys.all }),
                queryClient.invalidateQueries({ queryKey: statsKeys.all }),
            ]),
    });
}

export function useUpdateArtifact() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ repository, groupId, artifactId, ...body }: UpdateArtifactRequest & ArtifactRef) =>
            updateArtifact({ repository, groupId, artifactId }, body),
        onSuccess: (artifact) => {
            queryClient.setQueryData(artifactKeys.detail(artifact), artifact);
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
        mutationFn: ({ deleteFiles, ...id }: ArtifactRef & { deleteFiles: boolean }) => deleteArtifact(id, deleteFiles),
        onSuccess: (_, id) => {
            queryClient.removeQueries({ queryKey: artifactKeys.detail(id) });
            return Promise.all([
                queryClient.invalidateQueries({ queryKey: artifactKeys.list() }),
                queryClient.invalidateQueries({ queryKey: browseKeys.all }),
                queryClient.invalidateQueries({ queryKey: statsKeys.all }),
            ]);
        },
    });
}
