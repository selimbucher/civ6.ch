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

export function daily<T>(list: T[]): T {
	const day = Math.floor(Date.now() / 86_400_000);
	return list[day % list.length];
}
