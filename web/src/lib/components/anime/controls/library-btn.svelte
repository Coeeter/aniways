<script lang="ts">
	import { type } from 'arktype';
	import {
		Check,
		Clapperboard,
		ListChecks,
		LoaderCircle,
		Pencil,
		Plus,
		Save,
		Trash,
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { defaults, superForm } from 'sveltekit-superforms';
	import { arktype } from 'sveltekit-superforms/adapters';
	import { goto, invalidate } from '$app/navigation';
	import {
		captureListItemAdded,
		captureListItemRemoved,
		captureListItemUpdated,
	} from '$lib/analytics';
	import { apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/openapi';
	import { buttonVariants } from '$lib/components/ui/button';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Drawer from '$lib/components/ui/drawer';
	import * as Form from '$lib/components/ui/form';
	import { Input } from '$lib/components/ui/input';
	import * as Select from '$lib/components/ui/select';
	import Separator from '$lib/components/ui/separator/separator.svelte';
	import { getAppStateContext } from '$lib/context/state.svelte';
	import { IsMobile } from '$lib/hooks/is-mobile.svelte';
	import { language } from '$lib/stores/language';
	import { getLocalizedTitles } from '$lib/utils/anime-title';
	import { cn } from '$lib/utils';

	type LibraryResponse = components['schemas']['models.LibraryResponse'];
	type AnimeResponse = components['schemas']['models.AnimeResponse'];

	type Props = {
		animeId: string;
		libraryEntry: LibraryResponse | null;
		class?: string;
		iconOnly?: boolean;
		variations?: AnimeResponse[];
		currentAnimeName?: string;
	};

	let {
		animeId,
		libraryEntry,
		class: className,
		iconOnly = false,
		variations = [],
		currentAnimeName = 'Current Version',
	}: Props = $props();
	const appState = getAppStateContext();
	const languageValue = $derived($language);

	let isOpen = $state(false);
	let isAdding = $state(false);
	let isDeleting = $state(false);

	let isMobile = new IsMobile();

	const addToLibrary = async () => {
		if (isAdding) return;
		isAdding = true;

		try {
			const result = await apiClient.POST('/library/{animeID}', {
				body: { status: 'watching', watchedEpisodes: 0 },
				params: { path: { animeID: animeId } },
			});

			if (!result.response.ok) {
				toast.error('Failed to add to library');
				return;
			}

			captureListItemAdded({ anime_id: animeId, list_type: 'watching' });
			await invalidate('app:library');
			toast.success('Added to library');
			isOpen = false;
		} catch {
			toast.error('Failed to add to library');
		} finally {
			isAdding = false;
		}
	};

	const removeFromLibrary = async () => {
		if (isDeleting) return;
		isDeleting = true;

		try {
			const result = await apiClient.DELETE('/library/{animeID}', {
				params: { path: { animeID: animeId } },
			});

			if (!result.response.ok) {
				toast.error('Failed to remove from library');
				return;
			}

			captureListItemRemoved({ anime_id: animeId, list_type: libraryEntry?.status ?? 'unknown' });
			await invalidate('app:library');
			toast.success('Removed from library');
			isOpen = false;
		} catch {
			toast.error('Failed to remove from library');
		} finally {
			isDeleting = false;
		}
	};
	const checkIfLoggedIn = (fn: () => void) => {
		if (!appState.isLoggedIn) {
			toast.error('You must be logged in to use the library', {
				action: {
					label: 'Login',
					onClick: () => goto('/login'),
				},
			});
			return;
		}
		fn();
	};

	const updateFormSchema = type({
		watchedEpisodes: type('number>=0').describe('be a valid number of episodes'),
		status: type("'planning'|'watching'|'completed'|'dropped'|'paused'").describe(
			'be a valid library status',
		),
	});

	const form = superForm(defaults(arktype(updateFormSchema)), {
		SPA: true,
		validators: arktype(updateFormSchema),
		resetForm: false,
		onUpdate: async ({ form, cancel }) => {
			if (!form.valid) return;

			try {
				const result = await apiClient.PUT('/library/{animeID}', {
					body: form.data,
					params: { path: { animeID: animeId } },
				});
				if (!result.response.ok) {
					toast.error('Failed to update library');
					return;
				}

				captureListItemUpdated({
					anime_id: animeId,
					status: form.data.status,
					watched_episodes: form.data.watchedEpisodes,
				});
				await invalidate('app:library');
				toast.success('Library updated');
				isOpen = false;
			} catch {
				toast.error('Failed to update library');
				cancel();
			}
		},
	});

	const { form: formData, enhance, submitting, errors } = form;

	$effect(() => {
		if (!libraryEntry) return;

		formData.set({
			status: libraryEntry.status,
			watchedEpisodes: libraryEntry.watchedEpisodes,
		});
	});
</script>

{#if libraryEntry}
	{#if isMobile.current}
		<Drawer.Root bind:open={isOpen} direction="bottom">
			<Drawer.Trigger
				class={cn(
					buttonVariants({
						size: iconOnly ? 'icon' : 'default',
						variant: libraryEntry != null ? 'secondary' : 'outline',
					}),
					iconOnly ? '' : 'flex items-center gap-2 capitalize',
					className,
				)}
			>
				{#if iconOnly}
					<Pencil class="h-4 w-4" />
				{:else}
					<Check class="h-4 w-4" />
					{libraryEntry.status.split('_').join(' ')}
					{#if libraryEntry.watchedEpisodes > 0}
						· {libraryEntry.watchedEpisodes} Ep{libraryEntry.watchedEpisodes > 1 ? 's' : ''}
					{:else}
						· 0 Eps
					{/if}
				{/if}
			</Drawer.Trigger>
			<Drawer.Content>
				<Drawer.Header>
					<Drawer.Title>Update Library Entry</Drawer.Title>
					<Drawer.Description>Update your progress for this anime.</Drawer.Description>
				</Drawer.Header>

				<form use:enhance class="flex flex-col justify-center gap-2">
					{#if variations && variations.length > 0 && libraryEntry}
						<div class="px-4">
							svelte-ignore a11y_label_has_associated_control
							<label
								class="mb-2 block text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
								for="version-select-dialog"
							>
								Version
							</label>
							<Select.Root
								type="single"
								value={animeId}
								onValueChange={async (value) => {
									if (!value || value === animeId) return;
									const id = toast.loading('Switching version...');
									try {
										const result = await apiClient.PUT('/library/{animeID}/switch/{variationID}', {
											params: {
												path: { animeID: animeId, variationID: value },
											},
										});
										if (result.response.ok) {
											toast.success('Version switched', { id });
											await invalidate('app:library');
											await goto(`/anime/${value}`);
										} else {
											toast.error('Failed to switch version', { id });
										}
									} catch {
										toast.error('Failed to switch version', { id });
									}
								}}
							>
								<Select.Trigger class="w-full capitalize" id="version-select-dialog">
									{currentAnimeName}
								</Select.Trigger>
								<Select.Content>
									{#each variations as variation (variation.id)}
										{@const isCurrent = variation.id === animeId}
									<Select.Item value={variation.id} class="capitalize" disabled={isCurrent}>
										{getLocalizedTitles(variation, languageValue).main}
										{#if isCurrent}
											<span class="ml-2 text-xs text-muted-foreground">(Current)</span>
										{/if}
									</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<Separator orientation="horizontal" class="my-2" />
					{/if}

					<Form.Field {form} name="status" class="px-4">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Status</Form.Label>
								<div class="relative">
									<ListChecks
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Select.Root type="single" bind:value={$formData.status} name={props.name}>
										<Select.Trigger
											{...props}
											class={cn(
												'w-full pl-10 capitalize',
												$errors.status &&
													$errors.status.length > 0 &&
													'border-destructive focus-visible:ring-destructive',
											)}
										>
											{$formData.status || 'Select a status'}
										</Select.Trigger>
										<Select.Content>
											<Select.Item value="planning" class="capitalize">Planning</Select.Item>
											<Select.Item value="watching" class="capitalize">Watching</Select.Item>
											<Select.Item value="completed" class="capitalize">Completed</Select.Item>
											<Select.Item value="dropped" class="capitalize">Dropped</Select.Item>
											<Select.Item value="paused" class="capitalize">Paused</Select.Item>
										</Select.Content>
									</Select.Root>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field {form} name="watchedEpisodes" class="px-4">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Watched Episodes</Form.Label>
								<div class="relative">
									<Clapperboard
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										bind:value={$formData.watchedEpisodes}
										placeholder="0"
										type="number"
										class={cn(
											'pl-10',
											$errors.watchedEpisodes &&
												$errors.watchedEpisodes.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
									/>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Drawer.Footer class="flex-col-reverse gap-2 md:flex-row">
						<Button
							type="button"
							variant="destructive"
							class="md:mr-auto"
							onclick={removeFromLibrary}
							disabled={$submitting || isDeleting}
						>
							{#if isDeleting}
								<LoaderCircle class="size-4 animate-spin" />
								Removing...
							{:else}
								<Trash class="size-4" />
								Remove from Library
							{/if}
						</Button>
						<Separator orientation="horizontal" class="my-2 md:hidden" />
						<Drawer.Close class={buttonVariants({ variant: 'secondary' })} type="button">
							Cancel
						</Drawer.Close>
						<Form.Button disabled={$submitting}>
							{#if $submitting}
								<LoaderCircle class="size-4 animate-spin" />
								Saving...
							{:else}
								<Save class="size-4" />
								Save Changes
							{/if}</Form.Button
						>
					</Drawer.Footer>
				</form>
			</Drawer.Content>
		</Drawer.Root>
	{:else}
		<Dialog.Root bind:open={isOpen}>
			<Dialog.Trigger
				class={cn(
					buttonVariants({
						size: iconOnly ? 'icon' : 'default',
						variant: libraryEntry != null ? 'secondary' : 'outline',
					}),
					iconOnly ? '' : 'flex items-center gap-2 capitalize',
					className,
				)}
			>
				{#if iconOnly}
					<Pencil class="h-4 w-4" />
				{:else}
					<Check class="h-4 w-4" />
					{libraryEntry.status.split('_').join(' ')}
					{#if libraryEntry.watchedEpisodes > 0}
						· {libraryEntry.watchedEpisodes} Ep{libraryEntry.watchedEpisodes > 1 ? 's' : ''}
					{:else}
						· 0 Eps
					{/if}
				{/if}
			</Dialog.Trigger>
			<Dialog.Content>
				<Dialog.Header>
					<Dialog.Title>Update Library Entry</Dialog.Title>
					<Dialog.Description>Update your progress for this anime.</Dialog.Description>
				</Dialog.Header>
				<form use:enhance class="flex flex-col justify-center gap-2">
					{#if variations && variations.length > 0 && libraryEntry}
						<div>
							<label
								class="mb-2 block text-sm leading-none font-medium peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
								for="version-select"
							>
								Version
							</label>
							<Select.Root
								type="single"
								value={animeId}
								onValueChange={async (value) => {
									if (!value || value === animeId) return;
									const id = toast.loading('Switching version...');
									try {
										const result = await apiClient.PUT('/library/{animeID}/switch/{variationID}', {
											params: {
												path: { animeID: animeId, variationID: value },
											},
										});
										if (result.response.ok) {
											toast.success('Version switched', { id });
											await invalidate('app:library');
											await goto(`/anime/${value}`);
										} else {
											toast.error('Failed to switch version', { id });
										}
									} catch {
										toast.error('Failed to switch version', { id });
									}
								}}
							>
								<Select.Trigger class="w-full capitalize" id="version-select">
									{currentAnimeName}
								</Select.Trigger>
								<Select.Content>
									{#each variations as variation (variation.id)}
										{@const isCurrent = variation.id === animeId}
								<Select.Item value={variation.id} class="capitalize" disabled={isCurrent}>
									{getLocalizedTitles(variation, languageValue).main}
									{#if isCurrent}
										<span class="ml-2 text-xs text-muted-foreground">(Current)</span>
									{/if}
								</Select.Item>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>
						<Separator orientation="horizontal" class="my-2" />
					{/if}

					<Form.Field {form} name="status">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Status</Form.Label>
								<div class="relative">
									<ListChecks
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Select.Root type="single" bind:value={$formData.status} name={props.name}>
										<Select.Trigger
											{...props}
											class={cn(
												'w-full pl-10 capitalize',
												$errors.status &&
													$errors.status.length > 0 &&
													'border-destructive focus-visible:ring-destructive',
											)}
										>
											{$formData.status || 'Select a status'}
										</Select.Trigger>
										<Select.Content>
											<Select.Item value="planning" class="capitalize">Planning</Select.Item>
											<Select.Item value="watching" class="capitalize">Watching</Select.Item>
											<Select.Item value="completed" class="capitalize">Completed</Select.Item>
											<Select.Item value="dropped" class="capitalize">Dropped</Select.Item>
											<Select.Item value="paused" class="capitalize">Paused</Select.Item>
										</Select.Content>
									</Select.Root>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field {form} name="watchedEpisodes">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Watched Episodes</Form.Label>
								<div class="relative">
									<Clapperboard
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										bind:value={$formData.watchedEpisodes}
										placeholder="0"
										type="number"
										class={cn(
											'pl-10',
											$errors.watchedEpisodes &&
												$errors.watchedEpisodes.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
									/>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Dialog.Footer class="gap-2">
						<Button
							type="button"
							variant="destructive"
							class="md:mr-auto"
							onclick={removeFromLibrary}
							disabled={$submitting || isDeleting}
						>
							{#if isDeleting}
								<LoaderCircle class="size-4 animate-spin" />
								Removing...
							{:else}
								<Trash class="size-4" />
								Remove from Library
							{/if}
						</Button>
						<Separator orientation="horizontal" class="my-2 md:hidden" />
						<Dialog.Close class={buttonVariants({ variant: 'secondary' })} type="button">
							Cancel
						</Dialog.Close>
						<Form.Button disabled={$submitting}>
							{#if $submitting}
								<LoaderCircle class="size-4 animate-spin" />
								Saving...
							{:else}
								<Save class="size-4" />
								Save Changes
							{/if}</Form.Button
						>
					</Dialog.Footer>
				</form>
			</Dialog.Content>
		</Dialog.Root>
	{/if}
{:else}
	<Button
		variant="outline"
		size={iconOnly ? 'icon' : 'default'}
		class={cn(iconOnly ? '' : 'flex items-center gap-2', className)}
		onclick={() => checkIfLoggedIn(addToLibrary)}
		disabled={isAdding}
	>
		{#if isAdding}
			<LoaderCircle class="size-4 animate-spin" />
			{#if !iconOnly}Adding...{/if}
		{:else}
			<Plus class="h-4 w-4" />
			{#if !iconOnly}Add to Library{/if}
		{/if}
	</Button>
{/if}
