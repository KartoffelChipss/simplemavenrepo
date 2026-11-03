import { repositoryFileUrl } from '@/api/browse.ts';
import { javadocPageUrl } from '@/api/javadoc.ts';
import type { BrowseEntry, BrowseListing } from '@/api/types.ts';
import { browseUrl, isJavadocJar, join } from '@/components/browse/paths.ts';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { formatBytes, formatDateTime } from '@/lib/format.ts';
import { cn } from '@/lib/utils';
import { BookOpen, Download, File, Folder, Lock, Package } from 'lucide-react';
import { Link, useNavigate } from 'react-router';

const SIDECAR_EXTENSIONS = ['md5', 'sha1', 'sha256', 'sha512', 'asc'];

interface Row {
    entry: BrowseEntry;
    sidecars: string[];
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

export function EntryTable({ listing }: { listing: BrowseListing }) {
    const navigate = useNavigate();
    const rows = groupSidecars(listing.entries);

    if (rows.length === 0) {
        return <p className="text-sm text-muted-foreground">This directory is empty.</p>;
    }

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
                    {rows.map(({ entry, sidecars }) => {
                        const path = join(listing.path, entry.name);
                        const isDir = entry.type === 'directory';
                        return (
                            <TableRow
                                key={entry.name}
                                className={cn(isDir && 'cursor-pointer')}
                                onClick={isDir ? () => void navigate(browseUrl(path)) : undefined}
                            >
                                <TableCell className="whitespace-normal">
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
                                                download
                                                className="flex items-center gap-1 font-mono text-sm hover:underline"
                                            >
                                                {entry.name}
                                                <Download className="size-3.5 text-muted-foreground" />
                                            </a>
                                        )}
                                        {isJavadocJar(entry) && (
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
                                        <p className="mt-1 ml-6 text-sm text-muted-foreground">
                                            {entry.artifact.description}
                                        </p>
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
                    })}
                </TableBody>
            </Table>
        </div>
    );
}
