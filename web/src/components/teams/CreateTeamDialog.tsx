import { useCreateTeam } from '@/api/teams.ts';
import { teamNameError } from '@/components/users/validation.ts';
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
import { Field, FieldError, FieldGroup, FieldLabel } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { toast } from '@/components/ui/toast';
import { Plus } from 'lucide-react';
import { useState, type FormEvent } from 'react';
import { useNavigate } from 'react-router';

function CreateTeamForm() {
    const [name, setName] = useState('');
    const create = useCreateTeam();
    const navigate = useNavigate();

    const nameInvalid = teamNameError(name.trim());
    const canSubmit = !create.isPending && name.trim() && !nameInvalid;

    function onSubmit(e: FormEvent) {
        e.preventDefault();
        create.mutate(
            { name: name.trim() },
            {
                onSuccess: (team) => {
                    toast.add({ title: 'Team created', description: team.name });
                    void navigate(`/teams/${encodeURIComponent(team.name)}`);
                },
            }
        );
    }

    return (
        <form onSubmit={onSubmit}>
            <DialogHeader>
                <DialogTitle>Create team</DialogTitle>
                <DialogDescription>
                    Teams give their members grants together, e.g. read access for everyone in a project.
                </DialogDescription>
            </DialogHeader>
            <FieldGroup className="py-4">
                <Field data-invalid={!!nameInvalid}>
                    <FieldLabel htmlFor="team-name">Name</FieldLabel>
                    <Input
                        id="team-name"
                        autoComplete="off"
                        autoCapitalize="none"
                        spellCheck={false}
                        placeholder="e.g. backend"
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        aria-invalid={!!nameInvalid}
                        required
                        autoFocus
                    />
                    {nameInvalid && <FieldError>{nameInvalid}</FieldError>}
                </Field>
                {create.error && <FieldError>{create.error.message}</FieldError>}
            </FieldGroup>
            <DialogFooter>
                <DialogClose render={<Button variant="outline" />}>Cancel</DialogClose>
                <Button type="submit" disabled={!canSubmit}>
                    {create.isPending && <Spinner />}
                    Create team
                </Button>
            </DialogFooter>
        </form>
    );
}

export function CreateTeamDialog() {
    return (
        <Dialog>
            <DialogTrigger render={<Button />}>
                <Plus />
                Create team
            </DialogTrigger>
            <DialogContent className="sm:max-w-md">
                <CreateTeamForm />
            </DialogContent>
        </Dialog>
    );
}
