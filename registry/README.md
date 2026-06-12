# Jabline Package Registry 📦

This repository hosts the **official package registry** for [Jabline](https://github.com/Jabline-lang/Jabline) — a cloud-native, concurrent programming language.

## Usage

```bash
# Install a package
jabline get package_name

# Or use a full git URL
jabline get https://github.com/user/repo.git
```

## Available Packages

*The registry is empty — no packages have been published yet.*

Submit your package via `jabline publish` (requires `GITHUB_TOKEN`) or open a Pull Request adding your entry to `index.json`.

## Registering a Package

To add your package to the registry:

1. **Fork** this repository
2. **Add your package** to `index.json`:
   ```json
   {
     "your-package": {
       "name": "your-package",
       "description": "What your package does",
       "url": "https://github.com/your-org/your-package.git",
       "version": "0.1.0",
       "author": "Your Name",
       "category": "utility",
       "tags": ["http", "routing"],
       "license": "MIT",
       "repository": "https://github.com/your-org/your-package"
     }
   }
   ```
3. Open a **Pull Request**

### Requirements

- Your package must have a `jabline.toml` at root
- Your package must be open source (MIT, Apache 2.0, or similar)
- Your package should have at least basic test coverage

## Schema

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | ✅ | Package name (kebab-case) |
| `description` | string | ✅ | Short description (max 100 chars) |
| `url` | string | ✅ | Git clone URL |
| `version` | string | ✅ | SemVer version |
| `author` | string | ✅ | Author name or org |
| `category` | string | ❌ | Category: web, data, crypto, util, dev |
| `tags` | array | ❌ | Search tags |
| `license` | string | ❌ | SPDX license identifier |
| `repository` | string | ❌ | GitHub repo URL for docs |

## API

The registry is available as raw JSON:

```
https://raw.githubusercontent.com/Jabline-lang/registry/main/index.json
```

Jabline CLI caches this for 1 hour and falls back to the bundled registry when offline.

## License

MIT
