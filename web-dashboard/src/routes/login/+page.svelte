<script lang="ts">
    import { goto } from '$app/navigation';
    import { isAuthenticated } from '$lib/stores';
    import { onMount } from 'svelte';
    import { fade, fly, slide } from 'svelte/transition';
    import { quintOut } from 'svelte/easing';

    let email = 'admin@example.com';
    let password = 'admin123';
    let error = '';
    let loading = false;
    let mounted = false;
    let mouseX = 0;
    let mouseY = 0;
    let formShake = false;
    let isSubmitting = false;
    let particles = Array.from({ length: 50 }, (_, i) => ({
        id: i,
        x: Math.random() * 100,
        y: Math.random() * 100,
        size: Math.random() * 4 + 1,
        speedX: (Math.random() - 0.5) * 0.5,
        speedY: (Math.random() - 0.5) * 0.5
    }));

    onMount(() => {
        mounted = true;
        
        // Mouse tracking for parallax effects
        const handleMouseMove = (e: MouseEvent) => {
            mouseX = (e.clientX / window.innerWidth - 0.5) * 2;
            mouseY = (e.clientY / window.innerHeight - 0.5) * 2;
        };
        
        window.addEventListener('mousemove', handleMouseMove);
        
        // Particle animation
        const animateParticles = () => {
            particles = particles.map(particle => ({
                ...particle,
                x: (particle.x + particle.speedX + 100) % 100,
                y: (particle.y + particle.speedY + 100) % 100
            }));
        };
        
        const interval = setInterval(animateParticles, 50);
        
        return () => {
            window.removeEventListener('mousemove', handleMouseMove);
            clearInterval(interval);
        };
    });

  async function handleLogin(event: Event) {
        event.preventDefault();
        
        if (isSubmitting) return;
        
        isSubmitting = true;
        loading = true;
        error = '';

        const start = Date.now();
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

            const elapsed = Date.now() - start;
            if (elapsed < 800) await new Promise(r => setTimeout(r, 800 - elapsed));

            if (response.redirected && response.url.includes('/login')) {
                 // Add shake animation for failed login
                 formShake = true;
                 setTimeout(() => formShake = false, 600);
                 password = '';
            } else if (response.ok || response.redirected) {
                isAuthenticated.set(true);
                goto('/');
            } else {
                formShake = true;
                setTimeout(() => formShake = false, 600);
                password = '';
            }
        } catch (e) {
            formShake = true;
            setTimeout(() => formShake = false, 600);
        } finally {
            loading = false;
            isSubmitting = false;
        }
    }
</script>

