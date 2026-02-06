import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import type { Config } from '@sveltejs/kit';

const config: Config = {
  preprocess: vitePreprocess(),

  kit: {
    adapter: adapter({
      pages: 'build',
      assets: 'build',
      fallback: 'index.html',  // SPA fallback for client-side routing
      precompress: false,
      strict: false
    }),

    prerender: {
      entries: []  // Disable prerendering - pure SPA mode
    }
  }
};

export default config;
