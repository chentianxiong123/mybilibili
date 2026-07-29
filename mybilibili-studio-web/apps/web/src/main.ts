import { createApp } from "vue";
import App from "./App.vue";
import "./index.css";
import { initCustomFonts } from "./components/editor/inspector/font-options";

void initCustomFonts();

const root = document.getElementById("root")!;

const app = createApp(App);
app.mount(root);
