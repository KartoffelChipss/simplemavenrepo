import { badgeUrl, downloadsBadgeUrl, type BadgePeriod } from '@/api/browse.ts';
import type { Artifact } from '@/api/types.ts';
import { artifactBrowseUrl } from '@/components/artifacts/urls.ts';
import { CodeSnippet } from '@/components/CodeSnippet.tsx';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useState } from 'react';

type BadgeKind = 'version' | `downloads-${BadgePeriod}`;

const badgeKinds: { value: BadgeKind; label: string; alt: string }[] = [
    { value: 'version', label: 'Latest version', alt: 'Maven' },
    { value: 'downloads-week', label: 'Downloads per week', alt: 'Downloads' },
    { value: 'downloads-month', label: 'Downloads per month', alt: 'Downloads' },
    { value: 'downloads-year', label: 'Downloads per year', alt: 'Downloads' },
    { value: 'downloads-all', label: 'Total downloads', alt: 'Downloads' },
];

function badgePath(artifact: Artifact, kind: BadgeKind) {
    if (kind === 'version') return badgeUrl(artifact.groupId, artifact.artifactId);
    return downloadsBadgeUrl(artifact.groupId, artifact.artifactId, kind.slice('downloads-'.length) as BadgePeriod);
}

function badgeSnippets(artifact: Artifact, path: string, alt: string) {
    const origin = window.location.origin;
    const image = `${origin}${path}`;
    const page = `${origin}${artifactBrowseUrl(artifact)}`;
    return [
        { id: 'markdown', label: 'Markdown', language: 'markdown', code: `[![${alt}](${image})](${page})` },
        {
            id: 'html',
            label: 'HTML',
            language: 'html',
            code: `<a href="${page}"><img src="${image}" alt="${alt}"></a>`,
        },
    ];
}

export function ReadmeBadge({ artifact }: { artifact: Artifact }) {
    const [kind, setKind] = useState<BadgeKind>('version');
    const selected = badgeKinds.find((k) => k.value === kind) ?? badgeKinds[0];
    const path = badgePath(artifact, kind);
    const snippets = badgeSnippets(artifact, path, selected.alt);

    return (
        <div className="flex flex-col gap-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
                <img src={path} alt={`${selected.label} badge`} height={20} className="h-5 w-fit" />
                <Select items={badgeKinds} value={kind} onValueChange={(value) => value && setKind(value)}>
                    <SelectTrigger aria-label="Badge" className="w-48">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {badgeKinds.map((k) => (
                            <SelectItem key={k.value} value={k.value}>
                                {k.label}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>
            <Tabs defaultValue="markdown">
                <TabsList>
                    {snippets.map((s) => (
                        <TabsTrigger key={s.id} value={s.id}>
                            {s.label}
                        </TabsTrigger>
                    ))}
                </TabsList>
                {snippets.map((s) => (
                    <TabsContent key={s.id} value={s.id}>
                        <CodeSnippet code={s.code} language={s.language} label={`${s.label} badge`} />
                    </TabsContent>
                ))}
            </Tabs>
        </div>
    );
}
