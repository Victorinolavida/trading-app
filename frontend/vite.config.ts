import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "");

  const port = Number(env.PORT ?? 5173);
  const gatewayUrl = env.GATEWAY_URL ?? "http://localhost:3001";

  return {
    plugins: [react(), tailwindcss()],
    server: {
      port,
      proxy: {
        "/api": {
          target: gatewayUrl,
          changeOrigin: true,
        },
      },
    },
  };
});
