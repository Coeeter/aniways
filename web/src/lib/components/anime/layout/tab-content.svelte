<script lang="ts">
	import {
		ArrowRight,
		BookOpen,
		Check,
		ChevronDown,
		Film,
		Play,
		Users,
		Heart,
	} from 'lucide-svelte';
	import { flip } from 'svelte/animate';
	import { onNavigate } from '$app/navigation';
	import type { components } from '$lib/api/openapi';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import Input from '$lib/components/ui/input/input.svelte';
	import { language } from '$lib/stores/language';
	import { cn } from '$lib/utils';
	import { getLocalizedTitles } from '$lib/utils/anime-title';

	type AnimeResponse = components['schemas']['models.AnimeWithMetadataResponse'];
	type EpisodeResponse = components['schemas']['models.EpisodeResponse'];
	type RelationsResponse = components['schemas']['models.RelationsResponse'];
	type Character = components['schemas']['models.CharacterResponse'];

	interface Props {
		selectedTab: 'overview' | 'episodes' | 'relations';
		anime: AnimeResponse;
		episodes: EpisodeResponse[];
		franchise: RelationsResponse | null;
		characters: Character[] | null;
	}

	let { selectedTab, anime, episodes, franchise, characters }: Props = $props();
	const languageValue = $derived($language);

	let showFullDescription = $state(false);
	let showAllCharacters = $state(false);
	const onToggleDescription = () => {
		showFullDescription = !showFullDescription;
	};

	const onToggleAllCharacters = () => {
		showAllCharacters = !showAllCharacters;
	};

	onNavigate((params) => {
		if (params.from?.route.id !== '/anime/[id]') return;
		if (params.from?.params?.id !== params.to?.params?.id) {
			selectedTab = 'overview';
			showFullDescription = false;
			showAllCharacters = false;
		}
	});

	let episodesSearch = $state('');
	let sortOrder = $state<'asc' | 'desc'>('asc');
	let filteredEpisodes = $derived.by(() => {
		return episodes
			.filter((ep) => {
				if (!episodesSearch) return true;
				const textToSearch = ep.number.toString() + (ep.title || 'Episode ' + ep.number);
				return textToSearch.toLowerCase().includes(episodesSearch.toLowerCase());
			})
			.sort((a, b) => {
				if (sortOrder === 'asc') {
					return a.number - b.number;
				} else {
					return b.number - a.number;
				}
			});
	});

	let related = $derived.by(() => {
		return franchise?.related?.filter((rel) => rel.id !== anime.id) || [];
	});
</script>

