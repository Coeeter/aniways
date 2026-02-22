<script lang="ts">
	import { ArrowLeft, LogOut, Sparkles } from 'lucide-svelte';
	import posthog from 'posthog-js';
	import { toast } from 'svelte-sonner';
	import { goto, invalidate } from '$app/navigation';
	import { apiClient } from '$lib/api/client';
	import Button from '$lib/components/ui/button/button.svelte';

	let isLoggingOut = $state(false);

	async function handleLogout() {
		isLoggingOut = true;
		try {
			await apiClient.POST('/auth/logout');
			posthog.reset();
			await invalidate('app:user');
			await goto('/');
			toast.success('Successfully logged out');
		} catch (error) {
			console.error('Logout error:', error);
			toast.error('Failed to logout. Please try again.');
		} finally {
			isLoggingOut = false;
		}
	}
</script>

<svelte:head>
	<title>Logout - Aniways</title>
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
		<div class="mx-auto w-full max-w-md text-center">
			<div class="mb-8">
				<div class="mb-4 inline-flex items-center gap-2 rounded-full bg-primary/10 px-4 py-2">
					<Sparkles class="h-4 w-4 text-primary" />
					<span class="text-sm font-semibold tracking-wider text-primary uppercase">
						Sign Out
					</span>
				</div>
				<h1 class="mb-2 text-3xl font-bold tracking-tight">Sign Out</h1>
				<p class="text-muted-foreground">Are you sure you want to sign out of your account?</p>
			</div>

			<div class="rounded-2xl border border-primary/10 bg-card/80 p-8 shadow-2xl backdrop-blur-sm">
				<div class="flex flex-col gap-4">
					<Button size="lg" class="w-full gap-2" onclick={handleLogout} disabled={isLoggingOut}>
						{#if isLoggingOut}
							<div
								class="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
							></div>
							Signing Out...
						{:else}
							<LogOut class="h-4 w-4" />
							Yes, Sign Out
						{/if}
					</Button>

					<Button size="lg" variant="outline" href="/" class="w-full gap-2">
						<ArrowLeft class="h-4 w-4" />
						Cancel
					</Button>
				</div>
			</div>
		</div>
	</div>
</div>
