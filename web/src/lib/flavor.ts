// The league's small print: deadpan one-liners rotated once per day.
// Picked by UTC day so the server render and client hydration agree.

export const mottos = [
	'Strictly neutral. Aggressively so.',
	'No Gandhis were provoked in the making of this website.',
	'All denouncements are final.',
	'Powered by faith, gold, and lingering grievances.',
	'Warmonger penalties apply site-wide.',
	'This footer maintains open borders.',
	'Officially recognised by zero governing bodies.',
	'Loyalty pressure in this region is high.'
];

export const heroQuotes = [
	{ text: 'I came for the diplomacy. I stayed for the betrayal.', by: 'a founding member' },
	{ text: 'Just one more turn.', by: 'everyone, at 2 a.m.' },
	{ text: 'The early game is my strongest phase, in that I still have hope.', by: 'rating withheld' },
	{ text: 'We are a very serious league.', by: 'the very serious league' },
	{ text: 'History will judge us kindly. We maintain the records ourselves.', by: 'Hall of Records' }
];

export function daily<T>(list: T[]): T {
	const day = Math.floor(Date.now() / 86_400_000);
	return list[day % list.length];
}
