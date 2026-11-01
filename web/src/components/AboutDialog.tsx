import { useAbout } from '@/api/about.ts';
import { useServerName } from '@/api/settings.ts';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/components/ui/dialog';
import { Skeleton } from '@/components/ui/skeleton';
import { formatDateTime } from '@/lib/format.ts';
import { ExternalLink } from 'lucide-react';
import { LogoMark } from '@/components/Logo.tsx';
import type { ReactNode } from 'react';

const REPO_URL = 'https://github.com/KartoffelChipss/simplemavenrepo';

const links = [
    { label: 'Source code', href: REPO_URL },
    { label: 'Report an issue', href: `${REPO_URL}/issues` },
    { label: 'Releases', href: `${REPO_URL}/releases` },
];

function InfoRow({ label, children }: { label: string; children: ReactNode }) {
    return (
        <div className="flex items-baseline justify-between gap-4 py-2 text-sm">
            <dt className="text-muted-foreground">{label}</dt>
            <dd className="truncate text-right font-mono text-xs">{children}</dd>
        </div>
    );
}

function ServerInfo({ isAdmin }: { isAdmin: boolean }) {
    const { data: about, isPending } = useAbout({ enabled: isAdmin });

    if (!isAdmin) return null;
    if (isPending) return <Skeleton className="h-36" />;
    if (!about?.version) return null;

    return (
        <dl className="divide-y rounded-lg border px-3">
            <InfoRow label="Version">{about.version}</InfoRow>
            {about.goVersion && <InfoRow label="Go">{about.goVersion}</InfoRow>}
            {about.platform && <InfoRow label="Platform">{about.platform}</InfoRow>}
            {about.startedAt && <InfoRow label="Running since">{formatDateTime(about.startedAt)}</InfoRow>}
        </dl>
    );
}

export function AboutDialog({
    isAdmin,
    open,
    onOpenChange,
}: {
    isAdmin: boolean;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}) {
    const serverName = useServerName();

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="sm:max-w-md">
                <DialogHeader className="items-center text-center sm:items-center sm:text-center">
                    <LogoMark className="mb-2 size-12" />
                    <DialogTitle>{serverName}</DialogTitle>
                    <DialogDescription>
                        Powered by simplemavenrepo, a simple self-hosted Maven repository.
                    </DialogDescription>
                </DialogHeader>
                <div className="flex flex-col gap-4 py-2">
                    <ServerInfo isAdmin={isAdmin} />
                    <div className="flex flex-wrap justify-center gap-2">
                        {links.map((link) => (
                            <Button
                                key={link.href}
                                variant="outline"
                                size="sm"
                                render={<a href={link.href} target="_blank" rel="noreferrer" />}
                            >
                                {link.label}
                                <ExternalLink />
                            </Button>
                        ))}
                    </div>
                    <p className="text-center text-xs text-muted-foreground">
                        Free software licensed under the{' '}
                        <a
                            href={`${REPO_URL}/blob/main/LICENSE`}
                            target="_blank"
                            rel="noreferrer"
                            className="underline underline-offset-4 hover:text-foreground"
                        >
                            GNU GPL v3
                        </a>
                        .
                    </p>
                </div>
                <DialogFooter>
                    <DialogClose render={<Button variant="outline" />}>Close</DialogClose>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    );
}
