import type { Javadoc } from '@/api/types.ts';
import { search, type ResultKind } from '@/components/javadoc/search.ts';
import { containsPackage, packageTree, type PackageNode, type TypeNode } from '@/components/javadoc/tree.ts';
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible';
import { InputGroup, InputGroupAddon, InputGroupInput } from '@/components/ui/input-group';
import { cn } from '@/lib/utils';
import { Box, ChevronRight, Folder, FolderOpen, Search, SquareFunction } from 'lucide-react';
import { useDeferredValue, useEffect, useMemo, useRef, useState, type ReactNode } from 'react';

const pageOf = (url: string) => url.split('#')[0];

function NavLink({
    url,
    active,
    onOpen,
    className,
    children,
}: {
    url: string;
    active?: boolean;
    onOpen: (url: string) => void;
    className?: string;
    children: ReactNode;
}) {
    return (
        <button
            type="button"
            data-active={active || undefined}
            onClick={() => onOpen(url)}
            className={cn(
                'flex w-full min-w-0 items-center gap-2 rounded-md px-2 py-1 text-left text-sm hover:bg-muted',
                active && 'bg-accent font-medium text-accent-foreground hover:bg-accent',
                className
            )}
        >
            {children}
        </button>
    );
}

const kindIcons: Record<ResultKind, typeof Box> = { package: Folder, type: Box, member: SquareFunction };

/** A row of the tree: a chevron for nodes with children, then the link. */
function TreeRow({
    open,
    onOpenChange,
    label,
    expandable,
    children,
    nested,
}: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    label: string;
    expandable: boolean;
    children: ReactNode;
    nested: ReactNode;
}) {
    return (
        <Collapsible open={expandable && open} onOpenChange={onOpenChange}>
            <div className="flex items-center">
                {expandable ? (
                    <CollapsibleTrigger
                        aria-label={`${open ? 'Collapse' : 'Expand'} ${label}`}
                        className="flex size-6 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted"
                    >
                        <ChevronRight className={cn('size-3.5 transition-transform', open && 'rotate-90')} />
                    </CollapsibleTrigger>
                ) : (
                    <span className="size-6 shrink-0" />
                )}
                {children}
            </div>
            <CollapsibleContent>
                <ul className="ml-3 flex flex-col gap-0.5 border-l pl-1">{nested}</ul>
            </CollapsibleContent>
        </Collapsible>
    );
}

interface TreeState {
    current: string;
    /** The package of the current page. */
    activePackage?: string;
    onOpen: (url: string) => void;
    /** Nodes the user opened or closed, nodes on the way to the current page open by themselves. */
    toggled: Record<string, boolean>;
    setToggled: (key: string, open: boolean) => void;
}

const containsPage = (node: TypeNode, page: string): boolean =>
    node.type.url === page || node.children.some((c) => containsPage(c, page));

function TypeItem({ node, state }: { node: TypeNode; state: TreeState }) {
    const { current, onOpen, toggled, setToggled } = state;
    const key = `type ${node.type.url}`;
    const open = toggled[key] ?? node.children.some((c) => containsPage(c, current));
    return (
        <li>
            <TreeRow
                open={open}
                onOpenChange={(o) => setToggled(key, o)}
                label={node.type.name}
                expandable={node.children.length > 0}
                nested={node.children.map((c) => (
                    <TypeItem key={c.type.url} node={c} state={state} />
                ))}
            >
                <NavLink url={node.type.url} active={node.type.url === current} onOpen={onOpen}>
                    <Box className="size-3.5 shrink-0 text-muted-foreground" />
                    <span className="truncate font-mono" title={`${node.type.package}.${node.type.name}`}>
                        {node.label}
                    </span>
                </NavLink>
            </TreeRow>
        </li>
    );
}

function PackageItem({ node, state, onlyRoot }: { node: PackageNode; state: TreeState; onlyRoot?: boolean }) {
    const { current, activePackage, onOpen, toggled, setToggled } = state;
    const key = `package ${node.name}`;
    const open = toggled[key] ?? (onlyRoot || containsPackage(node, activePackage));
    const expandable = node.children.length + node.types.length > 0;
    const label = (
        <>
            {open && expandable ? (
                <FolderOpen className="size-3.5 shrink-0 text-muted-foreground" />
            ) : (
                <Folder className="size-3.5 shrink-0 text-muted-foreground" />
            )}
            <span className="truncate font-mono" title={node.name}>
                {node.label}
            </span>
        </>
    );
    return (
        <li>
            <TreeRow
                open={open}
                onOpenChange={(o) => setToggled(key, o)}
                label={node.name}
                expandable={expandable}
                nested={
                    <>
                        {node.children.map((c) => (
                            <PackageItem key={c.name} node={c} state={state} />
                        ))}
                        {node.types.map((t) => (
                            <TypeItem key={t.type.url} node={t} state={state} />
                        ))}
                    </>
                }
            >
                {node.url ? (
                    <NavLink url={node.url} active={node.url === current} onOpen={onOpen}>
                        {label}
                    </NavLink>
                ) : (
                    <span className="flex min-w-0 items-center gap-2 px-2 py-1 text-sm text-muted-foreground">
                        {label}
                    </span>
                )}
            </TreeRow>
        </li>
    );
}

