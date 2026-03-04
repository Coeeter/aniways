<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { components } from '$lib/api/openapi';
	import { Badge } from '$lib/components/ui/badge';
	import { Pagination } from '$lib/components/ui/pagination';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import type { FilterManager } from '$lib/utils/filter-manager.svelte';
	import LibraryBtn from '../controls/library-btn.svelte';
	import AnimeCard from './anime-card.svelte';

	type AnimeResponse = components['schemas']['models.AnimeResponse'];
	type AnimeWithLibraryResponse = components['schemas']['models.AnimeWithLibraryResponse'];

	interface Props {
		anime: AnimeResponse[] | AnimeWithLibraryResponse[];
		totalPages: number;
		children?: Snippet;
		empty: Snippet;
		filterManager: FilterManager;
	}

	let { anime, totalPages, children, empty, filterManager }: Props = $props();
	const languageValue = $derived($language);
</script>

<main class="flex-1">
	{#if children}
		{@render children()}
	{/if}

	<div class="px-4">
		{#if filterManager.isLoading}
			<div
				class={filterManager.viewMode === 'grid'
					? 'grid grid-cols-2 gap-3 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6'
					: 'space-y-3 sm:space-y-4'}
			>
				{#each Array(filterManager.filters.itemsPerPage) as _, i (i)}
					<div class="space-y-2">
						<Skeleton class="aspect-[3/4] rounded-xl" />
						<Skeleton class="h-4 w-3/4" />
						<Skeleton class="h-3 w-1/2" />
					</div>
				{/each}
			</div>
		{:else if anime && anime.length > 0}
			{#if filterManager.viewMode === 'grid'}
				<div
					class="grid grid-cols-2 gap-3 duration-500 sm:grid-cols-3 sm:gap-4 md:grid-cols-4 xl:grid-cols-5 2xl:grid-cols-6"
				>
					{#each anime as a (a.id)}
						<AnimeCard
							anime={a}
							libraryEntry={'library' in a && a.library
								? {
										id: a.library.id,
										anime: a,
										status: a.library.status,
										animeId: a.id,
										createdAt: '',
										updatedAt: a.library.updatedAt,
										userId: '',
										watchedEpisodes: a.library.watchedEpisodes,
									}
								: undefined}
						/>
					{/each}
				</div>
			{:else}
				<div class="space-y-4">
					{#each anime as a (a.id)}
						<a
							href="/anime/{a.id}"
							class="flex gap-3 rounded-lg border bg-card p-3 transition hover:border-primary hover:shadow-lg sm:p-4"
							onclick={(e) => {
								if ((e.target as HTMLElement).closest('button')) {
									e.preventDefault();
								}
							}}
						>
							<img
								src={a.imageUrl}
								alt={getLocalizedTitles(a, languageValue).main}
								class="h-24 w-20 rounded-md object-cover sm:h-32 sm:w-24"
							/>
							<div class="flex-1 space-y-2">
								<div>
									<h3 class="line-clamp-2 text-sm font-semibold sm:text-lg">
										{getLocalizedTitles(a, languageValue).main}
									</h3>
									{#if getLocalizedTitles(a, languageValue).sub}
										<p class="text-sm text-muted-foreground">
											{getLocalizedTitles(a, languageValue).sub}
										</p>
									{/if}
								</div>
								<div class="flex flex-wrap gap-2">
									{#each a.genre.split(', ').slice(0, 5) as genre (genre)}
										<Badge variant="secondary" class="text-xs">{genre}</Badge>
									{/each}
								</div>
								<div class="flex gap-4 text-sm text-muted-foreground">
									<span class="capitalize">{a.season} {a.seasonYear}</span>
									{#if a.lastEpisode}
										<span>Episode {a.lastEpisode}</span>
									{/if}
								</div>
								{#if 'library' in a && a.library}
									<LibraryBtn
										animeId={a.id}
										libraryEntry={{
											id: a.library.id,
											anime: a,
											status: a.library.status,
											animeId: a.id,
											createdAt: '',
											updatedAt: a.library.updatedAt,
											userId: '',
											watchedEpisodes: a.library.watchedEpisodes,
										}}
									/>
								{/if}
							</div>
						</a>
					{/each}
				</div>
			{/if}

			<Pagination
				{totalPages}
				currentPage={filterManager.filters.page}
				onPageChange={filterManager.setPage}
			/>
		{:else}
			{@render empty()}
		{/if}
	</div>
</main>
