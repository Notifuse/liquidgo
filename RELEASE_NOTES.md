# v5.11.0

Compatibility update matching [Shopify Liquid v5.11.0](https://github.com/Shopify/liquid/releases/tag/v5.11.0).

## Highlights

- **Thread-safe rendering** - Fixed concurrent map read/write panic ([#2](https://github.com/Notifuse/liquidgo/issues/2))
- **Shopify v5.11.0 compatibility** - `strict2` error mode, snippet tag removed, render tag simplified
- **Better error handling** - New `LiquidError` interface, all 16 error types preserved

## Breaking Changes

- `{% snippet %}` tag removed (reverted per Shopify Liquid v5.11.0)
- `{% render %}` now only accepts string literals: `{% render 'name' %}` not `{% render variable %}`

## Added

- `strict2` error mode as alias for `rigid` (backwards compatible)
- `LiquidError` interface for unified error handling

## Fixed

- Concurrent template rendering now thread-safe with mutex protection
- `strictFilters=true` correctly raises `UndefinedFilter` errors
- Error type preservation (7 types were incorrectly converted to `InternalError`)
