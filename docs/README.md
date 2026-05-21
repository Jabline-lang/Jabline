# 📚 Documentación de Jabline - Aplicación React

Documentación profesional, minimalista y tecnológica del lenguaje de programación Jabline, construida con React, Tailwind CSS y Vite.

## 🎯 Características

- ⚡ **React 18** - Componentes modernos con hooks
- 🎨 **Tailwind CSS** - Estilos minimalistas y responsivos
- 🔥 **Vite** - Build tool ultra rápido
- 💻 **Componentes Modulares** - Arquitectura escalable
- 🎭 **SPA** - Single Page Application sin recargas
- ✨ **Animaciones Suaves** - Gradientes y transiciones
- 📱 **100% Responsive** - Funciona en todos los dispositivos

## 📂 Estructura del Proyecto

```
docs/
├── src/
│   ├── components/
│   │   ├── common/              # Componentes reutilizables
│   │   │   ├── Button.jsx
│   │   │   ├── CodeBlock.jsx
│   │   │   ├── FeatureCard.jsx
│   │   │   └── Section.jsx
│   │   └── layout/              # Componentes de layout
│   │       ├── Container.jsx
│   │       ├── Footer.jsx
│   │       └── Navbar.jsx
│   ├── pages/                   # Páginas principales
│   │   ├── Docs.jsx
│   │   ├── Examples.jsx
│   │   ├── Guide.jsx
│   │   ├── Home.jsx
│   │   └── Stdlib.jsx
│   ├── styles/
│   │   └── globals.css          # Estilos globales
│   ├── App.jsx                  # Componente raíz
│   └── main.jsx                 # Punto de entrada
├── index.html                   # HTML base
├── package.json                 # Dependencias
├── vite.config.js              # Configuración Vite
├── tailwind.config.js          # Configuración Tailwind
└── postcss.config.js           # Configuración PostCSS
```

## 🚀 Inicio Rápido

### Instalación

```bash
# Navega a la carpeta docs
cd docs

# Instala dependencias
npm install
# o
yarn install
```

### Desarrollo

```bash
# Inicia servidor de desarrollo
npm run dev
# o
yarn dev
```

Se abrirá automáticamente en `http://localhost:3000`

### Build para Producción

```bash
# Construye para producción
npm run build
# o
yarn build

# Previsualizar build
npm run preview
# o
yarn preview
```

## 📁 Organización de Componentes

### Componentes Comunes (`src/components/common/`)

Componentes reutilizables en toda la aplicación:

- **Button.jsx** - Botón versátil con variantes (primary, secondary)
- **CodeBlock.jsx** - Bloque de código con highlight.js
- **FeatureCard.jsx** - Tarjeta de características
- **Section.jsx** - Sección de contenido

### Layout (`src/components/layout/`)

Estructura base de la aplicación:

- **Navbar.jsx** - Navegación principal con links activos
- **Footer.jsx** - Pie de página con enlaces
- **Container.jsx** - Contenedor con max-width y padding

### Páginas (`src/pages/`)

Cada página es un componente que representa una sección:

- **Home.jsx** - Página de inicio con hero section
- **Guide.jsx** - Guía de inicio rápido
- **Docs.jsx** - Documentación completa del lenguaje
- **Stdlib.jsx** - Referencia de Standard Library
- **Examples.jsx** - Ejemplos prácticos

## 🎨 Estilos y Personalización

### Tailwind CSS

Los estilos se aplican directamente en los componentes usando clases de Tailwind:

```jsx
<div className="max-w-7xl mx-auto px-4 py-20">
  <h1 className="text-4xl font-bold gradient-text">Título</h1>
</div>
```

### Clases Personalizadas

En `src/styles/globals.css` encontrarás clases reutilizables:

