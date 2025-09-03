# 📦 Ejemplos del Sistema de Módulos

Este directorio contiene ejemplos que demuestran el poderoso sistema de módulos de Jabline, incluyendo importación/exportación, módulos especializados y arquitecturas modulares escalables.

## 📋 Lista de Ejemplos

### 01. Módulos Básicos
**Archivo:** `01_basic_modules.jb`
- Importación completa de módulos
- Uso de funciones exportadas
- Integración de múltiples módulos (math, strings, arrays)
- Casos de uso fundamentales
- Validación de datos con módulos

### 02. Módulo Simple
**Archivo:** `02_simple_module.jb`
- Creación y uso de un módulo básico
- Exportación de constantes y funciones
- Importación completa
- Ejemplo minimalista pero completo

### 03. Ultra Simple
**Archivo:** `03_ultra_simple.jb`
- Demostración del módulo strings_minimal
- Funciones básicas de procesamiento de texto
- Validación de emails
- Formateo de nombres

### 04. Módulos Avanzados
**Archivo:** `04_advanced_modules.jb`
- Integración compleja de múltiples módulos
- Procesamiento de datos de usuarios
- Análisis estadístico con arrays
- Validación avanzada con strings
- Operaciones matemáticas complejas
- Sistema completo de análisis

### 05. Importaciones Selectivas
**Archivo:** `05_selective_imports.jb`
- Sintaxis `import { func1, func2 } from "module"`
- Importación solo de elementos necesarios
- Optimización del espacio de nombres
- Combinación de importaciones selectivas
- Mejores prácticas de importación

## 🏗️ Sistema de Módulos de Jabline

### Características Implementadas
- ✅ **Exportación**: `export let`, `export fn`
- ✅ **Importación Completa**: `import "modulo"`
- ✅ **Importación Selectiva**: `import { item1, item2 } from "modulo"`
- ✅ **Cache Automático**: Módulos se cargan una sola vez
- ✅ **Built-ins Integrados**: Funciones nativas disponibles en módulos
- ✅ **Manejo de Errores**: Validación y reporte de errores robusto

### Sintaxis de Módulos

#### Exportación (en el módulo)
```jabline
// Exportar constantes
export let PI = 3.14159;
export let VERSION = "1.0";

// Exportar funciones
export fn add(a, b) {
    return a + b;
}

export fn greet(name) {
    return "Hola " + name;
}
```

#### Importación Completa
```jabline
// Importar todo el módulo
import "math"

// Usar todas las funciones exportadas
echo(PI);           // 3.14159
echo(add(5, 3));    // 8
```

#### Importación Selectiva
```jabline
// Importar solo funciones específicas
import { add, multiply, PI } from "math"

// Solo estas funciones están disponibles
echo(add(10, 20));    // 30
echo(PI);            // 3.14159
// greet() NO está disponible
```

## 📚 Módulos Disponibles

### Módulo Math (`../../modules/math.jb`)
**Constantes:**
- `PI = 3`
- `E = 2`

**Funciones (12+):**
- `abs(x)` - Valor absoluto
- `max(a, b)` - Máximo de dos números  
- `min(a, b)` - Mínimo de dos números
- `pow(base, exp)` - Potencia
- `sqrt(x)` - Raíz cuadrada aproximada
- `factorial(n)` - Factorial
- `isEven(n)`, `isOdd(n)` - Verificar par/impar
- `random(min, max)` - Número aleatorio simple

### Módulo Strings Minimal (`../../modules/strings_minimal.jb`)
**Constantes:**
- `MESSAGE` - Mensaje del módulo

**Funciones (5+):**
- `capitalize(str)` - Capitaliza primera letra
- `formatName(first, last)` - Formatea nombre completo
- `isValidEmail(email)` - Valida formato de email
- `extractDomain(email)` - Extrae dominio del email
- `cleanSpaces(str)` - Limpia espacios extras

### Módulo Arrays (`../../modules/arrays.jb`)
**Funciones (20+):**
- `findIndex(arr, value)` - Encuentra índice
- `contains(arr, value)` - Verifica si contiene
- `remove(arr, value)` - Remueve elemento
- `sort(arr)` - Ordena elementos
- `reverseArray(arr)` - Invierte array
- `max(arr)`, `min(arr)` - Valor máximo/mínimo
- `sum(arr)` - Suma elementos
- `average(arr)` - Promedio
- `unique(arr)` - Elementos únicos
- Y muchas más...

## 🚀 Cómo Ejecutar

Para ejecutar cualquier ejemplo de módulos:

```bash
./jabline run examples/modules/[nombre_archivo].jb
```

Ejemplos específicos:
```bash
# Módulos básicos
./jabline run examples/modules/01_basic_modules.jb

# Importaciones selectivas
./jabline run examples/modules/05_selective_imports.jb

# Demo avanzado
./jabline run examples/modules/04_advanced_modules.jb
```

