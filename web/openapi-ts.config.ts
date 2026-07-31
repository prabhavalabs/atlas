import { defineConfig } from "@hey-api/openapi-ts"

export default defineConfig({
  input: "../api/openapi.yaml",
  output: "src/lib/api/generated",
})
