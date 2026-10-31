import { cn } from '@/lib/utils';
import { useIsDark } from '@/theme.ts';
import BoringAvatar from 'boring-avatars';

const LIGHT_COLORS = ['#1d6c96', '#83c4e7', '#144a67', '#2f7a5c', '#d9a24c'];
const DARK_COLORS = ['#7ab0d0', '#a8cfe3', '#5ba3cc', '#56ad84', '#e8b25c'];

export function UserAvatar({
    username,
    size = 32,
    className,
}: {
    username: string;
    size?: number;
    className?: string;
}) {
    const dark = useIsDark();

    return (
        <BoringAvatar
            name={username}
            variant="beam"
            colors={dark ? DARK_COLORS : LIGHT_COLORS}
            size={size}
            square
            aria-hidden
            className={cn('shrink-0 rounded-full', className)}
        />
    );
}
