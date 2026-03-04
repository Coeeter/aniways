<script lang="ts">
	import { Play } from 'lucide-svelte';
	import type { Snippet } from 'svelte';
	import type { components } from '$lib/api/openapi';
	import LibraryBtn from '$lib/components/anime/controls/library-btn.svelte';
	import { language } from '$lib/stores/language';
	import { cn } from '$lib/utils';
	import { getLocalizedTitles } from '$lib/utils/anime-title';

	type AnimeResponse = components['schemas']['models.AnimeResponse'];
	type LibraryResponse = components['schemas']['models.LibraryResponse'];

	interface Props {
		anime: AnimeResponse;
		topLeftBadge?: Snippet;
		libraryEntry?: LibraryResponse | null;
		episodeLink?: number | null;
		class?: string;
		onclick?: () => void;
	}

	let {
		anime,
		topLeftBadge,
		libraryEntry,
		episodeLink = null,
		class: className,
		onclick: onCardClick,
	}: Props = $props();
	let linkUrl = $derived(
		episodeLink ? `/anime/${anime.id}/watch?ep=${episodeLink}` : `/anime/${anime.id}`,
	);
	let titles = $derived.by(() => getLocalizedTitles(anime, $language));
</script>

<a
	href={linkUrl}
	class={cn(
		'group block transform transition-all duration-500 hover:-translate-y-2 hover:scale-105',
		className,
	)}
	onclick={(e) => {
		if ((e.target as HTMLElement).closest('button')) {
			e.preventDefault();
			return;
		}
		onCardClick?.();
	}}
>
	<div
		class="relative mb-4 aspect-[3/4] overflow-hidden rounded-xl bg-gradient-to-br from-muted to-muted/50 shadow-lg transition-all duration-500 group-hover:shadow-2xl group-hover:shadow-primary/20"
	>
		<img
			src={anime.imageUrl}
			alt={titles.main}
			class="h-full w-full object-cover transition-all duration-700 group-hover:scale-110"
		/>

		<div
			class="absolute inset-0 bg-gradient-to-t from-black/50 via-black/20 to-transparent transition-opacity duration-300 group-hover:opacity-0"
		></div>
		<div
			class="absolute inset-0 bg-gradient-to-t from-black/80 via-black/50 to-black/20 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
		></div>

		{#if topLeftBadge}
			<div class="absolute top-3 left-3">
				{@render topLeftBadge()}
			</div>
		{/if}

		{#if anime.lastEpisode}
			<div class="absolute top-3 right-3">
				<span
					class="rounded-md bg-primary/90 px-2 py-1 text-xs font-semibold text-primary-foreground backdrop-blur-sm"
				>
					EP {anime.lastEpisode}
				</span>
			</div>
		{/if}

		{#if libraryEntry}
			<div class="absolute top-3 left-3 z-10">
				<LibraryBtn animeId={anime.id} {libraryEntry} iconOnly={true} />
			</div>
		{/if}

		<div
			class="absolute right-3 bottom-3 left-3 translate-y-2 opacity-0 transition-all duration-300 group-hover:translate-y-0 group-hover:opacity-100"
		>
			<div class="flex gap-1">
				{#each anime.genre.split(', ').slice(0, 2) as genre (genre)}
					<span class="rounded-full bg-white/20 px-2 py-0.5 text-xs text-white backdrop-blur-sm">
						{genre}
					</span>
				{/each}
			</div>
		</div>

		<div
			class="absolute inset-0 flex items-center justify-center opacity-0 transition-opacity duration-300 group-hover:opacity-100"
		>
			<div class="rounded-full border border-white/30 bg-white/20 p-3">
				<Play class="h-6 w-6 text-white" />
			</div>
		</div>
	</div>

	<div class="space-y-1">
		<h3
			class="line-clamp-1 text-sm font-semibold transition-colors duration-300 group-hover:text-primary"
		>
			{titles.main}
		</h3>
		{#if episodeLink}
			<div class="text-xs text-muted-foreground">
				<span class="font-medium text-primary">Episode {episodeLink}</span>
			</div>
		{:else if libraryEntry}
			<div class="text-xs text-muted-foreground">
				<span class="capitalize">{libraryEntry.status.replace('_', ' ')}</span>
				<span class="font-bold">{libraryEntry.watchedEpisodes}</span> of
				<span class="font-bold">{anime.lastEpisode ?? '???'}</span>
				{anime.lastEpisode === 1 ? 'episode' : 'episodes'}
			</div>
		{:else}
			<div class="text-xs text-muted-foreground">
				<span class="capitalize">{anime.season} {anime.seasonYear}</span>
			</div>
		{/if}
	</div>
</a>
