import { useCreateArtifact } from '@/api/artifacts.ts';
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

function ArtifactForm({ onDone }: { onDone: () => void }) {
    const [groupId, setGroupId] = useState('');
    const [artifactId, setArtifactId] = useState('');
    const [description, setDescription] = useState('');
    const [isPublic, setIsPublic] = useState(false);
    const [allowReleaseOverwrite, setAllowReleaseOverwrite] = useState(false);

    const create = useCreateArtifact();

    const groupIdInvalid = groupIdError(groupId.trim());
    const artifactIdInvalid = artifactIdError(artifactId.trim());
    const canSubmit = !create.isPending && groupId.trim() && artifactId.trim() && !groupIdInvalid && !artifactIdInvalid;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        const coordinates = `${groupId.trim()}:${artifactId.trim()}`;
        create.mutate(
            {
                groupId: groupId.trim(),
                artifactId: artifactId.trim(),
                public: isPublic,
                allowReleaseOverwrite,
                description: description.trim() || null,
            },
            {
                onSuccess: () => {
                    toast.add({ title: 'Artifact registered', description: coordinates });
                    onDone();
                },
            }
        );
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>Register artifact</DialogTitle>
                <DialogDescription>
                    Only registered artifacts can be published and downloaded. Users still need a write grant to
                    publish.
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
                        aria-invalid={!!groupIdInvalid}
                        required
                        autoFocus
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
                {create.error && <FieldError>{create.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {create.isPending && <Spinner />}
                    Register
                </Button>
            </DialogFooter>
        </form>
    );
}

export function ArtifactDialog({ trigger }: { trigger: ReactElement }) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={trigger} />
            <DialogContent className="sm:max-w-md">
                <ArtifactForm onDone={() => setOpen(false)} />
            </DialogContent>
        </Dialog>
    );
}
