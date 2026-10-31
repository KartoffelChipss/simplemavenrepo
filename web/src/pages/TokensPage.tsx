import { useMyTokens } from '@/api/tokens.ts';
import { PageHeader } from '@/components/PageHeader.tsx';
import { CreateTokenDialog } from '@/components/tokens/CreateTokenDialog.tsx';
import { TokenTable } from '@/components/tokens/TokenTable.tsx';

export function TokensPage() {
    const { data: tokens, error, isPending } = useMyTokens();

    return (
        <>
            <div className="flex flex-wrap items-start justify-between gap-4">
                <PageHeader
                    title="Tokens"
                    description="Tokens let Maven, Gradle and scripts access the repository as you."
                />
                <CreateTokenDialog />
            </div>
            <TokenTable
                tokens={tokens}
                isPending={isPending}
                error={error}
                emptyDescription="Create a token to publish or download artifacts with your build tool."
            />
        </>
    );
}
