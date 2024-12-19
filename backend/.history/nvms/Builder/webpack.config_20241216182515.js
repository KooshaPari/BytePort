import path from "path";
import { fileURLToPath } from "url";
import SpinSdkPlugin from "@fermyon/spin-sdk/plugins/webpack/index.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

export default {
	entry: "./src/spin.ts",
	experiments: {
		outputModule: true,
		asyncWebAssembly: true,
		syncWebAssembly: true,
	},
	module: {
		rules: [
			{
				test: /\.tsx?$/,
				use: "ts-loader",
				exclude: /node_modules/,
			},
			{
				test: /\.wasm$/,
				type: "webassembly/async",
			},
		],
	},
	resolve: {
		extensions: [".tsx", ".ts", ".js", ".wasm"],
	},
	output: {
		path: path.resolve(__dirname, "./"),
		filename: "dist.js",
		module: true,
		library: {
			type: "module",
		},
		webassemblyModuleFilename: "[hash].wasm",
	},
	plugins: [new SpinSdkPlugin()],
	optimization: {
		minimize: false,
		usedExports: true,
	},
};