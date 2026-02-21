import createClient from 'openapi-fetch';
import { browser } from '$app/environment';
import { PUBLIC_API_URL } from '$env/static/public';
import { captureApiError } from '$lib/analytics';
import type { paths } from './openapi';

export const apiClient = createClient<paths>({
	baseUrl: PUBLIC_API_URL || 'http://localhost:8080',
	credentials: 'include',
});

// Track 5xx API errors client-side only (4xx are expected user-error flows)
if (browser) {
	apiClient.use({
		onResponse({ response, request }) {
			if (response.status >= 500) {
				captureApiError({
					error_message: `${request.method} ${new URL(request.url).pathname} → ${response.status}`,
					status_code: response.status,
					route: window.location.pathname,
				});
			}
			return undefined;
		},
	});
}
