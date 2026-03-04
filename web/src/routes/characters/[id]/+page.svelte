<script lang="ts">
	import { ArrowLeft, ExternalLink, Heart, Users, ChevronDown } from 'lucide-svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let character = $derived(data.character!);
	const languageValue = $derived($language);

	let filteredVAs = $derived.by(() => {
		if (!character.voices) return [];
		return character.voices.filter(
			(voice) =>
				voice.language.toLowerCase() === 'japanese' || voice.language.toLowerCase() === 'english',
		);
	});

	let otherVAs = $derived.by(() => {
		if (!character.voices) return [];
		return character.voices.filter(
			(voice) =>
				voice.language.toLowerCase() !== 'japanese' && voice.language.toLowerCase() !== 'english',
		);
	});

	let showAllVAs = $state(false);

	function goBack() {
		if (browser) {
			if (window.history.length > 1) {
				window.history.back();
			} else {
				goto('/');
			}
		}
	}
</script>

<svelte:head>
	<title>{character.name} - Character - Aniways</title>
	<meta
		name="description"
		content={`Learn more about ${character.name} and their anime appearances`}
	/>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<Button variant="ghost" size="sm" class="gap-2" onclick={goBack}>
			<ArrowLeft class="h-4 w-4" />
			Back
		</Button>
	</div>

	<div class="mb-8">
		<div class="flex flex-col items-center gap-8 md:flex-row md:items-end">
			<div class="flex-shrink-0">
				<img
					src={character.image}
					alt={character.name}
					class="h-64 w-48 rounded-lg object-cover shadow-lg md:h-80 md:w-56"
					loading="lazy"
				/>
			</div>

			<div class="flex-1 space-y-4 text-center md:text-left">
				<div>
					<h1 class="text-3xl font-bold md:text-4xl">{character.name}</h1>
					{#if character.nameKanji}
						<p class="mt-2 text-xl text-muted-foreground">{character.nameKanji}</p>
					{/if}
				</div>

				<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:gap-6">
					{#if character.favorites > 0}
						<div class="flex items-center justify-center gap-2 md:justify-start">
							<Heart class="h-5 w-5 text-red-500" />
							<span class="text-lg font-medium"
								>{character.favorites.toLocaleString()} favorites</span
							>
						</div>
					{/if}

					<div class="flex justify-center gap-2 md:justify-start">
						<Button
							variant="outline"
							size="sm"
							href={`https://myanimelist.net/character/${character.malId}`}
							target="_blank"
							class="gap-2"
						>
							<ExternalLink class="h-4 w-4" />
							MAL
						</Button>
					</div>
				</div>
			</div>
		</div>
	</div>

	{#if character.about}
		<section class="mb-8">
			<h2 class="mb-4 text-2xl font-bold">About</h2>
			<div class="max-w-none">
				<div class="leading-relaxed text-muted-foreground">
					{#each character.about.split(/\n+/).filter((line) => line.trim() !== '') as line, i (i)}
						{line}<br />
					{/each}
				</div>
			</div>
		</section>
	{/if}

	<div class="grid gap-8 lg:grid-cols-3">
		<div class="space-y-8 lg:col-span-2">
			{#if character.nicknames && character.nicknames.length > 0}
				<section class="space-y-4">
					<h2 class="text-2xl font-bold">Nicknames</h2>
					<div class="flex flex-wrap gap-2">
						{#each character.nicknames as nickname (nickname)}
							<Badge variant="outline">{nickname}</Badge>
						{/each}
					</div>
				</section>
			{/if}

			{#if character.anime && character.anime.length > 0}
				<section class="space-y-4">
					<h2 class="text-2xl font-bold">Anime Appearances</h2>
					<div class="grid gap-4 sm:grid-cols-2">
						{#each character.anime as animeAppearance (animeAppearance.anime.id + '-' + animeAppearance.role)}
							<a
								href="/anime/{animeAppearance.anime.id}"
								class="group flex gap-4 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent"
							>
								<img
									src={animeAppearance.anime.imageUrl}
									alt={getLocalizedTitles(animeAppearance.anime, languageValue).main}
									class="h-20 w-14 flex-shrink-0 rounded object-cover"
									loading="lazy"
								/>
								<div class="min-w-0 flex-1">
									<h3 class="line-clamp-2 font-semibold group-hover:text-primary">
										{getLocalizedTitles(animeAppearance.anime, languageValue).main}
									</h3>
									{#if getLocalizedTitles(animeAppearance.anime, languageValue).sub}
										<p class="line-clamp-1 text-sm text-muted-foreground">
											{getLocalizedTitles(animeAppearance.anime, languageValue).sub}
										</p>
									{/if}
									<div class="mt-2 flex items-center gap-2">
										<Badge variant="secondary" class="text-xs">
											{animeAppearance.role}
										</Badge>
										{#if animeAppearance.anime.genre}
											<span class="text-xs text-muted-foreground">
												{animeAppearance.anime.genre.split(', ').slice(0, 2).join(', ')}
											</span>
										{/if}
									</div>
								</div>
							</a>
						{/each}
					</div>
				</section>
			{/if}

			{#if character.voices && character.voices.length > 0}
				<section class="space-y-4">
					<h2 class="text-2xl font-bold">Voice Actors</h2>

					{#if filteredVAs.length > 0}
						<div class="grid gap-4 sm:grid-cols-2">
							{#each filteredVAs as voice (voice.person.malId + '-' + voice.language)}
								<a
									href="/va/{voice.person.malId}"
									class="group flex gap-3 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent"
								>
									{#if voice.person.image}
										<img
											src={voice.person.image}
											alt={voice.person.name}
											class="h-12 w-12 flex-shrink-0 rounded-full object-cover"
											loading="lazy"
										/>
									{:else}
										<div
											class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-muted"
										>
											<Users class="h-6 w-6 text-muted-foreground" />
										</div>
									{/if}
									<div class="min-w-0 flex-1">
										<h3 class="font-semibold group-hover:text-primary">{voice.person.name}</h3>
										<p class="text-sm text-muted-foreground">{voice.language}</p>
									</div>
								</a>
							{/each}
						</div>
					{/if}

					{#if otherVAs.length > 0}
						<div class="flex justify-center">
							<Button
								variant="outline"
								size="sm"
								onclick={() => (showAllVAs = !showAllVAs)}
								class="gap-2"
							>
								<ChevronDown
									class="h-4 w-4 transition-transform {showAllVAs ? 'rotate-180' : ''}"
								/>
								{showAllVAs ? 'Show Less' : `Show ${otherVAs.length} More Languages`}
							</Button>
						</div>

						{#if showAllVAs}
							<div class="grid gap-4 sm:grid-cols-2">
								{#each otherVAs as voice (voice.person.malId + '-' + voice.language)}
									<a
										href="/va/{voice.person.malId}"
										class="group flex gap-3 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent"
									>
										{#if voice.person.image}
											<img
												src={voice.person.image}
												alt={voice.person.name}
												class="h-12 w-12 flex-shrink-0 rounded-full object-cover"
												loading="lazy"
											/>
										{:else}
											<div
												class="flex h-12 w-12 flex-shrink-0 items-center justify-center rounded-full bg-muted"
											>
												<Users class="h-6 w-6 text-muted-foreground" />
											</div>
										{/if}
										<div class="min-w-0 flex-1">
											<h3 class="font-semibold group-hover:text-primary">{voice.person.name}</h3>
											<p class="text-sm text-muted-foreground">{voice.language}</p>
										</div>
									</a>
								{/each}
							</div>
						{/if}
					{/if}
				</section>
			{/if}
		</div>

		<div class="space-y-6">
			<Card>
				<CardHeader>
					<CardTitle>Character Info</CardTitle>
				</CardHeader>
				<CardContent class="space-y-3">
					<div class="flex justify-between">
						<span class="text-muted-foreground">MAL ID</span>
						<span class="font-medium">{character.malId}</span>
					</div>
					{#if character.favorites > 0}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Favorites</span>
							<span class="font-medium">{character.favorites.toLocaleString()}</span>
						</div>
					{/if}
					{#if character.anime && character.anime.length > 0}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Anime Count</span>
							<span class="font-medium">{character.anime.length}</span>
						</div>
					{/if}
					{#if character.voices && character.voices.length > 0}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Voice Actors</span>
							<span class="font-medium">{character.voices.length}</span>
						</div>
					{/if}
				</CardContent>
			</Card>
		</div>
	</div>
</div>
