import { apiClient } from '$lib/api/client';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	const res = await apiClient.GET('/home', {
		fetch,
	});

	return {
		...res.data,
		error: res.error?.error,
	};
};
