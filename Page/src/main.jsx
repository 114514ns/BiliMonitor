import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import {BrowserRouter} from "react-router-dom";
import {HeroUIProvider} from "@heroui/react";
import axios from "axios";

window.debug = true

//https://github.com/heroui-inc/heroui/discussions/2080?sort=top#discussioncomment-9207779
/*
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

 */
window.vhToPx = (vhPercent) =>{
    const vh = window.innerHeight;
    return (vhPercent / 100) * vh;
}

window.vwToPx= (vhPercent) =>{
    const vh = window.innerWidth;
    return (vhPercent / 100) * vh;
}

window.AVATAR_API = 'https://workers.vrp.moe/bilibili/avatar/'

axios.interceptors.request.use((config) => {
    if (import.meta.env.PROD) {
        config.url = config.url?.replace('/api', '');
        //config.url = config.url?.replace('live.ikun.dev', 'live-api.ikun.dev');
    }
    return config;
});
window.toSpace = UID => {
    window.open('https://space.bilibili.com/' + UID, '_blank');
}
window.host = location.hostname;

window.port = location.port;

window.protocol = location.protocol.replace(":","")
document.addEventListener("DOMContentLoaded", (event) => {
    window.page = 1
});
window.getGuardIcon= (level) =>{
    var array = ["","https://i1.hdslb.com/bfs/static/blive/blfe-live-room/static/img/logo-1.b718085..png","https://i1.hdslb.com/bfs/static/blive/blfe-live-room/static/img/logo-2.d43d078..png","https://i1.hdslb.com/bfs/static/blive/blfe-live-room/static/img/logo-3.6d2f428..png"]
    return array[level]
}
window.isMobile = ()=> {
    return /Mobi|Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent)  || window.innerWidth <= 768
}


createRoot(document.getElementById('root')).render(
      <HeroUIProvider>
          <BrowserRouter>
              <App />
          </BrowserRouter>
      </HeroUIProvider>
)
