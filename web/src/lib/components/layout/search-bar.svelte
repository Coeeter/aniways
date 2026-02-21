<script lang="ts">
	import { LoaderCircle, Search } from 'lucide-svelte';
	import { Debounced, resource, watch } from 'runed';
	import { onNavigate } from '$app/navigation';
	import { apiClient } from '$lib/api/client';
	import { captureSearch } from '$lib/analytics';
	import { Button } from '../ui/button';
	import * as Command from '../ui/command';

	let isSearchOpen = $state(false);
	let rawQuery = $state('');
	const debouncedQuery = new Debounced(() => rawQuery, 700);

	const trendingResource = resource(
		() => [],
		async () => {
			const response = await apiClient.GET('/anime/listings/trending');
			return response.data || [];
		},
		{ once: true },
	);

	let searchResource = resource(
		() => debouncedQuery.current,
		async (query, _, { signal }) => {
			if (!query || query.length < 3) return trendingResource.current || [];
			const response = await apiClient.GET('/anime/listings/search', {
				params: {
					query: {
						q: query,
						itemsPerPage: 5,
					},
				},
				signal,
			});
			const items = response.data?.items || [];
			captureSearch({ query, result_count: items.length, navigated_to_catalog: false });
			return items;
		},
		{
			debounce: 0,
			initialValue: trendingResource.current || [],
		},
	);

	onNavigate(() => {
		isSearchOpen = false;
	});

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'k' && (event.metaKey || event.ctrlKey)) {
			event.preventDefault();
			isSearchOpen = !isSearchOpen;
		}
	}

	watch(
		() => trendingResource.current,
		() => {
			if (rawQuery.length === 0) {
				searchResource.refetch();
			}
		},
	);

	watch(
		() => isSearchOpen,
		() => {
			if (isSearchOpen) return;
			rawQuery = '';
		},
	);

	watch(
		() => rawQuery,
		() => {
			if (rawQuery.length === 0) {
				debouncedQuery.setImmediately('');
			}
		},
	);
</script>

<svelte:window on:keydown={handleKeydown} />

<Button
	variant="ghost"
	class="flex w-10 items-center justify-center lg:hidden"
	onclick={() => (isSearchOpen = true)}
	aria-label="Open search"
>
	<Search class="h-5 w-5" />
</Button>

<Button
	variant="outline"
	class="hidden w-64 items-center justify-start space-x-2 border-muted-foreground/20 bg-transparent text-muted-foreground hover:border-primary/50 lg:flex"
	onclick={() => (isSearchOpen = true)}
>
	<Search class="h-4 w-4" />
	<span>Search anime...</span>
	<kbd
		class="pointer-events-none ml-auto inline-flex h-5 items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100 select-none"
	>
		<span class="text-xs">⌘</span>K
	</kbd>
</Button>

<Command.Dialog bind:open={isSearchOpen} shouldFilter={false}>
	<Command.Input placeholder="Search anime..." bind:value={rawQuery} />
	<Command.List class="max-h-[400px]">
		{#if rawQuery.length < 3 && rawQuery.length > 0}
			<Command.Empty>Please enter at least 3 characters to search.</Command.Empty>
		{:else if rawQuery.length !== 0 && (searchResource.loading || debouncedQuery.pending)}
			<Command.Empty class="flex w-full items-center justify-center">
				<LoaderCircle class="mr-2 inline-block h-4 w-4 animate-spin text-primary" />
				Searching...
			</Command.Empty>
		{:else}
			<Command.Group heading={rawQuery.length >= 3 ? 'Results' : 'Trending Anime'}>
				{#if searchResource.current.length === 0}
					<Command.Empty>No results found.</Command.Empty>
				{:else}
					{#each searchResource.current as anime (anime.id)}
						<Command.LinkItem
							value={anime.ename || anime.jname}
							class="flex cursor-pointer items-center gap-3 p-3 hover:bg-accent"
							href="/anime/{anime.id}"
						>
							<div class="relative h-16 w-12 flex-shrink-0 overflow-hidden rounded-md">
								<img
									src={anime.imageUrl}
									alt={anime.jname || anime.ename}
									class="h-full w-full object-cover"
								/>
							</div>
							<div class="min-w-0 flex-1">
								<div class="line-clamp-1 font-medium">
									{anime.jname || anime.ename}
								</div>
								{#if anime.jname && anime.ename}
									<div class="line-clamp-1 text-sm text-muted-foreground">
										{anime.ename}
									</div>
								{/if}
								<div class="flex items-center gap-2 text-xs text-muted-foreground">
									<span class="capitalize">{anime.season} {anime.seasonYear}</span>
									{#if anime.genre}
										<span>•</span>
										<span class="line-clamp-1"
											>{anime.genre.split(', ').slice(0, 2).join(', ')}</span
										>
									{/if}
								</div>
							</div>
						</Command.LinkItem>
					{/each}
				{/if}

				{#if rawQuery.length >= 3}
					{#if searchResource.current.length > 0}
						<Command.Separator />
					{/if}
					<Command.LinkItem
						value="view-all-{rawQuery}"
						class="flex cursor-pointer items-center justify-center gap-2 p-3 font-medium text-primary hover:bg-accent"
						href="/catalog?search={rawQuery}"
						onclick={() =>
							captureSearch({
								query: rawQuery,
								result_count: searchResource.current.length,
								navigated_to_catalog: true,
							})}
					>
						<Search class="h-4 w-4" />
						{searchResource.current.length > 0
							? `View all results for "${rawQuery}"`
							: `Search for "${rawQuery}" in catalog`}
					</Command.LinkItem>
				{/if}
			</Command.Group>
		{/if}
	</Command.List>
</Command.Dialog>
