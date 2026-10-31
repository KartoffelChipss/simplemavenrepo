const MATCH_ALL = '**';

export function normalizePattern(input: string): { pattern: string } | { error: string } {
    let p = input.trim().replace(/^\/+|\/+$/g, '');
    if (p === MATCH_ALL) return { pattern: MATCH_ALL };
    if (p.endsWith(`/${MATCH_ALL}`)) p = p.slice(0, -MATCH_ALL.length - 1);
    if (p === '' || p.includes('\\')) return { error: 'Enter a repository path like com/example' };
    for (const segment of p.split('/')) {
        if (segment === '' || segment === '.' || segment === '..' || segment.includes('*')) {
            return { error: 'Use a path like com/example/**, wildcards only work at the end' };
        }
    }
    return { pattern: `${p}/${MATCH_ALL}` };
}
