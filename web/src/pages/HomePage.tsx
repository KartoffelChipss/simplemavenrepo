import { useMe } from '@/api/auth.ts';
import { useServerName, useSettings } from '@/api/settings.ts';
import { Markdown } from '@/components/Markdown.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { BookOpen, FolderTree, Settings } from 'lucide-react';
import { Link } from 'react-router';

export function HomePage() {
    const { data: settings, error, isPending } = useSettings();
    const { data: me } = useMe();
    const serverName = useServerName();

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader title={serverName} description="Maven repository" />
                <div className="flex flex-wrap gap-2">
                    <Button variant="outline" render={<Link to="/browse" />}>
                        <FolderTree />
                        Browse
                    </Button>
                    <Button variant="outline" render={<Link to="/docs" />}>
                        <BookOpen />
                        Javadoc
                    </Button>
                </div>
            </div>

            {isPending && <Skeleton className="h-40" />}
            {error && <p className="text-destructive">{error.message}</p>}

            {settings?.description && (
                <Card>
                    <CardContent>
                        <Markdown>{settings.description}</Markdown>
                    </CardContent>
                </Card>
            )}

            {settings && !settings.description && me?.isAdmin && (
                <Card>
                    <CardContent className="flex flex-wrap items-center justify-between gap-4">
                        <p className="text-muted-foreground">
                            Add a description in the settings to tell visitors what this repository hosts and how to use
                            it.
                        </p>
                        <Button variant="outline" render={<Link to="/settings" />}>
                            <Settings />
                            Settings
                        </Button>
                    </CardContent>
                </Card>
            )}
        </>
    );
}
