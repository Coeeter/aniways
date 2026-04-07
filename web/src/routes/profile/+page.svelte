<script lang="ts">
	import { type } from 'arktype';
	import { formatDate } from 'date-fns';
	import {
		Calendar,
		Clock,
		Heart,
		LoaderCircle,
		Pencil,
		Play,
		Settings,
		Star,
		Trash,
		Upload,
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { defaults, setError, superForm } from 'sveltekit-superforms';
	import { arktype, arktypeClient } from 'sveltekit-superforms/adapters';
	import { invalidate } from '$app/navigation';
	import { apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/openapi';
	import PageHeader from '$lib/components/layout/page-header.svelte';
	import Button, { buttonVariants } from '$lib/components/ui/button/button.svelte';
	import * as Card from '$lib/components/ui/card';
	import * as Dropdown from '$lib/components/ui/dropdown-menu';
	import * as Form from '$lib/components/ui/form';
	import Input from '$lib/components/ui/input/input.svelte';
	import { Separator } from '$lib/components/ui/separator';
	import { cn } from '$lib/utils';
	import type { PageProps } from './$types';

	let { data }: PageProps = $props();

	const profileFormSchema = type({
		username: type('string').describe('a valid username'),
		email: type('string.email').describe('a valid email address'),
	});

	const sf = superForm(defaults(arktype(profileFormSchema)), {
		SPA: true,
		resetForm: false,
		validators: arktypeClient(profileFormSchema),
		onUpdate: async ({ form, cancel }) => {
			if (!form.valid) return;
			try {
				const res = await apiClient.PUT('/users', {
					body: form.data,
				});

				if (res.response.status === 200) {
					await invalidate('app:user');
					toast.success('Profile updated successfully');
					return;
				}

				if (res.response.status === 400) {
					const error = res.error as components['schemas']['models.ValidationErrorResponse'];
					if (error?.details) {
						Object.entries(error.details).forEach(([field, messages]) => {
							setError(form, field as 'username' | 'email', messages);
						});
						toast.error('Please fix the errors in the form and try again.');
						cancel();
						return;
					}
				}

				toast.error('Failed to update profile. Please try again.');
				cancel();
			} catch {
				toast.error('An error occurred. Please try again.');
				cancel();
			}
		},
	});

	const { form, enhance, submitting, errors } = sf;

	$effect(() => {
		if (!data.user) return;
		form.set({
			username: data.user.username || '',
			email: data.user.email || '',
		});
	});

	let hasChanges = $derived(
		$form.username !== data.user?.username || $form.email !== data.user?.email,
	);

	let isUploadingImage = $state(false);

	async function handleImageUpload(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;

		if (file.size > 5 * 1024 * 1024) {
			toast.error('Image size must be less than 5MB');
			return;
		}

		isUploadingImage = true;
		try {
			const formData = new FormData();
			formData.append('image', file);

			const res = await apiClient.PUT('/users/image', {
				// @ts-expect-error FormData is supported
				body: formData,
			});

			if (res.response.ok) {
				await invalidate('app:user');
				toast.success('Profile picture updated successfully');
			} else {
				toast.error('Failed to update profile picture');
			}
		} catch {
			toast.error('An error occurred while uploading image');
		} finally {
			isUploadingImage = false;
		}
	}

	async function removeProfilePicture() {
		await apiClient.DELETE('/users/image');
		await invalidate('app:user');
		toast.success('Profile picture removed successfully');
	}
</script>

<svelte:head>
	<title>Profile - Aniways</title>
	<meta name="description" content="Manage your profile and view your anime library statistics" />
</svelte:head>

<div class="min-h-screen bg-background">
	<PageHeader title="Profile" description="Manage your account and view your anime library" />

	<div class="container mx-auto px-4 py-8">
		<div class="grid gap-8 lg:grid-cols-3">
			<div class="lg:col-span-1">
				<Card.Root class="p-6">
					<div class="flex flex-col items-center space-y-4">
						<div class="relative">
							{#if data.user?.profilePicture}
								<img
									src={data.user.profilePicture}
									alt={data.user?.username}
									class="aspect-square h-48 rounded-full border-2 border-border object-cover"
								/>
							{:else}
								<div
									class="flex aspect-square h-48 items-center justify-center rounded-full border-2 border-border bg-primary/10 text-6xl font-bold text-primary"
								>
									{data.user?.username
										?.split(' ')
										.map((n) => n.charAt(0))
										.join('')
										.toUpperCase()
										.slice(0, 2)}
								</div>
							{/if}
							<Dropdown.Root>
								<Dropdown.Trigger
									class={cn(
										buttonVariants({ size: 'icon' }),
										'absolute right-0 bottom-0 flex size-8 -translate-x-1/2 -translate-y-1/2 rounded-full',
									)}
									disabled={isUploadingImage}
								>
									{#if isUploadingImage}
										<LoaderCircle class="size-4 animate-spin" />
									{:else}
										<Pencil class="size-4" />
									{/if}
								</Dropdown.Trigger>
								<Dropdown.Content align="end" class="w-48">
									<Dropdown.Item>
										<label for="file-input" class="flex w-full cursor-pointer items-center gap-2">
											<Upload class="h-4 w-4" />
											Upload Picture
										</label>
									</Dropdown.Item>
									{#if data.user?.profilePicture}
										<Dropdown.Item onclick={removeProfilePicture}>
											<Trash class="h-4 w-4" />
											Remove Picture
										</Dropdown.Item>
									{/if}
								</Dropdown.Content>
							</Dropdown.Root>
							<input
								id="file-input"
								type="file"
								accept="image/*"
								class="hidden"
								onchange={handleImageUpload}
								disabled={isUploadingImage}
							/>
						</div>

						<div class="text-center">
							<h2 class="text-lg font-semibold">{data.user?.username || 'Anonymous User'}</h2>
							<p class="text-sm text-muted-foreground">{data.user?.email}</p>
						</div>

						<Separator />

						<div class="w-full space-y-3">
							<div class="flex items-center gap-3 text-sm">
								<Calendar class="h-4 w-4 text-muted-foreground" />
								<span class="text-muted-foreground">Joined</span>
								<span class="font-medium"
									>{data.user?.createdAt
										? formatDate(data.user.createdAt, 'dd/MM/yyyy')
										: 'Unknown'}
								</span>
							</div>
						</div>
					</div>
				</Card.Root>

				<Card.Root class="mt-8 p-6">
					<h3 class="mb-4 font-semibold">Quick Actions</h3>
					<div class="space-y-3">
						<Button variant="outline" class="w-full justify-start" href="/settings">
							<Settings class="mr-2 h-4 w-4" />
							Account Settings
						</Button>
						<Button variant="outline" class="w-full justify-start" href="/my-list">
							<Heart class="mr-2 h-4 w-4" />
							My Library
						</Button>
						<Button variant="outline" class="w-full justify-start" href="/continue-watching">
							<Play class="mr-2 h-4 w-4" />
							Continue Watching
						</Button>
						<Button variant="outline" class="w-full justify-start" href="/planning">
							<Clock class="mr-2 h-4 w-4" />
							Planning List
						</Button>
					</div>
				</Card.Root>
			</div>

			<div class="lg:col-span-2">
				<Card.Root class="p-6">
					<h2 class="mb-4 text-xl font-semibold">Update Profile</h2>
					<form method="POST" use:enhance class="space-y-4">
						<Form.Field form={sf} name="username">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Username</Form.Label>
									<Input
										{...props}
										bind:value={$form.username}
										placeholder="Enter your username"
										class={cn(
											$errors.username &&
												$errors.username.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
									/>
								{/snippet}
							</Form.Control>
							<Form.Description>This is your public display name</Form.Description>
							<Form.FieldErrors />
						</Form.Field>

						<Form.Field form={sf} name="email">
							<Form.Control>
								{#snippet children({ props })}
									<Form.Label>Email</Form.Label>
									<Input
										{...props}
										type="email"
										bind:value={$form.email}
										placeholder="Enter your email"
										class={cn(
											$errors.email &&
												$errors.email.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
									/>
								{/snippet}
							</Form.Control>
							<Form.Description>We'll never share your email with anyone</Form.Description>
							<Form.FieldErrors />
						</Form.Field>

						<Button type="submit" disabled={$submitting || !hasChanges} class="w-full">
							{#if $submitting}
								<LoaderCircle class="mr-2 h-4 w-4 animate-spin" />
							{/if}
							Update Profile
						</Button>
					</form>
				</Card.Root>

				<div class="mt-8">
					<h2 class="mb-4 text-xl font-semibold">Library Statistics</h2>
					<div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						<Card.Root class="p-4">
							<div class="flex items-center gap-3">
								<div class="flex h-12 w-12 items-center justify-center rounded-full bg-blue-500/10">
									<Play class="h-6 w-6 text-blue-500" />
								</div>
								<div>
									<p class="text-sm font-medium text-muted-foreground">Currently Watching</p>
									<p class="text-2xl font-bold">{data.stats.watching}</p>
								</div>
							</div>
						</Card.Root>

						<Card.Root class="p-4">
							<div class="flex items-center gap-3">
								<div
									class="flex h-12 w-12 items-center justify-center rounded-full bg-orange-500/10"
								>
									<Clock class="h-6 w-6 text-orange-500" />
								</div>
								<div>
									<p class="text-sm font-medium text-muted-foreground">Planning to Watch</p>
									<p class="text-2xl font-bold">{data.stats.planning}</p>
								</div>
							</div>
						</Card.Root>

						<Card.Root class="p-4">
							<div class="flex items-center gap-3">
								<div
									class="flex h-12 w-12 items-center justify-center rounded-full bg-green-500/10"
								>
									<Star class="h-6 w-6 text-green-500" />
								</div>
								<div>
									<p class="text-sm font-medium text-muted-foreground">Completed</p>
									<p class="text-2xl font-bold">{data.stats.completed}</p>
								</div>
							</div>
						</Card.Root>
					</div>
				</div>

				<div class="mt-8">
					<h3 class="mb-4 text-lg font-semibold">Your Anime Journey</h3>
					<Card.Root class="p-6">
						<div class="py-8 text-center">
							<Heart class="mx-auto mb-4 h-12 w-12 text-muted-foreground" />
							<h4 class="mb-2 text-lg font-medium">Keep Exploring!</h4>
							<p class="mb-4 text-sm text-muted-foreground">
								You have {data.stats.watching + data.stats.planning + data.stats.completed} anime in your
								library
							</p>
							<div class="flex justify-center gap-2">
								<Button variant="outline" href="/my-list">
									<Heart class="mr-2 h-4 w-4" />
									View Library
								</Button>
								<Button href="/catalog">
									<Play class="mr-2 h-4 w-4" />
									Discover More
								</Button>
							</div>
						</div>
					</Card.Root>
				</div>
			</div>
		</div>
	</div>
</div>
