import { apiClient } from '$lib/api/client';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	try {
		const res = await apiClient.GET('/home', {
			fetch,
		});

		if (res.error || !res.data) {
			return {
				error: res.error?.error || 'Failed to load homepage data',
			};
		}

		return res.data;
	} catch {
		return {
			error: 'Failed to load homepage data',
		};
	}
};
