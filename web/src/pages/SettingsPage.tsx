import { DEFAULT_SERVER_NAME, useSettings, useUpdateSettings } from '@/api/settings.ts';
import type { Settings } from '@/api/types.ts';
import { Markdown } from '@/components/Markdown.tsx';
import { PageHeader } from '@/components/PageHeader.tsx';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldContent, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Skeleton } from '@/components/ui/skeleton';
import { Spinner } from '@/components/ui/spinner';
import { Switch } from '@/components/ui/switch';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Textarea } from '@/components/ui/textarea';
import { toast } from '@/components/ui/toast';
import { useState, type FormEvent } from 'react';
import { useLocation, useSearchParams } from 'react-router';

const MAX_NAME_LENGTH = 64;
const MAX_DESCRIPTION_LENGTH = 20000;
const DEFAULT_OSV_URL = 'https://api.osv.dev';

const TABS = [
    { value: 'general', label: 'General' },
    { value: 'vulnerabilities', label: 'Vulnerabilities' },
] as const;
type Tab = (typeof TABS)[number]['value'];

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

function VulnerabilitySettingsForm({ settings }: { settings: Settings }) {
    const [enabled, setEnabled] = useState(settings.osvEnabled);
    const [url, setUrl] = useState(settings.osvUrl ?? '');
    const update = useUpdateSettings();

    const normalizedUrl = url.trim().replace(/\/+$/, '');
    const dirty = enabled !== settings.osvEnabled || normalizedUrl !== (settings.osvUrl ?? '');

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        update.mutate(
            { osvEnabled: enabled, osvUrl: normalizedUrl },
            { onSuccess: () => toast.add({ title: 'Settings saved' }) }
        );
    }

    return (
        <Card>
            <form onSubmit={onSubmit} className="contents">
                <CardHeader>
                    <CardTitle>Vulnerability checks</CardTitle>
                    <CardDescription>
                        Check the direct dependencies of every version for known vulnerabilities using{' '}
                        <a href="https://osv.dev" target="_blank" rel="noreferrer" className="underline">
                            OSV
                        </a>
                        .
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <FieldGroup>
                        <Field orientation="horizontal">
                            <FieldContent>
                                <FieldLabel htmlFor="settings-osv-enabled">Check dependencies</FieldLabel>
                                <FieldDescription>
                                    The server sends the coordinates and versions of dependencies to the OSV API.
                                    Dependencies hosted on this server are never sent.
                                </FieldDescription>
                            </FieldContent>
                            <Switch id="settings-osv-enabled" checked={enabled} onCheckedChange={setEnabled} />
                        </Field>
                        <Field>
                            <FieldLabel htmlFor="settings-osv-url">OSV API URL</FieldLabel>
                            <Input
                                id="settings-osv-url"
                                type="url"
                                placeholder={DEFAULT_OSV_URL}
                                value={url}
                                onChange={(e) => setUrl(e.target.value)}
                            />
                            <FieldDescription>
                                Leave empty to use the public API, or point it at a mirror that serves the same API.
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
    const location = useLocation();
    const [searchParams, setSearchParams] = useSearchParams();
    const tabParam = searchParams.get('tab');
    const tab: Tab = TABS.find((t) => t.value === tabParam)?.value ?? 'general';

    const setTab = (t: string) =>
        setSearchParams(t === 'general' ? {} : { tab: t }, { replace: true, state: location.state as unknown });

    return (
        <>
            <PageHeader title="Settings" description="Server-wide settings." />
            {isPending && <Skeleton className="h-96" />}
            {error && <p className="text-destructive">{error.message}</p>}
            {settings && (
                <Tabs value={tab} onValueChange={setTab} className="gap-4">
                    <TabsList variant="line" className="w-full justify-start border-b">
                        {TABS.map((t) => (
                            <TabsTrigger key={t.value} value={t.value} className="flex-none">
                                {t.label}
                            </TabsTrigger>
                        ))}
                    </TabsList>
                    <TabsContent value="general" keepMounted className="max-w-3xl">
                        <SettingsForm key={`${settings.name}|${settings.description}`} settings={settings} />
                    </TabsContent>
                    <TabsContent value="vulnerabilities" keepMounted className="max-w-3xl">
                        <VulnerabilitySettingsForm
                            key={`${settings.osvEnabled}|${settings.osvUrl}`}
                            settings={settings}
                        />
                    </TabsContent>
                </Tabs>
            )}
        </>
    );
}