function Tree({ javadoc, current, onOpen }: { javadoc: Javadoc; current: string; onOpen: (url: string) => void }) {
    const roots = useMemo(() => packageTree(javadoc), [javadoc]);
    const [toggled, setToggledState] = useState<Record<string, boolean>>({});
    const state: TreeState = {
        current,
        activePackage:
            javadoc.packages.find((p) => p.url === current)?.name ??
            javadoc.types.find((t) => t.url === current)?.package,
        onOpen,
        toggled,
        setToggled: (key, open) => setToggledState((t) => ({ ...t, [key]: open })),
    };

    return (
        <ul className="flex flex-col gap-0.5">
            {roots.map((node) => (
                <PackageItem key={node.name} node={node} state={state} onlyRoot={roots.length === 1} />
            ))}
        </ul>
    );
}

function Results({ javadoc, query, onOpen }: { javadoc: Javadoc; query: string; onOpen: (url: string) => void }) {
    const results = useMemo(() => search(javadoc, query), [javadoc, query]);
    if (results.length === 0) {
        return <p className="px-2 py-4 text-center text-sm text-muted-foreground">Nothing found.</p>;
    }
    return (
        <ul className="flex flex-col gap-0.5">
            {results.map((r) => {
                const Icon = kindIcons[r.kind];
                return (
                    <li key={`${r.kind} ${r.url}`}>
                        <NavLink url={r.url} onOpen={onOpen} className="items-start">
                            <Icon className="mt-0.5 size-3.5 shrink-0 text-muted-foreground" />
                            <span className="flex min-w-0 flex-col">
                                <span className="truncate font-mono" title={r.name}>
                                    {r.name}
                                </span>
                                {r.context && (
                                    <span className="truncate font-mono text-xs text-muted-foreground">
                                        {r.context}
                                    </span>
                                )}
                            </span>
                        </NavLink>
                    </li>
                );
            })}
        </ul>
    );
}

/** Lists the packages and types of the javadoc, or search results while there is a query. */
export function JavadocNav({
    javadoc,
    page,
    onOpen,
}: {
    javadoc: Javadoc;
    page: string;
    onOpen: (url: string) => void;
}) {
    const [query, setQuery] = useState('');
    const deferredQuery = useDeferredValue(query);
    const inputRef = useRef<HTMLInputElement>(null);
    const listRef = useRef<HTMLDivElement>(null);
    const current = pageOf(page);

    // "/" focuses the search like in javadoc itself. Keys typed inside the frame don't reach this page.
    useEffect(() => {
        const onKeyDown = (e: KeyboardEvent) => {
            const target = e.target as HTMLElement;
            if (e.key !== '/' || target.closest('input, textarea, [contenteditable]')) return;
            e.preventDefault();
            inputRef.current?.focus();
        };
        document.addEventListener('keydown', onKeyDown);
        return () => document.removeEventListener('keydown', onKeyDown);
    }, []);

    useEffect(() => {
        if (query) return;
        listRef.current?.querySelector('[data-active]')?.scrollIntoView({ block: 'nearest' });
    }, [current, query]);

    return (
        <div className="flex min-h-0 flex-1 flex-col">
            <div className="border-b p-2">
                <InputGroup>
                    <InputGroupAddon>
                        <Search />
                    </InputGroupAddon>
                    <InputGroupInput
                        ref={inputRef}
                        type="search"
                        placeholder="Search (type /)"
                        aria-label="Search the javadoc"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        onKeyDown={(e) => {
                            if (e.key === 'Escape') setQuery('');
                            if (e.key === 'Enter') {
                                const first = search(javadoc, query)[0];
                                if (first) onOpen(first.url);
                            }
                        }}
                    />
                </InputGroup>
            </div>
            <div ref={listRef} className="min-h-0 flex-1 overflow-y-auto p-2">
                {deferredQuery.trim() ? (
                    <Results javadoc={javadoc} query={deferredQuery} onOpen={onOpen} />
                ) : (
                    <Tree javadoc={javadoc} current={current} onOpen={onOpen} />
                )}
            </div>
        </div>
    );
}
