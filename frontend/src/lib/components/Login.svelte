<script>
  import Logo from "./Logo.svelte"; // Login component
</script>

<div class="min-h-screen flex items-center justify-center bg-slate-50 p-4">
  <div class="max-w-md w-full glass-card p-10 rounded-2xl shadow-xl text-center">
    <div class="mb-8 flex justify-center">
      <div class="w-20 h-20 bg-gradient-to-br from-indigo-600 to-purple-600 rounded-2xl flex items-center justify-center transform rotate-3 shadow-lg">
        <Logo className="w-10 h-10 text-white" />
      </div>
    </div>
    
    
    <h1 class="text-4xl font-black text-transparent bg-clip-text bg-gradient-to-r from-indigo-600 to-purple-600 mb-3 tracking-tight">
      PackUp
    </h1>
    <p class="text-slate-500 mb-8 font-medium">
      Your journey, organized. 
      <br /> 
      Track tasks as you travel through life.
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
