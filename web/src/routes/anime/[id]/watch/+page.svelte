<script lang="ts">
	import {
		ArrowLeft,
		ChevronLeft,
		ChevronRight,
		CircleAlert,
		Film,
		Info,
		LoaderCircle,
		Play,
		RefreshCcw,
		Server,
		Star,
	} from 'lucide-svelte';
	import { resource } from 'runed';
	import { toast } from 'svelte-sonner';
	import { flip } from 'svelte/animate';
	import { invalidate } from '$app/navigation';
	import { captureServerSwitch, captureStreamError } from '$lib/analytics';
	import { apiClient } from '$lib/api/client';
	import LibraryBtn from '$lib/components/anime/controls/library-btn.svelte';
	import Player from '$lib/components/anime/player/index.svelte';
	import { Badge } from '$lib/components/ui/badge';
	import Button from '$lib/components/ui/button/button.svelte';
	import Input from '$lib/components/ui/input/input.svelte';
	import { getAppStateContext } from '$lib/context/state.svelte';
	import { language } from '$lib/stores/language';
	import { cn } from '$lib/utils';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	const appState = getAppStateContext();
	const languageValue = $derived($language);
	const titles = $derived.by(() => getLocalizedTitles(data.anime, languageValue));

	let selectedServer = $derived(data.servers[0] || null);

	// Track stream errors once per error instance
	let lastTrackedError = $state<string | null>(null);
	$effect(() => {
		const err = streamResource.error;
		if (!err) {
			lastTrackedError = null;
			return;
		}
		const key = `${selectedServer?.serverId}:${err.message}`;
		if (key === lastTrackedError) return;
		lastTrackedError = key;
		captureStreamError({
			anime_id: data.anime.id,
			episode_number: data.episodeNumber,
			server_name: selectedServer?.serverName ?? 'unknown',
			stream_type: selectedServer?.type?.toLowerCase() ?? 'unknown',
			error_message: err.message,
		});
	});

	const streamResource = resource(
		[() => data.anime.id, () => selectedServer?.serverId, () => selectedServer?.type],
		async ([animeId, serverId, type], _, { signal }) => {
			const server = data.servers.find((s) => s.serverId === serverId && s.type === type);
			if (!server) return null;

			const res = await apiClient.GET('/anime/{id}/episodes/servers/{serverID}', {
				params: {
					path: {
						id: animeId,
						serverID: server.serverId,
					},
					query: {
						server: server.serverName.toLowerCase().replace(/\s+/g, '-'),
						type: server.type.toLowerCase(),
					},
				},
				signal,
			});

			if (!res.data) throw new Error('No stream data received from server');
			return res.data;
		},
	);

	let episodesSearch = $state('');
	let filteredEpisodes = $derived.by(() => {
		if (!episodesSearch) return data.episodes;
		return data.episodes.filter((ep) => {
			const s = (ep.number.toString() + (ep.title || `Episode ${ep.number}`)).toLowerCase();
			return s.includes(episodesSearch.toLowerCase());
		});
	});

	let groupedServers = $derived.by(() => {
		return Object.fromEntries(
			data.servers
				.map((s) => s.type.toLowerCase())
				.filter((v, i, a) => a.indexOf(v) === i) // remove duplicates
				.map((type) => [type, data.servers.filter((s) => s.type.toLowerCase() === type)]),
		);
	});

	const episodeUrl = (epNum: number) => `/anime/${data.anime.id}/watch?ep=${epNum}`;
	let prevEpisodeUrl = $derived(data.episodeNumber > 1 ? episodeUrl(data.episodeNumber - 1) : null);
	let nextEpisodeUrl = $derived(
		data.episodeNumber < data.episodes.length ? episodeUrl(data.episodeNumber + 1) : null,
	);

	const updateLibrary = async () => {
		if (!appState.isLoggedIn) return;
		if (data.libraryEntry && data.episodeNumber <= data.libraryEntry.watchedEpisodes) return;

		const id = toast.loading('Updating library status...');
		try {
			await apiClient.PUT('/library/{animeID}', {
				params: { path: { animeID: data.anime.id } },
				body: { watchedEpisodes: data.episodeNumber, status: 'watching' },
			});

			await invalidate('app:library');
			toast.success('Library status updated');
		} catch {
			toast.error('Failed to update library status');
		} finally {
			toast.dismiss(id);
		}
	};
</script>

<svelte:head>
	<title>
		{titles.main} - Episode {data.episodeNumber} - Aniways
	</title>
	<meta name="description" content="Watch {titles.main} Episode {data.episodeNumber} on Aniways" />
</svelte:head>

