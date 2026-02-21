import type { Handle, HandleFetch, HandleServerError } from '@sveltejs/kit';
import { PostHog } from 'posthog-node';
import { PUBLIC_APP_ENV, PUBLIC_API_URL, PUBLIC_POSTHOG_HOST, PUBLIC_POSTHOG_KEY } from '$env/static/public';

export const handle: Handle = async ({ event, resolve }) => {
	return resolve(event, {
		filterSerializedResponseHeaders: (name) => name === 'content-type' || name === 'content-length',
	});
};

export const handleError: HandleServerError = async ({ error, status, message }) => {
	// 404s are expected — skip
	if (status === 404) return { message };

	if (PUBLIC_POSTHOG_KEY && !PUBLIC_POSTHOG_KEY.includes('REPLACE')) {
		const client = new PostHog(PUBLIC_POSTHOG_KEY, {
			host: PUBLIC_POSTHOG_HOST || 'https://us.i.posthog.com',
		});
		client.captureException(error, 'server', {
			environment: PUBLIC_APP_ENV || 'development',
		});
		await client.shutdown();
	}

	return { message };
};

export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	if (request.url.startsWith(PUBLIC_API_URL)) {
		const cookies = event.request.headers.get('cookie');
		if (cookies) {
			request.headers.set('cookie', cookies);
		}
	}

	return fetch(request);
};
