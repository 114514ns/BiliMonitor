import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import './index.css'
import App from './App.jsx'
import {BrowserRouter} from "react-router-dom";
import {HeroUIProvider, ToastProvider} from "@heroui/react";
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
window.formatNumber = (num) => {



    if (num === null || num === undefined || isNaN(num) || num === 0) return '0';

    if (num < 1) {
        return num
    }

    const units = ['', 'K', 'M', 'G', 'T', 'P'];
    const tier = Math.floor(Math.log10(Math.abs(num)) / 3);

    if (tier === 0) return num + '';

    const unit = units[tier];
    const scaled = num / Math.pow(10, tier * 3);
    return scaled.toFixed(1) + unit + '';
}
window.vhToPx = (vhPercent) =>{
    const vh = window.innerHeight;
    return (vhPercent / 100) * vh;
}

window.vwToPx= (vhPercent) =>{
    const vh = window.innerWidth;
    return (vhPercent / 100) * vh;
}

window.AVATAR_API = 'https://workers.vrp.moe/bilibili/avatar/'
document.title = "Vtuber数据台"
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
window.getOpacity = ()=> {
    var op = localStorage.getItem("opacity")
    if (op === null) {
        localStorage.setItem("opacity","100")
        return 100
    } else {
        return op
    }
}

const fetchGuild = async () => {
    if (!localStorage.getItem("guild")) {
        const response = await fetch('https://i0.hdslb.com/bfs/im_new/8e9a54c0fb86a1f22a5da2a457205fcf2.png',{
            referrerPolicy: "no-referrer"
        });
        const arrayBuffer = await response.arrayBuffer()
        var dec = new TextDecoder();
        localStorage.setItem("guild", dec.decode(arrayBuffer).substring(16569));
    }
}
const fetchMoney = async () => {
    if (!localStorage.getItem("money")) {
        const response = await fetch('https://i0.hdslb.com/bfs/im_new/de4a78b0e06d48d42eddd6f8a0483b1e2.png',{
            referrerPolicy: "no-referrer"
        });
        const arrayBuffer = await response.arrayBuffer()
        var dec = new TextDecoder();
        localStorage.setItem("money", dec.decode(arrayBuffer).substring(16569));
    }
}
axios.interceptors.request.use(function (config) {
    config.url = config.url.replaceAll('%20','').replaceAll(' ','') //UserPage的room不知道为啥，第一次请求的时候room会是一个空格而不是空字符串，先这样吧
    return config;
  }, function (error) {
    return Promise.reject(error);
  });
window.SEARCH_LIVER = ""
fetchGuild();
fetchMoney();
var html = document.querySelector('html')
var item = localStorage.getItem("background")
if (item != null) {
    html.style.backgroundImage = `url("${item}")`
    html.style.backgroundSize = 'cover'
    html.style.backgroundAttachment = 'fixed'
    html.style.backgroundPosition = 'center'
}

createRoot(document.getElementById('root')).render(
      <HeroUIProvider>
          <BrowserRouter>
                  <App />
          </BrowserRouter>
      </HeroUIProvider>
)
