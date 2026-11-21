import { Badge } from '@/components/ui/badge';
import { Warehouse } from 'lucide-react';

export function RepositoryBadge({ name }: { name: string }) {
    return (
        <Badge variant="outline" className="font-mono" title={`Repository ${name}`}>
            <Warehouse />
            {name}
        </Badge>
    );
}
