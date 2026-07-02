# Vben API Modules

This directory is reserved for hand-written business API modules.

Suggested convention:

- Keep each business domain in its own subdirectory.
- Register routes from `vbenapi/module_routes.go`.
- Reuse the root `vbenapi.Store` and shared response/auth helpers where practical.
