import { useArtifact, useDeleteArtifact, useUpdateArtifact } from '@/api/artifacts.ts';
import { isApiError } from '@/api/client.ts';
import type { Artifact } from '@/api/types.ts';
import { DeployLog } from '@/components/artifacts/DeployLog.tsx';
import { ReadmeBadge } from '@/components/artifacts/ReadmeBadge.tsx';
import { artifactBrowseUrl } from '@/components/artifacts/urls.ts';
import { MAX_DESCRIPTION_LENGTH } from '@/components/artifacts/validation.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Field, FieldContent, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Skeleton } from '@/components/ui/skeleton';
import { Spinner } from '@/components/ui/spinner';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { toast } from '@/components/ui/toast';
import { formatDate } from '@/lib/format.ts';
import { usePageTitle } from '@/title.ts';
import { ArrowLeft, FolderTree } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { Link, useNavigate, useParams } from 'react-router';

const coordinates = (a: Artifact) => `${a.groupId}:${a.artifactId}`;

function SettingsCard({ artifact }: { artifact: Artifact }) {
    const [description, setDescription] = useState(artifact.description ?? '');
    const [isPublic, setIsPublic] = useState(artifact.public);
    const [allowReleaseOverwrite, setAllowReleaseOverwrite] = useState(artifact.allowReleaseOverwrite);
    const update = useUpdateArtifact();

    const dirty =
        description.trim() !== (artifact.description ?? '') ||
        isPublic !== artifact.public ||
        allowReleaseOverwrite !== artifact.allowReleaseOverwrite;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        update.mutate(
            {
                groupId: artifact.groupId,
                artifactId: artifact.artifactId,
                public: isPublic,
                allowReleaseOverwrite,
                description: description.trim(),
            },
            { onSuccess: () => toast.add({ title: 'Artifact updated', description: coordinates(artifact) }) }
        );
    }

    return (
        <Card>
            <form onSubmit={onSubmit} className="contents">
                <CardHeader>
                    <CardTitle>Settings</CardTitle>
                    <CardDescription>The coordinates of a registered artifact can’t be changed.</CardDescription>
                </CardHeader>
                <CardContent>
                    <FieldGroup>
                        <Field>
                            <FieldLabel htmlFor="artifact-description">Description</FieldLabel>
                            <Textarea
                                id="artifact-description"
                                placeholder="Optional"
                                maxLength={MAX_DESCRIPTION_LENGTH}
                                value={description}
                                onChange={(e) => setDescription(e.target.value)}
                            />
                        </Field>
                        <Field orientation="horizontal">
                            <FieldContent>
                                <FieldLabel htmlFor="artifact-public">Public</FieldLabel>
                                <FieldDescription>
                                    Anyone can download it without a token. Publishing still needs one.
                                </FieldDescription>
                            </FieldContent>
                            <Switch id="artifact-public" checked={isPublic} onCheckedChange={setIsPublic} />
                        </Field>
                        <Field orientation="horizontal">
                            <FieldContent>
                                <FieldLabel htmlFor="artifact-overwrite">Allow overwriting releases</FieldLabel>
                                <FieldDescription>
                                    Published release versions can be deployed again. Snapshots can always be replaced.
                                </FieldDescription>
                            </FieldContent>
                            <Switch
                                id="artifact-overwrite"
                                checked={allowReleaseOverwrite}
                                onCheckedChange={setAllowReleaseOverwrite}
                            />
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

function BadgeCard({ artifact }: { artifact: Artifact }) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>README badge</CardTitle>
                <CardDescription>
                    {artifact.public
                        ? 'Shows the latest release and links to the artifact in the repository.'
                        : 'The badge only shows the version to people with read access, so it won’t work in public READMEs until the artifact is public.'}
                </CardDescription>
            </CardHeader>
            <CardContent>
                <ReadmeBadge artifact={artifact} />
            </CardContent>
        </Card>
    );
}

function DeployLogCard({ artifact }: { artifact: Artifact }) {
    return (
        <Card>
            <CardHeader>
                <CardTitle>Deploy log</CardTitle>
                <CardDescription>
                    Every upload of a .jar or .pom file, who made it and whether it succeeded. Checksums and metadata
                    aren’t listed.
                </CardDescription>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
                <DeployLog artifact={artifact} />
            </CardContent>
        </Card>
    );
}

