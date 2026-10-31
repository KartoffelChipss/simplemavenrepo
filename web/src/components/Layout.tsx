import { useServerName } from '@/api/settings.ts';
import { AppSidebar } from '@/components/AppSidebar.tsx';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { navTitle } from '@/nav.ts';
import { useEffect } from 'react';
import { Outlet, useLocation } from 'react-router';

export function Layout() {
    const { pathname } = useLocation();
    const title = navTitle(pathname);
    const serverName = useServerName();

    useEffect(() => {
        document.title = serverName;
    }, [serverName]);

    return (
        <SidebarProvider>
            <AppSidebar />
            <SidebarInset>
                <header className="flex h-14 shrink-0 items-center gap-2 border-b px-4">
                    <SidebarTrigger className="-ml-1" />
                    <Separator orientation="vertical" className="my-auto mr-2 data-[orientation=vertical]:h-4" />
                    {title && <span className="text-sm font-medium">{title}</span>}
                </header>
                <div className="flex flex-1 flex-col gap-4 p-6">
                    <Outlet />
                </div>
            </SidebarInset>
        </SidebarProvider>
    );
}
