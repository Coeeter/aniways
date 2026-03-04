<script lang="ts">
	import { Play } from 'lucide-svelte';
	import { captureResumeClicked } from '$lib/analytics';
	import AnimeCard from '$lib/components/anime/display/anime-card.svelte';
	import EmptyState from '$lib/components/anime/display/empty-state.svelte';
	import PageHeader from '$lib/components/layout/page-header.svelte';
	import { Pagination } from '$lib/components/ui/pagination';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const languageValue = $derived($language);
</script>

<svelte:head>
	<title>Continue Watching - Aniways</title>
	<meta name="description" content="Resume watching your anime where you left off" />
</svelte:head>

<div class="min-h-screen bg-background">
	<PageHeader
		title="Continue Watching"
		description="Resume watching your anime where you left off"
	/>

	<div class="container mx-auto px-4 py-8">
		{#if data.listings.items.length === 0}
			<EmptyState
				icon={Play}
				title="No anime to continue watching"
				description="Start watching some anime to see them appear here. You can resume from where you left off."
				action={{
					label: 'Browse Catalog',
					href: '/catalog',
					variant: 'default',
				}}
			/>
		{:else}
			<div
				class="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
			>
				{#each data.listings.items as item (item.id)}
					<AnimeCard
						anime={item.anime}
						episodeLink={item.watchedEpisodes + 1}
						onclick={() =>
							captureResumeClicked({
								anime_id: item.anime.id,
						anime_title: getLocalizedTitles(item.anime, languageValue).main,
								resume_episode: item.watchedEpisodes + 1,
							})}
					/>
				{/each}
			</div>

			{#if data.listings.pageInfo.totalPages > 1}
				<div class="mt-8 flex justify-center">
					<Pagination
						currentPage={data.listings.pageInfo.currentPage}
						totalPages={data.listings.pageInfo.totalPages}
					/>
				</div>
			{/if}
		{/if}
	</div>
</div>
