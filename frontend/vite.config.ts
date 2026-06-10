import path from "node:path";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      // "@/..." を frontend/src/... の別名にする（深い相対パスを避ける）
      "@": path.resolve(__dirname, "src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // ブラウザからは同一オリジンの /api を叩き、Vite が 8080 の backend へ転送する。
      // これで開発時の CORS を回避できる。
      "/api": "http://localhost:8080",
    },
  },
});
