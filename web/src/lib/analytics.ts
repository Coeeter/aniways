import posthog from 'posthog-js';
import { PUBLIC_APP_ENV, PUBLIC_POSTHOG_HOST, PUBLIC_POSTHOG_KEY } from '$env/static/public';

export type { PostHog } from 'posthog-js';

// ---------------------------------------------------------------------------
// Breadcrumb trail
// A circular buffer of the last N user actions. Every capture* call adds an
// entry. When an error fires, the full trail is attached so you can see
// exactly what the user did before the error hit.
// ---------------------------------------------------------------------------

type Breadcrumb = {
	timestamp: string; // ISO
	action: string; // event name, e.g. 'anime_viewed'
	page: string; // window.location.pathname at time of action
	detail?: string; // human-readable summary, e.g. "Naruto Shippuden ep 1"
};

const MAX_BREADCRUMBS = 15;
const breadcrumbs: Breadcrumb[] = [];

function addBreadcrumb(action: string, detail?: string) {
	if (typeof window === 'undefined') return;
	breadcrumbs.push({
		timestamp: new Date().toISOString(),
		action,
		page: window.location.pathname,
		detail,
	});
	// Keep only the last MAX_BREADCRUMBS entries
	if (breadcrumbs.length > MAX_BREADCRUMBS) {
		breadcrumbs.shift();
	}
}

export function getBreadcrumbs(): Breadcrumb[] {
	return [...breadcrumbs];
}

// ---------------------------------------------------------------------------
// Init
// ---------------------------------------------------------------------------

/**
 * Initialise PostHog once for the lifetime of the browser session.
 * Must only be called client-side (inside onMount / $effect).
 */