<div class="min-h-screen bg-background">
	<header
		class="sticky top-0 z-40 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60"
	>
		<div class="container mx-auto flex h-16 items-center gap-4 px-4">
			<Button href="/anime/{data.anime.id}" variant="ghost" size="icon">
				<ArrowLeft class="h-5 w-5" />
			</Button>
			<div class="min-w-0 flex-1">
				<h1 class="line-clamp-1 text-lg font-semibold">
					{titles.main}
				</h1>
				<p class="text-sm text-muted-foreground">
					Episode {data.episodeNumber}: {data.currentEpisode.title ||
						`Episode ${data.episodeNumber}`}
				</p>
			</div>
			<Button
				href="/anime/{data.anime.id}"
				variant="outline"
				size="sm"
				class="hidden md:inline-flex"
			>
				<Info class="mr-2 h-4 w-4" />
				Anime Details
			</Button>
		</div>
	</header>

	<div class="container mx-auto space-y-4 p-4">
		<div
			class={cn('relative aspect-video w-full overflow-hidden rounded-lg bg-black')}
			role="presentation"
		>
			{#if streamResource.error}
				<div
					class="absolute inset-0 flex flex-col items-center justify-center gap-4 p-4 text-white"
				>
					<CircleAlert class="h-12 w-12 text-destructive" />
					<p class="text-center text-lg font-medium">
						{streamResource.error.message || 'Failed to load video'}
					</p>
					<div class="flex gap-2">
						<Button
							onclick={() => {
								streamResource.refetch();
							}}
							variant="secondary"
							size="sm"
						>
							<RefreshCcw class="mr-2 h-4 w-4" />
							Try Again
						</Button>
						<Button
							onclick={() => {
								const servers = groupedServers[selectedServer?.type.toLowerCase() ?? ''] || [];
								const index = servers.findIndex((s) => s.serverId === selectedServer?.serverId);
								const next = servers[(index + 1) % servers.length] || null;
								if (next && next.serverId !== selectedServer?.serverId) {
									captureServerSwitch({
										anime_id: data.anime.id,
										anime_title: titles.main,
										episode_number: data.episodeNumber,
										from_server: selectedServer?.serverName ?? 'unknown',
										to_server: next.serverName,
										stream_type: next.type.toLowerCase(),
										reason: 'error_fallback',
									});
								}
								selectedServer = next;
							}}
							variant="secondary"
							size="sm"
							disabled={data.servers.length <= 1}
						>
							Try Different Server
						</Button>
					</div>
				</div>
			{:else if streamResource.loading || !streamResource.current}
				<div class="absolute inset-0 grid place-items-center">
					<LoaderCircle class="h-12 w-12 animate-spin text-primary" />
				</div>
			{:else}
				{#key `${data.anime.id}:${data.episodeNumber}:${selectedServer?.serverId ?? ''}`}
					<Player
						playerId={`anime-${data.anime.id}-ep-${data.episodeNumber}`}
						info={streamResource.current}
						{nextEpisodeUrl}
						{updateLibrary}
						animeId={data.anime.id}
						animeTitle={titles.main}
						episodeNumber={data.episodeNumber}
						serverName={selectedServer?.serverName ?? 'unknown'}
						streamType={selectedServer?.type?.toLowerCase() ?? 'unknown'}
					/>
				{/key}
			{/if}
		</div>

		<div class="flex items-center justify-between rounded-lg border bg-card p-4">
			<div class="flex items-center gap-2">
				<Button size="sm" variant="outline" disabled={!prevEpisodeUrl} href={prevEpisodeUrl}>
					<ChevronLeft class="mr-1 h-4 w-4" />
					Previous
				</Button>

				<Button size="sm" variant="outline" disabled={!nextEpisodeUrl} href={nextEpisodeUrl}>
					Next
					<ChevronRight class="ml-1 h-4 w-4" />
				</Button>
			</div>
		</div>

		{#if Object.keys(groupedServers).length > 0}
			<div class="space-y-3 rounded-lg border bg-card p-4">
				<h3 class="text-base font-semibold">Servers</h3>
				{#each Object.entries(groupedServers) as [type, servers], i (i)}
					<div class="space-y-2">
						<p class="text-sm font-medium text-muted-foreground uppercase">{type}</p>
						<div class="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4">
							{#each servers as server (server.serverId)}
								<Button
									variant={selectedServer?.serverId === server.serverId &&
									selectedServer.type === server.type
										? 'default'
										: 'outline'}
									size="sm"
									onclick={() => {
										if (
											server.serverId === selectedServer?.serverId &&
											server.type === selectedServer?.type
										)
											return;
										captureServerSwitch({
											anime_id: data.anime.id,
											anime_title: titles.main,
											episode_number: data.episodeNumber,
											from_server: selectedServer?.serverName ?? 'unknown',
											to_server: server.serverName,
											stream_type: server.type.toLowerCase(),
											reason: 'manual',
										});
										selectedServer = server;
									}}
									class="justify-start"
								>
									<Server class="mr-2 h-3 w-3" />
									<span class="truncate capitalize">{server.serverName}</span>
								</Button>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<div class="rounded-lg border bg-card p-4">
			<div class="flex items-start gap-4">
				<img src={data.anime.imageUrl} alt={titles.main} class="h-24 w-16 rounded object-cover" />
				<div class="flex-1 space-y-2">
					<div>
						<h2 class="line-clamp-1 text-lg font-bold">
							{titles.main}
						</h2>
						{#if titles.sub}
							<p class="line-clamp-1 text-sm text-muted-foreground">{titles.sub}</p>
						{/if}
					</div>
					<div class="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
						{#if data.anime.metadata?.mean}
							<div class="flex items-center gap-1">
								<Star class="h-3 w-3 fill-yellow-500 text-yellow-500" />
								<span>{data.anime.metadata.mean.toFixed(1)}</span>
							</div>
							<span>•</span>
						{/if}
						<span class="capitalize">{data.anime.season} {data.anime.seasonYear}</span>
						<span>•</span>
						<span>{data.anime.metadata?.totalEpisodes || data.anime.lastEpisode} Episodes</span>
					</div>
					<div class="mt-2 flex flex-wrap gap-2">
						<Button href="/anime/{data.anime.id}" variant="outline">
							<Info class="mr-2 h-4 w-4" />
							View Details
						</Button>
						<LibraryBtn animeId={data.anime.id} libraryEntry={data.libraryEntry ?? null} />
					</div>
				</div>
			</div>
		</div>

		<div class="rounded-lg border bg-card p-4">
			<h3 class="text-base font-bold">
				Episode {data.episodeNumber}: {data.currentEpisode.title || `Episode ${data.episodeNumber}`}
			</h3>
			{#if data.currentEpisode.isFiller}
				<span class="mt-2 inline-block rounded bg-muted px-2 py-1 text-xs font-medium">
					Filler Episode
				</span>
			{/if}
		</div>

		<div class="space-y-3 rounded-lg border bg-card p-4">
			<div class="flex items-center justify-between">
				<h3 class="font-semibold">Episodes</h3>
				<span class="text-sm text-muted-foreground">
					{data.episodes.length} Total
				</span>
			</div>

			<Input
				type="text"
				placeholder="Search episodes..."
				bind:value={episodesSearch}
				class="mb-6 h-9 w-full"
			/>

			{#if filteredEpisodes.length === 0}
				<div class="rounded-lg border bg-muted/30 p-8 text-center">
					<Film class="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
					<p class="text-sm text-muted-foreground">No episodes match your search...</p>
				</div>
			{/if}

			<div class="max-h-96 w-full space-y-3 overflow-y-auto">
				{#each filteredEpisodes as episode (episode.id)}
					<div
						class="group w-full"
						animate:flip={{ duration: 500 }}
						{@attach (node: HTMLElement) => {
							if (episode.id !== data.currentEpisode.id) return;
							const container = node.parentElement;
							if (!container) return;
							container.scrollTo({
								top:
									node.offsetTop -
									container.offsetTop -
									container.clientHeight / 2 +
									node.clientHeight / 2,
							});
						}}
					>
						<Button
							href={episodeUrl(episode.number)}
							variant={episode.id === data.currentEpisode.id ? 'default' : 'outline'}
							class="h-fit w-full"
						>
							<div
								class={cn(
									'flex aspect-square w-10 flex-shrink-0 items-center justify-center rounded text-xs font-bold',
									'bg-accent text-accent-foreground',
								)}
							>
								{episode.number}
							</div>
							<div class="min-w-0 flex-1 text-left">
								<p class="line-clamp-1 text-sm">
									{episode.title || `Episode ${episode.number}`}
								</p>
							</div>
							<div class="flex items-center gap-2">
								{#if episode.isFiller}
									<Badge variant="secondary" class="text-xs">Filler</Badge>
								{/if}
								<Play
									class={cn(
										'h-3 w-3 flex-shrink-0',
										episode.id === data.currentEpisode.id || 'opacity-0 group-hover:opacity-100',
									)}
								/>
							</div>
						</Button>
					</div>
				{/each}
			</div>
		</div>
	</div>
</div>
