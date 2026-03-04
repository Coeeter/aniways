<script lang="ts">
	import { ArrowLeft, ExternalLink, Mic, Music } from 'lucide-svelte';
	import { browser } from '$app/environment';
	import { goto } from '$app/navigation';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Pagination } from '$lib/components/ui/pagination';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();
	let person = $derived(data.person!);
	const languageValue = $derived($language);

	let currentCharacterPage = $state(1);
	let currentAnimePage = $state(1);
	const itemsPerPage = 12;

	let paginatedCharacters = $derived.by(() => {
		if (!person.characters) return [];
		const start = (currentCharacterPage - 1) * itemsPerPage;
		const end = start + itemsPerPage;
		return person.characters.slice(start, end);
	});

	let characterTotalPages = $derived.by(() => {
		return Math.ceil((person.characters?.length || 0) / itemsPerPage);
	});

	let paginatedAnime = $derived.by(() => {
		if (!person.anime) return [];
		const start = (currentAnimePage - 1) * itemsPerPage;
		const end = start + itemsPerPage;
		return person.anime.slice(start, end);
	});

	let animeTotalPages = $derived.by(() => {
		return Math.ceil((person.anime?.length || 0) / itemsPerPage);
	});

	function goBack() {
		if (browser) {
			if (window.history.length > 1) {
				window.history.back();
			} else {
				goto('/');
			}
		}
	}

	function goToCharacterPage(page: number) {
		if (page >= 1 && page <= characterTotalPages) {
			currentCharacterPage = page;
		}
	}

	function goToAnimePage(page: number) {
		if (page >= 1 && page <= animeTotalPages) {
			currentAnimePage = page;
		}
	}
</script>

