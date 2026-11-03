import type { Artifact } from '@/api/types.ts';
import { CodeSnippet } from '@/components/CodeSnippet.tsx';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

function snippets(artifact: Artifact, version: string, repositoryUrl: string) {
    const { groupId, artifactId } = artifact;
    return [
        {
            id: 'maven',
            label: 'Maven',
            language: 'xml',
            code: `<repositories>
  <repository>
    <id>smr</id>
    <url>${repositoryUrl}</url>
  </repository>
</repositories>

<dependency>
  <groupId>${groupId}</groupId>
  <artifactId>${artifactId}</artifactId>
  <version>${version}</version>
</dependency>`,
        },
        {
            id: 'gradle-kotlin',
            label: 'Gradle (Kotlin)',
            language: 'kotlin',
            code: `repositories {
    maven("${repositoryUrl}")
}

dependencies {
    implementation("${groupId}:${artifactId}:${version}")
}`,
        },
        {
            id: 'gradle-groovy',
            label: 'Gradle (Groovy)',
            language: 'groovy',
            code: `repositories {
    maven { url '${repositoryUrl}' }
}

dependencies {
    implementation '${groupId}:${artifactId}:${version}'
}`,
        },
    ];
}

export function UsageSnippets({ artifact, version }: { artifact: Artifact; version: string }) {
    const items = snippets(artifact, version, `${window.location.origin}/releases`);

    return (
        <Tabs defaultValue="maven">
            <TabsList>
                {items.map((s) => (
                    <TabsTrigger key={s.id} value={s.id}>
                        {s.label}
                    </TabsTrigger>
                ))}
            </TabsList>
            {items.map((s) => (
                <TabsContent key={s.id} value={s.id}>
                    <CodeSnippet code={s.code} language={s.language} label={s.label} />
                </TabsContent>
            ))}
            {!artifact.public && (
                <p className="mt-2 text-sm text-muted-foreground">
                    This artifact is private: your build tool needs a username and token with read access.
                </p>
            )}
        </Tabs>
    );
}