## 📚 Prerrequisitos

Antes de explorar el sistema de módulos:

- ✅ Dominio de conceptos básicos (variables, funciones, arrays)
- ✅ Comprensión de estructuras de datos
- ✅ Familiaridad con funciones built-in
- ✅ Experiencia con ejemplos básicos y avanzados

## 🎓 Orden Recomendado

1. **Módulo Simple** - Entender conceptos básicos
2. **Ultra Simple** - Ver módulos especializados
3. **Módulos Básicos** - Integración de múltiples módulos
4. **Importaciones Selectivas** - Optimización de imports
5. **Módulos Avanzados** - Casos de uso complejos

## 💡 Casos de Uso Reales

### Sistema de Validación
```jabline
import { isValidEmail, capitalize } from "strings_minimal"
import { contains } from "arrays"

fn validateUserData(userData) {
    let errors = [];
    
    if (isValidEmail(userData.email) == false) {
        errors = push(errors, "Email inválido");
    }
    
    return errors;
}
```

### Análisis de Datos
```jabline
import { sum, average, max, min } from "arrays"
import { abs } from "math"

fn analyzeMetrics(data) {
    return {
        "total": sum(data),
        "average": average(data),
        "range": max(data) - min(data)
    };
}
```

### Procesamiento de Texto
```jabline
import { capitalize, cleanSpaces } from "strings_minimal"

fn processUserInput(input) {
    let cleaned = cleanSpaces(input);
    return capitalize(cleaned);
}
```

## 🔧 Arquitectura del Sistema

### Flujo de Importación
```
Archivo .jb → Lexer → Parser → AST → ModuleSystem → Cache → Evaluator
```

### Componentes Clave
- **ModuleSystem**: Gestión y cache de módulos
- **AST Nodes**: ImportStatement, ExportStatement
- **Parser**: Parsing de sintaxis import/export
- **Evaluator**: Evaluación e integración

### Rutas de Búsqueda
1. `./` - Directorio actual
2. `./modules/` - Directorio de módulos
3. `./lib/` - Directorio de librerías

## 🌟 Ventajas del Sistema de Módulos

### Organización
- **Separación de responsabilidades**: Cada módulo tiene un propósito específico
- **Código reutilizable**: Funciones compartibles entre proyectos
- **Estructura clara**: Organización lógica del código

### Escalabilidad
- **Proyectos grandes**: Manejo de aplicaciones complejas
- **Colaboración**: Equipos trabajando en módulos independientes
- **Mantenimiento**: Actualizaciones centralizadas

### Eficiencia
- **Cache automático**: Módulos se cargan una sola vez
- **Importación selectiva**: Solo cargar lo necesario
- **Optimización**: Mejor rendimiento de carga

## 💼 Mejores Prácticas

### ✅ Hacer
- Usar importación selectiva cuando sea posible
- Organizar módulos por funcionalidad
- Documentar funciones exportadas
- Nombrar módulos descriptivamente
- Validar datos de entrada en módulos

### ❌ Evitar
- Importar módulos completos si solo necesitas pocas funciones
- Crear dependencias circulares
- Módulos demasiado grandes
- Conflictos de nombres entre módulos
- Funciones con efectos secundarios no documentados

## 🛠️ Solución de Problemas

### Error: "failed to load module"
- ✅ Verificar que el archivo existe
- ✅ Comprobar la ruta del módulo
- ✅ Asegurar extensión `.jb`

### Error: "X is not exported by module Y"
- ✅ Verificar que esté marcado con `export`
- ✅ Revisar el nombre exacto (case-sensitive)
- ✅ Confirmar que el módulo se cargó correctamente

## ➡️ Siguientes Pasos

Una vez que domines el sistema de módulos:

- **Crear tus propios módulos** especializados
- **Desarrollar librerías reutilizables** para tus proyectos
- **Aplicar arquitecturas modulares** en aplicaciones complejas
- **Contribuir módulos** a la comunidad Jabline

## 🏆 Objetivos de Aprendizaje

Al completar estos ejemplos, serás capaz de:

- ✅ Crear y exportar módulos personalizados
- ✅ Importar módulos completa y selectivamente  
- ✅ Integrar múltiples módulos en aplicaciones complejas
- ✅ Aplicar mejores prácticas de organización modular
- ✅ Desarrollar arquitecturas escalables y mantenibles
- ✅ Crear librerías reutilizables para la comunidad

## 🚀 Impacto en el Desarrollo

El sistema de módulos transforma Jabline en un lenguaje apto para:

- **Desarrollo profesional**: Aplicaciones de nivel empresarial
- **Colaboración**: Trabajo en equipo eficiente
- **Escalabilidad**: Proyectos que crecen en complejidad
- **Reutilización**: Bibliotecas compartibles
- **Mantenimiento**: Código fácil de actualizar y extender

¡El sistema de módulos convierte a Jabline en una plataforma de desarrollo moderna y completa!