<svelte:head>
	<title>{person.name} - Voice Actor - Aniways</title>
	<meta
		name="description"
		content={`Learn more about ${person.name} and their voice acting roles`}
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
					src={person.image}
					alt={person.name}
					class="h-64 w-48 rounded-lg object-cover shadow-lg md:h-80 md:w-56"
					loading="lazy"
				/>
			</div>

			<div class="flex-1 space-y-4 text-center md:text-left">
				<div>
					<h1 class="text-3xl font-bold md:text-4xl">{person.name}</h1>
					{#if person.givenName}
						<p class="mt-2 text-xl text-muted-foreground">{person.givenName}</p>
					{/if}
				</div>

				<div class="flex justify-center gap-2 md:justify-start">
					<Button
						variant="outline"
						size="sm"
						href={`https://myanimelist.net/people/${person.malId}`}
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

	{#if person.about}
		<section class="mb-8">
			<h2 class="mb-4 text-2xl font-bold">About</h2>
			<div class="max-w-none">
				<div class="leading-relaxed text-muted-foreground">
					{#each person.about.split(/\n+/).filter((line) => line.trim() !== '') as line, i (i)}
						{line}<br />
					{/each}
				</div>
			</div>
		</section>
	{/if}

	<div class="grid gap-8 lg:grid-cols-3">
		<div class="space-y-8 lg:col-span-2">
			{#if person.characters && person.characters.length > 0}
				<section class="space-y-4">
					<div class="flex items-center justify-between">
						<h2 class="flex items-center gap-2 text-2xl font-bold">
							<Mic class="h-6 w-6" />
							Voice Acting Roles
						</h2>
						<span class="text-sm text-muted-foreground">
							{person.characters.length} total roles
						</span>
					</div>
					<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						{#each paginatedCharacters as character (character.character.malId + '-' + character.anime.id)}
							<a
								href="/characters/{character.character.malId}"
								class="group flex gap-4 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent"
							>
								<img
									src={character.character.image}
									alt={character.character.name}
									class="h-20 w-14 flex-shrink-0 rounded object-cover"
									loading="lazy"
								/>
								<div class="min-w-0 flex-1 overflow-hidden">
									<h3 class="line-clamp-2 font-semibold group-hover:text-primary">
										{character.character.name}
									</h3>
									<p class="line-clamp-1 text-sm text-muted-foreground">
									{getLocalizedTitles(character.anime, languageValue).main}
									</p>
									<div class="mt-2 flex flex-wrap items-center gap-2">
										<Badge variant="secondary" class="text-xs">
											{character.role}
										</Badge>
										{#if character.anime.genre}
											<span class="line-clamp-1 text-xs text-muted-foreground">
												{character.anime.genre.split(', ').slice(0, 2).join(', ')}
											</span>
										{/if}
									</div>
								</div>
							</a>
						{/each}
					</div>

					<Pagination
						totalPages={characterTotalPages}
						currentPage={currentCharacterPage}
						onPageChange={goToCharacterPage}
					/>
				</section>
			{/if}

			{#if person.anime && person.anime.length > 0}
				<section class="space-y-4">
					<div class="flex items-center justify-between">
						<h2 class="flex items-center gap-2 text-2xl font-bold">
							<Music class="h-6 w-6" />
							Anime Work
						</h2>
						<span class="text-sm text-muted-foreground">
							{person.anime.length} total works
						</span>
					</div>
					<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						{#each paginatedAnime as animeWork (animeWork.anime.id)}
							<a
								href="/anime/{animeWork.anime.id}"
								class="group flex gap-4 rounded-lg border bg-card p-4 transition-all hover:border-primary/50 hover:bg-accent"
							>
								<img
									src={animeWork.anime.imageUrl}
								alt={getLocalizedTitles(animeWork.anime, languageValue).main}
									class="h-20 w-14 flex-shrink-0 rounded object-cover"
									loading="lazy"
								/>
								<div class="min-w-0 flex-1 overflow-hidden">
									<h3 class="line-clamp-2 font-semibold group-hover:text-primary">
									{getLocalizedTitles(animeWork.anime, languageValue).main}
									</h3>
								{#if getLocalizedTitles(animeWork.anime, languageValue).sub}
									<p class="line-clamp-1 text-sm text-muted-foreground">
										{getLocalizedTitles(animeWork.anime, languageValue).sub}
									</p>
								{/if}
									<p class="my-3 text-xs text-muted-foreground">
										{animeWork.position.replace('add', '').trim()}
									</p>
									<div class="mt-2 flex flex-wrap items-center gap-2">
										{#if animeWork.anime.genre}
											<span class="line-clamp-1 text-xs text-muted-foreground">
												{animeWork.anime.genre.split(', ').slice(0, 2).join(', ')}
											</span>
										{/if}
									</div>
								</div>
							</a>
						{/each}
					</div>
					<Pagination
						totalPages={animeTotalPages}
						currentPage={currentAnimePage}
						onPageChange={goToAnimePage}
					/>
				</section>
			{/if}
		</div>

		<div class="space-y-6">
			<Card>
				<CardHeader>
					<CardTitle>Voice Actor Info</CardTitle>
				</CardHeader>
				<CardContent class="space-y-3">
					<div class="flex justify-between">
						<span class="text-muted-foreground">MAL ID</span>
						<span class="font-medium">{person.malId}</span>
					</div>
					{#if person.characters && person.characters.length > 0}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Voice Roles</span>
							<span class="font-medium">{person.characters.length}</span>
						</div>
					{/if}
					{#if person.anime && person.anime.length > 0}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Anime Work</span>
							<span class="font-medium">{person.anime.length}</span>
						</div>
					{/if}
					{#if characterTotalPages > 1}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Voice Roles Page</span>
							<span class="font-medium">{currentCharacterPage} of {characterTotalPages}</span>
						</div>
					{/if}
					{#if animeTotalPages > 1}
						<div class="flex justify-between">
							<span class="text-muted-foreground">Anime Work Page</span>
							<span class="font-medium">{currentAnimePage} of {animeTotalPages}</span>
						</div>
					{/if}
				</CardContent>
			</Card>
		</div>
	</div>
</div>
