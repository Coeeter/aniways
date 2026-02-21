<script lang="ts">
	import type { components } from '$lib/api/openapi';
	import { createArtPlayer } from '$lib/components/anime/player/create-player.svelte';
	import { getAppStateContext } from '$lib/context/state.svelte';

	type StreamInfo = components['schemas']['models.StreamingDataResponse'];

	type Props = {
		playerId: string;
		info: StreamInfo;
		nextEpisodeUrl: string | null;
		updateLibrary: () => Promise<void>;
		animeId: string;
		animeTitle: string;
		episodeNumber: number;
		serverName: string;
		streamType: string;
	};

	let {
		info,
		playerId,
		nextEpisodeUrl,
		updateLibrary,
		animeId,
		animeTitle,
		episodeNumber,
		serverName,
		streamType,
	}: Props = $props();
	const appState = getAppStateContext();

	let element: HTMLDivElement | null = $state(null);

	$effect(() => {
		if (!element) return;

		const player = createArtPlayer({
			id: playerId,
			appState,
			container: element,
			source: info,
			nextEpisodeUrl,
			updateLibrary,
			animeId,
			animeTitle,
			episodeNumber,
			serverName,
			streamType,
		});

		return () => {
			player.destroy();
		};
	});
</script>

<div class="h-full w-full bg-card" bind:this={element}></div>
