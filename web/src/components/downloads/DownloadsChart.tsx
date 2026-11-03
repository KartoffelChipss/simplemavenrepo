import type { DayDownloads, DownloadDays } from '@/api/types.ts';
import { ChartContainer, ChartTooltip, ChartTooltipContent, type ChartConfig } from '@/components/ui/chart';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts';

const periods = [
    { value: '7', label: 'Last 7 days' },
    { value: '30', label: 'Last 30 days' },
    { value: '90', label: 'Last 90 days' },
    { value: '365', label: 'Last 365 days' },
];

export function PeriodSelect({ days, onChange }: { days: DownloadDays; onChange: (days: DownloadDays) => void }) {
    return (
        <Select
            items={periods}
            value={String(days)}
            onValueChange={(value) => value && onChange(Number(value) as DownloadDays)}
        >
            <SelectTrigger aria-label="Period" className="w-36">
                <SelectValue />
            </SelectTrigger>
            <SelectContent>
                {periods.map((p) => (
                    <SelectItem key={p.value} value={p.value}>
                        {p.label}
                    </SelectItem>
                ))}
            </SelectContent>
        </Select>
    );
}

// Days are UTC dates, formatting them in the local time zone could show the day before
const tickFormat = new Intl.DateTimeFormat(undefined, { month: 'short', day: 'numeric', timeZone: 'UTC' });
const tooltipFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeZone: 'UTC' });
const parseDay = (date: string) => new Date(`${date}T00:00:00Z`);

const chartConfig = {
    count: { label: 'Downloads', color: 'var(--chart-1)' },
} satisfies ChartConfig;

export function DownloadsChart({ daily }: { daily: DayDownloads[] }) {
    return (
        <ChartContainer config={chartConfig} className="aspect-auto h-48 w-full">
            <BarChart data={daily} margin={{ top: 4, left: 0, right: 0 }} barCategoryGap={daily.length > 90 ? 0 : 2}>
                <CartesianGrid vertical={false} />
                <XAxis
                    dataKey="date"
                    tickLine={false}
                    axisLine={false}
                    tickMargin={8}
                    minTickGap={24}
                    tickFormatter={(date: string) => tickFormat.format(parseDay(date))}
                />
                <YAxis width={40} tickLine={false} axisLine={false} allowDecimals={false} />
                <ChartTooltip
                    cursor={{ fill: 'var(--muted)', opacity: 0.6 }}
                    content={
                        <ChartTooltipContent labelFormatter={(date) => tooltipFormat.format(parseDay(String(date)))} />
                    }
                />
                {/* Animating from one period to another morphs unrelated bars into each other */}
                <Bar dataKey="count" fill="var(--color-count)" radius={[4, 4, 0, 0]} isAnimationActive={false} />
            </BarChart>
        </ChartContainer>
    );
}

export function DownloadTotals({
    periodTotal,
    total,
    days,
}: {
    periodTotal: number;
    total: number;
    days: DownloadDays;
}) {
    return (
        <div className="grid gap-3 sm:grid-cols-2">
            <div className="flex flex-col gap-1 rounded-lg border p-3">
                <span className="text-muted-foreground">Last {days} days</span>
                <span className="font-heading text-2xl font-medium tabular-nums">{periodTotal.toLocaleString()}</span>
            </div>
            <div className="flex flex-col gap-1 rounded-lg border p-3">
                <span className="text-muted-foreground">All time</span>
                <span className="font-heading text-2xl font-medium tabular-nums">{total.toLocaleString()}</span>
            </div>
        </div>
    );
}
