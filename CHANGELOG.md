# Changelog

## [5.11.0]

Compatibility update matching [Shopify Liquid v5.11.0](https://github.com/Shopify/liquid/releases/tag/v5.11.0).

### Added
- `strict2` error mode as alias for `rigid` (backwards compatible)
- `LiquidError` interface for unified error handling
- Integration tests for `strict_filters` mode

### Removed
- **Snippet tag**: `{% snippet %}` tag removed (reverted per Shopify Liquid v5.11.0)
- `SnippetDrop` type removed

### Changed
- **Render tag**: Now only accepts string literals for template names
  ```liquid
  {% render 'template' %}    <!-- Works -->
  {% render variable %}      <!-- Error: invalid syntax -->
  ```
- Simplified `HandleError` using `LiquidError` interface (~100 → ~25 lines)
- Simplified panic recovery in variable.go and template.go

### Fixed
- **Concurrent template rendering** ([#2](https://github.com/Notifuse/liquidgo/issues/2)): Added mutex protection for thread-safe concurrent renders
- **Strict filters**: `strictFilters=true` now correctly raises `UndefinedFilter` errors
- **Error type preservation**: All 16 error types now preserved (7 were incorrectly converted to `InternalError`)
