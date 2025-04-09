// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },

  // Runtime config
  runtimeConfig: {
    // Private keys are only available on the server
    // Public keys will be exposed to the client
    public: {
      apiUrl: 'http://localhost:8080/api/v1'
    }
  },

  modules: [
    '@nuxt/fonts',
    '@nuxt/icon',
    '@nuxt/image',
    '@nuxt/eslint',
    '@nuxtjs/tailwindcss',    
    '@nuxtjs/color-mode',
    'shadcn-nuxt',
  ],
  
  colorMode: {
    preference: 'system',
    fallback: 'light',
    classSuffix: '',
    storage: 'cookie',
    storageKey: 'nuxt-color-mode',
  },
  
  shadcn: {
    /**
     * Prefix for all the imported component
     */
    prefix: '',
    /**
     * Directory that the component lives in.
     * @default "./components/ui"
     */
    componentDir: './components/ui'
  }
})