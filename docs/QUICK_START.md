# ⚡ Guía Rápida - Documentación Jabline

## 🚀 Inicio en 30 segundos

### 1️⃣ Instala Dependencias
```bash
cd docs
npm install
```

### 2️⃣ Inicia el Servidor
```bash
npm run dev
```

### 3️⃣ Abre en el Navegador
```
http://localhost:3000
```

¡Eso es todo! 🎉

---

## 📖 Secciones Disponibles

| Sección | Descripción |
|---------|-------------|
| **Inicio** | Presentación y características principales de Jabline |
| **Guía** | Instalación, variables, funciones, control de flujo |
| **Documentación** | Structs, Genéricos, Concurrencia, Módulos, CLI |
| **Stdlib** | Referencia completa de módulos estándar |
| **Ejemplos** | Código práctico: REST API, Concurrencia, DB, etc. |

---

## 💻 Comandos Principales

```bash
# Desarrollo
npm run dev              # Servidor con hot-reload en :3000

# Producción
npm run build            # Construye optimizado en dist/
npm run preview          # Vista previa del build

# Linting
npm run lint             # Verifica código con ESLint
```

---

## 📁 Agregar Nueva Página (5 min)

### Paso 1: Crea el archivo
**`src/pages/FAQ.jsx`**
```jsx
import Container from '../components/layout/Container'
import Section from '../components/common/Section'

export default function FAQ() {
  return (
    <Container>
      <Section title="Preguntas Frecuentes">
        <p className="text-gray-400">Tu contenido aquí</p>
      </Section>
    </Container>
  )
}
```

### Paso 2: Importa en App.jsx
```jsx
import FAQ from './pages/FAQ'
```

### Paso 3: Agrega el caso
```jsx
case 'faq':
  return <FAQ />
```

### Paso 4: Agrega al menú
En `Navbar.jsx`:
```jsx
{ id: 'faq', label: 'FAQ' }
```

✅ ¡Listo!

---

## 🎨 Componentes Comunes

### Button
```jsx
<Button>Click Me</Button>
<Button variant="secondary">Secundario</Button>
<Button href="https://..." external>Link</Button>
```

### CodeBlock
```jsx
<CodeBlock code={myCode} language="javascript" />
```

### FeatureCard
```jsx
<FeatureCard 
  icon="⚡" 
  title="Título" 
  description="Descripción"
/>
```

### Section
```jsx
<Section title="Sección" description="Descripción opcional">
  {/* contenido */}
</Section>
```

### Container
```jsx
<Container>
  {/* contenido centrado */}
</Container>
```

---

## 📦 Estructura de Archivos

```
docs/
├── src/
│   ├── components/
│   │   ├── common/        ← Componentes reutilizables
│   │   └── layout/        ← Layout (Navbar, Footer)
│   ├── pages/             ← Páginas (Home, Guide, Docs, etc.)
│   ├── styles/            ← Estilos CSS
│   ├── constants/         ← Valores constantes
│   ├── hooks/             ← Custom hooks
│   ├── utils/             ← Funciones helper
│   ├── App.jsx            ← Componente raíz
│   └── main.jsx           ← Punto de entrada
├── package.json
├── vite.config.js
├── tailwind.config.js
└── index.html
```

---

## 🎨 Personalización de Colores

En `tailwind.config.js`, modifica los colores:

```js
theme: {
  colors: {
    purple: {
      500: '#a855f7',
      600: '#9333ea',
      // ...
    }
  }
}
```

---

## 🚀 Deploy

### GitHub Pages
```bash
npm run build
# Sube dist/ a gh-pages
```

### Netlify
1. Conecta repo a Netlify
2. Build command: `npm run build`
3. Publish directory: `dist`

### Vercel
```bash
npm install -g vercel
vercel
```

---

## 🆘 Troubleshooting

### Puerto 3000 en uso
```bash
# Cambiar puerto en vite.config.js
server: { port: 3001 }
```

### Tailwind no funciona
```bash
rm -rf node_modules dist
npm install
npm run dev
```

### Highlighting de código no funciona
```jsx
<CodeBlock code={code} language="javascript" />
```

---

## 📚 Docs Útiles

- [React Docs](https://react.dev)
- [Tailwind CSS](https://tailwindcss.com)
- [Vite](https://vitejs.dev)
- [Highlight.js](https://highlightjs.org/)

---

**¿Necesitas ayuda?** Ver [README.md](./README.md) o [DEVELOPMENT.md](./DEVELOPMENT.md)
