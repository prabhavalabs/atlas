# Atlas web application

The Vite workspace contains Atlas's anonymous public application and authenticated
administration portal. It uses React, TypeScript, TanStack Router and Query,
Zustand for local map selection, shadcn/ui, Tailwind CSS, and lazy MapLibre GL JS.

```sh
pnpm install --frozen-lockfile
pnpm test
pnpm lint
pnpm typecheck
pnpm build
```

Set `VITE_API_BASE_URL` to the Atlas API origin for cross-origin development or
production builds. Leave it empty when the Go application serves the API on the
same origin.

The API contract is `../api/openapi.yaml`. Run `pnpm generate:api` after changing
it; generated files under `src/lib/api/generated` are never edited manually.

New UI primitives must be added with the shadcn CLI and reviewed for accessible
labelling, keyboard behavior, loading/error/empty states, and mobile layout.
