<script lang="ts">
	import {
		Dice6,
		Download,
		Heart,
		Library,
		LogOut,
		Menu,
		Settings,
		Swords,
		User,
	} from 'lucide-svelte';
	import { onNavigate } from '$app/navigation';
	import { page } from '$app/state';
	import UserProfileDropdown from '$lib/components/layout/user-profile-dropdown.svelte';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Sheet from '$lib/components/ui/sheet';
	import { getAppStateContext } from '$lib/context/state.svelte';
	import { cn } from '$lib/utils';
	import BrandText from './brand-text.svelte';
	import ProfilePicture from './profile-picture.svelte';
	import SearchBar from './search-bar.svelte';

	const appState = getAppStateContext();
	const showDownload = false;

	let links = $derived.by(() => {
		const base = [
			{ label: 'Catalog', link: '/catalog', Icon: Library },
			{ label: 'Genres', link: '/genres', Icon: Swords },
			{ label: 'Random', link: '/random', Icon: Dice6 },
		];

		if (appState.isLoggedIn) {
			base.push({ label: 'My List', link: '/my-list', Icon: Heart });
		}

		if (showDownload) {
			base.push({ label: 'Download', link: '/download', Icon: Download });
		}

		return base;
	});

	let sheetLinks = $derived.by(() => {
		if (!appState.isLoggedIn) {
			return links;
		}

		return [
			...links,
			{ label: 'Profile', link: '/profile', Icon: User },
			{ label: 'Settings', link: '/settings', Icon: Settings },
		];
	});

	let isSheetOpen = $state(false);

	onNavigate(() => {
		isSheetOpen = false;
	});
</script>

<header
	id="navbar"
	class="flex h-16 items-center border-b border-border bg-background/95 backdrop-blur-md supports-[backdrop-filter]:bg-background/60 md:h-20"
>
	<div class="container mx-auto p-4">
		<div class="flex items-center justify-between">
			<div class="flex items-center space-x-8">
				<a href="/" class="flex items-center">
					<BrandText size="lg" variant="anime" />
				</a>

				<nav class="hidden space-x-6 lg:flex">
					{#each links as link (link.link)}
						<a
							href={link.link}
							class={cn(
								'font-medium text-muted-foreground transition-colors hover:text-primary',
								page.url.pathname === link.link && 'text-foreground',
							)}
						>
							{link.label}
						</a>
					{/each}
				</nav>
			</div>

			<div class="flex items-center gap-4">
				<SearchBar />

				<Button variant="outline" class="lg:hidden" onclick={() => (isSheetOpen = true)}>
					<Menu class="h-5 w-5" />
					<span class="sr-only">Menu</span>
				</Button>

				{#if !appState.isLoggedIn}
					<Button href="/login" variant="outline" class="hidden lg:inline-flex">Sign In</Button>
					<Button href="/register" class="hidden lg:inline-flex">Register</Button>
				{:else}
					<UserProfileDropdown class="hidden lg:flex" />
				{/if}
			</div>
		</div>
	</div>
</header>

<Sheet.Root bind:open={isSheetOpen}>
	<Sheet.Content side="right">
		<Sheet.Header>
			<Sheet.Title>Menu</Sheet.Title>
		</Sheet.Header>

		{#if appState.isLoggedIn}
			<div class="px-4 pb-4">
				<div class="flex items-center gap-3 rounded-lg bg-muted/50 p-3">
					<UserProfileDropdown class="hidden" />
					<ProfilePicture />
					<div class="flex flex-col">
						<p class="text-sm font-medium">{appState.user?.username}</p>
						<p class="text-xs text-muted-foreground">{appState.user?.email}</p>
					</div>
				</div>
			</div>
		{/if}

		<div class="flex flex-col gap-2 px-4">
			{#each sheetLinks as link (link.link)}
				<a
					href={link.link}
					class={cn(
						'flex items-center font-medium text-muted-foreground transition-colors',
						page.url.pathname === link.link && 'text-foreground',
					)}
				>
					<link.Icon class="mr-2 h-4 w-4" />
					{link.label}
				</a>
			{/each}
		</div>

		<div class="mt-auto mb-4 flex flex-col gap-2 px-4">
			{#if !appState.isLoggedIn}
				<Button href="/login" variant="outline" class="w-full">Sign In</Button>
				<Button href="/register" class="w-full">Register</Button>
			{:else}
				<Button href="/logout" variant="destructive" class="w-full">
					<LogOut class="h-4 w-4" />
					Log out
				</Button>
			{/if}
		</div>
	</Sheet.Content>
</Sheet.Root>
