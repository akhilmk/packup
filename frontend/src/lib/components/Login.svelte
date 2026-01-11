<script lang="ts">
  // Login component
</script>

<div class="min-h-screen flex items-center justify-center bg-slate-50 p-4">
  <div class="max-w-md w-full glass-card p-10 rounded-2xl shadow-xl text-center">
    <div class="mb-8 flex justify-center">
      <div class="w-16 h-16 bg-indigo-600 rounded-2xl flex items-center justify-center transform rotate-6">
        <svg xmlns="http://www.w3.org/2000/svg" class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
      </div>
    </div>
    
    <h1 class="text-3xl font-black text-slate-800 mb-3 tracking-tight">
      Focus Flow
    </h1>
    <p class="text-slate-500 mb-8 font-medium">
      Your distraction-free task master. <br /> Sign in to sync your productivity.
    </p>

    <a 
      href="/api/auth/google/login"
      class="flex items-center justify-center w-full px-4 py-3 border border-slate-200 rounded-xl shadow-sm bg-white hover:bg-slate-50 text-slate-700 font-bold transition-all duration-200 transform hover:-translate-y-0.5"
    >
      <img src="https://www.svgrepo.com/show/475656/google-color.svg" class="w-5 h-5 mr-3" alt="Google logo" />
      <span>Continue with Google</span>
    </a>
    
    <p class="mt-8 text-xs text-slate-400 font-medium max-w-xs mx-auto">
      By continuing, you agree to organize your life one task at a time.
    </p>
  </div>
</div>

<style>
  .glass-card {
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
  }
</style>
