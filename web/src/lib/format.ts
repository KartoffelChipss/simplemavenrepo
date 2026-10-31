const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' });
const dateTimeFormat = new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' });

export function formatDate(iso: string): string {
    return dateFormat.format(new Date(iso));
}

export function formatDateTime(iso: string): string {
    return dateTimeFormat.format(new Date(iso));
}
