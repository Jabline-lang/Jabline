# Contributing to the Jabline Package Registry

Thank you for your interest in contributing a package! This registry is how Jabline users discover and install packages.

## Adding a Package

1. **Add your package** to `index.json` in alphabetical order
2. **Open a Pull Request** with your change
3. The CI will validate:
   - JSON is valid
   - Required fields are present
   - URL is a valid git URL
   - Version is valid SemVer
   - No duplicate entries

## Package Requirements

- ✅ Must have a `jabline.toml` with `name`, `version`, `description`
- ✅ Must have a public git repository
- ✅ Must use a permissive open source license (MIT, Apache 2.0, BSD, MPL, or Unlicense)
- ✅ Should have at least basic tests
- ✅ Should follow Jabline naming conventions (kebab-case)

## Quality Guidelines

- ⭐ Well-documented with examples
- ⭐ Tested with `jabline test`
- ⭐ Follows Jabline idioms and style
- ⭐ Published with a SemVer version

## PR Review Process

1. Automated checks run (JSON validation, schema check)
2. Maintainers review the package quality
3. If approved, the PR is merged and the package is live within minutes
