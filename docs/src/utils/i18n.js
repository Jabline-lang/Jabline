// i18n configuration - Structure for future implementation (v0.9.0+)

export const SUPPORTED_LANGUAGES = {
  es: 'Español',
  en: 'English',
}

export const DEFAULT_LANGUAGE = 'es'

export const translations = {
  es: {
    nav: {
      home: 'Inicio',
      guide: 'Guía',
      docs: 'Documentación',
      stdlib: 'Standard Library',
      examples: 'Ejemplos',
      blog: 'Blog',
      community: 'Comunidad',
    },
    home: {
      title: 'Jabline',
      subtitle: 'Un lenguaje de programación compilado y cloud-native con máquina virtual bytecode personalizada.',
      cta_start: 'Comenzar',
      cta_github: 'Ver en GitHub',
      version: 'v0.6.0 • Production Ready',
    },
    blog: {
      title: 'Blog de Jabline',
      subtitle: 'Artículos, tutoriales y noticias sobre Jabline.',
      newsletter: 'Suscríbete al Newsletter',
      newsletter_desc: 'Recibe las últimas noticias sobre Jabline directamente en tu email.',
    },
    community: {
      title: 'Comunidad de Jabline',
      contributors: 'Contribuidores',
      code_of_conduct: 'Código de Conducta',
      get_involved: 'Cómo Participar',
    },
    footer: {
      rights: 'Todos los derechos reservados.',
      license: 'Licencia MIT',
    },
  },
  en: {
    nav: {
      home: 'Home',
      guide: 'Guide',
      docs: 'Documentation',
      stdlib: 'Standard Library',
      examples: 'Examples',
      blog: 'Blog',
      community: 'Community',
    },
    home: {
      title: 'Jabline',
      subtitle: 'A compiled and cloud-native programming language with custom bytecode virtual machine.',
      cta_start: 'Get Started',
      cta_github: 'View on GitHub',
      version: 'v0.6.0 • Production Ready',
    },
    blog: {
      title: 'Jabline Blog',
      subtitle: 'Articles, tutorials and news about Jabline.',
      newsletter: 'Subscribe to Newsletter',
      newsletter_desc: 'Get the latest news about Jabline directly in your email.',
    },
    community: {
      title: 'Jabline Community',
      contributors: 'Contributors',
      code_of_conduct: 'Code of Conduct',
      get_involved: 'How to Participate',
    },
    footer: {
      rights: 'All rights reserved.',
      license: 'MIT License',
    },
  },
}

// Hook to get current language from localStorage
export function useLanguage() {
  const [lang, setLang] = React.useState(() => {
    const saved = localStorage.getItem('language')
    return saved || DEFAULT_LANGUAGE
  })

  const changeLang = (newLang) => {
    if (SUPPORTED_LANGUAGES[newLang]) {
      setLang(newLang)
      localStorage.setItem('language', newLang)
    }
  }

  return { lang, changeLang, t: translations[lang] }
}

// Utility to get translation key value
export function t(lang, path) {
  const keys = path.split('.')
  let value = translations[lang]
  
  for (const key of keys) {
    value = value?.[key]
  }
  
  return value || path
}
