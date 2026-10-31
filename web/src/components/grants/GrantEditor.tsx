import { useCreateGrant, useDeleteGrant } from '@/api/grants.ts';
import type { Grant, Permission } from '@/api/types.ts';
import { normalizePattern } from '@/components/grants/pattern.ts';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Field, FieldDescription, FieldError, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { toast } from '@/components/ui/toast';
import { Plus, Trash2 } from 'lucide-react';
import { useState, type FormEvent } from 'react';

export type GrantTarget = { team: string } | { user: string };

const permissionOptions: { value: Permission; label: string }[] = [
    { value: 'read', label: 'Read' },
    { value: 'write', label: 'Read & write' },
];

export function PermissionBadge({ permission }: { permission: Permission }) {
    return permission === 'write' ? <Badge>Read & write</Badge> : <Badge variant="outline">Read</Badge>;
}

function DeleteGrantButton({ grant }: { grant: Grant }) {
    const remove = useDeleteGrant();

    function onDelete() {
        remove.mutate(grant.id, {
            onSuccess: () => toast.add({ title: 'Grant removed', description: grant.pattern }),
            onError: (err) => toast.add({ title: 'Could not remove grant', description: err.message }),
        });
    }

    return (
        <Button
            variant="ghost"
            size="icon-sm"
            aria-label={`Remove grant ${grant.pattern}`}
            disabled={remove.isPending}
            onClick={onDelete}
        >
            <Trash2 />
        </Button>
    );
}

function AddGrantForm({ target }: { target: GrantTarget }) {
    const [pattern, setPattern] = useState('');
    const [permission, setPermission] = useState<Permission>('read');
    const create = useCreateGrant();

    const normalized = pattern.trim() ? normalizePattern(pattern) : undefined;
    const patternInvalid = normalized && 'error' in normalized ? normalized.error : undefined;
    const canSubmit = !create.isPending && normalized && !patternInvalid;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        if (!normalized || 'error' in normalized) return;
        create.mutate(
            { ...target, pattern: normalized.pattern, permission },
            {
                onSuccess: (grant) => {
                    toast.add({ title: 'Grant added', description: grant.pattern });
                    setPattern('');
                },
            }
        );
    }

    return (
        <form onSubmit={onSubmit} className="flex flex-col gap-2">
            <div className="flex flex-wrap items-end gap-2">
                <Field data-invalid={!!patternInvalid} className="min-w-48 flex-1">
                    <FieldLabel htmlFor="grant-pattern">Repository path</FieldLabel>
                    <Input
                        id="grant-pattern"
                        className="font-mono"
                        placeholder="com/example/**"
                        autoCapitalize="none"
                        spellCheck={false}
                        value={pattern}
                        onChange={(e) => setPattern(e.target.value)}
                        aria-invalid={!!patternInvalid}
                    />
                </Field>
                <Field className="w-auto">
                    <FieldLabel htmlFor="grant-permission">Permission</FieldLabel>
                    <Select
                        items={permissionOptions}
                        value={permission}
                        onValueChange={(v) => setPermission(v ?? 'read')}
                    >
                        <SelectTrigger id="grant-permission" className="w-36">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {permissionOptions.map((o) => (
                                <SelectItem key={o.value} value={o.value}>
                                    {o.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </Field>
                <Button type="submit" disabled={!canSubmit}>
                    {create.isPending ? <Spinner /> : <Plus />}
                    Add
                </Button>
            </div>
            {patternInvalid ? (
                <FieldError>{patternInvalid}</FieldError>
            ) : normalized && 'pattern' in normalized && normalized.pattern !== pattern.trim() ? (
                <FieldDescription>
                    Saved as <span className="font-mono">{normalized.pattern}</span>
                </FieldDescription>
            ) : (
                <FieldDescription>
                    A group path covers everything below it. Use <span className="font-mono">**</span> for the whole
                    repository.
                </FieldDescription>
            )}
            {create.error && <FieldError>{create.error.message}</FieldError>}
        </form>
    );
}

export function GrantEditor({ grants, target }: { grants: Grant[]; target: GrantTarget }) {
    return (
        <div className="flex flex-col gap-4">
            {grants.length === 0 ? (
                <p className="text-sm text-muted-foreground">No grants yet.</p>
            ) : (
                <div className="rounded-lg border">
                    <Table>
                        <TableHeader>
                            <TableRow>
                                <TableHead>Repository path</TableHead>
                                <TableHead>Permission</TableHead>
                                <TableHead className="w-0">
                                    <span className="sr-only">Actions</span>
                                </TableHead>
                            </TableRow>
                        </TableHeader>
                        <TableBody>
                            {grants.map((grant) => (
                                <TableRow key={grant.id}>
                                    <TableCell className="font-mono">{grant.pattern}</TableCell>
                                    <TableCell>
                                        <PermissionBadge permission={grant.permission} />
                                    </TableCell>
                                    <TableCell>
                                        <DeleteGrantButton grant={grant} />
                                    </TableCell>
                                </TableRow>
                            ))}
                        </TableBody>
                    </Table>
                </div>
            )}
            <AddGrantForm target={target} />
        </div>
    );
}
