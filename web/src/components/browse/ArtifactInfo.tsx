import { useMe } from '@/api/auth.ts';
import type { Artifact } from '@/api/types.ts';
import { artifactSettingsUrl } from '@/components/artifacts/urls.ts';
import { CodeSnippet } from '@/components/CodeSnippet.tsx';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardAction, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { BookOpen, Globe, Lock, Settings } from 'lucide-react';
import { Link } from 'react-router';

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

export function ArtifactInfo({
    artifact,
    version,
    javadocUrl,
}: {
    artifact: Artifact;
    version?: string;
    javadocUrl?: string;
}) {
    const repositoryUrl = `${window.location.origin}/releases`;
    const { data: me } = useMe();

    return (
        <Card>
            <CardHeader>
                <CardTitle className="flex flex-wrap items-center gap-2 font-mono">
                    <span>
                        <span className="text-muted-foreground">{artifact.groupId}:</span>
                        {artifact.artifactId}
                        {version && <span className="text-muted-foreground">:{version}</span>}
                    </span>
                    {artifact.public ? (
                        <Badge variant="secondary" className="font-sans">
                            <Globe />
                            Public
                        </Badge>
                    ) : (
                        <Badge variant="outline" className="font-sans">
                            <Lock />
                            Private
                        </Badge>
                    )}
                </CardTitle>
                {artifact.description && <CardDescription>{artifact.description}</CardDescription>}
                {(javadocUrl || me?.isAdmin) && (
                    <CardAction className="flex gap-2">
                        {javadocUrl && (
                            <Button variant="outline" size="sm" render={<Link to={javadocUrl} />}>
                                <BookOpen />
                                Javadoc
                            </Button>
                        )}
                        {me?.isAdmin && (
                            <Button variant="outline" size="sm" render={<Link to={artifactSettingsUrl(artifact)} />}>
                                <Settings />
                                Settings
                            </Button>
                        )}
                    </CardAction>
                )}
            </CardHeader>
            <CardContent>
                {version ? (
                    <Tabs defaultValue="maven">
                        <TabsList>
                            {snippets(artifact, version, repositoryUrl).map((s) => (
                                <TabsTrigger key={s.id} value={s.id}>
                                    {s.label}
                                </TabsTrigger>
                            ))}
                        </TabsList>
                        {snippets(artifact, version, repositoryUrl).map((s) => (
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
                ) : (
                    <p className="text-sm text-muted-foreground">No version has been published yet.</p>
                )}
            </CardContent>
        </Card>
    );
}