<div class="space-y-8 lg:col-span-2">
	{#if selectedTab === 'overview'}
		{#if anime.metadata?.description}
			<section class="hidden space-y-4 md:block">
				<h3 class="text-xl font-bold">Synopsis</h3>
				<div class="max-w-none">
					<p
						class={cn(
							'leading-relaxed text-muted-foreground',
							!showFullDescription && 'line-clamp-4',
						)}
					>
						{anime.metadata.description}
					</p>
					{#if anime.metadata.description.length > 300}
						<button
							onclick={onToggleDescription}
							class="mt-2 inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline"
						>
							{showFullDescription ? 'Show Less' : 'Read More'}
							<ChevronDown
								class={cn('h-4 w-4 transition-transform', showFullDescription && 'rotate-180')}
							/>
						</button>
					{/if}
				</div>
			</section>
		{/if}

		<section class="space-y-4">
			<h3 class="text-xl font-bold">Main Characters</h3>
			{#if characters && characters.length > 0}
				<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{#each showAllCharacters ? characters : characters.slice(0, 6) as character (character.malId)}
						<a
							href="/characters/{character.malId}"
							class="group flex gap-3 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent"
						>
							<img
								src={character.image}
								alt={character.name}
								class="h-16 w-12 flex-shrink-0 rounded object-cover"
								loading="lazy"
							/>
							<div class="min-w-0 flex-1">
								<h4 class="line-clamp-2 font-semibold group-hover:text-primary">
									{character.name}
								</h4>
								<div class="mt-1 flex items-center gap-2 text-sm text-muted-foreground">
									<Badge
										variant={character.role.toLowerCase() === 'main' ? 'default' : 'outline'}
										class="text-xs"
									>
										{character.role}
									</Badge>
									{#if character.favorites > 0}
										<span class="flex items-center gap-1 text-xs">
											<Heart class="h-3 w-3" />
											{character.favorites.toLocaleString()}
										</span>
									{/if}
								</div>
							</div>
						</a>
					{/each}
				</div>
				{#if characters.length > 6}
					<div class="flex justify-center">
						<Button variant="outline" size="sm" class="gap-2" onclick={onToggleAllCharacters}>
							<Users class="h-4 w-4" />
							{showAllCharacters ? 'Show Less' : `View All Characters (${characters.length})`}
						</Button>
					</div>
				{/if}
			{:else}
				<div class="rounded-lg border bg-muted/30 p-8 text-center">
					<Users class="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
					<p class="text-sm text-muted-foreground">No character information available</p>
				</div>
			{/if}
		</section>
	{:else if selectedTab === 'episodes'}
		<section class="space-y-4">
			<div class="flex flex-col justify-between md:flex-row md:items-center">
				<h3 class="text-xl font-bold">Episodes</h3>

				<div class="flex flex-col gap-2 md:flex-row md:items-center">
					<span class="text-sm whitespace-nowrap text-muted-foreground">
						{episodes.length}
						{episodes.length === 1 ? 'Episode' : 'Episodes'}
					</span>

					<Button
						variant="outline"
						size="icon"
						class="h-9 w-9"
						onclick={() => {
							sortOrder = sortOrder === 'asc' ? 'desc' : 'asc';
						}}
					>
						{sortOrder === 'asc' ? '↑' : '↓'}
					</Button>

					<Input
						type="text"
						placeholder="Search episodes..."
						bind:value={episodesSearch}
						class="mb-2 max-w-xs md:mb-0"
					/>
				</div>
			</div>

			{#if episodes.length > 0}
				<div class="grid gap-3">
					{#each filteredEpisodes as episode (episode.id)}
						<div animate:flip={{ duration: 500 }} class="group">
							<Button
								href="/anime/{anime.id}/watch?ep={episode.number}"
								class="flex h-fit items-center gap-4"
								variant="outline"
							>
								<div
									class="flex aspect-square w-10 flex-shrink-0 items-center justify-center rounded-lg bg-accent"
								>
									<span class="text-sm font-bold text-accent-foreground">{episode.number}</span>
								</div>
								<div class="min-w-0 flex-1">
									<h4 class="line-clamp-1 font-semibold">
										{episode.title || `Episode ${episode.number}`}
									</h4>
									<div class="flex items-center gap-3 text-sm text-muted-foreground">
										{#if episode.isFiller}
											<span
												class="rounded bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground"
											>
												Filler
											</span>
										{/if}
									</div>
								</div>
								<Play
									class="h-5 w-5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
								/>
							</Button>
						</div>
					{/each}
				</div>
				{#if filteredEpisodes.length === 0}
					<div class="rounded-lg border bg-muted/30 p-8 text-center">
						<Film class="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
						<p class="text-sm text-muted-foreground">No episodes match your search...</p>
					</div>
				{/if}
			{:else}
				<div class="rounded-lg border bg-muted/30 p-8 text-center">
					<Film class="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
					<p class="text-sm text-muted-foreground">No episodes available</p>
				</div>
			{/if}
		</section>
	{:else if selectedTab === 'relations'}
		<section class="space-y-6">
			{#if franchise?.watchOrder && franchise.watchOrder.length > 0}
				<div class="space-y-4">
					<div class="flex items-center gap-2">
						<h3 class="text-xl font-bold">Watch Order</h3>
						<span class="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
							{franchise.watchOrder.length}
							{franchise.watchOrder.length === 1 ? 'Entry' : 'Entries'}
						</span>
					</div>
					<div class="space-y-3">
						{#each franchise.watchOrder as relatedAnime, index (relatedAnime.id)}
							{@const isCurrent = relatedAnime.id === anime.id}
							<a
								href={isCurrent ? undefined : `/anime/${relatedAnime.id}`}
								class={cn(
									'group flex items-center gap-4 rounded-lg border p-4 transition-all',
									isCurrent
										? 'cursor-default border-primary bg-primary/10'
										: 'bg-card hover:border-primary/50 hover:bg-accent',
								)}
							>
								<div
									class={cn(
										'flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-full text-sm font-bold',
										isCurrent ? 'bg-primary text-primary-foreground' : 'bg-primary/10 text-primary',
									)}
								>
									{index + 1}
								</div>
								<img
									src={relatedAnime.imageUrl}
									alt={getLocalizedTitles(relatedAnime, languageValue).main}
									class="h-20 w-14 rounded object-cover"
								/>
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<h4
											class={cn(
												'line-clamp-1 font-semibold',
												!isCurrent && 'group-hover:text-primary',
											)}
										>
											{getLocalizedTitles(relatedAnime, languageValue).main}
										</h4>
										{#if isCurrent}
											<span
												class="rounded bg-primary px-2 py-0.5 text-xs font-medium text-primary-foreground"
											>
												Currently Viewing
											</span>
										{/if}
									</div>
									{#if getLocalizedTitles(relatedAnime, languageValue).sub}
										<p class="line-clamp-1 text-sm text-muted-foreground">
											{getLocalizedTitles(relatedAnime, languageValue).sub}
										</p>
									{/if}
									<div class="mt-1 flex items-center gap-3 text-xs text-muted-foreground">
										<span class="capitalize">{relatedAnime.season} {relatedAnime.seasonYear}</span>
										{#if relatedAnime.lastEpisode}
											<span>• {relatedAnime.lastEpisode} Episodes</span>
										{/if}
									</div>
								</div>
								{#if !isCurrent}
									<ArrowRight
										class="h-5 w-5 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
									/>
								{:else}
									<Check class="h-5 w-5 text-primary" />
								{/if}
							</a>
						{/each}
					</div>
				</div>
			{/if}

			{#if related && related.length > 0}
				<div class="space-y-4">
					<h3 class="text-xl font-bold">Related Anime</h3>
					<div class="grid gap-4 sm:grid-cols-2">
						{#each related as relatedAnime (relatedAnime.id)}
							{@const isCurrent = relatedAnime.id === anime.id}
							<a
								href={isCurrent ? undefined : `/anime/${relatedAnime.id}`}
								class={cn(
									'group flex gap-4 rounded-lg border p-4 transition-all',
									isCurrent
										? 'relative cursor-default border-primary bg-primary/10'
										: 'bg-card hover:border-primary/50 hover:bg-accent',
								)}
							>
								{#if isCurrent}
									<div class="absolute top-2 right-2">
										<span
											class="rounded bg-primary px-2 py-0.5 text-xs font-medium text-primary-foreground"
										>
											Current
										</span>
									</div>
								{/if}
								<img
									src={relatedAnime.imageUrl}
									alt={getLocalizedTitles(relatedAnime, languageValue).main}
									class="h-24 w-16 rounded object-cover"
								/>
								<div class="min-w-0 flex-1">
									<h4
										class={cn(
											'line-clamp-2 font-semibold',
											!isCurrent && 'group-hover:text-primary',
										)}
									>
										{getLocalizedTitles(relatedAnime, languageValue).main}
									</h4>
									<div class="mt-1 flex flex-wrap gap-2 text-xs text-muted-foreground">
										<span class="capitalize">{relatedAnime.season} {relatedAnime.seasonYear}</span>
									</div>
									<div class="mt-2 flex flex-wrap gap-1">
										{#each relatedAnime.genre.split(', ').slice(0, 2) as genre (genre)}
											<span class="rounded bg-muted px-2 py-0.5 text-xs">
												{genre}
											</span>
										{/each}
									</div>
								</div>
							</a>
						{/each}
					</div>
				</div>
			{/if}

			{#if !franchise?.watchOrder?.length && !related?.length}
				<div class="rounded-lg border bg-muted/30 p-8 text-center">
					<BookOpen class="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
					<p class="text-sm text-muted-foreground">No related anime found</p>
				</div>
			{/if}
		</section>
	{/if}
</div>