export function initAnalytics() {
	if (typeof window === 'undefined') return;
	if (!PUBLIC_POSTHOG_KEY || PUBLIC_POSTHOG_KEY.includes('REPLACE')) return;

	posthog.init(PUBLIC_POSTHOG_KEY, {
		api_host: PUBLIC_POSTHOG_HOST || 'https://us.i.posthog.com',
		// Capture pageviews manually via afterNavigate so SPA route changes are tracked
		capture_pageview: false,
		// Keep sessions alive for 30 minutes of inactivity
		session_idle_timeout_seconds: 1800,
		// Persist the anonymous device ID across sessions
		persistence: 'localStorage+cookie',
	});

	// Register environment as a super property — attached to every event automatically.
	// Filter by environment = 'production' vs 'development' in the PostHog dashboard.
	posthog.register({
		environment: PUBLIC_APP_ENV || 'development',
	});
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

/**
 * Identify a logged-in user. Call after login or on app boot when already
 * authenticated. Pass null to reset (logout).
 */
export function identifyUser(user: { id: string; username: string; email: string } | null) {
	if (typeof window === 'undefined') return;

	if (!user) {
		posthog.reset();
		return;
	}

	posthog.identify(user.id, {
		username: user.username,
		email: user.email,
	});
}

// ---------------------------------------------------------------------------
// Helpers — current user context (attached to every error event)
// ---------------------------------------------------------------------------

function getCurrentUserContext() {
	// posthog.get_distinct_id() returns either the user's ID (after identify())
	// or the anonymous device ID. Either way it's useful for error correlation.
	try {
		return {
			distinct_id: posthog.get_distinct_id?.() ?? 'unknown',
			is_identified: posthog.get_property?.('$user_id') != null,
			username: posthog.get_property?.('username') ?? null,
		};
	} catch {
		return { distinct_id: 'unknown', is_identified: false, username: null };
	}
}

// ---------------------------------------------------------------------------
// Event capture helpers — all strongly typed, all add a breadcrumb
// ---------------------------------------------------------------------------

export function capturePageview(url: string) {
	addBreadcrumb('pageview', url);
	posthog.capture('$pageview', { $current_url: url });
}

export function captureAnimeViewed(props: {
	anime_id: string;
	anime_title: string;
	genres: string[];
	season: string;
	season_year: number | null;
	rating: string;
}) {
	addBreadcrumb('anime_viewed', `${props.anime_title} (${props.anime_id})`);
	posthog.capture('anime_viewed', props);
}

export function captureEpisodePlayed(props: {
	anime_id: string;
	anime_title: string;
	episode_number: number;
	server_name: string;
	stream_type: string;
}) {
	addBreadcrumb(
		'episode_played',
		`${props.anime_title} ep${props.episode_number} [${props.stream_type}/${props.server_name}]`,
	);
	posthog.capture('episode_played', props);
}

export function captureEpisodeCompleted(props: {
	anime_id: string;
	anime_title: string;
	episode_number: number;
}) {
	addBreadcrumb('episode_completed', `${props.anime_title} ep${props.episode_number}`);
	posthog.capture('episode_completed', props);
}

export function captureSearch(props: {
	query: string;
	result_count: number;
	navigated_to_catalog: boolean;
}) {
	addBreadcrumb('search_performed', `"${props.query}" → ${props.result_count} results`);
	posthog.capture('search_performed', props);
}

export function captureUserLoggedIn() {
	addBreadcrumb('user_logged_in');
	posthog.capture('user_logged_in');
}

export function captureUserRegistered() {
	addBreadcrumb('user_registered');
	posthog.capture('user_registered');
}

export function captureListItemAdded(props: { anime_id: string; list_type: string }) {
	addBreadcrumb('list_item_added', `${props.anime_id} → ${props.list_type}`);
	posthog.capture('list_item_added', props);
}

export function captureListItemRemoved(props: { anime_id: string; list_type: string }) {
	addBreadcrumb('list_item_removed', `${props.anime_id} from ${props.list_type}`);
	posthog.capture('list_item_removed', props);
}

export function captureListItemUpdated(props: {
	anime_id: string;
	status: string;
	watched_episodes: number;
}) {
	addBreadcrumb('list_item_updated', `${props.anime_id} → ${props.status} (${props.watched_episodes} eps)`);
	posthog.capture('list_item_updated', props);
}

// ---------------------------------------------------------------------------
// Error capture — the rich version
// Attaches: page, user context, last N breadcrumbs, stack trace if available
// ---------------------------------------------------------------------------

export function captureServerSwitch(props: {
	anime_id: string;
	anime_title: string;
	episode_number: number;
	from_server: string;
	to_server: string;
	stream_type: string;
	reason: 'manual' | 'error_fallback';
}) {
	addBreadcrumb(
		'server_switched',
		`${props.anime_title} ep${props.episode_number}: ${props.from_server} → ${props.to_server} (${props.reason})`,
	);
	posthog.capture('server_switched', props);
}

export function captureStreamError(props: {
	anime_id: string;
	episode_number: number;
	server_name: string;
	stream_type: string;
	error_message: string;
}) {
	addBreadcrumb(
		'stream_error',
		`${props.anime_id} ep${props.episode_number} on ${props.server_name}: ${props.error_message}`,
	);
	posthog.capture('stream_error', props);
}

export function captureSettingToggled(props: {
	setting: 'autoPlayEpisode' | 'autoNextEpisode' | 'autoResumeEpisode' | 'incognitoMode';
	new_value: boolean;
}) {
	addBreadcrumb('setting_toggled', `${props.setting} → ${props.new_value}`);
	posthog.capture('setting_toggled', props);
}

export function captureThemeChanged(props: { theme_id: number; theme_name: string }) {
	addBreadcrumb('theme_changed', props.theme_name);
	posthog.capture('theme_changed', props);
}

export function captureAccountDeleted() {
	addBreadcrumb('account_deleted');
	posthog.capture('account_deleted');
}

export function captureLibraryCleared() {
	addBreadcrumb('library_cleared');
	posthog.capture('library_cleared');
}

export function captureRandomAnimeUsed(props: { anime_id: string }) {
	addBreadcrumb('random_anime_used', props.anime_id);
	posthog.capture('random_anime_used', props);
}

export function captureResumeClicked(props: {
	anime_id: string;
	anime_title: string;
	resume_episode: number;
}) {
	addBreadcrumb('resume_clicked', `${props.anime_title} ep${props.resume_episode}`);
	posthog.capture('resume_clicked', props);
}

/**
 * Capture an API-level error (5xx responses from the backend).
 * Uses posthog.captureException() so it appears in PostHog's Error Tracking UI
 * alongside JS exceptions, with the breadcrumb trail attached as context.
 */
export function captureApiError(props: {
	error_message: string;
	status_code: number;
	route?: string;
}) {
	const trail = getBreadcrumbs();
	const lastCrumb = trail[trail.length - 1];

	posthog.captureException(new Error(props.error_message), {
		error_type: 'api_error',
		status_code: props.status_code,
		error_route: props.route ?? (typeof window !== 'undefined' ? window.location.pathname : null),
		last_action: lastCrumb?.action ?? null,
		last_action_detail: lastCrumb?.detail ?? null,
		breadcrumb_trail: trail,
	});
}
