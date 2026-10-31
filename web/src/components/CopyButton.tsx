import { Button } from '@/components/ui/button';
import { toast } from '@/components/ui/toast';
import { Check, Copy } from 'lucide-react';
import { useState, type ComponentProps } from 'react';

async function copyText(text: string): Promise<boolean> {
    try {
        await navigator.clipboard.writeText(text);
        return true;
    } catch {
        // clipboard not available
    }
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    document.body.append(textarea);
    textarea.select();
    const ok = document.execCommand('copy');
    textarea.remove();
    return ok;
}

export function CopyButton({
    text,
    label = 'Copy',
    ...props
}: { text: string; label?: string } & ComponentProps<typeof Button>) {
    const [copied, setCopied] = useState(false);

    async function onClick() {
        if (await copyText(text)) {
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        } else {
            toast.add({ title: 'Could not copy', description: 'Select the text and copy it manually.' });
        }
    }

    return (
        <Button variant="ghost" size="icon-sm" aria-label={label} onClick={() => void onClick()} {...props}>
            {copied ? <Check /> : <Copy />}
        </Button>
    );
}
