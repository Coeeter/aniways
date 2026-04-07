import type { Handle, HandleFetch } from '@sveltejs/kit';
import { PUBLIC_API_URL } from '$env/static/public';

export const handle: Handle = async ({ event, resolve }) => {
	return resolve(event, {
		filterSerializedResponseHeaders: (name) => name === 'content-type' || name === 'content-length',
	});
};

export const handleFetch: HandleFetch = async ({ event, request, fetch }) => {
	if (request.url.startsWith(PUBLIC_API_URL)) {
		const cookies = event.request.headers.get('cookie');
		if (cookies) {
			request.headers.set('cookie', cookies);
		}

		// Forward the real client IP so the API rate limiter keys by user IP,
		// not the shared SSR server IP (which would cause all SSR requests from
		// multiple users to share a single rate limit bucket).
		const clientIp =
			event.request.headers.get('cf-connecting-ip') ||
			event.request.headers.get('x-forwarded-for')?.split(',')[0]?.trim() ||
			event.getClientAddress();
		if (clientIp) {
			request.headers.set('cf-connecting-ip', clientIp);
		}
	}

	return fetch(request);
};
