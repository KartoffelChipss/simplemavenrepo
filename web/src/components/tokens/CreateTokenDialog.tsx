import { useCreateMyToken } from '@/api/tokens.ts';
import type { CreatedToken } from '@/api/types.ts';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
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
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { toast } from '@/components/ui/toast';
import { Check, Copy, Plus, TriangleAlert } from 'lucide-react';
import { useRef, useState, type FormEvent } from 'react';

const DAY = 24 * 60 * 60 * 1000;

const expiryOptions = [
    { value: '30', label: '30 days' },
    { value: '90', label: '90 days' },
    { value: '365', label: '1 year' },
    { value: 'never', label: 'Never' },
];

function CreateForm({ onCreated }: { onCreated: (token: CreatedToken) => void }) {
    const [name, setName] = useState('');
    const [expiry, setExpiry] = useState('90');
    const create = useCreateMyToken();

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        const expiresAt = expiry === 'never' ? null : new Date(Date.now() + Number(expiry) * DAY).toISOString();
        create.mutate({ name: name.trim(), expiresAt }, { onSuccess: onCreated });
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>Create token</DialogTitle>
                <DialogDescription>
                    Tokens authenticate Maven, Gradle and scripts. Use one as the password in your build tool's
                    credentials.
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                <Field>
                    <FieldLabel htmlFor="token-name">Name</FieldLabel>
                    <Input
                        id="token-name"
                        placeholder="e.g. CI publishing"
                        maxLength={100}
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        required
                        autoFocus
                    />
                    <FieldDescription>Helps you recognize the token later.</FieldDescription>
                </Field>
                <Field>
                    <FieldLabel htmlFor="token-expiry">Expires after</FieldLabel>
                    <Select items={expiryOptions} value={expiry} onValueChange={(v) => setExpiry(v ?? '90')}>
                        <SelectTrigger id="token-expiry" className="w-full">
                            <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                            {expiryOptions.map((o) => (
                                <SelectItem key={o.value} value={o.value}>
                                    {o.label}
                                </SelectItem>
                            ))}
                        </SelectContent>
                    </Select>
                </Field>
                {create.error && <FieldError>{create.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={create.isPending || !name.trim()}>
                    {create.isPending && <Spinner />}
                    Create token
                </Button>
            </DialogFooter>
        </form>
    );
}

function CreatedTokenView({ token }: { token: CreatedToken }) {
    const [copied, setCopied] = useState(false);

    const input = useRef<HTMLInputElement>(null);

    async function copy() {
        try {
            await navigator.clipboard.writeText(token.token);
            setCopied(true);
            return;
        } catch {
            // clipboard not available
        }
        input.current?.select();
        if (document.execCommand('copy')) {
            setCopied(true);
        } else {
            toast.add({ title: 'Could not copy', description: 'The token is selected, copy it manually.' });
        }
    }

    return (
        <>
            <DialogHeader>
                <DialogTitle>Token created</DialogTitle>
                <DialogDescription>
                    <span className="font-medium text-foreground">{token.name}</span> is ready to use.
                </DialogDescription>
            </DialogHeader>
            <div className="flex flex-col gap-4 py-4">
                <Alert>
                    <TriangleAlert />
                    <AlertTitle>Copy it now</AlertTitle>
                    <AlertDescription>The token is only shown once. If you lose it, create a new one.</AlertDescription>
                </Alert>
                <InputGroup>
                    <InputGroupInput
                        ref={input}
                        aria-label="Token"
                        readOnly
                        value={token.token}
                        className="font-mono"
                        onFocus={(e) => e.currentTarget.select()}
                    />
                    <InputGroupAddon align="inline-end">
                        <InputGroupButton aria-label="Copy token" size="icon-xs" onClick={() => void copy()}>
                            {copied ? <Check /> : <Copy />}
                        </InputGroupButton>
                    </InputGroupAddon>
                </InputGroup>
            </div>
            <DialogFooter>
                <DialogClose render={<Button />}>Done</DialogClose>
            </DialogFooter>
        </>
    );
}

export function CreateTokenDialog() {
    const [open, setOpen] = useState(false);
    const [created, setCreated] = useState<CreatedToken | null>(null);

    return (
        <Dialog open={open} onOpenChange={setOpen} onOpenChangeComplete={(isOpen) => !isOpen && setCreated(null)}>
            <DialogTrigger render={<Button />}>
                <Plus />
                Create token
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
                {created ? <CreatedTokenView token={created} /> : <CreateForm onCreated={setCreated} />}
            </DialogContent>
        </Dialog>
    );
}
