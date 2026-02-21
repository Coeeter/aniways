<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { captureAnimeViewed } from '$lib/analytics';
	import HeroSection from '$lib/components/anime/layout/hero-section.svelte';
	import InfoSidebar from '$lib/components/anime/layout/info-sidebar.svelte';
	import TabContent from '$lib/components/anime/layout/tab-content.svelte';
	import { cn } from '$lib/utils';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const tabs = ['overview', 'episodes', 'relations'] as const;
	let selectedTab = $derived.by(() => {
		const urlTab = page.url.searchParams.get('tab') || 'overview';
		return urlTab as (typeof tabs)[number];
	});

	onMount(() => {
		captureAnimeViewed({
			anime_id: data.anime.id,
			anime_title: data.anime.jname || data.anime.ename || '',
			genres: data.anime.genre ? data.anime.genre.split(', ') : [],
			season: data.anime.season || '',
			season_year: data.anime.seasonYear ?? null,
			rating: data.ratingLabel,
		});
	});
</script>

<svelte:head>
	<title>{data.anime.jname || data.anime.ename} - Aniways</title>
	<meta name="description" content={`Watch ${data.anime.jname || data.anime.ename} on Aniways`} />
</svelte:head>

<HeroSection
	anime={data.anime}
	banner={data.banner?.url ?? null}
	trailer={data.trailer?.trailer ?? null}
	ratingLabel={data.ratingLabel}
	totalEpisodes={data.episodes?.length}
	libraryEntry={data.libraryStatus ?? null}
	variations={data.variations}
/>

<div
	class="sticky top-17 z-20 w-full border-b bg-background/95 backdrop-blur-md supports-[backdrop-filter]:bg-background/60"
>
	<div class="container mx-auto px-4">
		<nav class="mx-auto flex justify-center gap-8 md:justify-start">
			{#each tabs as tab (tab)}
				<a
					href={`?tab=${tab}`}
					class={cn(
						'relative py-4 text-sm font-medium transition-colors hover:text-primary',
						selectedTab === tab ? 'text-primary' : 'text-muted-foreground',
					)}
					data-sveltekit-noscroll
					data-sveltekit-replacestate
				>
					{tab.charAt(0).toUpperCase() + tab.slice(1)}
					{#if selectedTab === tab}
						<div class="absolute right-0 bottom-0 left-0 h-0.5 bg-primary"></div>
					{/if}
				</a>
			{/each}
		</nav>
	</div>
</div>
<div class="container mx-auto px-4 py-8">
	<div class={cn('grid gap-8 lg:grid-cols-3')}>
		<TabContent
			{selectedTab}
			anime={data.anime}
			episodes={data.episodes ?? []}
			franchise={data.franchise ?? null}
			characters={data.characters ?? null}
		/>

		<InfoSidebar
			anime={data.anime}
			ratingLabel={data.ratingLabel}
			{selectedTab}
			variations={data.variations}
		/>
	</div>
</div>
