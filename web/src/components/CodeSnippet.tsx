import { CopyButton } from '@/components/CopyButton.tsx';
import { CodeBlock, CodeBlockActions, CodeBlockContent, CodeBlockLine } from '@/components/ui/code-block';
import { useHighlightedCode } from '@/hooks/use-highlighted-code';

export function CodeSnippet({ code, language, label }: { code: string; language: string; label: string }) {
    const lines = useHighlightedCode(code, language);

    return (
        <CodeBlock size="sm">
            <CodeBlockActions>
                <CopyButton text={code} label={`Copy ${label} snippet`} size="icon-xs" />
            </CodeBlockActions>
            <CodeBlockContent>
                {lines.map((line, i) => (
                    <CodeBlockLine key={i}>{line}</CodeBlockLine>
                ))}
            </CodeBlockContent>
        </CodeBlock>
    );
}
