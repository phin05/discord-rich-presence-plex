import react from "@vitejs/plugin-react-swc";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [react()],
	define: {
		APP_VERSION: JSON.stringify(process.env.APP_VERSION ?? "0.0.0-dev"),
	},
	resolve: {
		alias: {
			"@": "/src",
		},
	},
});
