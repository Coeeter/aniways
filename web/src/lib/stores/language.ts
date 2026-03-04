import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export type Language = 'jp' | 'en';

const STORAGE_KEY = 'aniways-language';
const DEFAULT_LANGUAGE: Language = 'jp';

const getInitialLanguage = () => {
	if (!browser) return DEFAULT_LANGUAGE;
	const stored = localStorage.getItem(STORAGE_KEY);
	if (stored === 'en' || stored === 'jp') {
		return stored;
	}
	return DEFAULT_LANGUAGE;
};

const store = writable<Language>(getInitialLanguage());

if (browser) {
	store.subscribe((value) => {
		localStorage.setItem(STORAGE_KEY, value);
	});
}

export const language = store;
export const setLanguage = (value: Language) => {
	store.set(value);
};
