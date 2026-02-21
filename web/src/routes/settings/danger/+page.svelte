<script lang="ts">
	import { type } from 'arktype';
	import { Database, Eye, EyeOff, LoaderCircle, Trash2, TriangleAlert } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { defaults, setError, superForm } from 'sveltekit-superforms';
	import { arktype, arktypeClient } from 'sveltekit-superforms/adapters';
	import { goto, invalidate } from '$app/navigation';
	import { captureAccountDeleted, captureLibraryCleared } from '$lib/analytics';
	import { apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/openapi';
	import Button from '$lib/components/ui/button/button.svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Dialog from '$lib/components/ui/dialog';
	import * as Form from '$lib/components/ui/form';
	import Input from '$lib/components/ui/input/input.svelte';
	import { cn } from '$lib/utils';

	const deleteFormSchema = type({
		password: type('string').describe('current password'),
		confirmation: type('string').describe('confirmation text'),
	});

	const sf = superForm(defaults(arktype(deleteFormSchema)), {
		SPA: true,
		resetForm: true,
		validators: arktypeClient(deleteFormSchema),
		onUpdate: async ({ form, cancel }) => {
			if (!form.valid) return;

			if (form.data.confirmation !== 'DELETE') {
				setError(form, 'confirmation', ['Please type DELETE to confirm']);
				toast.error('Please type DELETE to confirm');
				cancel();
				return;
			}

			try {
				const res = await apiClient.DELETE('/users', {
					body: {
						password: form.data.password,
					},
				});

				if (res.response.status === 204) {
					captureAccountDeleted();
					await invalidate('app:user');
					await goto('/');
					toast.success('Account deleted successfully');
					return;
				}

				if (res.response.status === 400) {
					const error = res.error as components['schemas']['models.ValidationErrorResponse'];
					if (error?.details) {
						Object.entries(error.details).forEach(([field, messages]) => {
							setError(form, field as 'password', messages);
						});
						toast.error('Please fix the errors in the form and try again.');
						cancel();
						return;
					}
				}

				toast.error('Failed to delete account. Please check your password.');
				cancel();
			} catch {
				toast.error('An error occurred. Please try again.');
				cancel();
			}
		},
	});

	const { form, enhance, submitting, errors } = sf;

	let showDeleteDialog = $state(false);
	let showClearDialog = $state(false);
	let showPassword = $state(false);
	let isClearingLibrary = $state(false);

	async function clearLibrary() {
		isClearingLibrary = true;
		try {
			const response = await apiClient.DELETE('/library');
			if (response.response.status === 200) {
				captureLibraryCleared();
				toast.success('Library cleared successfully');
				showClearDialog = false;
			} else {
				toast.error('Failed to clear library. Please try again.');
			}
		} catch (error) {
			console.error('Failed to clear library:', error);
			toast.error('An error occurred while clearing your library.');
		}
		isClearingLibrary = false;
	}
</script>

<Card.Root class="border-red-800">
	<Card.Header>
		<Card.Title class="flex items-center gap-2 text-red-400">
			<TriangleAlert class="h-5 w-5" />
			Danger Zone
		</Card.Title>
		<Card.Description>Irreversible and destructive actions</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="rounded-lg border border-red-800 bg-red-950 p-4">
			<div class="flex items-start gap-3">
				<Database class="mt-0.5 h-5 w-5 text-red-400" />
				<div class="space-y-2">
					<h4 class="font-medium text-red-200">Clear Library</h4>
					<p class="text-sm text-red-300">
						Remove all anime from your library. This action cannot be undone.
					</p>
					<Button variant="destructive" size="sm" onclick={() => (showClearDialog = true)}>
						<Database class="mr-2 h-4 w-4" />
						Clear Library
					</Button>
				</div>
			</div>
		</div>

		<div class="rounded-lg border border-red-800 bg-red-950 p-4">
			<div class="flex items-start gap-3">
				<Trash2 class="mt-0.5 h-5 w-5 text-red-400" />
				<div class="space-y-2">
					<h4 class="font-medium text-red-200">Delete Account</h4>
					<p class="text-sm text-red-300">
						Permanently delete your account and all associated data. This action cannot be undone.
					</p>
					<Button variant="destructive" size="sm" onclick={() => (showDeleteDialog = true)}>
						<Trash2 class="mr-2 h-4 w-4" />
						Delete Account
					</Button>
				</div>
			</div>
		</div>
	</Card.Content>
</Card.Root>

<Dialog.Root bind:open={showDeleteDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title class="text-red-400">Delete Account</Dialog.Title>
			<Dialog.Description>
				This action cannot be undone. All your data will be permanently deleted.
			</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4">
			<div class="rounded-lg border border-red-800 bg-red-950 p-4">
				<div class="flex items-start gap-3">
					<TriangleAlert class="mt-0.5 h-5 w-5 text-red-400" />
					<div class="space-y-2">
						<h4 class="font-medium text-red-200">What will be deleted:</h4>
						<ul class="space-y-1 text-sm text-red-300">
							<li>• Your profile and account information</li>
							<li>• Your entire anime library and watch history</li>
							<li>• All your settings and preferences</li>
							<li>• Connected OAuth accounts</li>
						</ul>
					</div>
				</div>
			</div>

			<form method="POST" use:enhance class="space-y-4">
				<Form.Field form={sf} name="password">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>Current Password</Form.Label>
							<div class="relative">
								<Input
									{...props}
									type={showPassword ? 'text' : 'password'}
									bind:value={$form.password}
									placeholder="Enter your current password"
									class={cn(
										$errors.password &&
											$errors.password.length > 0 &&
											'border-destructive focus-visible:ring-destructive',
									)}
								/>
								<Button
									variant="ghost"
									size="sm"
									type="button"
									class="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
									onclick={() => (showPassword = !showPassword)}
								>
									{#if showPassword}
										<EyeOff class="h-4 w-4" />
									{:else}
										<Eye class="h-4 w-4" />
									{/if}
								</Button>
							</div>
						{/snippet}
					</Form.Control>
					<Form.Description>Enter your current password to verify your identity</Form.Description>
					<Form.FieldErrors />
				</Form.Field>

				<Form.Field form={sf} name="confirmation">
					<Form.Control>
						{#snippet children({ props })}
							<Form.Label>
								Type <span class="font-mono font-bold">DELETE</span> to confirm:
							</Form.Label>
							<Input
								{...props}
								bind:value={$form.confirmation}
								placeholder="DELETE"
								class={cn(
									'font-mono',
									$errors.confirmation &&
										$errors.confirmation.length > 0 &&
										'border-destructive focus-visible:ring-destructive',
								)}
							/>
						{/snippet}
					</Form.Control>
					<Form.Description>This confirms you understand the consequences</Form.Description>
					<Form.FieldErrors />
				</Form.Field>

				<Button
					type="submit"
					variant="destructive"
					disabled={$submitting || $form.confirmation !== 'DELETE'}
					class="w-full"
				>
					{#if $submitting}
						<LoaderCircle class="mr-2 h-4 w-4 animate-spin" />
					{/if}
					Delete Account
				</Button>
			</form>
		</div>
		<Dialog.Footer>
			<Button variant="outline" onclick={() => (showDeleteDialog = false)}>Cancel</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={showClearDialog}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title class="text-red-400">Clear Library</Dialog.Title>
			<Dialog.Description>
				This action cannot be undone. All anime in your library will be permanently removed.
			</Dialog.Description>
		</Dialog.Header>
		<div class="space-y-4">
			<div class="rounded-lg border border-red-800 bg-red-950 p-4">
				<div class="flex items-start gap-3">
					<TriangleAlert class="mt-0.5 h-5 w-5 text-red-400" />
					<div class="space-y-2">
						<h4 class="font-medium text-red-200">What will be removed:</h4>
						<ul class="space-y-1 text-sm text-red-300">
							<li>• All anime in your library (watching, completed, dropped, etc.)</li>
							<li>• Watch progress and episode counts</li>
							<li>• Personal ratings and notes</li>
							<li>• Continue watching list</li>
						</ul>
					</div>
				</div>
			</div>

			<div class="flex gap-3">
				<Button variant="outline" onclick={() => (showClearDialog = false)} class="flex-1">
					Cancel
				</Button>
				<Button
					variant="destructive"
					onclick={clearLibrary}
					disabled={isClearingLibrary}
					class="flex-1"
				>
					{#if isClearingLibrary}
						<LoaderCircle class="mr-2 h-4 w-4 animate-spin" />
					{/if}
					Clear Library
				</Button>
			</div>
		</div>
	</Dialog.Content>
</Dialog.Root>
