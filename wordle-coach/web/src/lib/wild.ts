/**
 * Openers for players who would rather have fun than win.
 *
 * Every word here is a legal guess, and most of them are terrible ones: they
 * hoard rare letters, double up, and tell you almost nothing. That is the
 * point. The meaning is the reward for playing one.
 */
export type WildWord = {
	word: string;
	/** Reads after "it means", so it is a phrase and not a sentence. */
	meaning: string;
};

/**
 * Curated by hand rather than pulled from the corpus: rarity alone gets you a
 * page of obsolete plurals, and what makes a word worth typing is the story
 * behind it. All are in the allowed list, so any of them can actually be played.
 */
export const WILD: WildWord[] = [
	{ word: "crwth", meaning: "an ancient Celtic lyre — and a word with no a, e, i, o or u" },
	{ word: "cwtch", meaning: "a Welsh hug, and also the cubbyhole you keep things in" },
	{ word: "phpht", meaning: "a snort of mild disgust, written down" },
	{ word: "schwa", meaning: "the lazy 'uh' ending sofa — the commonest vowel sound in English" },
	{ word: "kudzu", meaning: "the vine that ate the American South, a foot of it a day" },
	{ word: "xebec", meaning: "a three-masted ship, once the favourite of Mediterranean corsairs" },
	{ word: "umiak", meaning: "an open Inuit boat of stretched skins, big enough for the family" },
	{ word: "sylph", meaning: "a spirit of the air, invented whole by a 16th-century alchemist" },
	{ word: "golem", meaning: "a figure of clay woken by a written word placed in its mouth" },
	{ word: "djinn", meaning: "a spirit made of smokeless fire, and the ancestor of the genie" },
	{ word: "dryad", meaning: "a nymph who lives inside one tree and dies when it falls" },
	{ word: "nixie", meaning: "a water sprite who sings swimmers under" },
	{ word: "kvell", meaning: "to burst with pride over somebody else's success" },
	{ word: "plotz", meaning: "to collapse from an excess of feeling" },
	{ word: "schmo", meaning: "a hapless nobody, and never the hero of the story" },
	{ word: "yenta", meaning: "a woman who can keep neither a secret nor her matchmaking to herself" },
	{ word: "quoit", meaning: "a heavy ring you fling at a stake and hope it catches" },
	{ word: "quirt", meaning: "a riding whip with a short braided lash" },
	{ word: "gonzo", meaning: "reporting so personal the reporter becomes the story" },
	{ word: "hokum", meaning: "sentimental nonsense that plays beautifully to a crowd" },
	{ word: "blurb", meaning: "the puff on a book jacket — coined as a joke in 1907" },
	{ word: "bumph", meaning: "paperwork so useless it is named after lavatory paper" },
	{ word: "snook", meaning: "the thumb-to-nose gesture you cock at someone" },
	{ word: "zilch", meaning: "nothing at all, said with a certain relish" },
	{ word: "pawky", meaning: "drily and slyly funny, in the Scots manner" },
	{ word: "fubsy", meaning: "small, plump and pleasingly shaped" },
	{ word: "moxie", meaning: "nerve and know-how — named after a bitter American soda" },
	{ word: "nabob", meaning: "someone who came home from India indecently rich" },
	{ word: "ratel", meaning: "the honey badger, celebrated for not caring" },
	{ word: "oribi", meaning: "a small antelope that whistles when it is alarmed" },
	{ word: "fetor", meaning: "a stench bad enough to have earned its own noun" },
	{ word: "knurl", meaning: "the little milled ridges on a knob, there so your fingers grip" },
	{ word: "wodge", meaning: "a thick, clumsy lump of something" },
	{ word: "huzza", meaning: "a shout of triumph, older than hooray" },
	{ word: "zayin", meaning: "the seventh letter of the Hebrew alphabet, named for a weapon" },
	{ word: "bwana", meaning: "boss, in Swahili" },
	{ word: "kraal", meaning: "a ring of huts inside a stockade, cattle and all" },
	{ word: "sabot", meaning: "a wooden shoe — the root sabotage grew from" },
	{ word: "quale", meaning: "the redness of red: what a thing is like from the inside" },
	{ word: "mungo", meaning: "cheap cloth shredded back out of old rags" },
	{ word: "yobbo", meaning: "a lout, from 'boy' spelled backwards" },
	{ word: "jorum", meaning: "a very large drinking bowl, and everything in it" },
	{ word: "vuggy", meaning: "riddled with small crystal-lined cavities, said of rock" },
	{ word: "boffo", meaning: "a runaway hit, in the language of show business" },
	{ word: "flump", meaning: "to drop down heavily, and audibly" },
	{ word: "mufti", meaning: "the plain clothes of someone who normally wears a uniform" },
	{ word: "pukka", meaning: "genuine and first class, from the Hindi for cooked" },
	{ word: "xysts", meaning: "the covered walks where ancient athletes trained out of the rain" },
	{ word: "yclad", meaning: "clothed — a fossil of Old English still sitting in the dictionary" },
	{ word: "hajji", meaning: "one who has made the pilgrimage to Mecca" },
	{ word: "skoal", meaning: "a Nordic toast: raise the bowl and drink" },
	{ word: "zebus", meaning: "humped cattle, sacred across much of India" },
	{ word: "nudzh", meaning: "to pester somebody, patiently, until they give in" },
	{ word: "quaff", meaning: "to drink deep and with obvious enjoyment" },
];

/**
 * Picks a word at random, never the one already on offer. Rerolling and getting
 * the same word back reads as a broken button, however fair the coin was.
 */
export function randomWild(except?: string): WildWord {
	const pool = WILD.filter((w) => w.word !== except);
	return pool[Math.floor(Math.random() * pool.length)];
}
