const groupIdPattern = /^[A-Za-z0-9_-]+(\.[A-Za-z0-9_-]+)*$/;
const artifactIdPattern = /^[A-Za-z0-9_-][A-Za-z0-9_.-]*$/;

export const MAX_DESCRIPTION_LENGTH = 1000;

export function groupIdError(groupId: string): string | undefined {
    if (groupId && !groupIdPattern.test(groupId)) {
        return "Use dot-separated parts of letters, digits, '_' or '-', e.g. com.example";
    }
}

export function artifactIdError(artifactId: string): string | undefined {
    if (artifactId && !artifactIdPattern.test(artifactId)) {
        return "Use letters, digits, '.', '_' or '-', not starting with '.'";
    }
}
