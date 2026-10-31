const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' });
const dateTimeFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });

export function formatDate(iso: string): string {
    return dateFormat.format(new Date(iso));
}

export function formatDateTime(iso: string): string {
    return dateTimeFormat.format(new Date(iso));
}

const sizeUnits = ['B', 'KB', 'MB', 'GB', 'TB'];

export function formatBytes(bytes: number): string {
    let value = bytes;
    let unit = 0;
    while (value >= 1024 && unit < sizeUnits.length - 1) {
        value /= 1024;
        unit++;
    }
    return `${unit === 0 ? value : value.toFixed(value < 10 ? 1 : 0)} ${sizeUnits[unit]}`;
}