function DeleteArtifactCard({ artifact }: { artifact: Artifact }) {
    const remove = useDeleteArtifact();
    const navigate = useNavigate();
    const [deleteFiles, setDeleteFiles] = useState(false);

    function onDelete() {
        remove.mutate(
            { ...artifact, deleteFiles },
            {
                onSuccess: () => {
                    toast.add({
                        title: deleteFiles ? 'Artifact and files deleted' : 'Artifact deleted',
                        description: coordinates(artifact),
                    });
                    void navigate('/artifacts');
                },
                onError: (err) => toast.add({ title: 'Could not delete artifact', description: err.message }),
            }
        );
    }

    return (
        <Card className="border-destructive/40">
            <CardHeader>
                <CardTitle>Delete artifact</CardTitle>
                <CardDescription>
                    It can no longer be published or downloaded. Files already uploaded stay on disk and become
                    available again if you register the artifact again, unless you choose to delete them too.
                </CardDescription>
            </CardHeader>
            <CardContent>
                <AlertDialog onOpenChange={(open) => open && setDeleteFiles(false)}>
                    <AlertDialogTrigger render={<Button variant="destructive" />}>Delete artifact</AlertDialogTrigger>
                    <AlertDialogContent>
                        <AlertDialogHeader>
                            <AlertDialogTitle>Delete {coordinates(artifact)}?</AlertDialogTitle>
                            <AlertDialogDescription>
                                It can no longer be published or downloaded until it is registered again.
                            </AlertDialogDescription>
                        </AlertDialogHeader>
                        <Field orientation="horizontal">
                            <Checkbox
                                id="artifact-delete-files"
                                checked={deleteFiles}
                                onCheckedChange={setDeleteFiles}
                            />
                            <FieldContent>
                                <FieldLabel htmlFor="artifact-delete-files">Also delete files from disk</FieldLabel>
                                <FieldDescription>
                                    Permanently removes every uploaded version of the artifact. This cannot be undone.
                                </FieldDescription>
                            </FieldContent>
                        </Field>
                        <AlertDialogFooter>
                            <AlertDialogCancel>Cancel</AlertDialogCancel>
                            <AlertDialogAction variant="destructive" disabled={remove.isPending} onClick={onDelete}>
                                {remove.isPending && <Spinner />}
                                {deleteFiles ? 'Delete artifact and files' : 'Delete'}
                            </AlertDialogAction>
                        </AlertDialogFooter>
                    </AlertDialogContent>
                </AlertDialog>
            </CardContent>
        </Card>
    );
}

export function ArtifactPage() {
    const { groupId = '', artifactId = '' } = useParams();
    const { data: artifact, error, isPending } = useArtifact({ groupId, artifactId });
    usePageTitle(`${groupId}:${artifactId}`);

    return (
        <>
            <Link
                to="/artifacts"
                className="flex w-fit items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
            >
                <ArrowLeft className="size-4" />
                Artifacts
            </Link>

            {isPending && (
                <div className="flex flex-col gap-4">
                    <Skeleton className="h-8 w-64" />
                    <Skeleton className="h-60" />
                </div>
            )}

            {isApiError(error, 404) && (
                <PageHeader
                    title="Artifact not found"
                    description={`There is no registered artifact “${groupId}:${artifactId}”.`}
                />
            )}
            {error && !isApiError(error, 404) && <p className="text-destructive">{error.message}</p>}

            {artifact && (
                <>
                    <div className="flex flex-wrap items-start justify-between gap-4">
                        <PageHeader
                            title={coordinates(artifact)}
                            description={`Registered ${formatDate(artifact.createdAt)}`}
                        />
                        <Button variant="outline" render={<Link to={artifactBrowseUrl(artifact)} />}>
                            <FolderTree />
                            Browse files
                        </Button>
                    </div>
                    <div className="grid items-start gap-4 xl:grid-cols-2">
                        <SettingsCard
                            key={`${artifact.description}|${artifact.public}|${artifact.allowReleaseOverwrite}`}
                            artifact={artifact}
                        />
                        <BadgeCard artifact={artifact} />
                    </div>
                    <DeployLogCard artifact={artifact} />
                    <DeleteArtifactCard artifact={artifact} />
                </>
            )}
        </>
    );
}
