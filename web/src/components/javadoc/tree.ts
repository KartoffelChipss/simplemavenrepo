import type { Javadoc, JavadocType } from '@/api/types.ts';

export interface TypeNode {
    type: JavadocType;
    label: string;
    children: TypeNode[];
}

export interface PackageNode {
    name: string;
    label: string;
    url?: string;
    types: TypeNode[];
    children: PackageNode[];
}

const byLabel = (a: { label: string }, b: { label: string }) => a.label.localeCompare(b.label);

function typeTree(types: JavadocType[]): TypeNode[] {
    const nodes = new Map(types.map((type) => [type.name, { type, label: type.name.split('.').pop()!, children: [] }]));
    const roots: TypeNode[] = [];
    for (const node of nodes.values()) {
        const outer = nodes.get(node.type.name.split('.').slice(0, -1).join('.'));
        (outer ? outer.children : roots).push(node);
    }
    for (const node of nodes.values()) node.children.sort(byLabel);
    return roots.sort(byLabel);
}

export function packageTree(javadoc: Javadoc): PackageNode[] {
    const urls = new Map(javadoc.packages.map((p) => [p.name, p.url]));
    const typesByPackage = new Map<string, JavadocType[]>();
    for (const t of javadoc.types) typesByPackage.set(t.package, [...(typesByPackage.get(t.package) ?? []), t]);

    const root: PackageNode = { name: '', label: '', types: [], children: [] };
    const nodes = new Map<string, PackageNode>([['', root]]);
    const nodeFor = (name: string): PackageNode => {
        const existing = nodes.get(name);
        if (existing) return existing;
        const cut = name.startsWith('<') ? -1 : name.lastIndexOf('.');
        const parent = nodeFor(cut < 0 ? '' : name.slice(0, cut));
        const node: PackageNode = { name, label: name.slice(cut + 1), types: [], children: [] };
        parent.children.push(node);
        nodes.set(name, node);
        return node;
    };
    for (const name of new Set([...urls.keys(), ...typesByPackage.keys()])) {
        const node = nodeFor(name);
        node.url = urls.get(name);
        node.types = typeTree(typesByPackage.get(name) ?? []);
    }

    const compact = (node: PackageNode): PackageNode => {
        while (!node.url && node.types.length === 0 && node.children.length === 1) {
            const child = node.children[0];
            node = { ...child, label: `${node.label}.${child.label}` };
        }
        return { ...node, children: node.children.map(compact).sort(byLabel) };
    };
    return root.children.map(compact).sort(byLabel);
}

export function containsPackage(node: PackageNode, pkg: string | undefined) {
    return pkg !== undefined && (pkg === node.name || pkg.startsWith(`${node.name}.`));
}
