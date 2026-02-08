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
document.title = "Vtuber 数据"
axios.interceptors.request.use((config) => {
    if (import.meta.env.PROD) {
        config.url = config.url?.replace('/api', '')
        if (config.url.startsWith('/')) {
            config.url = 'https://api.vtb.cat' + config.url;
        }
        var url = new URL(config.url);
        url.host = 'api.vtb.cat'
        config.url = url.toString()
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

window.inspectGuard = (obj) => {

    //返回true表示是刷的舰长

    if (obj.MessageCount > 20) return false

    if (obj.MessageCount === 0 && obj.GuardCount === 0 && obj.Level > 21) {
        return false
    }

    if (obj.Level > 24) {
        return false
    }

    if (obj.OverEnter === true) {
        return false
    }

    if (obj.MessageCount === 0) {
        return true
    }

    if (obj.TimeOut === false) {
        return true
    }

    if (obj.GuardCount / (obj.MessageCount + obj.GuardCount) > 0.5) {
        return true
    }





    return false

}
const fetchGuild = async (force) => {
    if (!localStorage.getItem("guild") || force) {
        const response = await fetch('https://storage.ikun.dev/d/Microsoft365/static/bili_guild_infos.json?sign=Jgd-iZ5deklFU3Jbjq4lp2-TVdD1h44aNA5XUsi79n4=:0',{
            referrerPolicy: "no-referrer"
        });
        const arrayBuffer = await response.arrayBuffer()
        var dec = new TextDecoder();
        localStorage.setItem("guild", dec.decode(arrayBuffer).substring(0));
    }
}
const fetchMoney = async (force) => {
    if (!localStorage.getItem("money") || force) {
        const response = await fetch('https://storage.ikun.dev/d/Microsoft365/static/rank.json?sign=bOejKONpD3-QV8TtS64DwgtZEwbZy2yt3uCkNn2yolc=:0',{
            referrerPolicy: "no-referrer"
        });
        const arrayBuffer = await response.arrayBuffer()
        var dec = new TextDecoder();
        localStorage.setItem("money", dec.decode(arrayBuffer).substring(0));
    }
}
window.fetchMoney = fetchMoney
window.fetchGuild = fetchGuild
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

var session = localStorage.getItem("session")
if (session === null || session === undefined || session === '') {
    function UUID() {
        return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function(c) {
            const r = Math.random() * 16 | 0;
            const v = c === "x" ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    localStorage.setItem("session", UUID())

}

if (localStorage.getItem("defaultPageSize") === null) {
    localStorage.setItem("defaultPageSize", "10");
}

window.prefetch = (url) => {
    const link = document.createElement("link");
    link.rel = "prefetch";
    if (import.meta.env.PROD) {
        url = url?.replace('/api', '');
        //config.url = config.url?.replace('live.ikun.dev', 'live-api.ikun.dev');
    }
    link.href = url;
    document.head.appendChild(link);
}

createRoot(document.getElementById('root')).render(
      <HeroUIProvider>
          <BrowserRouter>
                  <App />
          </BrowserRouter>
      </HeroUIProvider>
)
