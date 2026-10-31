export interface About {
    status: string;
    version: string;
}

export interface User {
    id: number;
    username: string;
    isAdmin: boolean;
    hasPassword: boolean;
    createdAt: string;
}

export interface UserDetails extends User {
    teams: Team[];
    grants: Grant[];
}

export interface Team {
    id: number;
    name: string;
    createdAt: string;
}

export interface TeamDetails extends Team {
    members: User[];
    grants: Grant[];
}

export type Permission = 'read' | 'write';

export interface Grant {
    id: number;
    user?: string;
    team?: string;
    pattern: string;
    /** write implies read. */
    permission: Permission;
    createdAt: string;
}

export interface Token {
    id: number;
    userId: number;
    name: string;
    prefix: string;
    createdAt: string;
    expiresAt: string | null;
    lastUsedAt: string | null;
}

export interface CreatedToken extends Token {
    token: string;
}

export interface Artifact {
    id: number;
    groupId: string;
    artifactId: string;
    path: string;
    public: boolean;
    description: string | null;
    createdAt: string;
}

// Request bodies

export interface LoginRequest {
    username: string;
    password: string;
}

export interface ChangePasswordRequest {
    currentPassword?: string;
    newPassword: string;
}

export interface CreateUserRequest {
    username: string;
    isAdmin?: boolean;
    password?: string;
}

export interface UpdateUserRequest {
    isAdmin?: boolean;
    password?: string;
}

export interface CreateTokenRequest {
    name: string;
    expiresAt: string | null;
}

export interface CreateTeamRequest {
    name: string;
}

export type CreateGrantRequest = { pattern: string; permission: Permission } & (
    { user: string; team?: never } | { team: string; user?: never }
);

export interface GrantFilter {
    user?: string;
    team?: string;
}

export interface CreateArtifactRequest {
    groupId: string;
    artifactId: string;
    public?: boolean;
    description?: string | null;
}

export interface UpdateArtifactRequest {
    public?: boolean;
    description?: string;
}

export type BrowseEntryType = 'directory' | 'file';

export interface BrowseEntry {
    name: string;
    type: BrowseEntryType;
    /** Only set for files. */
    size?: number;
    /** Only set for files. */
    modifiedAt?: string;
    /** Set when the directory is the root of a registered artifact. */
    artifact?: Artifact;
}

/** A directory of the repository, containing only what the caller may read. */
export interface BrowseListing {
    /** The normalized path without leading or trailing slashes, "" for the repository root. */
    path: string;
    /** The artifact the path is inside of, if any. */
    artifact?: Artifact;
    entries: BrowseEntry[];
}
