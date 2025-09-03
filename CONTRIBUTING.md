# 🤝 Contribuyendo a Jabline Programming Language

¡Gracias por tu interés en contribuir a Jabline! Este documento proporciona pautas y información sobre cómo contribuir efectivamente al proyecto.

## 📋 Tabla de Contenidos

- [Código de Conducta](#código-de-conducta)
- [Cómo Contribuir](#cómo-contribuir)
- [Configuración del Entorno](#configuración-del-entorno)
- [Estructura del Proyecto](#estructura-del-proyecto)
- [Estándares de Codificación](#estándares-de-codificación)
- [Proceso de Pull Request](#proceso-de-pull-request)
- [Reportar Bugs](#reportar-bugs)
- [Sugerir Nuevas Características](#sugerir-nuevas-características)
- [Documentación](#documentación)
- [Pruebas](#pruebas)

## 📜 Código de Conducta

Este proyecto se adhiere a un código de conducta. Al participar, se espera que mantengas este código. Por favor reporta comportamientos inaceptables.

## 🚀 Cómo Contribuir

Hay muchas formas de contribuir a Jabline:

### 🐛 Reportando Bugs
- Usa el sistema de issues de GitHub
- Incluye información detallada del error
- Proporciona pasos para reproducir el problema
- Incluye información del sistema operativo y versión de Go

### ✨ Sugiriendo Mejoras
- Abre un issue describiendo la mejora
- Explica por qué sería útil
- Proporciona ejemplos de uso si es posible

### 🔧 Contribuciones de Código
- Correcciones de bugs
- Nuevas características
- Optimizaciones de rendimiento
- Mejoras en la documentación

### 📚 Documentación
- Mejorar README y guías
- Crear nuevos ejemplos
- Corregir errores tipográficos
- Traducir contenido

## 🛠️ Configuración del Entorno

### Prerrequisitos
- **Go 1.19+**: [Instalar Go](https://golang.org/doc/install)
- **Git**: Para control de versiones
- **Editor**: VS Code, GoLand, Vim, etc.

### Configuración Local

1. **Fork el repositorio**
   ```bash
   # En GitHub, haz click en "Fork"
   ```

2. **Clonar tu fork**
   ```bash
   git clone https://github.com/TU-USUARIO/jabline.git
   cd jabline
   ```

3. **Configurar upstream**
   ```bash
   git remote add upstream https://github.com/REPO-ORIGINAL/jabline.git
   ```

4. **Instalar dependencias**
   ```bash
   go mod download
   ```

5. **Compilar el proyecto**
   ```bash
   go build -o jabline main.go
   ```

6. **Ejecutar pruebas**
   ```bash
   # Pruebas básicas
   ./jabline run examples/basic/01_variables_operadores.jb
   
   # Sistema de módulos
   ./jabline run examples/modules/01_basic_modules.jb
   ```

## 📁 Estructura del Proyecto

```
jabline/
├── cmd/                    # Comandos CLI (Cobra)
├── pkg/                    # Paquetes principales
│   ├── lexer/             # Análisis léxico
│   ├── parser/            # Análisis sintáctico
│   ├── ast/               # Árbol de sintaxis abstracta
│   ├── evaluator/         # Evaluador e intérprete
│   ├── object/            # Sistema de objetos
│   └── token/             # Definiciones de tokens
├── modules/               # Módulos estándar (.jb)
│   ├── math.jb
│   ├── strings_minimal.jb
│   └── arrays.jb
├── examples/              # Ejemplos organizados
│   ├── basic/            # Ejemplos básicos
│   ├── advanced/         # Ejemplos avanzados
│   ├── modern/           # Características modernas
│   └── modules/          # Sistema de módulos
├── docs/                  # Documentación (si existe)
├── main.go               # Punto de entrada
├── go.mod                # Dependencias Go
└── README.md             # Documentación principal
```

## 📝 Estándares de Codificación

### Go Code Style
- Seguir las convenciones estándar de Go (`gofmt`, `golint`)
- Usar nombres descriptivos para variables y funciones
- Incluir comentarios para funciones públicas
- Mantener funciones pequeñas y enfocadas

### Ejemplo de Función Bien Documentada
```go
// ParseImportStatement parses an import statement and returns an ImportStatement AST node.
// It supports both complete imports ("import module") and selective imports 
// ("import { item1, item2 } from module").
//
// Returns nil if the import statement is malformed.
func (p *Parser) ParseImportStatement() *ast.ImportStatement {
    // Implementation...
}
```

### Convenciones de Nombres
- **Archivos**: `snake_case.go` (ej: `import_statement.go`)
- **Funciones públicas**: `PascalCase` (ej: `ParseExpression`)
- **Funciones privadas**: `camelCase` (ej: `parseIdentifier`)
- **Constantes**: `UPPER_CASE` (ej: `MAX_DEPTH`)
- **Variables**: `camelCase` (ej: `tokenType`)

### Jabline Code Style
- Usar espacios consistentes para indentación
- Incluir comentarios explicativos en ejemplos
- Seguir convenciones de nombres del lenguaje
- Documentar funciones exportadas en módulos

## 🔄 Proceso de Pull Request

### 1. Preparación
```bash
# Actualizar tu fork
git checkout main
git pull upstream main
git push origin main

# Crear rama para tu feature
git checkout -b feature/nueva-caracteristica
```

### 2. Desarrollo
- Implementar cambios siguiendo los estándares
- Escribir o actualizar pruebas
- Actualizar documentación si es necesario
- Probar localmente

### 3. Commit
```bash
# Commits descriptivos
git add .
git commit -m "feat: agregar soporte para operador ternario

- Implementar parsing del operador ?:
- Agregar evaluación en el intérprete
- Incluir pruebas y ejemplos
- Actualizar documentación"
```

### 4. Push y PR
```bash
git push origin feature/nueva-caracteristica
```

Luego crear el PR en GitHub con:
- **Título descriptivo**
- **Descripción detallada** de los cambios
- **Referencias a issues** relacionados
- **Pruebas realizadas**

### 5. Revisión
- Responder a comentarios constructivamente
- Realizar cambios solicitados
- Mantener la rama actualizada con main

## 🐛 Reportar Bugs

### Template de Bug Report

```markdown
**Descripción del Bug**
Descripción clara y concisa del problema.

**Pasos para Reproducir**
1. Ejecutar comando '...'
2. Usar código '....'
3. Observar error

**Comportamiento Esperado**
Qué debería haber pasado.

**Comportamiento Actual**
Qué pasó en realidad.

**Información del Sistema**
- OS: [ej: Ubuntu 20.04]
- Go Version: [ej: 1.21.0]
- Jabline Version: [ej: commit hash]

**Código de Ejemplo**
```jabline
// Código que causa el problema
let variable = "valor";
```

**Información Adicional**
Cualquier otro contexto relevante.
```

## ✨ Sugerir Nuevas Características

### Template de Feature Request

```markdown
**Descripción de la Característica**
Descripción clara de qué quieres que se agregue.

**Problema que Resuelve**
¿Qué problema actual resuelve esta característica?

**Solución Propuesta**
Descripción detallada de cómo funcionaría.

**Alternativas Consideradas**
Otras soluciones que has considerado.

**Ejemplos de Uso**
```jabline
// Cómo se usaría la nueva característica
nueva_caracteristica(parametros);
```

**Información Adicional**
Contexto adicional, screenshots, etc.
```

## 📖 Documentación

### Tipos de Documentación
- **README**: Información general y getting started
- **Ejemplos**: Código funcional con comentarios
- **Comentarios**: Documentación inline en el código
- **Guías**: Tutoriales y best practices

### Estándares de Documentación
- Usar Markdown para archivos de documentación
- Incluir ejemplos de código funcionales
- Proporcionar explicaciones paso a paso
- Mantener documentación actualizada con el código

## 🧪 Pruebas

### Tipos de Pruebas
1. **Pruebas de Ejemplos**: Ejecutar todos los ejemplos
2. **Pruebas de Integración**: Funcionalidades completas
3. **Pruebas de Regresión**: Evitar romper código existente

### Ejecutar Pruebas

```bash
# Pruebas básicas
for file in examples/basic/*.jb; do 
    echo "Testing $file"
    ./jabline run "$file" || echo "FAILED: $file"
done

# Pruebas de módulos
for file in examples/modules/*.jb; do 
    echo "Testing $file"
    ./jabline run "$file" || echo "FAILED: $file"
done

# Pruebas modernas
for file in examples/modern/*.jb; do 
    echo "Testing $file"
    ./jabline run "$file" || echo "FAILED: $file"
done
```

### Agregar Nuevas Pruebas
- Crear ejemplos que demuestren la funcionalidad
- Incluir casos de error cuando sea apropiado
- Documentar el comportamiento esperado
- Probar en diferentes escenarios

## 🎯 Áreas de Contribución Priorizadas

### Alto Impacto
- **Correcciones de bugs críticos**
- **Mejoras de rendimiento**
- **Nuevos built-ins útiles**
- **Documentación mejorada**

### Características Deseadas
- **Async/await support**
- **Package manager**
- **More built-in functions**
- **REPL interactive**
- **Better error messages**

### Contribuciones Bienvenidas
- **Más ejemplos de uso real**
- **Módulos especializados**
- **Optimizaciones del intérprete**
- **Herramientas de desarrollo**

## 🏷️ Convenciones de Commits

Usamos [Conventional Commits](https://www.conventionalcommits.org/):

```
tipo(scope): descripción breve

Descripción más detallada si es necesaria

- Cambio específico 1
- Cambio específico 2

Fixes #123
```

### Tipos de Commit
- `feat`: Nueva característica
- `fix`: Corrección de bug
- `docs`: Solo documentación
- `style`: Cambios de formato
- `refactor`: Refactoring sin cambios funcionales
- `test`: Agregar o corregir pruebas
- `chore`: Mantenimiento general

### Ejemplos
```bash
feat(parser): agregar soporte para operador ternario
fix(evaluator): corregir evaluación de optional chaining
docs(examples): agregar ejemplo de async/await
refactor(lexer): optimizar tokenización de strings
```

## 🤔 ¿Necesitas Ayuda?

- **Documentación**: Consulta README y ejemplos
- **Issues**: Busca issues existentes similares
- **Discusiones**: Usa GitHub Discussions para preguntas generales
- **Código**: Revisa implementaciones existentes como referencia

## 📞 Contacto

- **GitHub Issues**: Para bugs y feature requests
- **GitHub Discussions**: Para preguntas y discusiones
- **Email**: [si aplica]

## 🙏 Reconocimientos

Todas las contribuciones son valoradas y reconocidas. Los contribuidores aparecen en:
- README del proyecto
- Release notes
- Hall of fame (si existe)

¡Gracias por hacer que Jabline sea mejor para todos! 🚀

---

**Jabline Programming Language** - Un lenguaje moderno, expresivo y colaborativo.