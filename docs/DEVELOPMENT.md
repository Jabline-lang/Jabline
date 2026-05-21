# Documentación de Jabline

Documentación profesional para Jabline, un lenguaje de programación compilado, cloud-native con máquina virtual personalizada.

## 🚀 Inicio Rápido

### Requisitos Previos

- Node.js 16+ 
- npm o yarn

### Instalación y Ejecución

```bash
# Instalar dependencias
npm install

# Iniciar servidor de desarrollo
npm run dev

# Build para producción
npm run build

# Previsualizar build
npm run preview
```

## 📖 Documentación Disponible

La aplicación incluye:

- **Inicio**: Presentación de Jabline con características principales
- **Guía de Inicio Rápido**: Instalación, variables, funciones y control de flujo
- **Documentación Completa**: Structs, Genéricos, Concurrencia, Módulos, Sistema de Tipos y CLI
- **Standard Library**: Referencia de módulos disponibles (HTTP, JSON, Crypto, DB, IO, etc.)
- **Ejemplos Prácticos**: Código real - Servidor REST, Concurrencia, Procesamiento de Datos, API Client, DB

## 🏗️ Arquitectura

### Modular y Escalable

```
src/
├── components/
│   ├── common/          ← Componentes reutilizables
│   └── layout/          ← Layout components
├── pages/               ← Páginas principales
├── styles/              ← Estilos globales
└── App.jsx              ← Root component
```

### Componentes Principales

- **Button**: Botón reutilizable con variantes
- **CodeBlock**: Bloque de código con syntax highlighting
- **FeatureCard**: Tarjeta de características
- **Section**: Wrapper de sección
- **Container**: Contenedor responsivo
- **Navbar**: Navegación principal
- **Footer**: Pie de página

## 🎨 Stack Tecnológico

- **React 18**: Componentes funcionales y hooks
- **Tailwind CSS**: Utility-first CSS framework
- **Vite**: Herramienta de build ultrarrápida
- **Highlight.js**: Resaltado de sintaxis
- **PostCSS**: Procesamiento de CSS

## 📱 Features

- ✨ Diseño minimalista y profesional
- 🎭 Single Page Application (SPA)
- 📱 100% Responsive
- 🎨 Tema oscuro optimizado
- ⚡ Rendimiento optimizado
- 🔍 Sintaxis resaltada en bloques de código
- 🚀 Animaciones suaves
- ♿ Accesible

## 🔧 Configuración

### Vite (`vite.config.js`)
```js
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    open: true
  }
})
```

### Tailwind (`tailwind.config.js`)
Configurado con extensiones de animaciones personalizadas.

### ESLint (`.eslintrc.json`)
Configurado para React con buenas prácticas.

## 📚 Cómo Agregar Contenido

### Nueva Página

1. Crea `src/pages/NewPage.jsx`
2. Actualiza el switch en `src/App.jsx`
3. Agrega el link en `src/components/layout/Navbar.jsx`

### Nuevo Componente

1. Crea `src/components/common/NewComponent.jsx`
2. Importa en tus páginas
3. Reutiliza en toda la app

## 🌐 Despliegue

### GitHub Pages
```bash
npm run build
# Push dist/ a tu rama gh-pages
```

### Netlify
Conecta el repo a Netlify, configura build command `npm run build` y publish directory `dist`.

### Vercel
```bash
vercel
```

## 📄 Licencia

MIT

---

Para más información, visita [Jabline Repository](https://github.com/Jabline-lang/Jabline)
