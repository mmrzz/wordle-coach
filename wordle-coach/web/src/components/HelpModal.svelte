<script lang="ts">
	import Modal from "./Modal.svelte";

	type Props = { open: boolean; onclose: () => void };
	let { open, onclose }: Props = $props();
</script>

<Modal {open} title="How to use this" {onclose}>
	<p>
		This is a coach, not the game. Play Wordle somewhere else and relay what
		happens back here; it never knows the answer, so it works out what you know
		purely from the guesses and colours you feed it.
	</p>

	<ol class="steps">
		<li>
			<strong>The row is already filled</strong> with our best guess, shown
			faintly. Press <kbd>Enter</kbd> to take it, or start typing to use your
			own word instead.
		</li>
		<li>
			<strong>Tap <span class="qmark">?</span></strong> to see the five best guesses,
			what each is worth, and to play one straight from the list.
		</li>
		<li>
			<strong>Enter locks the word in.</strong> Now click each tile to match the
			colours the real game showed you: grey, then yellow, then green.
		</li>
		<li>
			<strong>Enter again</strong> confirms those colours, and the next row is
			filled with the new best guess.
		</li>
	</ol>

	<p class="aside">
		Colours you have already proved are filled in for you. A green square stays
		green, and a letter shown yellow comes back at least yellow. Anything it
		cannot prove is left grey, so a tile is only ever wrong in the direction
		that needs raising, never lowering.
	</p>

	<p class="aside">
		Win by colouring a row all green, by pressing the
		<span class="tickmark">✓</span> beside a locked row, or by narrowing the
		field to one word — at that point the answer is settled whether or not you
		have played it.
	</p>

	<p class="aside">
		On an empty board there is a <strong>go wild</strong> offer above the grid:
		one deliberately strange opener, and what it means. Nearly all of them are
		bad guesses and good words. Click the line to play it, or
		<span class="tickmark">↻</span> for another.
	</p>

	<h3>Bits, and what "leaves" means</h3>
	<p>
		Both numbers say the same thing two ways: how much a guess narrows the
		field.
	</p>
	<dl>
		<dt>bits</dt>
		<dd>
			How much the guess tells you, on average. <strong>Each bit halves the
			field.</strong> From 2,315 answers, a 5.89-bit opener leaves about 40 —
			because halving 2,315 nearly six times over gets you there. Higher is
			better, and the practical ceiling is around 6.
		</dd>
		<dt>leaves</dt>
		<dd>
			The same idea without the maths: how many answers you should expect to
			still be standing afterwards. "leaves ~62" means a typical outcome
			leaves about 62 candidates. Lower is better.
		</dd>
	</dl>
	<p>
		They usually agree, but not always, and that is the interesting part. Bits
		reward a guess that splits the field into many groups; "leaves" punishes a
		guess that might dump you into one big group. A guess can score well on one
		and poorly on the other, which is exactly what the temperature dial lets
		you choose between.
	</p>

	<h3>The stars beside each row</h3>
	<p>
		A mark out of five, measured against the best guess available at the time:
		five stars means nothing else would have told you more, and two means you
		left most of the information on the table. Underneath it is where the word
		ranked among all 12,972 legal guesses.
	</p>
	<p>
		Both are scored from the position you played it in, so they grade the
		decision and not the luck — a guess can be excellent and still get you
		nowhere. Hover for the bits behind the rating.
	</p>

	<h3>The two side panels</h3>
	<p>
		On the right, every letter with the chance it is in the answer and the spot
		it most likely sits in, vowels kept apart from consonants. It is there so
		you can work the word out yourself rather than be handed one; an italic
		position means the letter is scattered and that spot is barely a hint.
	</p>
	<p>
		On the left, once four or fewer answers survive, the shortlist itself. At
		that point ranking guesses is beside the point — you can simply read the
		words and pick.
	</p>

	<h3>The temperature dial</h3>
	<p>
		Bottom left. It changes what "best" means — which is to say, whether bits or
		"leaves" wins when the two disagree. In the middle it maximises bits.
		Turned towards <strong>Cautious</strong> it would rather avoid a bad outcome
		than chase a great one, so it minimises "leaves" and then the worst case;
		towards <strong>Bold</strong> it chases the widest split and accepts the
		risk of occasionally landing in a big group.
	</p>

	<h3>If you mis-click a colour</h3>
	<p>
		Press <kbd>Backspace</kbd> while colouring to reopen the word, or use the
		undo arrow to step back a whole turn. If the colours you entered cannot all
		be true, you will be told and sent back to fix them.
	</p>

	<h3>When the answer is not on the list</h3>
	<p>
		We solve against the official list of 2,315 answers, which is what the real
		game draws from. Some clones do not: they let any of the 12,972 legal words
		be the answer.
	</p>
	<p>
		So when nothing on the list can fit your colours, you are asked rather than
		told. Look again first — a mis-clicked tile explains it far more often. If
		you are sure, take the <strong>go off the books</strong> button and the
		whole guess list becomes fair game for the answer. Expect more guesses to
		be needed: a wider field is a slower one to narrow. The board says
		<em>off the books</em> beside the count while it is on, and a new game
		starts back on the list.
	</p>
</Modal>

<style>
	p {
		margin: 0 0 12px;
		font-size: 0.86rem;
		line-height: 1.55;
		color: var(--fg);
	}

	h3 {
		margin: 18px 0 6px;
		font-size: 0.78rem;
		font-weight: 700;
		letter-spacing: 0.07em;
		text-transform: uppercase;
		color: var(--fg-muted);
	}

	.steps {
		margin: 0 0 4px;
		padding-left: 20px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		font-size: 0.86rem;
		line-height: 1.5;
	}

	kbd {
		padding: 1px 5px;
		border: 1px solid var(--border);
		border-radius: 3px;
		background: color-mix(in srgb, var(--fg) 8%, transparent);
		font-family: inherit;
		font-size: 0.78em;
		font-weight: 700;
	}

	.qmark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.15em;
		height: 1.15em;
		border: 1.5px solid var(--border-strong);
		border-radius: 50%;
		font-size: 0.85em;
	}

	.tickmark {
		color: var(--color-correct);
		font-weight: 700;
	}

	.aside {
		margin: 10px 0 0;
		padding: 8px 10px;
		border-radius: 6px;
		background: color-mix(in srgb, var(--fg) 6%, transparent);
		font-size: 0.79rem;
		line-height: 1.5;
		color: var(--fg-muted);
	}

	dl {
		margin: 0 0 12px;
		font-size: 0.86rem;
		line-height: 1.55;
	}

	dt {
		font-weight: 700;
		font-variant: small-caps;
		letter-spacing: 0.04em;
		margin-top: 8px;
	}

	dd {
		margin: 2px 0 0;
		padding-left: 12px;
		border-left: 2px solid var(--border);
		color: var(--fg-muted);
	}

	dd strong {
		color: var(--fg);
	}
</style>
