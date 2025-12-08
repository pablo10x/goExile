<script lang="ts">
    import { goto } from '$app/navigation';
    import { isAuthenticated } from '$lib/stores';

    let email = 'admin@example.com';
    let password = 'admin123';
    let error = '';
    let loading = false;

    async function handleLogin(event: Event) {
        event.preventDefault();
        loading = true;
        error = '';

        const formData = new URLSearchParams();
        formData.append('email', email);
        formData.append('password', password);

        try {
            const response = await fetch('/login', {
                method: 'POST',
                body: formData,
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                }
            });

            if (response.redirected && response.url.includes('/login')) {
                 // Redirected back to login implies failure
                 error = 'Invalid credentials';
            } else if (response.ok || response.redirected) {
                // Success
                isAuthenticated.set(true);
                goto('/');
            } else {
                error = 'Login failed';
            }
        } catch (e) {
            error = 'An unexpected error occurred';
        } finally {
            loading = false;
        }
    }
</script>

<div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
    <div class="card p-8 w-full max-w-md">
        <h1 class="text-3xl font-bold mb-6 text-center text-slate-50">GoExile Admin</h1>
        
        {#if error}
            <div class="bg-red-500/10 border border-red-500/50 text-red-400 px-4 py-3 rounded mb-4 text-sm">
                {error}
            </div>
        {/if}

        <form onsubmit={handleLogin} class="space-y-6">
            <div>
                <label for="email" class="block text-sm font-medium text-slate-300 mb-2">Email</label>
                <input 
                    type="email" 
                    id="email" 
                    bind:value={email}
                    required
                    class="w-full px-4 py-2 bg-slate-800/50 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                    placeholder="admin@example.com"
                />
            </div>
            
            <div>
                <label for="password" class="block text-sm font-medium text-slate-300 mb-2">Password</label>
                <input 
                    type="password" 
                    id="password" 
                    bind:value={password}
                    required
                    class="w-full px-4 py-2 bg-slate-800/50 border border-slate-600 rounded-lg text-slate-200 placeholder-slate-500 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500 transition-colors"
                    placeholder="••••••••"
                />
            </div>

            <button 
                type="submit" 
                disabled={loading}
                class="btn-primary w-full justify-center disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {loading ? 'Logging in...' : 'Login'}
            </button>
        </form>
    </div>
</div>
