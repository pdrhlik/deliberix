import { TranslateService } from "@ngx-translate/core";

/**
 * Resolve an HTTP error to a localized, user-facing string.
 *
 * Strategy:
 * 1. If the response carries a `code`, look up `errors.{code}` in i18n.
 * 2. Otherwise fall back to the server's `message` (English).
 * 3. Otherwise fall back to a generic `common.error` translation.
 */
export function apiErrorMessage(err: unknown, translate: TranslateService): string {
  const body = (err as { error?: { code?: string; message?: string } } | undefined)?.error;
  const code = body?.code;
  if (code) {
    const key = `errors.${code}`;
    const translated = translate.instant(key);
    if (translated && translated !== key) return translated;
  }
  if (body?.message) return body.message;
  return translate.instant("common.error");
}
