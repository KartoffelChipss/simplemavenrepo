import { request, seg } from '@/api/client.ts';
import { grantKeys, meKeys, teamKeys, userKeys } from '@/api/keys.ts';
import type { CreateTeamRequest, Team, TeamDetails } from '@/api/types.ts';
import { queryOptions, useMutation, useQuery, useQueryClient, type QueryClient } from '@tanstack/react-query';

const teamPath = (name: string) => `/teams/${seg(name)}`;
const memberPath = (team: string, username: string) => `${teamPath(team)}/members/${seg(username)}`;

export const listTeams = (signal?: AbortSignal) => request<Team[]>('GET', '/teams', undefined, signal);
export const getTeam = (name: string, signal?: AbortSignal) =>
    request<TeamDetails>('GET', teamPath(name), undefined, signal);
export const createTeam = (body: CreateTeamRequest) => request<Team>('POST', '/teams', body);
export const deleteTeam = (name: string) => request<void>('DELETE', teamPath(name));
export const addTeamMember = (team: string, username: string) => request<void>('PUT', memberPath(team, username));
export const removeTeamMember = (team: string, username: string) => request<void>('DELETE', memberPath(team, username));

export const teamsQuery = () =>
    queryOptions({
        queryKey: teamKeys.list(),
        queryFn: ({ signal }) => listTeams(signal),
    });

export const teamQuery = (name: string) =>
    queryOptions({
        queryKey: teamKeys.detail(name),
        queryFn: ({ signal }) => getTeam(name, signal),
    });

export function useTeams() {
    return useQuery(teamsQuery());
}

export function useTeam(name: string) {
    return useQuery(teamQuery(name));
}

export function useCreateTeam() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: createTeam,
        onSuccess: () => queryClient.invalidateQueries({ queryKey: teamKeys.list() }),
    });
}

function invalidateMembership(queryClient: QueryClient) {
    return Promise.all([
        queryClient.invalidateQueries({ queryKey: teamKeys.all }),
        queryClient.invalidateQueries({ queryKey: userKeys.all }),
        queryClient.invalidateQueries({ queryKey: meKeys.all, exact: true }),
    ]);
}

export function useDeleteTeam() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: deleteTeam,
        onSuccess: (_, name) => {
            queryClient.removeQueries({ queryKey: teamKeys.detail(name) });
            return Promise.all([
                invalidateMembership(queryClient),
                queryClient.invalidateQueries({ queryKey: grantKeys.all }),
            ]);
        },
    });
}

export function useAddTeamMember() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ team, username }: { team: string; username: string }) => addTeamMember(team, username),
        onSuccess: () => invalidateMembership(queryClient),
    });
}

export function useRemoveTeamMember() {
    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: ({ team, username }: { team: string; username: string }) => removeTeamMember(team, username),
        onSuccess: () => invalidateMembership(queryClient),
    });
}
