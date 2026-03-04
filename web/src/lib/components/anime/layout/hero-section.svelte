<script lang="ts">
	import { CirclePlay, Play, Share2, Star } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import type { components } from '$lib/api/openapi';
	import LibraryBtn from '$lib/components/anime/controls/library-btn.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import { Dialog, DialogContent, DialogHeader, DialogTitle } from '$lib/components/ui/dialog';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import { cn } from '$lib/utils';

	type AnimeResponse = components['schemas']['models.AnimeWithMetadataResponse'];
	type LibraryReponse = components['schemas']['models.LibraryResponse'];
	type AnimeVariation = components['schemas']['models.AnimeResponse'];

	interface Props {
		anime: AnimeResponse;
		banner: string | null;
		trailer: string | null;
		ratingLabel: string;
		totalEpisodes?: number;
		libraryEntry: LibraryReponse | null;
		variations?: AnimeVariation[];
	}

	let {
		anime,
		banner,
		trailer,
		totalEpisodes,
		ratingLabel,
		libraryEntry,
		variations = [],
	}: Props = $props();
	const languageValue = $derived($language);
	const titles = $derived.by(() => getLocalizedTitles(anime, languageValue));

	let mediaType = $derived.by(() => {
		const type = anime.metadata?.mediaType || 'tv';
		return type === 'tv'
			? 'TV'
			: type
					.split('_')
					.map((w) => w.charAt(0).toUpperCase() + w.slice(1))
					.join(' ');
	});

	let isTrailerDialogOpen = $state(false);
</script>

<section class="relative -mt-32 bg-background md:-mt-17">
	<div class="relative h-[400px] overflow-hidden">
		{#if banner}
			<img
				src={banner}
				alt="Banner"
				class="absolute inset-0 h-full w-full object-cover object-top"
			/>
		{:else}
			<img
				src={anime.metadata?.mainPictureUrl || anime.imageUrl}
				alt="Banner"
				class="absolute inset-0 h-full w-full object-cover opacity-50 blur-lg"
			/>
		{/if}
		<div
			class="absolute inset-0 bg-gradient-to-t from-background via-background/50 to-transparent"
		></div>
	</div>

	<div class="container mx-auto px-4 pb-4 md:pb-8">
		<div class="flex flex-col items-center gap-8 md:flex-row md:items-start">
			<div class="flex-shrink-0">
				<div class="group relative -mt-48 overflow-hidden rounded-lg shadow-2xl md:-mt-24">
					<img
						src={anime.metadata?.mainPictureUrl || anime.imageUrl}
					alt={titles.main}
						class="aspect-[2/3] w-48 object-cover lg:w-56"
					/>
					{#if trailer}
						<button
							onclick={() => (isTrailerDialogOpen = true)}
							class="absolute inset-0 flex cursor-pointer flex-col items-center justify-center bg-black/60 opacity-0 transition-opacity duration-300 group-hover:opacity-100"
						>
							<CirclePlay class="h-12 w-12 text-white" />
							<span class="mt-2 text-sm font-medium text-white">Watch Trailer</span>
						</button>
					{/if}
				</div>
			</div>

			<div class="flex-1 space-y-6 md:space-y-4">
				<div>
					<h1 class="text-3xl font-bold lg:text-4xl">
						{titles.main}
					</h1>
					{#if titles.sub}
						<h2 class="mt-1 text-lg text-muted-foreground">
							{titles.sub}
						</h2>
					{/if}
				</div>

				<div class="flex flex-wrap items-center gap-2 text-sm md:gap-3">
					<div class="flex w-full items-center gap-1 md:w-fit">
						<Star class="h-4 w-4 fill-yellow-500 text-yellow-500" />
						<span class={cn(anime.metadata?.mean ? 'font-semibold' : 'text-muted-foreground')}>
							{#if anime.metadata?.mean}
								{anime.metadata.mean.toFixed(1)}
							{:else}
								N/A
							{/if}
						</span>
					</div>
					<span class="hidden text-muted-foreground md:inline">|</span>
					<span class="text-muted-foreground">{mediaType}</span>
					<span class="text-muted-foreground">|</span>
					{#if anime.metadata?.totalEpisodes}
						<span class="text-muted-foreground">{anime.metadata.totalEpisodes} Episodes</span>
						<span class="text-muted-foreground">|</span>
					{/if}
					<span class="text-muted-foreground capitalize">{anime.season} {anime.seasonYear}</span>
					<span class="text-muted-foreground">|</span>
					<span class="text-muted-foreground">{ratingLabel}</span>
				</div>

				{#if anime.genre}
					<div class="flex flex-wrap gap-2">
						{#each anime.genre.split(', ') as genre (genre)}
							<Button size="sm" variant="outline" href="/catalog?genres={genre}">
								{genre}
							</Button>
						{/each}
					</div>
				{/if}

				<div class="grid grid-cols-2 gap-3 pt-2 md:flex md:flex-wrap">
					{#if totalEpisodes && totalEpisodes > 0}
						<Button size="default" class="gap-2" href="/anime/{anime.id}/watch">
							<Play class="h-4 w-4" />
							Start Watching
						</Button>
					{/if}

					<LibraryBtn
						{libraryEntry}
						animeId={anime.id}
						{variations}
						currentAnimeName={titles.main || 'Current Version'}
					/>

					{#if trailer}
						<Button
							size="default"
							variant="outline"
							class="col-span-2 gap-2 md:col-auto"
							onclick={() => (isTrailerDialogOpen = true)}
						>
							<CirclePlay class="h-4 w-4" />
							Trailer
						</Button>
					{/if}

					<Button
						size="default"
						variant="ghost"
						class="hidden gap-2 md:inline-flex"
						onclick={() => {
							navigator.clipboard.writeText(window.location.href);
							toast.success('Link copied to clipboard');
						}}
					>
						<Share2 class="h-4 w-4" />
					</Button>
				</div>

				{#if anime.metadata}
					<div class="hidden flex-wrap gap-6 border-t pt-4 text-sm md:flex">
						{#if anime.metadata.rank}
							<div>
								<span class="text-muted-foreground">Ranked</span>
								<p class="font-semibold">#{anime.metadata.rank}</p>
							</div>
						{/if}
						{#if anime.metadata.popularity}
							<div>
								<span class="text-muted-foreground">Popularity</span>
								<p class="font-semibold">#{anime.metadata.popularity}</p>
							</div>
						{/if}
						{#if anime.metadata.scoringUsers}
							<div>
								<span class="text-muted-foreground">Members</span>
								<p class="font-semibold">{anime.metadata.scoringUsers.toLocaleString()}</p>
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	</div>
</section>

<Dialog bind:open={isTrailerDialogOpen}>
	<DialogContent class="max-w-none sm:max-w-none">
		<DialogHeader>
			<DialogTitle>
				Trailer - {titles.main}
			</DialogTitle>
		</DialogHeader>
		{#if trailer}
			<div class="aspect-video w-full">
				<iframe
					src={trailer.replace('watch?v=', 'embed/')}
					title="YouTube video player"
					frameborder="0"
					allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
					allowfullscreen
					class="h-full w-full"
				></iframe>
			</div>
		{:else}
			<p>Trailer not available.</p>
		{/if}
	</DialogContent>
</Dialog>
