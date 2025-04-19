import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import {BrowserRouter} from "react-router-dom";
import {HeroUIProvider} from "@heroui/react";

window.debug = true

//https://github.com/heroui-inc/heroui/discussions/2080?sort=top#discussioncomment-9207779
const originalWarn = console.warn
console.warn = (...args) => {
    const [firstArg] = args;
    if (
        typeof firstArg === "string" &&
        firstArg.includes(
            "An aria-label or aria-labelledby prop is required for accessibility."
        )
    ) {
        return;
    }

    originalWarn(...args);
};
window.toSpace = UID => {
    window.open('https://space.bilibili.com/' + UID, '_blank');
}
const host = location.hostname;

const port = location.port;

const protocol = location.protocol.replace(":","")
createRoot(document.getElementById('root')).render(
  <StrictMode>
      <HeroUIProvider>
          <BrowserRouter>
              <App />
          </BrowserRouter>
      </HeroUIProvider>
  </StrictMode>,
)
