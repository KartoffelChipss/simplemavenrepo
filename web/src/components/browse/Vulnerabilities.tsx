import type {
    PomDependency,
    SkippedDependency,
    Vulnerability,
    VulnerabilityReport,
    VulnerabilitySeverity,
} from '@/api/types.ts';
import { dependencyKey, type indexReport } from '@/components/browse/vulnerabilityReport.ts';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { formatDate } from '@/lib/format.ts';
import { cn } from '@/lib/utils';
import { ChevronDown, ChevronRight, ShieldAlert, ShieldCheck, ShieldQuestionMark } from 'lucide-react';

const SEVERITIES: VulnerabilitySeverity[] = ['critical', 'high', 'medium', 'low', 'unknown'];

const severityClass: Record<VulnerabilitySeverity, string> = {
    critical: 'bg-red-600 text-white dark:bg-red-500',
    high: 'bg-destructive/10 text-destructive dark:bg-destructive/20',
    medium: 'bg-amber-500/15 text-amber-700 dark:text-amber-400',
    low: 'border-border text-muted-foreground',
    unknown: 'border-border text-muted-foreground',
};

export function SeverityBadge({ severity, count }: { severity: VulnerabilitySeverity; count?: number }) {
    return (
        <Badge variant="outline" className={cn('capitalize', severityClass[severity])}>
            {count !== undefined && <span className="tabular-nums">{count}</span>}
            {severity}
        </Badge>
    );
}

export function VulnerabilityCountBadge({ report }: { report?: VulnerabilityReport }) {
    const all = report?.dependencies.flatMap((d) => d.vulnerabilities) ?? [];
    if (all.length === 0) return null;
    const highest = SEVERITIES.find((severity) => all.some((v) => v.severity === severity)) ?? 'unknown';
    return (
        <Badge
            variant="outline"
            className={cn('tabular-nums', severityClass[highest])}
            aria-label={`${all.length} known vulnerabilities, the most severe is ${highest}`}
        >
            <ShieldAlert />
            {all.length}
        </Badge>
    );
}

function countBySeverity(vulnerabilities: Vulnerability[]) {
    return SEVERITIES.map((severity) => ({
        severity,
        count: vulnerabilities.filter((v) => v.severity === severity).length,
    })).filter((c) => c.count > 0);
}

const skipReasons: Record<SkippedDependency['reason'], string> = {
    hosted: 'Hosted on this server, not checked',
    unresolved: 'The version is a range or unknown, not checked',
};

export function VulnerabilityCell({
    dependency,
    index,
    pending,
    expanded,
    onToggle,
}: {
    dependency: PomDependency;
    index: ReturnType<typeof indexReport>;
    pending: boolean;
    expanded: boolean;
    onToggle: () => void;
}) {
    if (pending) return <Skeleton className="h-5 w-16" />;
    const key = dependencyKey(dependency);
    const vulnerabilities = index.vulnerabilities.get(key);
    if (!vulnerabilities) {
        const reason = index.skipped.get(key);
        if (!reason) return null;
        return (
            <Tooltip>
                <TooltipTrigger render={<span className="inline-flex text-muted-foreground" />}>
                    <ShieldQuestionMark className="size-4" aria-label={skipReasons[reason]} />
                </TooltipTrigger>
                <TooltipContent>{skipReasons[reason]}</TooltipContent>
            </Tooltip>
        );
    }
    if (vulnerabilities.length === 0) {
        return (
            <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
                <ShieldCheck className="size-4" />
                None known
            </span>
        );
    }
    const Chevron = expanded ? ChevronDown : ChevronRight;
    return (
        <button
            type="button"
            onClick={onToggle}
            aria-expanded={expanded}
            className="inline-flex items-center gap-1 rounded-md outline-none focus-visible:ring-[3px] focus-visible:ring-ring/50"
        >
            <Chevron className="size-4 text-muted-foreground" />
            <span className="flex flex-wrap gap-1">
                {countBySeverity(vulnerabilities).map((c) => (
                    <SeverityBadge key={c.severity} severity={c.severity} count={c.count} />
                ))}
            </span>
        </button>
    );
}

export function VulnerabilityList({ vulnerabilities }: { vulnerabilities: Vulnerability[] }) {
    return (
        <ul className="flex flex-col gap-3 py-1">
            {vulnerabilities.map((v) => {
                const cve = v.aliases.filter((a) => a.startsWith('CVE-'));
                return (
                    <li key={v.id} className="flex flex-col gap-1">
                        <div className="flex flex-wrap items-center gap-2">
                            <SeverityBadge severity={v.severity} />
                            <a
                                href={v.url}
                                target="_blank"
                                rel="noreferrer noopener"
                                className="font-mono text-sm hover:underline"
                            >
                                {v.id}
                            </a>
                            {cve.length > 0 && (
                                <span className="font-mono text-xs text-muted-foreground">{cve.join(', ')}</span>
                            )}
                            {v.published && (
                                <span className="text-xs text-muted-foreground">{formatDate(v.published)}</span>
                            )}
                        </div>
                        {v.summary && <p className="text-sm whitespace-normal">{v.summary}</p>}
                        <p className="text-xs whitespace-normal text-muted-foreground">
                            {v.fixed.length > 0 ? (
                                <>
                                    Fixed in <span className="font-mono">{v.fixed.join(', ')}</span>
                                </>
                            ) : (
                                'No fixed version known'
                            )}
                        </p>
                    </li>
                );
            })}
        </ul>
    );
}

export function VulnerabilitySummary({
    report,
    pending,
    error,
}: {
    report?: VulnerabilityReport;
    pending: boolean;
    error: Error | null;
}) {
    if (error) {
        return (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
                <ShieldQuestionMark className="size-4" />
                Vulnerability check failed: {error.message}
            </p>
        );
    }
    if (pending || !report) {
        return (
            <p className="flex items-center gap-2 text-sm text-muted-foreground">
                <ShieldQuestionMark className="size-4" />
                Checking dependencies for known vulnerabilities…
            </p>
        );
    }
    const all = report.dependencies.flatMap((d) => d.vulnerabilities);
    const affected = report.dependencies.filter((d) => d.vulnerabilities.length > 0).length;
    const checked = report.dependencies.length;
    return (
        <div className="flex flex-wrap items-center gap-x-2 gap-y-1 text-sm">
            {all.length > 0 ? (
                <>
                    <ShieldAlert className="size-4 text-destructive" />
                    <span>
                        {all.length} known {all.length === 1 ? 'vulnerability' : 'vulnerabilities'} in {affected}{' '}
                        {affected === 1 ? 'dependency' : 'dependencies'}
                    </span>
                    <span className="flex flex-wrap gap-1">
                        {countBySeverity(all).map((c) => (
                            <SeverityBadge key={c.severity} severity={c.severity} count={c.count} />
                        ))}
                    </span>
                </>
            ) : (
                <>
                    <ShieldCheck className="size-4 text-muted-foreground" />
                    <span>
                        No known vulnerabilities in {checked} direct {checked === 1 ? 'dependency' : 'dependencies'}
                    </span>
                </>
            )}
            <span className="text-muted-foreground">
                · Only direct dependencies are checked, using{' '}
                <a href="https://osv.dev" target="_blank" rel="noreferrer noopener" className="hover:underline">
                    OSV
                </a>
            </span>
        </div>
    );
}
