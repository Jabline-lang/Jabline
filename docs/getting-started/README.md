# Docs - Getting Started

## Instalación

```bash
# Desde releases de GitHub
# Descargar el binario para tu plataforma

# O compilar desde fuente
git clone https://github.com/Jabline-lang/jabline
cd jabline
go build -o jabline main.go
```

## Primer programa

Crea `hello.jb`:

```jabline
echo("¡Hola, mundo!")
```

Ejecuta:

```bash
jabline run hello.jb
```

## CLI básico

```
jabline run <archivo>      # Ejecutar script
jabline repl               # REPL interactivo
jabline fmt <archivo>      # Formatear código
jabline test <archivo>    # Ejecutar tests
jabline build <archivo>    # Compilar a bytecode
```

Ver ejemplos en la carpeta `getting-started/`.
