import type { Permission, TokenScope } from '@/api/types.ts';
import { PermissionBadge } from '@/components/grants/GrantEditor.tsx';
import { normalizePattern, permissionOptions } from '@/components/grants/pattern.ts';
import { Button } from '@/components/ui/button';
import { FieldDescription, FieldError } from '@/components/ui/field';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, X } from 'lucide-react';
import { useState, type KeyboardEvent } from 'react';

const sameScope = (a: TokenScope, b: TokenScope) => a.pattern === b.pattern && a.permission === b.permission;

export function ScopeEditor({ scopes, onChange }: { scopes: TokenScope[]; onChange: (scopes: TokenScope[]) => void }) {
    const [pattern, setPattern] = useState('');
    const [permission, setPermission] = useState<Permission>('write');

    const normalized = pattern.trim() ? normalizePattern(pattern) : undefined;
    const patternInvalid = normalized && 'error' in normalized ? normalized.error : undefined;

    function add() {
        if (!normalized || 'error' in normalized) return;
        const scope = { pattern: normalized.pattern, permission };
        if (!scopes.some((s) => sameScope(s, scope))) onChange([...scopes, scope]);
        setPattern('');
    }

    function onKeyDown(e: KeyboardEvent) {
        if (e.key !== 'Enter') return;
        e.preventDefault();
        add();
    }

    return (
        <div className="flex flex-col gap-2">
            {scopes.length > 0 && (
                <ul className="flex flex-col divide-y rounded-lg border">
                    {scopes.map((scope) => (
                        <li
                            key={`${scope.pattern} ${scope.permission}`}
                            className="flex items-center gap-2 py-1 pr-1 pl-3"
                        >
                            <span className="min-w-0 flex-1 truncate font-mono text-sm">{scope.pattern}</span>
                            <PermissionBadge permission={scope.permission} />
                            <Button
                                type="button"
                                variant="ghost"
                                size="icon-sm"
                                aria-label={`Remove scope ${scope.pattern}`}
                                onClick={() => onChange(scopes.filter((s) => !sameScope(s, scope)))}
                            >
                                <X />
                            </Button>
                        </li>
                    ))}
                </ul>
            )}
            <div className="flex gap-2">
                <Input
                    aria-label="Repository path"
                    className="min-w-0 flex-1 font-mono"
                    placeholder="com/example/lib"
                    autoCapitalize="none"
                    spellCheck={false}
                    value={pattern}
                    onChange={(e) => setPattern(e.target.value)}
                    onKeyDown={onKeyDown}
                    aria-invalid={!!patternInvalid}
                />
                <Select items={permissionOptions} value={permission} onValueChange={(v) => setPermission(v ?? 'write')}>
                    <SelectTrigger aria-label="Permission" className="w-36 shrink-0">
                        <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                        {permissionOptions.map((o) => (
                            <SelectItem key={o.value} value={o.value}>
                                {o.label}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    aria-label="Add scope"
                    disabled={!normalized || !!patternInvalid}
                    onClick={add}
                >
                    <Plus />
                </Button>
            </div>
            {patternInvalid ? (
                <FieldError>{patternInvalid}</FieldError>
            ) : (
                <FieldDescription>
                    An artifact path like <span className="font-mono">com/example/lib</span> or a group path that covers
                    everything below it. The token never gets more access than its user.
                </FieldDescription>
            )}
        </div>
    );
}
