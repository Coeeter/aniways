import type { Language } from '$lib/stores/language';

export type TitlePair = {
	ename?: string | null;
	jname?: string | null;
};

export const getMainTitle = (titles: TitlePair, language: Language) => {
	if (language === 'en') {
		return titles.ename || titles.jname || '';
	}
	return titles.jname || titles.ename || '';
};

export const getSubTitle = (titles: TitlePair, language: Language) => {
	if (!titles.ename || !titles.jname) return null;
	if (language === 'en') return titles.jname;
	return titles.ename;
};

export const getLocalizedTitles = (titles: TitlePair, language: Language) => ({
	main: getMainTitle(titles, language),
	sub: getSubTitle(titles, language),
});
