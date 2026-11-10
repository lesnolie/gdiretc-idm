import { defineConfig } from "mygo-cli";
export default defineConfig({
 name: "gdirect IDM", slug: "gdiretc-idm", identifier: "dev.lesnolie.gdiretcidm", version: "0.1.0",
 devUrl: "http://localhost:5173", devCommand: "bun run dev:web",
 buildCommand: "bun run build:web", frontendDist: "dist",
 bindings: "src/mygo.ts", out: "build"
});
