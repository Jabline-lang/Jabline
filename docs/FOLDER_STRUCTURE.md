# 📋 Estructura de Carpetas - Documentación de Jabline

```
docs/
│
├── 📄 Archivos de Configuración
├── index.html                    # HTML principal (punto de entrada)
├── package.json                  # Dependencias del proyecto
├── vite.config.js               # Configuración de Vite
├── tailwind.config.js           # Configuración de Tailwind CSS
├── postcss.config.js            # Configuración de PostCSS
├── .eslintrc.json               # Configuración de ESLint
├── .gitignore                   # Archivos ignorados por Git
├── .env.example                 # Ejemplo de variables de entorno
│
├── 📁 src/
│   │
│   ├── 📄 main.jsx              # Punto de entrada de React
│   ├── 📄 App.jsx               # Componente raíz
│   │
│   ├── 📁 components/           # Componentes reutilizables
│   │   ├── 📁 common/
│   │   │   ├── Button.jsx       # Componente de botón versátil
│   │   │   ├── CodeBlock.jsx    # Bloque de código con highlighting
│   │   │   ├── FeatureCard.jsx  # Tarjeta de características
│   │   │   └── Section.jsx      # Wrapper de sección
│   │   │
│   │   └── 📁 layout/
│   │       ├── Container.jsx    # Contenedor responsivo
│   │       ├── Navbar.jsx       # Barra de navegación
│   │       └── Footer.jsx       # Pie de página
│   │
│   ├── 📁 pages/                # Páginas principales
│   │   ├── Home.jsx             # Página de inicio
│   │   ├── Guide.jsx            # Guía de inicio rápido
│   │   ├── Docs.jsx             # Documentación completa
│   │   ├── Stdlib.jsx           # Referencia de stdlib
│   │   └── Examples.jsx         # Ejemplos prácticos
│   │
│   ├── 📁 styles/               # Estilos globales
│   │   └── globals.css          # CSS personalizado con Tailwind
│   │
│   ├── 📁 constants/            # Constantes de la aplicación
│   │   └── config.js            # Configuración de constantes
│   │
│   ├── 📁 hooks/                # Custom React hooks
│   │   └── useCustom.js         # Hooks personalizados
│   │
│   └── 📁 utils/                # Funciones utilitarias
│       └── helpers.js           # Funciones auxiliares
│
├── 📁 dist/                     # Build generado (ignorado por Git)
├── 📁 node_modules/             # Dependencias (ignorado por Git)
│
├── 📚 Documentación
├── README.md                    # Documentación principal del proyecto
├── DEVELOPMENT.md              # Guía de desarrollo
└── FOLDER_STRUCTURE.md         # Este archivo
```

## 📋 Descripción de Directorios Principales

### `/src/components/`
Componentes React reutilizables organizados por categoría:
- **common/**: Componentes genéricos reutilizables en toda la app
- **layout/**: Componentes de estructura de página (header, footer, etc.)

### `/src/pages/`
Páginas principales del sitio, cada una representa una sección:
- Cada página es un componente React independiente
- Se renderizan dinámicamente según la ruta/sección seleccionada

### `/src/styles/`
Estilos globales usando Tailwind CSS:
- `globals.css` contiene clases personalizadas que extienden Tailwind

### `/src/constants/`
Valores constantes reutilizables:
- Configuración de la app
- Arrays de datos estáticos

### `/src/hooks/`
Custom hooks de React para lógica reutilizable:
- Manejo de highlighting de código
- Scroll automático
- Eventos de teclado

### `/src/utils/`
Funciones utilitarias puras:
- Helpers de formato
- Funciones de manipulación de strings

## 🔄 Flujo de la Aplicación

```
index.html (punto de entrada)
    ↓
src/main.jsx (renderiza React)
    ↓
src/App.jsx (componente raíz con enrutamiento)
    ↓
Navbar.jsx ← selecciona página ← Footer.jsx
    ↓
Renderiza página (Home.jsx, Guide.jsx, etc.)
    ↓
Componentes comunes (Button, CodeBlock, Section, etc.)
```

## 📦 Cómo Agregar Contenido

### Agregar una Nueva Página

1. Crea `src/pages/NewPage.jsx`
2. Importa en `src/App.jsx`
3. Agrega caso en el switch de `App.jsx`
4. Agrega item de navegación en `src/components/layout/Navbar.jsx`

### Agregar un Componente

1. Crea archivo en `src/components/common/` o `src/components/layout/`
2. Exporta como default
3. Importa donde sea necesario

### Agregar Constantes

1. Edita `src/constants/config.js`
2. Importa donde lo necesites: `import { CONSTANT } from '../constants/config'`

### Agregar Utilidades

1. Agrega función a `src/utils/helpers.js`
2. Importa donde lo necesites: `import { helper } from '../utils/helpers'`

## 🎨 Nomenclatura

- **Componentes**: PascalCase (`Button.jsx`, `CodeBlock.jsx`)
- **Funciones**: camelCase (`useHighlight()`, `formatCode()`)
- **Constantes**: UPPER_SNAKE_CASE (`NAVIGATION_ITEMS`, `APP_VERSION`)
- **Variables**: camelCase (`isActive`, `userData`)

## ✅ Checklist para Nuevos Archivos

Cuando crees un nuevo archivo, verifica:
- [ ] Usa nomenclatura correcta
- [ ] Incluye comentarios/documentación
- [ ] Proporciona props typing en comentarios
- [ ] Está en la carpeta correcta
- [ ] Se importa donde es necesario
- [ ] Sin dependencias innecesarias

---

**Última actualización**: Mayo 8, 2026
**Versión**: 0.6.0
