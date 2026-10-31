import { DEFAULT_SERVER_NAME, useSettings, useUpdateSettings } from '@/api/settings.ts';
import type { Settings } from '@/api/types.ts';
import { Markdown } from '@/components/Markdown.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { Spinner } from '@/components/ui/spinner';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { toast } from '@/components/ui/toast';
import { useState, type FormEvent } from 'react';

const MAX_NAME_LENGTH = 64;
const MAX_DESCRIPTION_LENGTH = 20000;

function SettingsForm({ settings }: { settings: Settings }) {
    const [name, setName] = useState(settings.name ?? '');
    const [description, setDescription] = useState(settings.description ?? '');
    const update = useUpdateSettings();

    const dirty = name.trim() !== (settings.name ?? '') || description.trim() !== (settings.description ?? '');

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        update.mutate(
            { name: name.trim(), description: description.trim() },
            { onSuccess: () => toast.add({ title: 'Settings saved' }) }
        );
    }

    return (
        <Card>
            <form onSubmit={onSubmit} className="contents">
                <CardHeader>
                    <CardTitle>General</CardTitle>
                    <CardDescription>How the server presents itself to everyone who opens it.</CardDescription>
                </CardHeader>
                <CardContent>
                    <FieldGroup>
                        <Field>
                            <FieldLabel htmlFor="settings-name">Name</FieldLabel>
                            <Input
                                id="settings-name"
                                placeholder={DEFAULT_SERVER_NAME}
                                maxLength={MAX_NAME_LENGTH}
                                value={name}
                                onChange={(e) => setName(e.target.value)}
                            />
                            <FieldDescription>
                                Shown in the sidebar, the browser tab and on the home page.
                            </FieldDescription>
                        </Field>
                        <Field>
                            <FieldLabel htmlFor="settings-description">Description</FieldLabel>
                            <Tabs defaultValue="write">
                                <TabsList>
                                    <TabsTrigger value="write">Write</TabsTrigger>
                                    <TabsTrigger value="preview">Preview</TabsTrigger>
                                </TabsList>
                                <TabsContent value="write">
                                    <Textarea
                                        id="settings-description"
                                        className="min-h-64 font-mono text-sm"
                                        placeholder="Optional"
                                        maxLength={MAX_DESCRIPTION_LENGTH}
                                        value={description}
                                        onChange={(e) => setDescription(e.target.value)}
                                    />
                                </TabsContent>
                                <TabsContent value="preview">
                                    <div className="min-h-64 rounded-md border px-4 py-3">
                                        {description.trim() ? (
                                            <Markdown>{description}</Markdown>
                                        ) : (
                                            <p className="text-sm text-muted-foreground">Nothing to preview.</p>
                                        )}
                                    </div>
                                </TabsContent>
                            </Tabs>
                            <FieldDescription>
                                Shown on the home page. Supports GitHub flavored markdown, raw HTML is not rendered.
                            </FieldDescription>
                        </Field>
                        {update.error && <FieldError>{update.error.message}</FieldError>}
                    </FieldGroup>
                </CardContent>
                <CardFooter>
                    <Button type="submit" disabled={!dirty || update.isPending}>
                        {update.isPending && <Spinner />}
                        Save
                    </Button>
                </CardFooter>
            </form>
        </Card>
    );
}

export function SettingsPage() {
    const { data: settings, error, isPending } = useSettings();

    return (
        <>
            <PageHeader title="Settings" description="Server-wide settings." />
            {isPending && <Skeleton className="h-96" />}
            {error && <p className="text-destructive">{error.message}</p>}
            {settings && <SettingsForm key={`${settings.name}|${settings.description}`} settings={settings} />}
        </>
    );
}
