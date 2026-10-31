import { useServerName } from '@/api/settings.ts';
import { AppSidebar } from '@/components/AppSidebar.tsx';
import { Separator } from '@/components/ui/separator';
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar';
import { navTitle } from '@/nav.ts';
import { useEffect, useState } from 'react';
import { Outlet, useLocation } from 'react-router';

const SIDEBAR_STORAGE_KEY = 'smr.sidebar';

function readSidebarOpen(): boolean {
    try {
        return localStorage.getItem(SIDEBAR_STORAGE_KEY) !== 'collapsed';
    } catch {
        // Storage unavailable
        return true;
    }
}

function storeSidebarOpen(open: boolean) {
    try {
        if (open) localStorage.removeItem(SIDEBAR_STORAGE_KEY);
        else localStorage.setItem(SIDEBAR_STORAGE_KEY, 'collapsed');
    } catch {
        // Storage unavailable
    }
}

export function Layout() {
    const { pathname } = useLocation();
    const title = navTitle(pathname);
    const serverName = useServerName();
    const [sidebarOpen, setSidebarOpen] = useState(readSidebarOpen);

    function onSidebarOpenChange(open: boolean) {
        setSidebarOpen(open);
        storeSidebarOpen(open);
    }

    useEffect(() => {
        document.title = serverName;
    }, [serverName]);

    return (
        <SidebarProvider open={sidebarOpen} onOpenChange={onSidebarOpenChange}>
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
