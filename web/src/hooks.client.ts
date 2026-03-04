import type { HandleClientError } from '@sveltejs/kit';
import posthog from 'posthog-js';
import { getBreadcrumbs } from '$lib/analytics';

export const handleError: HandleClientError = ({ error, status, message }) => {
	// 404s are expected — don't log them as exceptions
	if (status === 404) return { message };

	posthog.captureException(error, {
		// Attach our breadcrumb trail so PostHog's error view shows
		// exactly what the user was doing before this error fired
		breadcrumb_trail: getBreadcrumbs(),
		sveltekit_status: status,
		sveltekit_message: message,
	});

	return { message };
};
