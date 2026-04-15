/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import react from '@vitejs/plugin-react';
import { defineConfig, transformWithEsbuild } from 'vite';
import pkg from '@douyinfe/vite-plugin-semi';
import path from 'path';
import { codeInspectorPlugin } from 'code-inspector-plugin';
const { vitePluginSemi } = pkg;

const resolveVendorChunk = (id) => {
  if (!id.includes('node_modules')) {
    return undefined;
  }
  const normalizedId = id.split(path.sep).join('/');
  // Keep Semi UI and i18n runtime together to avoid circular chunk warnings.
  if (
    normalizedId.includes('@douyinfe/semi-icons') ||
    normalizedId.includes('@douyinfe/semi-ui') ||
    normalizedId.includes('i18next-browser-languagedetector') ||
    normalizedId.includes('react-i18next') ||
    normalizedId.includes('/i18next/')
  ) {
    return 'ui-runtime';
  }
  if (normalizedId.includes('@lobehub/icons')) {
    return 'icon-runtime';
  }
  return undefined;
};

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  plugins: [
    codeInspectorPlugin({
      bundler: 'vite',
    }),
    {
      name: 'treat-js-files-as-jsx',
      async transform(code, id) {
        if (!/src\/.*\.js$/.test(id)) {
          return null;
        }

        // Use the exposed transform from vite, instead of directly
        // transforming with esbuild
        return transformWithEsbuild(code, id, {
          loader: 'jsx',
          jsx: 'automatic',
        });
      },
    },
    react(),
    vitePluginSemi({
      cssLayer: true,
    }),
  ],
  optimizeDeps: {
    force: true,
    esbuildOptions: {
      loader: {
        '.js': 'jsx',
        '.json': 'json',
      },
    },
  },
  build: {
    chunkSizeWarningLimit: 1800,
    rollupOptions: {
      output: {
        manualChunks: resolveVendorChunk,
      },
    },
  },
  server: {
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/mj': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
      '/pg': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
});
