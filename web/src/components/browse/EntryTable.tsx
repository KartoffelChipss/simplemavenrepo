import { repositoryFileUrl } from '@/api/browse.ts';
import { javadocPageUrl } from '@/api/javadoc.ts';
import type { BrowseEntry, BrowseListing } from '@/api/types.ts';
import { browseUrl, isJavadocJar, join } from '@/components/browse/paths.ts';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatBytes, formatDateTime } from '@/lib/format.ts';
import { cn } from '@/lib/utils';
import { BookOpen, ChevronRight, Download, File, Folder, Lock, Package } from 'lucide-react';
import { useState } from 'react';
import { Link, useNavigate } from 'react-router';

const SIDECAR_EXTENSIONS = ['md5', 'sha1', 'sha256', 'sha512', 'asc'];

interface Row {
    entry: BrowseEntry;
    sidecars: string[];
    label?: string;
}

function groupSidecars(entries: BrowseEntry[]): Row[] {
    const files = new Set(entries.filter((e) => e.type === 'file').map((e) => e.name));
    const isSidecar = (name: string) =>
        SIDECAR_EXTENSIONS.some((ext) => name.endsWith(`.${ext}`) && files.has(name.slice(0, -ext.length - 1)));
    return entries
        .filter((e) => e.type === 'directory' || !isSidecar(e.name))
        .map((entry) => ({
            entry,
            sidecars:
                entry.type === 'file' ? SIDECAR_EXTENSIONS.filter((ext) => files.has(`${entry.name}.${ext}`)) : [],
        }));
}

function EntryIcon({ entry }: { entry: BrowseEntry }) {
    const className = 'size-4 shrink-0 text-muted-foreground';
    if (entry.artifact) return <Package className={className} />;
    if (entry.type === 'directory') return <Folder className={className} />;
    return <File className={className} />;
}

interface Build {
    /** e.g. 20261121.112050-3, the part of the file names that identifies the build. */
    id: string;
    number: number;
    /** ISO timestamp of the deploy. */
    deployedAt: string;
    rows: Row[];
}

function groupBuilds(listing: BrowseListing, rows: Row[]): { rest: Row[]; builds: Build[] } | undefined {
    const version = listing.path.split('/').at(-1) ?? '';
    const artifact = listing.artifact;
    if (!artifact || !version.endsWith('-SNAPSHOT') || listing.path !== join(artifact.path, version)) return undefined;

    const prefix = `${artifact.artifactId}-${version.slice(0, -'SNAPSHOT'.length)}`;
    const builds = new Map<string, Build>();
    const rest: Row[] = [];
    for (const row of rows) {
        const match =
            row.entry.type === 'file' && row.entry.name.startsWith(prefix)
                ? /^(\d{4})(\d{2})(\d{2})\.(\d{2})(\d{2})(\d{2})-(\d+)(?=[-.])/.exec(
                      row.entry.name.slice(prefix.length)
                  )
                : null;
        if (!match) {
            rest.push(row);
            continue;
        }
        const [stamp, year, month, day, hour, minute, second, number] = match;
        let build = builds.get(stamp);
        if (!build) {
            build = {
                id: stamp,
                number: Number(number),
                deployedAt: `${year}-${month}-${day}T${hour}:${minute}:${second}Z`,
                rows: [],
            };
            builds.set(stamp, build);
        }
        build.rows.push({ ...row, label: `…${row.entry.name.slice(prefix.length + match[0].length)}` });
    }
    if (builds.size === 0) return undefined;
    return { rest, builds: [...builds.values()].sort((a, b) => b.number - a.number || b.id.localeCompare(a.id)) };
}

