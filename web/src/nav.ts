import type { User } from '@/api/types.ts';
import { BookOpen, Boxes, FolderTree, KeyRound, UserGroup, UserRound, type LucideIcon } from 'lucide-react';
import { matchPath } from 'react-router';

export interface NavItem {
    title: string;
    to: string;
    icon: LucideIcon;
    end?: boolean;
}

export interface NavGroup {
    label: string;
    visible: (user: User | null) => boolean;
    items: NavItem[];
}

export const navGroups: NavGroup[] = [
    {
        label: 'Repository',
        visible: () => true,
        items: [
            { title: 'Browse', to: '/browse', icon: FolderTree },
            { title: 'Javadoc', to: '/docs', icon: BookOpen },
        ],
    },
    {
        label: 'Account',
        visible: (user) => user !== null,
        items: [{ title: 'Tokens', to: '/tokens', icon: KeyRound }],
    },
    {
        label: 'Administration',
        visible: (user) => user?.isAdmin ?? false,
        items: [
            { title: 'Artifacts', to: '/artifacts', icon: Boxes },
            { title: 'Users', to: '/users', icon: UserRound },
            { title: 'Teams', to: '/teams', icon: UserGroup },
        ],
    },
];

/** Returns the title of the navigation item the path belongs to, for the page header. */
export function navTitle(pathname: string): string | undefined {
    if (matchPath({ path: '/browse/*' }, pathname)) return 'Browse';
    for (const group of navGroups) {
        for (const item of group.items) {
            if (matchPath({ path: item.to, end: item.end ?? false }, pathname)) return item.title;
        }
    }
    return undefined;
}
