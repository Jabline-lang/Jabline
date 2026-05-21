// Constantes para la aplicación

export const NAVIGATION_ITEMS = [
  { id: 'home', label: 'Inicio' },
  { id: 'guide', label: 'Guía' },
  { id: 'docs', label: 'Documentación' },
  { id: 'stdlib', label: 'Stdlib' },
  { id: 'examples', label: 'Ejemplos' },
]

export const SOCIAL_LINKS = [
  { label: 'GitHub', url: 'https://github.com/Jabline-lang/Jabline' },
  { label: 'Registry', url: 'https://github.com/Jabline-lang/registry' },
  { label: 'Issues', url: 'https://github.com/Jabline-lang/Jabline/issues' },
]

export const FEATURES = [
  {
    icon: '⚡',
    title: 'Compilado a Bytecode',
    description: 'Más rápido que intérpretes tradicionales. Ejecuta en JBVM.',
  },
  {
    icon: '🔀',
    title: 'Concurrencia Nativa',
    description: 'Modelo CSP con spawn y canales. Escala sin fricción.',
  },
  {
    icon: '📦',
    title: 'Toolchain Completo',
    description: 'LSP, formateador, testing, package manager integrados.',
  },
]

export const APP_VERSION = '0.6.0'
export const APP_TAGLINE = 'Un lenguaje de programación compilado, cloud-native, con máquina virtual bytecode personalizada.'
