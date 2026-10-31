import type { Javadoc } from '@/api/types.ts';

export type ResultKind = 'package' | 'type' | 'member';

export interface SearchResult {
    kind: ResultKind;
    name: string;
    context: string;
    url: string;
}

const LIMIT = 100;

export function search(javadoc: Javadoc, query: string): SearchResult[] {
    const q = query.trim().toLowerCase();
    if (!q) return [];

    const scored: { result: SearchResult; score: number }[] = [];
    const add = (result: SearchResult, name: string, qualified: string, rank: number) => {
        const n = name.toLowerCase();
        let score: number;
        if (n === q) score = 0;
        else if (n.startsWith(q)) score = 1;
        else if (n.includes(q)) score = 2;
        else if (qualified.toLowerCase().includes(q)) score = 3;
        else return;
        scored.push({ result, score: score * 3 + rank });
    };

    for (const t of javadoc.types) {
        add({ kind: 'type', name: t.name, context: t.package, url: t.url }, t.name, `${t.package}.${t.name}`, 0);
    }
    for (const p of javadoc.packages) {
        add({ kind: 'package', name: p.name, context: '', url: p.url }, p.name, p.name, 1);
    }
    for (const m of javadoc.members) {
        // Matches the member's name without its parameters, so "get" ranks get(int) as an exact match
        const name = m.name.split('(')[0];
        add({ kind: 'member', name: m.name, context: m.type, url: m.url }, name, `${m.type}.${m.name}`, 2);
    }

    scored.sort((a, b) => a.score - b.score || a.result.name.length - b.result.name.length);
    return scored.slice(0, LIMIT).map((s) => s.result);
}
