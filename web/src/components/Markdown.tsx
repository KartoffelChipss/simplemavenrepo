import { cn } from '@/lib/utils';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';

export function Markdown({ children, className }: { children: string; className?: string }) {
    return (
        <div
            className={cn(
                'prose max-w-none prose-neutral dark:prose-invert',
                'prose-code:before:content-none prose-code:after:content-none',
                '[&_.contains-task-list]:list-none [&_.contains-task-list]:pl-0',
                className
            )}
        >
            <ReactMarkdown remarkPlugins={[remarkGfm]}>{children}</ReactMarkdown>
        </div>
    );
}
