<script lang="ts">
	import { onMount } from 'svelte';
	import { afterNavigate } from '$app/navigation';
	import Footer from '$lib/components/layout/footer.svelte';
	import NavBar from '$lib/components/layout/nav-bar.svelte';
	import TopLoader from '$lib/components/layout/top-loader.svelte';
	import Sonner from '$lib/components/ui/sonner/sonner.svelte';
	import { capturePageview, identifyUser, initAnalytics } from '$lib/analytics';
	import { setLayoutStateContext } from '$lib/context/layout.svelte';
	import { setAppStateContext } from '$lib/context/state.svelte';
	import { getFontUrlsForTheme } from '$lib/themes';
	import { cn } from '$lib/utils';
	import '../app.css';
	import type { LayoutProps } from './$types';

	let { children, data }: LayoutProps = $props();
	const appState = setAppStateContext(data.user, data.settings);
	const layoutState = setLayoutStateContext();

	let theme = $derived(appState.settings?.theme);

	$effect(() => {
		if (!theme) return;
		document.documentElement.className = cn('dark', theme.className);
	});

	$effect(() => {
		appState.setUser(data.user);
		appState.setSettings(data.settings);
	});

	// Identify / de-identify when the user state changes
	$effect(() => {
		const user = appState.user;
		if (user) {
			identifyUser({ id: user.id, username: user.username, email: user.email });
		} else {
			identifyUser(null);
		}
	});

	onMount(() => {
		initAnalytics();
		// Capture the first pageview after PostHog is ready
		capturePageview(window.location.href);
		// JS errors and unhandled rejections are captured via hooks.client.ts
		// using posthog.captureException(), which feeds PostHog's Error Tracking UI
	});

	// Track SPA pageviews on every client-side navigation
	afterNavigate(({ to }) => {
		if (to?.url) {
			capturePageview(to.url.href);
		}
	});
</script>

<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link rel="stylesheet" href={getFontUrlsForTheme(theme?.className ?? '')} />
</svelte:head>

<TopLoader />
<div class="sticky top-0 z-50" {@attach layoutState.setHeight('navbar')}>
	<NavBar />
</div>
<div class="flex min-h-screen flex-col">
	<div class="flex-1 pb-4">
		{@render children?.()}
	</div>
	<Footer />
</div>
<Sonner richColors />