function EntryRow({
    listing,
    row: { entry, sidecars, label },
    showJavadoc = true,
    nested = false,
}: {
    listing: BrowseListing;
    row: Row;
    showJavadoc?: boolean;
    nested?: boolean;
}) {
    const navigate = useNavigate();
    const path = join(listing.path, entry.name);
    const isDir = entry.type === 'directory';
    return (
        <TableRow
            className={cn(isDir && 'cursor-pointer')}
            onClick={isDir ? () => void navigate(browseUrl(path)) : undefined}
        >
            <TableCell className={cn('whitespace-normal', nested && 'pl-9')}>
                <div className="flex items-center gap-2">
                    <EntryIcon entry={entry} />
                    {isDir ? (
                        <Link
                            to={browseUrl(path)}
                            className="font-mono text-sm hover:underline"
                            onClick={(e) => e.stopPropagation()}
                        >
                            {entry.name}
                        </Link>
                    ) : (
                        <a
                            href={repositoryFileUrl(path)}
                            download={entry.name}
                            title={label ? entry.name : undefined}
                            className="flex items-center gap-1 font-mono text-sm hover:underline"
                        >
                            {label ?? entry.name}
                            <Download className="size-3.5 text-muted-foreground" />
                        </a>
                    )}
                    {showJavadoc && isJavadocJar(entry) && (
                        <Link
                            to={javadocPageUrl(listing.path)}
                            className="flex items-center gap-1 rounded border px-1 text-[0.7rem] text-muted-foreground hover:text-foreground"
                        >
                            <BookOpen className="size-3" />
                            View
                        </Link>
                    )}
                    {entry.artifact && !entry.artifact.public && (
                        <Badge variant="outline">
                            <Lock />
                            Private
                        </Badge>
                    )}
                    {sidecars.length > 0 && (
                        <span className="flex gap-1">
                            {sidecars.map((ext) => (
                                <a
                                    key={ext}
                                    href={repositoryFileUrl(`${path}.${ext}`)}
                                    download
                                    className="rounded border px-1 font-mono text-[0.7rem] text-muted-foreground hover:text-foreground"
                                >
                                    {ext}
                                </a>
                            ))}
                        </span>
                    )}
                </div>
                {entry.artifact?.description && (
                    <p className="mt-1 ml-6 text-sm text-muted-foreground">{entry.artifact.description}</p>
                )}
            </TableCell>
            <TableCell className="text-right text-muted-foreground tabular-nums">
                {entry.size !== undefined ? formatBytes(entry.size) : ''}
            </TableCell>
            <TableCell className="hidden text-muted-foreground sm:table-cell">
                {entry.modifiedAt && entry.type === 'file' ? formatDateTime(entry.modifiedAt) : ''}
            </TableCell>
        </TableRow>
    );
}

function BuildRows({ listing, build, latest }: { listing: BrowseListing; build: Build; latest: boolean }) {
    const [open, setOpen] = useState(latest);
    const size = build.rows.reduce((sum, { entry }) => sum + (entry.size ?? 0), 0);
    const files = build.rows.length;

    return (
        <>
            <TableRow className="cursor-pointer bg-muted/40 hover:bg-muted/60" onClick={() => setOpen(!open)}>
                <TableCell>
                    <button
                        type="button"
                        aria-expanded={open}
                        className="flex items-center gap-2 text-sm font-medium"
                        onClick={(e) => {
                            e.stopPropagation();
                            setOpen(!open);
                        }}
                    >
                        <ChevronRight
                            className={cn('size-4 text-muted-foreground transition-transform', open && 'rotate-90')}
                        />
                        Build {build.number}
                        <span className="font-mono text-xs font-normal text-muted-foreground">{build.id}</span>
                        {latest && <Badge variant="secondary">Latest</Badge>}
                        <span className="text-xs font-normal text-muted-foreground">
                            {files} {files === 1 ? 'file' : 'files'}
                        </span>
                    </button>
                </TableCell>
                <TableCell className="text-right text-muted-foreground tabular-nums">{formatBytes(size)}</TableCell>
                <TableCell className="hidden text-muted-foreground sm:table-cell">
                    {formatDateTime(build.deployedAt)}
                </TableCell>
            </TableRow>
            {open &&
                build.rows.map((row) => (
                    <EntryRow key={row.entry.name} listing={listing} row={row} showJavadoc={latest} nested />
                ))}
        </>
    );
}

export function EntryTable({ listing }: { listing: BrowseListing }) {
    const rows = groupSidecars(listing.entries);

    if (rows.length === 0) {
        return <p className="text-sm text-muted-foreground">This directory is empty.</p>;
    }
    const grouped = groupBuilds(listing, rows);

    return (
        <div className="rounded-lg border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>Name</TableHead>
                        <TableHead className="text-right">Size</TableHead>
                        <TableHead className="hidden sm:table-cell">Modified</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {(grouped?.rest ?? rows).map((row) => (
                        <EntryRow key={row.entry.name} listing={listing} row={row} />
                    ))}
                    {grouped?.builds.map((build, i) => (
                        <BuildRows key={build.id} listing={listing} build={build} latest={i === 0} />
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
