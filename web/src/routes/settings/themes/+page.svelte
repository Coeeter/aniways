<script lang="ts">
	import { Check, Palette } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import type { components } from '$lib/api/openapi';
	import { captureThemeChanged } from '$lib/analytics';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import * as Card from '$lib/components/ui/card';
	import { getAppStateContext } from '$lib/context/state.svelte';
	import { cn } from '$lib/utils';
	import type { PageProps } from './$types';

	type Theme = components['schemas']['models.Theme'];

	let { data }: PageProps = $props();
	const appState = getAppStateContext();

	let currentTheme = $derived(appState.settings?.theme?.id);

	function updateTheme(theme: Theme) {
		appState.updateTheme(theme);
		captureThemeChanged({ theme_id: theme.id, theme_name: theme.name });
		toast.success(`Theme changed to ${theme.name}`);
	}
</script>

<svelte:head>
	<title>Themes - Settings - Aniways</title>
	<meta
		name="description"
		content="Choose your preferred theme and customize the appearance of Aniways"
	/>
</svelte:head>

<div class="space-y-6">
	<div class="space-y-2">
		<h2 class="flex items-center text-2xl font-bold">
			Themes
			<Badge variant="secondary" class="ml-2">Beta</Badge>
		</h2>
		<p class="text-muted-foreground">
			Choose a theme that matches your style. Each theme comes with its own color palette and
			typography.
		</p>
	</div>

	<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
		{#each data.themes as theme (theme.id)}
			<Card.Root
				class={cn(
					'group cursor-pointer transition-all duration-200 hover:shadow-lg',
					currentTheme === theme.id
						? 'shadow-lg ring-2 ring-primary'
						: 'hover:scale-[1.02] hover:shadow-md',
				)}
				onclick={() => updateTheme(theme)}
			>
				<Card.Header class="pb-3">
					<div class="flex items-start justify-between">
						<div class="space-y-1">
							<Card.Title class="text-lg">{theme.name}</Card.Title>
							<Card.Description>{theme.description}</Card.Description>
						</div>
						{#if currentTheme === theme.id}
							<Badge variant="default" class="gap-1">
								<Check class="h-3 w-3" />
								Active
							</Badge>
						{/if}
					</div>
				</Card.Header>
				<Card.Content class="space-y-4">
					<!-- Theme Preview -->
					<div
						class={cn(
							'dark overflow-hidden rounded-lg border bg-background',
							theme.className || 'default',
						)}
					>
						<!-- Mini App Interface -->
						<div class="space-y-3 p-3">
							<!-- Header -->
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<div class="h-5 w-5 rounded bg-primary"></div>
									<div class="h-3 w-16 rounded bg-foreground"></div>
								</div>
								<div class="flex gap-1">
									<div class="h-5 w-5 rounded bg-muted"></div>
									<div class="h-5 w-5 rounded bg-muted"></div>
								</div>
							</div>

							<!-- Content Area -->
							<div class="space-y-2">
								<!-- Anime Card 1 -->
								<div class="rounded-lg border bg-card p-2">
									<div class="flex gap-2">
										<div class="h-10 w-7 rounded bg-muted"></div>
										<div class="flex-1 space-y-1">
											<div class="h-2.5 w-4/5 rounded bg-foreground"></div>
											<div class="h-2 w-3/5 rounded bg-muted-foreground"></div>
											<div class="mt-1 flex gap-1">
												<div class="h-2 w-8 rounded-full bg-primary"></div>
												<div class="h-2 w-8 rounded-full bg-secondary"></div>
											</div>
										</div>
									</div>
								</div>

								<!-- Anime Card 2 -->
								<div class="rounded-lg border bg-card p-2">
									<div class="flex gap-2">
										<div class="h-10 w-7 rounded bg-muted"></div>
										<div class="flex-1 space-y-1">
											<div class="h-2.5 w-3/4 rounded bg-foreground"></div>
											<div class="h-2 w-2/3 rounded bg-muted-foreground"></div>
											<div class="mt-1 flex gap-1">
												<div class="h-2 w-8 rounded-full bg-accent"></div>
												<div class="h-2 w-8 rounded-full bg-muted"></div>
											</div>
										</div>
									</div>
								</div>
							</div>

							<!-- Action Buttons -->
							<div class="flex gap-2">
								<div class="flex h-6 w-14 items-center justify-center rounded bg-primary">
									<div class="h-1.5 w-8 rounded bg-primary-foreground"></div>
								</div>
								<div class="flex h-6 w-14 items-center justify-center rounded border border-border">
									<div class="h-1.5 w-8 rounded bg-foreground"></div>
								</div>
							</div>
						</div>
					</div>

					<Button
						variant={currentTheme === theme.id ? 'default' : 'outline'}
						size="sm"
						class="w-full gap-2"
					>
						{#if currentTheme === theme.id}
							<Check class="h-3 w-3" />
						{/if}
						{currentTheme === theme.id ? 'Active' : 'Select'}
					</Button>
				</Card.Content>
			</Card.Root>
		{/each}
	</div>

	{#if data.themes.length === 0}
		<Card.Root>
			<Card.Content class="p-8 text-center">
				<Palette class="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
				<h3 class="mb-2 text-lg font-semibold">No themes available</h3>
				<p class="text-muted-foreground">
					There are no themes available at the moment. Please try again later.
				</p>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
