import { badgeUrl } from '@/api/browse.ts';
import type { Artifact } from '@/api/types.ts';
import { artifactBrowseUrl } from '@/components/artifacts/urls.ts';
import { CodeSnippet } from '@/components/CodeSnippet.tsx';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

function badgeSnippets(artifact: Artifact) {
    const origin = window.location.origin;
    const image = `${origin}${badgeUrl(artifact.groupId, artifact.artifactId)}`;
    const page = `${origin}${artifactBrowseUrl(artifact)}`;
    return [
        { id: 'markdown', label: 'Markdown', language: 'markdown', code: `[![Maven](${image})](${page})` },
        { id: 'html', label: 'HTML', language: 'html', code: `<a href="${page}"><img src="${image}" alt="Maven"></a>` },
    ];
}

export function ReadmeBadge({ artifact }: { artifact: Artifact }) {
    const snippets = badgeSnippets(artifact);
    return (
        <div className="flex flex-col gap-3">
            <img
                src={badgeUrl(artifact.groupId, artifact.artifactId)}
                alt="Latest version badge"
                height={20}
                className="h-5 w-fit"
            />
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
