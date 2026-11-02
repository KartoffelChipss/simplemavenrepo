export interface About {
    status: string;
    version?: string;
    goVersion?: string;
    platform?: string;
    startedAt?: string;
}

/** Server-wide settings, null when not set. */
export interface Settings {
    name: string | null;
    /** Markdown shown on the home page. */
    description: string | null;
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
    scopes: TokenScope[];
}

export interface TokenScope {
    pattern: string;
    permission: Permission;
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
    allowReleaseOverwrite: boolean;
    description: string | null;
    createdAt: string;
}

/** An artifact with the disk space its files take, not counting artifacts nested inside its directory. */
export interface ArtifactDetails extends Artifact {
    /** Size in bytes. */
    size: number;
}

export interface StorageStats {
    totalSize: number;
    fileCount: number;
    artifactCount: number;
    unregisteredSize: number;
    largestArtifacts: ArtifactDetails[];
}

/** Length of a download period in days, ending today. */
export type DownloadDays = 7 | 30 | 90 | 365;

/** Downloads on one UTC day. */
export interface DayDownloads {
    /** Formatted as 2026-01-31. */
    date: string;
    count: number;
}

/** Downloads count fetches of the main file of a version, not of POMs, checksums, sources or javadoc. */
export interface ArtifactDownloads {
    /** All time. */
    total: number;
    periodTotal: number;
    days: DownloadDays;
    /** Every day of the period, oldest first, including days without downloads. */
    daily: DayDownloads[];
    /** Downloads in the period per version, most downloaded first. */
    versions: { version: string; count: number }[];
}

export interface DownloadStats {
    /** All time, including artifacts that are no longer registered. */
    total: number;
    periodTotal: number;
    days: DownloadDays;
    daily: DayDownloads[];
    /** Most downloaded registered artifacts in the period. */
    topArtifacts: (Artifact & { downloads: number })[];
}

export interface Deploy {
    id: number;
    version: string;
    file: string;
    username: string;
    status: number;
    success: boolean;
    error: string | null;
    size: number | null;
    createdAt: string;
}

export interface DeployPage {
    deploys: Deploy[];
    total: number;
    /** Starts at 1. */
    page: number;
    pageSize: number;
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
    scopes?: TokenScope[];
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
    allowReleaseOverwrite?: boolean;
    description?: string | null;
}

export interface UpdateArtifactRequest {
    public?: boolean;
    allowReleaseOverwrite?: boolean;
    description?: string;
}

/** An empty string clears a setting, a missing field leaves it unchanged. */
export interface UpdateSettingsRequest {
    name?: string;
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

export interface JavadocPackage {
    name: string;
    /** Path of the page in the javadoc jar. */
    url: string;
}

export interface JavadocType {
    package: string;
    /** Simple name, nested types are qualified by their enclosing type, e.g. Map.Entry. */
    name: string;
    url: string;
}

export interface JavadocMember {
    package: string;
    type: string;
    /** Signature as javadoc shows it, e.g. get(int). */
    name: string;
    /** Page and anchor in the javadoc jar. */
    url: string;
}

/** An artifact that has javadoc. */
export interface JavadocArtifact {
    artifact: Artifact;
    /** Versions with javadoc, newest first. Never empty. */
    versions: string[];
}

/** The contents of the javadoc jar of an artifact version. */
export interface Javadoc {
    artifact: Artifact;
    version: string;
    /** Versions of the artifact that have javadoc, newest first. */
    versions: string[];
    /** Where the files of the jar are served, ends with a slash. Contains a token that expires after a day or two. */
    baseUrl: string;
    startPage: string;
    /** Pages about the whole documentation, like the overview, tree and index. */
    pages: { title: string; url: string }[];
    packages: JavadocPackage[];
    types: JavadocType[];
    members: JavadocMember[];
}
