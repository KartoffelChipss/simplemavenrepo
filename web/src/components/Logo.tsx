import { cn } from '@/lib/utils';
import { useTheme } from '../theme';

export const Logo = () => {
    const theme = useTheme();

    return theme === 'dark' ? (
        <img src="/logo_dark.svg" alt="simplemavenrepo" className="h-5" />
    ) : (
        <img src="/logo.svg" alt="simplemavenrepo" className="h-5" />
    );
};

export const LogoMark = ({ className }: { className?: string }) => {
    return <img src="/logo_mark.svg" alt="simplemavenrepo" className={cn('h-5', className)} />;
};
