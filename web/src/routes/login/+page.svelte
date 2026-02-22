<script lang="ts">
	import { type } from 'arktype';
	import { ArrowRight, Eye, EyeOff, LoaderCircle, Lock, Mail, Sparkles } from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { defaults, setError, superForm } from 'sveltekit-superforms';
	import { arktype, arktypeClient } from 'sveltekit-superforms/adapters';
	import { goto, invalidate } from '$app/navigation';
	import { captureUserLoggedIn } from '$lib/analytics';
	import { apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/openapi';
	import * as Form from '$lib/components/ui/form';
	import Input from '$lib/components/ui/input/input.svelte';
	import { cn } from '$lib/utils';

	const loginFormSchema = type({
		email: type('string.email').describe('a valid email address'),
		password: type('string'),
	});

	const sf = superForm(defaults(arktype(loginFormSchema)), {
		SPA: true,
		validators: arktypeClient(loginFormSchema),
		onUpdate: async ({ form, cancel }) => {
			if (!form.valid) return;
			try {
				const res = await apiClient.POST('/auth/login', {
					body: form.data,
				});

				if (res.response.status === 200 && res.data) {
					captureUserLoggedIn();
					await invalidate('app:user');
					await goto('/');
					toast.success('Successfully logged in');
					return;
				}

				if (res.response.status === 400) {
					const error = res.error as components['schemas']['models.ValidationErrorResponse'];
					if (error?.details) {
						Object.entries(error.details).forEach(([field, messages]) => {
							setError(form, field as 'email' | 'password', messages);
						});
						toast.error('Please fix the errors in the form and try again.');
						cancel();
						return;
					}
				}

				if (res.response.status === 401) {
					toast.error('Invalid email or password. Please try again.');
					cancel();
					return;
				}

				throw new Error('Unexpected response from server');
			} catch (error) {
				console.error('Login error:', error);
				toast.error('Login failed. Please check your credentials and try again.');
				cancel();
			}
		},
	});

	const { form, enhance, submitting, errors } = sf;
	let showPassword = $state(false);
</script>

<svelte:head>
	<title>Sign In - Aniways</title>
	<meta
		name="description"
		content="Sign in to your Aniways account to access your personalized anime experience."
	/>
</svelte:head>

<div
	class="relative min-h-screen overflow-hidden bg-gradient-to-br from-background via-background to-primary/10"
>
	<div class="pointer-events-none absolute inset-0 overflow-hidden">
		<div
			class="absolute -top-20 -right-20 h-64 w-64 animate-pulse rounded-full bg-primary/5 blur-3xl"
		></div>
		<div
			class="absolute top-1/2 -left-20 h-48 w-48 animate-pulse rounded-full bg-secondary/10 blur-2xl"
			style="animation-delay: 1s;"
		></div>
	</div>

	<div class="relative z-10 container mx-auto px-4 py-8">
		<div class="mx-auto w-full max-w-md">
			<div class="mb-8 text-center">
				<div class="mb-4 inline-flex items-center gap-2 rounded-full bg-primary/10 px-4 py-2">
					<Sparkles class="h-4 w-4 text-primary" />
					<span class="text-sm font-semibold tracking-wider text-primary uppercase">
						Welcome Back
					</span>
				</div>
				<h1 class="mb-2 text-3xl font-bold tracking-tight">Sign In</h1>
				<p class="text-muted-foreground">Continue your anime journey with Aniways</p>
			</div>

			<div class="rounded-2xl border border-primary/10 bg-card/80 p-8 shadow-2xl backdrop-blur-sm">
				<form method="POST" use:enhance class="space-y-6">
					<Form.Field form={sf} name="email">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Email Address</Form.Label>
								<div class="relative">
									<Mail
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										bind:value={$form.email}
										placeholder="your@email.com"
										class={cn(
											'pl-10',
											$errors.email &&
												$errors.email.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
									/>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field form={sf} name="password">
						<Form.Control>
							{#snippet children({ props })}
								<div class="flex items-center justify-between">
									<Form.Label>Password</Form.Label>
									<a href="/forgot-password" class="text-sm text-primary hover:underline">
										Forgot password?
									</a>
								</div>
								<div class="relative">
									<Lock
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										type={showPassword ? 'text' : 'password'}
										bind:value={$form.password}
										placeholder="Enter your password"
										class={cn(
											'pr-10 pl-10',
											$errors.password &&
												$errors.password.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
									/>
									<button
										type="button"
										onclick={() => (showPassword = !showPassword)}
										class="absolute top-1/2 right-3 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
										disabled={$submitting}
									>
										{#if showPassword}
											<EyeOff class="h-4 w-4" />
										{:else}
											<Eye class="h-4 w-4" />
										{/if}
									</button>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Button size="lg" class="group w-full gap-2" disabled={$submitting}>
						{#if $submitting}
							<LoaderCircle class="size-4 animate-spin" />
							Signing In...
						{:else}
							Sign In
							<ArrowRight class="size-4 transition group-hover:translate-x-1" />
						{/if}
					</Form.Button>
				</form>

				<div class="my-8 flex items-center">
					<div class="flex-1 border-t border-muted"></div>
					<div class="px-4 text-sm text-muted-foreground">or</div>
					<div class="flex-1 border-t border-muted"></div>
				</div>

				<div class="text-center">
					<p class="text-sm text-muted-foreground">
						Don't have an account?
						<a href="/register" class="font-medium text-primary transition-colors hover:underline">
							Create Account
						</a>
					</p>
				</div>
			</div>
		</div>
	</div>
</div>
