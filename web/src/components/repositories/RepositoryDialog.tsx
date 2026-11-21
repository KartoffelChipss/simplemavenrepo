import { useCreateRepository, useUpdateRepository, versionPolicyLabels } from '@/api/repositories.ts';
import type { Repository, VersionPolicy } from '@/api/types.ts';
import { MAX_DESCRIPTION_LENGTH } from '@/components/artifacts/validation.ts';
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
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { Textarea } from '@/components/ui/textarea';
import { toast } from '@/components/ui/toast';
import { useState, type FormEvent, type ReactElement } from 'react';

const NAME_PATTERN = /^[a-z0-9][a-z0-9_-]{0,63}$/;
const RESERVED_NAMES = [
    'api',
    'assets',
    'javadoc',
    'livez',
    'readyz',
    'login',
    'browse',
    'docs',
    'tokens',
    'artifacts',
    'repositories',
    'users',
    'teams',
    'stats',
    'settings',
];

function nameError(name: string): string | undefined {
    if (!name) return undefined;
    if (!NAME_PATTERN.test(name)) {
        return 'Use up to 64 lowercase letters, digits, “_” or “-”, starting with a letter or digit';
    }
    if (RESERVED_NAMES.includes(name)) return `“${name}” is reserved for the server`;
    return undefined;
}

const policyItems = (Object.keys(versionPolicyLabels) as VersionPolicy[]).map((value) => ({
    value,
    label: versionPolicyLabels[value],
}));

const policyDescriptions: Record<VersionPolicy, string> = {
    release: 'Deploying a -SNAPSHOT version is rejected.',
    snapshot: 'Only -SNAPSHOT versions can be deployed.',
    mixed: 'Any version can be deployed.',
};

function RepositoryForm({ repository, onDone }: { repository?: Repository; onDone: () => void }) {
    const editing = repository !== undefined;
    const [name, setName] = useState(repository?.name ?? '');
    const [versionPolicy, setVersionPolicy] = useState<VersionPolicy>(repository?.versionPolicy ?? 'release');
    const [description, setDescription] = useState(repository?.description ?? '');
    const create = useCreateRepository();
    const update = useUpdateRepository();
    const mutation = editing ? update : create;

    const nameInvalid = editing ? undefined : nameError(name.trim());
    const canSubmit = !mutation.isPending && name.trim() && !nameInvalid;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        const onSuccess = () => {
            toast.add({ title: editing ? 'Repository updated' : 'Repository created', description: name.trim() });
            onDone();
        };
        if (editing) {
            update.mutate({ name: repository.name, versionPolicy, description: description.trim() }, { onSuccess });
        } else {
            create.mutate({ name: name.trim(), versionPolicy, description: description.trim() || null }, { onSuccess });
        }
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>{editing ? `Edit ${repository.name}` : 'Create repository'}</DialogTitle>
                <DialogDescription>
                    {editing
                        ? 'The name is part of every URL and permission path, so it can’t be changed.'
                        : 'Its artifacts are served under /<name>. Grants and token scopes start with the name too.'}
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                {!editing && (
                    <Field data-invalid={!!nameInvalid}>
                        <FieldLabel htmlFor="repository-name">Name</FieldLabel>
                        <Input
                            id="repository-name"
                            className="font-mono"
                            placeholder="snapshots"
                            autoCapitalize="none"
                            spellCheck={false}
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            aria-invalid={!!nameInvalid}
                            required
                            autoFocus
                        />
                        {nameInvalid ? (
                            <FieldError>{nameInvalid}</FieldError>
                        ) : (
                            <FieldDescription className="font-mono break-all">
                                {window.location.origin}/{name.trim() || '<name>'}
                            </FieldDescription>
                        )}
                    </Field>
                )}
                <Field>
                    <FieldLabel htmlFor="repository-policy">Versions</FieldLabel>
                    <Select
                        items={policyItems}
                        value={versionPolicy}
                        onValueChange={(v) => setVersionPolicy(v ?? 'release')}
                    >
                        <SelectTrigger id="repository-policy">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {policyItems.map((o) => (
                                <SelectItem key={o.value} value={o.value}>
                                    {o.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                    <FieldDescription>{policyDescriptions[versionPolicy]}</FieldDescription>
                </Field>
                <Field>
                    <FieldLabel htmlFor="repository-description">Description</FieldLabel>
                    <Textarea
                        id="repository-description"
                        placeholder="Optional"
                        maxLength={MAX_DESCRIPTION_LENGTH}
                        value={description}
                        onChange={(e) => setDescription(e.target.value)}
                    />
                </Field>
                {mutation.error && <FieldError>{mutation.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {mutation.isPending && <Spinner />}
                    {editing ? 'Save' : 'Create'}
                </Button>
            </DialogFooter>
        </form>
    );
}

export function RepositoryDialog({ repository, trigger }: { repository?: Repository; trigger: ReactElement }) {
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={trigger} />
            <DialogContent className="sm:max-w-md">
                <RepositoryForm repository={repository} onDone={() => setOpen(false)} />
            </DialogContent>
        </Dialog>
    );
}
