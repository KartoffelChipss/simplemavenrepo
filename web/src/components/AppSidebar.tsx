import { useLogout, useMe } from '@/api/auth.ts';
import { useServerName } from '@/api/settings.ts';
import { UserAvatar } from '@/components/UserAvatar.tsx';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuGroup,
    DropdownMenuItem,
    DropdownMenuLabel,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
    DropdownMenuSeparator,
    DropdownMenuSub,
    DropdownMenuSubContent,
    DropdownMenuSubTrigger,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
    Sidebar,
    SidebarContent,
    SidebarFooter,
    SidebarGroup,
    SidebarGroupLabel,
    SidebarHeader,
    SidebarMenu,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
    useSidebar,
} from '@/components/ui/sidebar';
import { navGroups, type NavItem } from '@/nav.ts';
import { setTheme, useTheme, type Theme } from '@/theme.ts';
import { ChevronsUpDown, LogIn, LogOut, Monitor, Moon, Package, Sun, SunMoon } from 'lucide-react';
import { Link, useMatch, useNavigate } from 'react-router';

function NavMenuItem({ item }: { item: NavItem }) {
    const active = useMatch({ path: item.to, end: item.end ?? false }) !== null;
    const { setOpenMobile } = useSidebar();

    return (
        <SidebarMenuItem>
            <SidebarMenuButton
                isActive={active}
                tooltip={item.title}
                render={<Link to={item.to} onClick={() => setOpenMobile(false)} />}
            >
                <item.icon />
                <span>{item.title}</span>
            </SidebarMenuButton>
        </SidebarMenuItem>
    );
}

const themes: { value: Theme; label: string; icon: typeof Sun }[] = [
    { value: 'system', label: 'System', icon: Monitor },
    { value: 'light', label: 'Light', icon: Sun },
    { value: 'dark', label: 'Dark', icon: Moon },
];

function ThemeSubmenu() {
    const theme = useTheme();

    return (
        <DropdownMenuSub>
            <DropdownMenuSubTrigger>
                <SunMoon />
                Theme
            </DropdownMenuSubTrigger>
            <DropdownMenuSubContent>
                <DropdownMenuRadioGroup value={theme} onValueChange={(value: Theme) => setTheme(value)}>
                    {themes.map((t) => (
                        <DropdownMenuRadioItem key={t.value} value={t.value}>
                            <t.icon />
                            {t.label}
                        </DropdownMenuRadioItem>
                    ))}
                </DropdownMenuRadioGroup>
            </DropdownMenuSubContent>
        </DropdownMenuSub>
    );
}

function UserMenu() {
    const { data: user, isPending } = useMe();
    const logout = useLogout();
    const { isMobile } = useSidebar();
    const navigate = useNavigate();

    if (isPending) return null;
    if (!user) {
        return (
            <SidebarMenu>
                <SidebarMenuItem>
                    <SidebarMenuButton tooltip="Log in" render={<Link to="/login" />}>
                        <LogIn />
                        <span>Log in</span>
                    </SidebarMenuButton>
                </SidebarMenuItem>
            </SidebarMenu>
        );
    }

    function onLogout() {
        logout.mutate(undefined, { onSuccess: () => void navigate('/') });
    }

    const avatar = <UserAvatar username={user.username} size={32} className="rounded-lg" />;

    return (
        <SidebarMenu>
            <SidebarMenuItem>
                <DropdownMenu>
                    <DropdownMenuTrigger
                        render={<SidebarMenuButton size="lg" className="data-popup-open:bg-sidebar-accent" />}
                    >
                        {avatar}
                        <div className="grid flex-1 text-left text-sm leading-tight">
                            <span className="truncate font-medium">{user.username}</span>
                            <span className="truncate text-xs text-muted-foreground">
                                {user.isAdmin ? 'Administrator' : 'User'}
                            </span>
                        </div>
                        <ChevronsUpDown className="ml-auto size-4" />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent
                        className="min-w-56"
                        side={isMobile ? 'bottom' : 'right'}
                        align="end"
                        sideOffset={4}
                    >
                        <DropdownMenuGroup>
                            <DropdownMenuLabel className="flex items-center gap-2 font-normal">
                                {avatar}
                                <span className="truncate font-medium text-foreground">{user.username}</span>
                            </DropdownMenuLabel>
                        </DropdownMenuGroup>
                        <DropdownMenuSeparator />
                        <ThemeSubmenu />
                        <DropdownMenuSeparator />
                        <DropdownMenuItem onClick={onLogout}>
                            <LogOut />
                            Log out
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>
            </SidebarMenuItem>
        </SidebarMenu>
    );
}

export function AppSidebar() {
    const { data: user = null } = useMe();
    const groups = navGroups.filter((g) => g.visible(user));
    const serverName = useServerName();

    return (
        <Sidebar collapsible="icon">
            <SidebarHeader>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton size="lg" render={<Link to="/" />}>
                            <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                                <Package className="size-4" />
                            </div>
                            <div className="grid flex-1 text-left text-sm leading-tight">
                                <span className="truncate font-semibold">{serverName}</span>
                                <span className="truncate text-xs text-muted-foreground">Maven repository</span>
                            </div>
                        </SidebarMenuButton>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarHeader>
            <SidebarContent>
                {groups.map((group) => (
                    <SidebarGroup key={group.label}>
                        <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
                        <SidebarMenu>
                            {group.items.map((item) => (
                                <NavMenuItem key={item.to} item={item} />
                            ))}
                        </SidebarMenu>
                    </SidebarGroup>
                ))}
            </SidebarContent>
            <SidebarFooter>
                <UserMenu />
            </SidebarFooter>
            <SidebarRail />
        </Sidebar>
    );
}
