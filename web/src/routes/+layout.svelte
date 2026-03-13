<script lang="ts">
	import { CircleAlert } from 'lucide-svelte';
	import posthog from 'posthog-js';
	import { onMount } from 'svelte';
	import { afterNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import { PUBLIC_APP_ENV } from '$env/static/public';
	import { capturePageview, initAnalytics } from '$lib/analytics';
	import Footer from '$lib/components/layout/footer.svelte';
	import NavBar from '$lib/components/layout/nav-bar.svelte';
	import TopLoader from '$lib/components/layout/top-loader.svelte';
	import Sonner from '$lib/components/ui/sonner/sonner.svelte';
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
	let lastIdentifiedUserId: string | null = null;

	$effect(() => {
		if (!theme) return;
		document.documentElement.className = cn('dark', theme.className);
	});

	$effect(() => {
		appState.setUser(data.user);
		appState.setSettings(data.settings);
	});

	onMount(() => {
		initAnalytics();
	});

	$effect(() => {
		if (!appState.user) return;

		if (lastIdentifiedUserId === appState.user.id) return; // Avoid re-identifying the same user

		posthog.identify(appState.user.id, {
			username: appState.user.username,
			email: appState.user.email,
			environment: PUBLIC_APP_ENV || 'development',
		});

		lastIdentifiedUserId = appState.user.id;
	});

	afterNavigate(({ to }) => {
		const toUrl = to?.url?.href;
		if (!toUrl) return;

		capturePageview(toUrl);
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
<div class="relative flex min-h-screen flex-col">
	{#if page.url.pathname === '/'}
		<div
			class="absolute top-0 left-0 z-30 flex w-full items-center border border-red-400 bg-red-800 p-4 text-foreground"
		>
			<CircleAlert class="mr-2 inline size-16" />
			The website is unable to stream content due to parent server limitations. I suggest using another
			website in the meantime. I'm actively working on a solution to this issue. I apologize for the
			inconvenience and appreciate your understanding.
		</div>
	{/if}
	<div class="flex-1 pb-4">
		{@render children?.()}
	</div>
	<Footer />
</div>
<Sonner richColors />