{#if mounted}
<!-- Spectacular Background -->
<div class="fixed inset-0 -z-50 overflow-hidden">
    <!-- Animated Gradient Background -->
    <div class="absolute inset-0 bg-gradient-to-br from-purple-900 via-black to-blue-900 animate-gradient-shift"></div>
    
    <!-- Moving Nebula Clouds -->
    <div class="absolute inset-0">
        <div class="absolute top-0 left-0 w-[60%] h-[40%] bg-blue-600/20 rounded-full blur-[150px] animate-blob" style="animation-delay: 0s;"></div>
        <div class="absolute top-1/4 right-0 w-[50%] h-[35%] bg-purple-600/20 rounded-full blur-[150px] animate-blob" style="animation-delay: 3s;"></div>
        <div class="absolute bottom-0 left-1/3 w-[45%] h-[30%] bg-cyan-600/15 rounded-full blur-[120px] animate-blob" style="animation-delay: 5s;"></div>
    </div>
    
    <!-- Particle Field -->
    <div class="absolute inset-0">
        {#each particles as particle (particle.id)}
            <div 
                class="absolute bg-white rounded-full animate-pulse"
                style="left: {particle.x}%; top: {particle.y}%; width: {particle.size}px; height: {particle.size}px; opacity: 0.6; animation-delay: {particle.id * 0.1}s;"
            ></div>
        {/each}
    </div>
    
    <!-- Grid Pattern Overlay -->
    <div class="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj4KICA8ZyBmaWxsPSJub25lIiBmaWxsLXJ1bGU9ImV2ZW5vZGQiPgogICAgPGcgZmlsbD0iIzY0NzQ4YiIgZmlsbC1vcGFjaXR5PSIwLjAzIj4KICAgICAgPHBhdGggZD0iTTM2IDM0di00aC0ydjRoLTR2Mmg0djRoMnYtNGg0di0yaC00em0wLTMwVjBoLTJ2NGgtNHYyaDR2NGgyVjZoNFY0aC00ek02IDM0di00SDR2NEgwdjJoNHY0aDJ2LTRoNHYtMkg2ek02IDRWMEg0djRIMHYyaDR2NGgyVjZoNFY0SDZ6Ii8+CiAgICA8L2c+CiAgPC9nPgo8L3N2Zz4=')] opacity-20"></div>
    
    <!-- Vignette Effect -->
    <div class="absolute inset-0 bg-radial-gradient from-transparent via-black/20 to-black/60"></div>
</div>

<!-- Main Content -->
<div class="min-h-screen w-full flex items-center justify-center p-4 relative z-10">
    
    <!-- Glass Card Container -->
    <div 
        class="w-full max-w-6xl mx-auto backdrop-blur-xl bg-white/5 rounded-3xl border border-white/10 shadow-2xl overflow-hidden transform transition-all duration-1000 {mounted ? 'translate-y-0 opacity-100 scale-100' : 'translate-y-10 opacity-0 scale-95'}"
        style="transform: perspective(1000px) rotateX({mouseY * 2}deg) rotateY({mouseX * 2}deg);"
    >
        <div class="grid lg:grid-cols-2 gap-0">
            
            <!-- Left Side: Brand Showcase -->
            <div class="relative p-12 lg:p-16 flex flex-col items-center justify-center bg-gradient-to-br from-blue-600/20 to-purple-600/20 border-r border-white/10">
                <!-- Geometric Background Elements -->
                <div class="absolute inset-0 overflow-hidden">
                    <div class="absolute top-10 left-10 w-32 h-32 border-2 border-blue-400/20 rounded-lg transform rotate-45 animate-float"></div>
                    <div class="absolute bottom-10 right-10 w-24 h-24 border-2 border-purple-400/20 rounded-full animate-pulse"></div>
                    <div class="absolute top-1/3 right-1/4 w-20 h-20 border-2 border-cyan-400/20 transform rotate-12 animate-spin" style="animation-duration: 20s;"></div>
                </div>
                
                <div class="relative z-10 text-center space-y-8" in:fly={{ y: -30, duration: 1200, delay: 200 }}>
                    <!-- Floating Logo -->
                    <div class="relative inline-flex">
                        <div class="absolute inset-0 bg-gradient-to-r from-blue-500 to-purple-500 rounded-2xl blur-xl animate-pulse"></div>
                        <div class="relative p-6 bg-white/10 backdrop-blur-lg rounded-2xl border border-white/20 shadow-2xl transform transition-all duration-300 hover:scale-110 hover:rotate-3">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-20 h-20 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                                <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"></path>
                            </svg>
                        </div>
                    </div>
                    
                    <div class="space-y-4">
                        <h1 class="text-6xl font-black text-white tracking-tight">
                            <span class="bg-gradient-to-r from-blue-400 via-cyan-400 to-purple-400 bg-clip-text text-transparent animate-gradient">GoExile</span>
                        </h1>
                        <p class="text-xl text-white/80 font-light max-w-sm mx-auto leading-relaxed">
                            Advanced Game Server Management Registry
                        </p>
                        <div class="flex items-center justify-center gap-6 text-sm text-white/60">
                            <div class="flex items-center gap-2">
                                <div class="w-2 h-2 bg-green-400 rounded-full animate-pulse"></div>
                                <span>Real-time</span>
                            </div>
                            <div class="flex items-center gap-2">
                                <div class="w-2 h-2 bg-blue-400 rounded-full animate-pulse"></div>
                                <span>Secured</span>
                            </div>
                            <div class="flex items-center gap-2">
                                <div class="w-2 h-2 bg-purple-400 rounded-full animate-pulse"></div>
                                <span>Distributed</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <!-- Right Side: Login Form -->
            <div class="relative p-12 lg:p-16 flex flex-col justify-center">
                <!-- Background Decoration -->
                <div class="absolute inset-0 bg-gradient-to-b from-transparent via-white/5 to-transparent"></div>
                
                <div class="relative z-10 space-y-8" in:fly={{ x: 30, duration: 1000, delay: 400 }}>
                    <div class="space-y-2">
                        <h2 class="text-4xl font-bold text-white">Welcome Back</h2>
                        <p class="text-white/60 text-lg">Authenticate to access your dashboard</p>
                    </div>

                    <form onsubmit={handleLogin} class="space-y-6" class:animate-shake={formShake}>
                        <!-- Email Field -->
                        <div class="space-y-2 group">
                            <label for="email" class="text-sm font-medium text-white/80 uppercase tracking-wider">Email Address</label>
                            <div class="relative">
                                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-white/50 group-focus-within:text-blue-400 transition-colors duration-300">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
                                        <polyline points="22,6 12,13 2,6"></polyline>
                                    </svg>
                                </div>
                                <input 
                                    type="email" 
                                    id="email" 
                                    bind:value={email}
                                    required
                                    class="w-full pl-12 pr-4 py-4 bg-white/5 backdrop-blur-sm border border-white/10 rounded-xl text-white placeholder-white/40 focus:outline-none focus:border-blue-400/50 focus:ring-2 focus:ring-blue-400/20 transition-all duration-300 hover:bg-white/10"
                                    placeholder="name@example.com"
                                />
                            </div>
                        </div>

                        <!-- Password Field -->
                        <div class="space-y-2 group">
                            <label for="password" class="text-sm font-medium text-white/80 uppercase tracking-wider">Password</label>
                            <div class="relative">
                                <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none text-white/50 group-focus-within:text-blue-400 transition-colors duration-300">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect>
                                        <path d="M7 11V7a5 5 0 0 1 10 0v4"></path>
                                    </svg>
                                </div>
                                <input 
                                    type="password" 
                                    id="password" 
                                    bind:value={password}
                                    required
                                    class="w-full pl-12 pr-4 py-4 bg-white/5 backdrop-blur-sm border border-white/10 rounded-xl text-white placeholder-white/40 focus:outline-none focus:border-blue-400/50 focus:ring-2 focus:ring-blue-400/20 transition-all duration-300 hover:bg-white/10"
                                    placeholder="••••••••"
                                />
                            </div>
                        </div>

                        <!-- Submit Button -->
                        <button 
                            type="submit" 
                            disabled={loading}
                            class="w-full py-4 px-8 bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-500 hover:to-purple-500 text-white font-bold rounded-xl shadow-xl shadow-blue-500/25 transition-all duration-300 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-3 group relative overflow-hidden"
                        >
                            <!-- Button Background Effect -->
                            <div class="absolute inset-0 bg-gradient-to-r from-white/0 via-white/20 to-white/0 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-700"></div>
                            
                            <div class="relative z-10 flex items-center gap-3">
                                {#if loading}
                                    <div class="w-6 h-6 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                                    <span>Authenticating...</span>
                                {:else}
                                    <span>Secure Login</span>
                                    <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 transition-transform duration-300 group-hover:translate-x-2" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                        <line x1="5" y1="12" x2="19" y2="12"></line>
                                        <polyline points="12 5 19 12 12 19"></polyline>
                                    </svg>
                                {/if}
                            </div>
                        </button>
                    </form>

                    <!-- Security Notice -->
                    <div class="pt-6 border-t border-white/10">
                        <div class="flex items-center gap-3 text-sm text-white/50">
                            <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
                            </svg>
                            <span>Secure connection with end-to-end encryption</span>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</div>

<!-- Floating Elements -->
{#if mounted}
    <div class="fixed inset-0 pointer-events-none z-20">
        <div class="absolute top-20 left-20 w-16 h-16 border-2 border-blue-400/30 rounded-lg animate-float" style="animation-delay: 0s;"></div>
        <div class="absolute top-40 right-32 w-12 h-12 border-2 border-purple-400/30 rounded-full animate-pulse" style="animation-delay: 1s;"></div>
        <div class="absolute bottom-32 left-1/3 w-20 h-20 border-2 border-cyan-400/30 transform rotate-45 animate-spin" style="animation-delay: 2s; animation-duration: 15s;"></div>
    </div>
{/if}

<style>
    @keyframes gradient-shift {
        0%, 100% { background-position: 0% 50%; }
        50% { background-position: 100% 50%; }
    }
    
    @keyframes blob {
        0% { transform: translate(0px, 0px) scale(1); }
        33% { transform: translate(30px, -50px) scale(1.1); }
        66% { transform: translate(-20px, 20px) scale(0.9); }
        100% { transform: translate(0px, 0px) scale(1); }
    }
    
    @keyframes float {
        0%, 100% { transform: translateY(0px) rotate(0deg); }
        50% { transform: translateY(-20px) rotate(5deg); }
    }
    
    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        10%, 30%, 50%, 70%, 90% { transform: translateX(-10px); }
        20%, 40%, 60%, 80% { transform: translateX(10px); }
    }
    
    .animate-gradient-shift {
        background-size: 200% 200%;
        animation: gradient-shift 8s ease infinite;
    }
    
    .animate-blob {
        animation: blob 7s infinite;
    }
    
    .animate-float {
        animation: float 3s ease-in-out infinite;
    }
    
    .animate-shake {
        animation: shake 0.5s ease-in-out;
    }
    
    .bg-radial-gradient {
        background: radial-gradient(circle at center, transparent 0%, rgba(0,0,0,0.2) 50%, rgba(0,0,0,0.6) 100%);
    }
    
    .animate-gradient {
        background-size: 200% 200%;
        animation: gradient-shift 3s ease infinite;
    }
</style>
{/if}