```css
.gradient-text       /* Gradiente de purple a pink */
.glass-effect       /* Efecto vidrio translúcido */
.nav-link          /* Link de navegación con subrayado */
.feature-card      /* Tarjeta de feature con hover */
.code-block        /* Bloque de código estilizado */
.fade-in           /* Animación de entrada */
```

## 🔧 Desarrollo

### Agregar una Nueva Página
� Buenas Prácticas Implementadas

✅ **Componentes Funcionales** - Todos los componentes usan hooks de React  
✅ **Props Drilling Mínimo** - Estructura plana de componentes  
✅ **Reutilización** - Componentes comunes y utilities  
✅ **Separación de Concerns** - Lógica separada del UI  
✅ **CSS-in-JS con Tailwind** - Sin archivos CSS separados  
✅ **Responsive Design** - Mobile-first approach  
✅ **Accesibilidad** - Semántica HTML correcta  
✅ **Performance** - SPA sin recargas innecesarias  

## 🐛 Troubleshooting

### Puerto 3000 ya está en uso

```bash
# Cambiar puerto en vite.config.js
server: {
  port: 3001,
  open: true
}
```

### Estilos Tailwind no se aplican

```bash
# Limpia cache de Vite y reinstala
rm -rf node_modules dist
npm install
npm run dev
```

### Highlight.js no funciona

Verifica que los códigos tengan el atributo `language-xxxx` correcto:

```jsx
<CodeBlock code={myCode} language="javascript" />
```

## 📚 Recursos

- [React Docs](https://react.dev)
- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [Vite Docs](https://vitejs.dev)
- [Highlight.js](https://highlightjs.org/)
- [Jabline Repository](https://github.com/Jabline-lang/Jabline)

## 🤝 Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/amazing`)
3. Commit tus cambios (`git commit -am 'Add amazing feature'`)
4. Push a la rama (`git push origin feature/amazing`)
5. Abre un Pull Request

## 📄 Licencia

MIT - Ver [LICENSE](../LICENSE)

---

**Última actualización**: Mayo 8, 2026  
**Versión**: 0.6.0  
**Built with React + Tailwind CSS + Vite**
  )
}
```

2. Actualiza `src/App.jsx`:

```jsx
import MyPage from './pages/MyPage'

// En el switch:
case 'mypage':
  return <MyPage />
```

3. Agrega el botón en `src/components/layout/Navbar.jsx`:

```jsx
const navItems = [
  // ...
  { id: 'mypage', label: 'Mi Página' },
]
```

### Agregar un Componente Reutilizable

1. Crea `src/components/common/MyComponent.jsx`:

```jsx
export default function MyComponent({ prop1, prop2 }) {
  return (
    <div className="glass-effect p-6 rounded-xl">
      {/* Contenido */}
    </div>
  )
}
```

2. Importa en tus páginas:

```jsx
import MyComponent from '../components/common/MyComponent'
```

## 📦 Dependencias

- **react** `^18.2.0` - Librería UI
- **react-dom** `^18.2.0` - Renderizado DOM
- **tailwindcss** `^3.3.0` - Framework CSS
- **vite** `^5.0.0` - Build tool
- **highlight.js** `^11.9.0` - Resaltado de código

## 🌐 Despliegue

### GitHub Pages

```bash
# Build
npm run build

# Los archivos en dist/ se despliegan automáticamente
# Configura en Settings > Pages > dist directory
```

### Netlify

```bash
# Conecta tu repositorio a Netlify
# Configura:
# - Build command: npm run build
# - Publish directory: dist
```

### Vercel

```bash
npm install -g vercel
vercel
```

## 📚 Recursos

- [Tailwind CSS Docs](https://tailwindcss.com/docs)
- [Highlight.js](https://highlightjs.org/)
- [Jabline GitHub](https://github.com/Jabline-lang/Jabline)

## 📝 Licencia

Esta documentación es parte del proyecto Jabline y está bajo la misma licencia (MIT).

---

**Última actualización**: Mayo 8, 2026
**Versión**: 0.6.0
