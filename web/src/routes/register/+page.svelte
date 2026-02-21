<script lang="ts">
	import { type } from 'arktype';
	import {
		ArrowRight,
		Check,
		Eye,
		EyeOff,
		LoaderCircle,
		Lock,
		Mail,
		Sparkles,
		User,
	} from 'lucide-svelte';
	import { toast } from 'svelte-sonner';
	import { defaults, setError, superForm } from 'sveltekit-superforms';
	import { arktype, arktypeClient } from 'sveltekit-superforms/adapters';
	import { goto } from '$app/navigation';
	import { apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/openapi';
	import { captureUserRegistered } from '$lib/analytics';
	import * as Form from '$lib/components/ui/form';
	import Input from '$lib/components/ui/input/input.svelte';
	import { cn } from '$lib/utils';

	const registerSchema = type({
		email: type('string.email').describe('a valid email address'),
		username: type('3<=string<=20').describe('a username between 3-20 characters'),
		password: type('string>=8&/^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).*$/').describe(
			'a strong password with uppercase, lowercase, and number',
		),
		confirmPassword: 'string',
	}).narrow((data, ctx) => {
		if (data.password !== data.confirmPassword) {
			return ctx.mustBe('Passwords do not match');
		}
		return true;
	});

	const sf = superForm(defaults(arktype(registerSchema)), {
		SPA: true,
		validators: arktypeClient(registerSchema),
		onUpdate: async ({ form, cancel }) => {
			if (!form.valid) return;
			try {
				const res = await apiClient.POST('/users', {
					body: {
						email: form.data.email,
						username: form.data.username,
						password: form.data.password,
					},
				});

				if (res.response.status === 200 && res.data) {
					captureUserRegistered();
					toast.success('Account created successfully!');
					await goto('/login');
					return;
				}

				if (res.response.status === 400) {
					const error = res.error as components['schemas']['models.ValidationErrorResponse'];
					if (error?.details) {
						Object.entries(error.details).forEach(([field, messages]) => {
							setError(form, field as keyof typeof form.data, messages);
						});
						toast.error('Please fix the errors in the form and try again.');
						cancel();
						return;
					}
				}

				if (res.response.status === 409) {
					toast.error('An account with this email or username already exists');
					cancel();
					return;
				}

				throw new Error('Unexpected response from server');
			} catch (error) {
				console.error('Registration error:', error);
				toast.error('Registration failed. Please try again.');
				cancel();
			}
		},
	});

	const { form, enhance, submitting, errors } = sf;

	let showPassword = $state(false);
	let showConfirmPassword = $state(false);

	// Password validation for visual feedback
	let passwordValidation = $derived.by(() => ({
		length: $form.password.length >= 8,
		hasUppercase: /[A-Z]/.test($form.password),
		hasLowercase: /[a-z]/.test($form.password),
		hasNumber: /\d/.test($form.password),
	}));

	let passwordMatch = $derived(
		$form.password === $form.confirmPassword && $form.confirmPassword.length > 0,
	);
</script>

<svelte:head>
	<title>Create Account - Aniways</title>
	<meta
		name="description"
		content="Join Aniways to discover, watch, and track your favorite anime series and movies."
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
		<div
			class="absolute right-1/4 bottom-20 h-32 w-32 animate-pulse rounded-full bg-accent/10 blur-xl"
			style="animation-delay: 2s;"
		></div>
	</div>

	<div class="relative z-10 container mx-auto px-4 py-8">
		<div class="mx-auto w-full max-w-md">
			<div class="mb-8 text-center">
				<div class="mb-4 inline-flex items-center gap-2 rounded-full bg-primary/10 px-4 py-2">
					<Sparkles class="h-4 w-4 text-primary" />
					<span class="text-sm font-semibold tracking-wider text-primary uppercase">
						Join Aniways
					</span>
				</div>
				<h1 class="mb-2 text-3xl font-bold tracking-tight">Create Account</h1>
				<p class="text-muted-foreground">Start your personalized anime journey today</p>
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
										type="email"
										bind:value={$form.email}
										placeholder="your@email.com"
										class={cn(
											'pl-10',
											$errors.email &&
												$errors.email.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
										disabled={$submitting}
									/>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field form={sf} name="username">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Username</Form.Label>
								<div class="relative">
									<User
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										type="text"
										bind:value={$form.username}
										placeholder="otaku_master"
										class={cn(
											'pl-10',
											$errors.username &&
												$errors.username.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
										disabled={$submitting}
									/>
								</div>
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Field form={sf} name="password">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Password</Form.Label>
								<div class="relative">
									<Lock
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										type={showPassword ? 'text' : 'password'}
										bind:value={$form.password}
										placeholder="Create a strong password"
										class={cn(
											'pr-10 pl-10',
											$errors.password &&
												$errors.password.length > 0 &&
												'border-destructive focus-visible:ring-destructive',
										)}
										disabled={$submitting}
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

								{#if $form.password.length > 0}
									<div class="mt-3 space-y-2 rounded-lg bg-muted/30 p-4 backdrop-blur-sm">
										<p class="text-sm font-medium text-muted-foreground">Password Requirements:</p>
										<div class="grid grid-cols-2 gap-2 text-sm">
											<div class="flex items-center gap-2">
												<Check
													class="h-4 w-4 {passwordValidation.length
														? 'text-green-500'
														: 'text-muted-foreground'}"
												/>
												<span
													class={passwordValidation.length
														? 'text-green-600'
														: 'text-muted-foreground'}
												>
													8+ characters
												</span>
											</div>
											<div class="flex items-center gap-2">
												<Check
													class="h-4 w-4 {passwordValidation.hasUppercase
														? 'text-green-500'
														: 'text-muted-foreground'}"
												/>
												<span
													class={passwordValidation.hasUppercase
														? 'text-green-600'
														: 'text-muted-foreground'}
												>
													Uppercase
												</span>
											</div>
											<div class="flex items-center gap-2">
												<Check
													class="h-4 w-4 {passwordValidation.hasLowercase
														? 'text-green-500'
														: 'text-muted-foreground'}"
												/>
												<span
													class={passwordValidation.hasLowercase
														? 'text-green-600'
														: 'text-muted-foreground'}
												>
													Lowercase
												</span>
											</div>
											<div class="flex items-center gap-2">
												<Check
													class="h-4 w-4 {passwordValidation.hasNumber
														? 'text-green-500'
														: 'text-muted-foreground'}"
												/>
												<span
													class={passwordValidation.hasNumber
														? 'text-green-600'
														: 'text-muted-foreground'}
												>
													Number
												</span>
											</div>
										</div>
									</div>
								{/if}
							{/snippet}
						</Form.Control>
					</Form.Field>

					<Form.Field form={sf} name="confirmPassword">
						<Form.Control>
							{#snippet children({ props })}
								<Form.Label>Confirm Password</Form.Label>
								<div class="relative">
									<Lock
										class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground"
									/>
									<Input
										{...props}
										type={showConfirmPassword ? 'text' : 'password'}
										bind:value={$form.confirmPassword}
										placeholder="Confirm your password"
										class={cn(
											'pr-10 pl-10',
											($errors.confirmPassword && $errors.confirmPassword.length > 0) ||
												($form.confirmPassword.length > 0 && !passwordMatch)
												? 'border-destructive focus-visible:ring-destructive'
												: passwordMatch
													? 'border-green-500 focus-visible:ring-green-500'
													: '',
										)}
										disabled={$submitting}
									/>
									<button
										type="button"
										onclick={() => (showConfirmPassword = !showConfirmPassword)}
										class="absolute top-1/2 right-3 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
										disabled={$submitting}
									>
										{#if showConfirmPassword}
											<EyeOff class="h-5 w-5" />
										{:else}
											<Eye class="h-5 w-5" />
										{/if}
									</button>
								</div>
								{#if $form.confirmPassword.length > 0}
									<div class="mt-2 flex items-center gap-2 text-sm">
										<Check
											class="h-4 w-4 {passwordMatch ? 'text-green-500' : 'text-muted-foreground'}"
										/>
										<span class={passwordMatch ? 'text-green-600' : 'text-muted-foreground'}>
											Passwords match
										</span>
									</div>
								{/if}
							{/snippet}
						</Form.Control>
						<Form.FieldErrors />
					</Form.Field>

					<Form.Button type="submit" size="lg" class="group w-full gap-2" disabled={$submitting}>
						{#if $submitting}
							<LoaderCircle class="size-4 animate-spin" />
							Creating Account...
						{:else}
							Create Account
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
						Already have an account?
						<a href="/login" class="font-medium text-primary transition-colors hover:underline">
							Sign In
						</a>
					</p>
				</div>
			</div>
		</div>
	</div>
</div>
