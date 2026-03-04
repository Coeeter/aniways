<script lang="ts">
	import { formatDate } from 'date-fns';
	import { ExternalLink, TrendingUp } from 'lucide-svelte';
	import type { components } from '$lib/api/openapi';
	import Button from '$lib/components/ui/button/button.svelte';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import { cn } from '$lib/utils';

	type AnimeResponse = components['schemas']['models.AnimeWithMetadataResponse'];
	interface Props {
		anime: AnimeResponse;
		ratingLabel: string;
		selectedTab?: string;
		variations?: AnimeResponse[];
	}

	let { anime, ratingLabel, selectedTab, variations = [] }: Props = $props();
	const languageValue = $derived($language);
	let mediaType = $derived.by(() => {
		const type = anime.metadata?.mediaType || 'tv';
		return type === 'tv'
			? 'TV'
			: type
					.split('_')
					.map((w) => w.charAt(0).toUpperCase() + w.slice(1))
					.join(' ');
	});
</script>

<div class={cn('sticky top-36 h-fit space-y-6', selectedTab !== 'overview' && 'hidden md:block')}>
	<div class="rounded-xl border bg-card p-6 shadow-sm">
		<h3 class="mb-4 text-lg font-bold">Information</h3>
		<dl class="space-y-3 text-sm">
			<div class="border-b pb-3">
				<dt class="mb-2 font-semibold text-muted-foreground">Media</dt>
				<div class="space-y-2">
					<div class="flex justify-between">
						<span class="text-muted-foreground">Type</span>
						<span class="font-medium">{mediaType}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">Episodes</span>
						<span class="font-medium">{anime.metadata?.totalEpisodes || '???'}</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">Duration</span>
						<span class="font-medium">
							{#if !anime.metadata?.avgEpisodeDuration}
								???
							{:else if anime.metadata.avgEpisodeDuration < 60}
								{anime.metadata.avgEpisodeDuration} sec
							{:else if anime.metadata.avgEpisodeDuration < 3600}
								{Math.floor(anime.metadata.avgEpisodeDuration / 60)} min {anime.metadata
									.avgEpisodeDuration % 60} sec
							{:else}
								{Math.floor(anime.metadata.avgEpisodeDuration / 3600)} hr
							{/if}
						</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">Status</span>
						<span class={cn('font-medium')}>
							{#if !anime.metadata?.airingStatus}
								???
							{:else}
								{anime.metadata.airingStatus
									.replace(/_/g, ' ')
									.replace(/\b\w/g, (l) => l.toUpperCase())}
							{/if}
						</span>
					</div>
				</div>
			</div>

			<div class="border-b pb-3">
				<dt class="mb-2 font-semibold text-muted-foreground">Airing</dt>
				<div class="space-y-2">
					<div class="flex justify-between">
						<span class="text-muted-foreground">Season</span>
						<span class="font-medium capitalize">
							{#if anime.metadata?.season && anime.metadata?.seasonYear}
								{anime.metadata.season} {anime.metadata.seasonYear}
							{:else if anime.metadata?.season}
								{anime.metadata.season} ???
							{:else if anime.metadata?.seasonYear}
								??? {anime.metadata.seasonYear}
							{:else}
								???
							{/if}
						</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">Start Date</span>
						<span class="font-medium">
							{#if anime.metadata?.airingStartDate}
								{formatDate(anime.metadata.airingStartDate, 'dd/MM/yyyy')}
							{:else}
								???
							{/if}
						</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">End Date</span>
						<span class="font-medium">
							{#if !anime.metadata?.airingEndDate}
								???
							{:else}
								{formatDate(anime.metadata.airingEndDate, 'dd/MM/yyyy')}
							{/if}
						</span>
					</div>
				</div>
			</div>

			<div class="border-b pb-3">
				<dt class="mb-2 font-semibold text-muted-foreground">Production</dt>
				<div class="space-y-2">
					<div class="flex justify-between">
						<span class="text-muted-foreground">Studio</span>
						<span class="font-medium capitalize">
							{#if anime.metadata?.studio}
								{anime.metadata.studio}
							{:else}
								???
							{/if}
						</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">Source</span>
						<span class="font-medium capitalize">
							{#if anime.metadata?.source}
								{anime.metadata.source.replace(/_/g, ' ').toLowerCase()}
							{:else}
								???
							{/if}
						</span>
					</div>
					<div class="flex justify-between">
						<span class="text-muted-foreground">Rating</span>
						<span class="font-medium">{ratingLabel}</span>
					</div>
				</div>
			</div>

			<div class="border-b pb-3 md:hidden">
				<dt class="mb-2 font-semibold text-muted-foreground">Synopsis</dt>
				<div class="space-y-2">
					{#if anime.metadata?.description}
						<p class="text-sm text-muted-foreground">{anime.metadata.description}</p>
					{:else}
						<p class="text-sm text-muted-foreground italic">No synopsis available.</p>
					{/if}
				</div>
			</div>

			{#if anime.malId || anime.anilistId}
				<div>
					<dt class="mb-2 font-semibold text-muted-foreground">External Links</dt>
					<div class="flex gap-2">
						{#if anime.malId}
							<Button
								size="sm"
								variant="outline"
								href={`https://myanimelist.net/anime/${anime.malId}`}
								target="_blank"
								class="gap-2"
							>
								<ExternalLink class="h-3 w-3" />
								MAL
							</Button>
						{/if}
						{#if anime.anilistId}
							<Button
								size="sm"
								variant="outline"
								href={`https://anilist.co/anime/${anime.anilistId}`}
								target="_blank"
								class="gap-2"
							>
								<ExternalLink class="h-3 w-3" />
								AniList
							</Button>
						{/if}
					</div>
				</div>
			{/if}
		</dl>
	</div>

	{#if variations && variations.length > 0}
		<div class="rounded-xl border bg-card p-6 shadow-sm">
			<h3 class="mb-4 text-lg font-bold">Other Versions</h3>
			<div class="space-y-2">
				{#each variations as variation (variation.id)}
					{@const isCurrent = variation.id === anime.id}
					<a
						href={isCurrent ? undefined : `/anime/${variation.id}`}
						class={cn(
							'group flex items-center gap-3 rounded-lg border p-3 transition-all',
							isCurrent
								? 'cursor-default border-primary bg-primary/10'
								: 'bg-background hover:border-primary/50 hover:bg-accent',
						)}
					>
						<img
							src={variation.imageUrl}
							alt={getLocalizedTitles(variation, languageValue).main}
							class="h-16 w-12 rounded object-cover"
						/>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<h4
									class={cn(
										'line-clamp-1 text-sm font-semibold',
										!isCurrent && 'group-hover:text-primary',
									)}
								>
								{getLocalizedTitles(variation, languageValue).main}
								</h4>
								{#if isCurrent}
									<span
										class="rounded bg-primary px-2 py-0.5 text-xs font-medium text-primary-foreground"
									>
										Current
									</span>
								{/if}
							</div>
							{#if getLocalizedTitles(variation, languageValue).sub}
								<p class="line-clamp-1 text-xs text-muted-foreground">
									{getLocalizedTitles(variation, languageValue).sub}
								</p>
							{/if}
						</div>
						{#if !isCurrent}
							<ExternalLink
								class="h-4 w-4 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100"
							/>
						{/if}
					</a>
				{/each}
			</div>
		</div>
	{/if}

	<div class="rounded-xl border bg-card p-6 shadow-sm">
		<h3 class="mb-4 text-lg font-bold">You Might Also Like</h3>
		<div class="text-center text-sm text-muted-foreground">
			<TrendingUp class="mx-auto mb-2 h-8 w-8" />
			<p>Recommendations coming soon</p>
		</div>
	</div>
</div>
