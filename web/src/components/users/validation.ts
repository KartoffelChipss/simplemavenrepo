const usernamePattern = /^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$/;

export const MIN_PASSWORD_LENGTH = 8;
export const MAX_PASSWORD_BYTES = 72;

export function usernameError(username: string): string | undefined {
    if (username && !usernamePattern.test(username)) {
        return "Use up to 64 letters, digits, '.', '_' or '-', starting with a letter or digit";
    }
}

export function passwordError(password: string): string | undefined {
    if (!password) return;
    if (password.length < MIN_PASSWORD_LENGTH) return `Use at least ${MIN_PASSWORD_LENGTH} characters`;
    if (new TextEncoder().encode(password).length > MAX_PASSWORD_BYTES) return 'This password is too long';
}
