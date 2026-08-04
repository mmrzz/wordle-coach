/**
 * Where a word's definition is looked up.
 *
 * Wiktionary rather than a standard dictionary because the guess list contains
 * thousands of obscure words a solver will happily recommend. SOARE, ROATE and
 * RAILE are all top openers and none of them appear in Merriam-Webster;
 * Wiktionary has them. Coverage is still not total, so a link can land on a
 * "no entry" page.
 */
const BASE = "https://en.wiktionary.org/wiki/";

/** Builds the definition link for a word. */
export function definitionUrl(word: string): string {
	return `${BASE}${encodeURIComponent(word.toLowerCase())}`;
}
