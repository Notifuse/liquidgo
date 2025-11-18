# Liquid Go Implementation

A Go implementation of Shopify's Liquid template engine with full feature parity.

## Project Structure

```
liquidgo/
├── .cursorrules              # AI agent instructions
├── .cursor/
│   └── instructions.md       # Detailed implementation guide
├── IMPLEMENTATION.md         # This file
├── liquid/                   # Main Go package (maps to lib/liquid/)
│   ├── tags/                # Tag implementations (maps to lib/liquid/tags/)
│   ├── tag/                 # Tag base classes (maps to lib/liquid/tag/)
│   └── profiler/            # Profiler (maps to lib/liquid/profiler/)
├── reference-liquid/         # Cloned Ruby repository for reference
└── README.md                 # Project README
```

## File Naming Conventions

This project maintains **exact file name parity** with the Ruby implementation to enable easy application of changelog updates.

### Rules:

1. **Base names match exactly** (only extension changes)

   - Ruby: `template.rb` → Go: `template.go`
   - Ruby: `if.rb` → Go: `if.go`

2. **Directory structure mirrors Ruby**

   - Ruby: `lib/liquid/tags/if.rb` → Go: `liquid/tags/if.go`
   - Ruby: `lib/liquid/tag/disableable.rb` → Go: `liquid/tag/disableable.go`

3. **Test files follow Go conventions**
   - Integration tests: `*_integration_test.go`
   - Unit tests: `*_unit_test.go`
   - Regular tests: `*_test.go`

### Why This Matters

When Shopify releases a new version:

1. Check `reference-liquid/History.md` for changes
2. Identify affected Ruby files
3. Find corresponding Go files using the same base name
4. Apply equivalent changes
5. Update version to match

This naming convention makes it trivial to track which files need updates.

## Versioning

**Version numbers match the Ruby repository exactly.**

- Current version: Tracked in `liquid/version.go`
- Maps to: `reference-liquid/lib/liquid/version.rb`
- Format: Semantic versioning (e.g., "5.10.0")
- When Ruby releases a new version, update Go version to match

## Reference Repository

The Ruby implementation is cloned in `reference-liquid/` directory. This serves as:

- **Implementation reference**: Understand how features work
- **Test reference**: Mirror test cases
- **Changelog source**: Track new releases via `History.md`
- **API reference**: Maintain API compatibility

**Do not modify** the reference repository. It should remain a clean clone of the upstream.

## Implementation Guidelines

### 1. Feature Parity

Goal: Achieve 100% feature parity with Ruby implementation.

- All tags must be implemented
- All filters must be implemented
- All error modes must be supported
- All edge cases must be handled

### 2. Go Idioms

While maintaining parity, use Go idioms:

- Structs for classes
- Interfaces for polymorphism
- Methods for behavior
- Error returns instead of exceptions
- Go naming conventions (exported = CapitalCase)

### 3. Testing

- Write tests for every feature
- Mirror Ruby test cases
- Use Go's `testing` package
- Aim for same coverage as Ruby version

### 4. Documentation

- Document public APIs with Go doc comments
- Reference Ruby implementation in complex logic
- Keep this file updated with architectural decisions

## Applying Updates from Ruby Repository

### Process

1. **Check for Updates**

   ```bash
   cd reference-liquid
   git fetch origin
   git log HEAD..origin/main --oneline
   ```

2. **Review Changelog**

   - Read `reference-liquid/History.md`
   - Identify new features/fixes

3. **Identify Affected Files**

   - Check git diff for specific changes
   - Map Ruby files to Go files using naming convention

4. **Implement Changes**

   - Read Ruby changes
   - Implement equivalent Go changes
   - Maintain same behavior

5. **Update Version**

   - Update `liquid/version.go` to match Ruby version
   - Commit with version number

6. **Update Tests**
   - Add/modify tests as needed
   - Ensure all tests pass

### Example

Ruby release 5.10.0 adds inline snippets:

- Changelog mentions `snippet.rb` changes
- Map to `liquid/tags/snippet.go`
- Check git diff: `git show v5.10.0 -- lib/liquid/tags/snippet.rb`
- Implement equivalent changes
- Update version to "5.10.0"
- Add tests

## Development Workflow

### Starting a New Feature

1. Identify Ruby file: `reference-liquid/lib/liquid/feature.rb`
2. Create Go file: `liquid/feature.go`
3. Read Ruby implementation thoroughly
4. Implement Go equivalent
5. Write tests: `liquid/feature_test.go`
6. Verify behavior matches Ruby

### Writing Tests

1. Find Ruby test: `reference-liquid/test/integration/feature_test.rb`
2. Create Go test: `liquid/feature_integration_test.go`
3. Translate test cases to Go
4. Use Go testing patterns
5. Ensure same coverage

### Code Review Checklist

- [ ] File name matches Ruby convention
- [ ] Implementation matches Ruby behavior
- [ ] Tests mirror Ruby tests
- [ ] Version updated if needed
- [ ] Documentation updated
- [ ] All tests pass

## Key Files

### Core Implementation

- `liquid/template.go` - Main API (maps to `lib/liquid/template.rb`)
- `liquid/environment.go` - Configuration (maps to `lib/liquid/environment.rb`)
- `liquid/parser.go` - Parsing logic (maps to `lib/liquid/parser.rb`)
- `liquid/lexer.go` - Lexical analysis (maps to `lib/liquid/lexer.rb`)

### Tags

- `liquid/tags/` - All tag implementations
- See `.cursor/instructions.md` for complete mapping

### Tests

- `liquid/*_integration_test.go` - Integration tests
- `liquid/*_unit_test.go` - Unit tests

## Resources

- **Ruby Implementation**: https://github.com/Shopify/liquid
- **Liquid Documentation**: https://shopify.github.io/liquid/
- **Reference Code**: `reference-liquid/` directory
- **Detailed Guide**: `.cursor/instructions.md`

## Contributing

When contributing:

1. Follow file naming conventions strictly
2. Reference Ruby implementation
3. Write comprehensive tests
4. Update version if needed
5. Document any deviations

## License

This project follows the same license as the Ruby implementation (MIT).
