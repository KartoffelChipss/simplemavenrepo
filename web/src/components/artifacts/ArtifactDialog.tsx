import { useCreateArtifact, useUpdateArtifact } from '@/api/artifacts.ts';
import type { Artifact } from '@/api/types.ts';
import { artifactIdError, groupIdError, MAX_DESCRIPTION_LENGTH } from '@/components/artifacts/validation.ts';
import { Button } from '@/components/ui/button';
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from '@/components/ui/dialog';
import { Field, FieldContent, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { toast } from '@/components/ui/toast';
import { useState, type FormEvent, type ReactElement } from 'react';

function ArtifactForm({ artifact, onDone }: { artifact?: Artifact; onDone: () => void }) {
    const editing = artifact !== undefined;
    const [groupId, setGroupId] = useState(artifact?.groupId ?? '');
    const [artifactId, setArtifactId] = useState(artifact?.artifactId ?? '');
    const [description, setDescription] = useState(artifact?.description ?? '');
    const [isPublic, setIsPublic] = useState(artifact?.public ?? false);

    const create = useCreateArtifact();
    const update = useUpdateArtifact();
    const mutation = editing ? update : create;

    const groupIdInvalid = editing ? undefined : groupIdError(groupId.trim());
    const artifactIdInvalid = editing ? undefined : artifactIdError(artifactId.trim());
    const canSubmit =
        !mutation.isPending && groupId.trim() && artifactId.trim() && !groupIdInvalid && !artifactIdInvalid;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        const coordinates = `${groupId.trim()}:${artifactId.trim()}`;
        const options = {
            onSuccess: () => {
                toast.add({ title: editing ? 'Artifact updated' : 'Artifact registered', description: coordinates });
                onDone();
            },
        };
        if (editing) {
            update.mutate(
                {
                    groupId: artifact.groupId,
                    artifactId: artifact.artifactId,
                    public: isPublic,
                    description: description.trim(),
                },
                options
            );
        } else {
            create.mutate(
                {
                    groupId: groupId.trim(),
                    artifactId: artifactId.trim(),
                    public: isPublic,
                    description: description.trim() || null,
                },
                options
            );
        }
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>{editing ? 'Edit artifact' : 'Register artifact'}</DialogTitle>
                <DialogDescription>
                    {editing
                        ? 'The coordinates of a registered artifact can’t be changed.'
                        : 'Only registered artifacts can be published and downloaded. Users still need a write grant to publish.'}
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                <Field data-invalid={!!groupIdInvalid}>
                    <FieldLabel htmlFor="artifact-group">Group ID</FieldLabel>
                    <Input
                        id="artifact-group"
                        className="font-mono"
                        placeholder="com.example"
                        autoCapitalize="none"
                        spellCheck={false}
                        value={groupId}
                        onChange={(e) => setGroupId(e.target.value)}
                        disabled={editing}
                        aria-invalid={!!groupIdInvalid}
                        required
                        autoFocus={!editing}
                    />
                    {groupIdInvalid && <FieldError>{groupIdInvalid}</FieldError>}
                </Field>
                <Field data-invalid={!!artifactIdInvalid}>
                    <FieldLabel htmlFor="artifact-id">Artifact ID</FieldLabel>
                    <Input
                        id="artifact-id"
                        className="font-mono"
                        placeholder="my-library"
                        autoCapitalize="none"
                        spellCheck={false}
                        value={artifactId}
                        onChange={(e) => setArtifactId(e.target.value)}
                        disabled={editing}
                        aria-invalid={!!artifactIdInvalid}
                        required
                    />
                    {artifactIdInvalid && <FieldError>{artifactIdInvalid}</FieldError>}
                </Field>
                <Field>
                    <FieldLabel htmlFor="artifact-description">Description</FieldLabel>
                    <Textarea
                        id="artifact-description"
                        placeholder="Optional"
                        maxLength={MAX_DESCRIPTION_LENGTH}
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                        autoFocus={editing}
                    />
                </Field>
                <Field orientation="horizontal">
                    <FieldContent>
                        <FieldLabel id="artifact-public-label" htmlFor="artifact-public">
                            Public
                        </FieldLabel>
                        <FieldDescription>
                            Anyone can download it without a token. Publishing still needs one.
                        </FieldDescription>
                    </FieldContent>
                    <Switch id="artifact-public" checked={isPublic} onCheckedChange={setIsPublic} />
                </Field>
                {mutation.error && <FieldError>{mutation.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {mutation.isPending && <Spinner />}
                    {editing ? 'Save' : 'Register'}
                </Button>
            </DialogFooter>
        </form>
    );
}

export function ArtifactDialog({ artifact, trigger }: { artifact?: Artifact; trigger: ReactElement }) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={trigger} />
            <DialogContent className="sm:max-w-md">
                <ArtifactForm artifact={artifact} onDone={() => setOpen(false)} />
            </DialogContent>
        </Dialog>
    );
}
