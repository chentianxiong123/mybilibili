import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import vue from "@vitejs/plugin-vue";
import path from "path";

// https://vitejs.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiTarget =
    env.VITE_MYBILIBILI_API_PROXY_TARGET || "http://127.0.0.1:8080";

  return {
    plugins: [react(), vue()],
    assetsInclude: ["**/*.wasm"],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "./src"),
        "@mybilibili-studio/core": path.resolve(__dirname, "../../packages/core/src"),
      },
    },
    worker: {
      format: "es",
    },
    optimizeDeps: {
      exclude: ["@ffmpeg/ffmpeg", "@ffmpeg/util", "@ffmpeg/core", "@ffmpeg/core-mt"],
    },
    build: {
      target: "esnext",
      rollupOptions: {
        output: {
          manualChunks: (id) => {
            if (id.includes("node_modules/react") || id.includes("node_modules/react-dom")) {
              return "react";
            }
            if (id.includes("node_modules/zustand")) {
              return "zustand";
            }
            if (id.includes("node_modules/three")) {
              return "three";
            }
            if (id.includes("node_modules/@radix-ui")) {
              return "radix";
            }
          },
        },
      },
    },
    server: {
      headers: {
        "Cross-Origin-Opener-Policy": "same-origin",
        "Cross-Origin-Embedder-Policy": "require-corp",
      },
      proxy: {
        "/api": {
          target: apiTarget,
          changeOrigin: true,
        },
      },
    },
    preview: {
      headers: {
        "Cross-Origin-Opener-Policy": "same-origin",
        "Cross-Origin-Embedder-Policy": "require-corp",
      },
    },
  };
});

