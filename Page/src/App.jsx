import React, {useEffect, useState} from 'react';
import {Route, Routes, useNavigate} from 'react-router-dom';
import './App.css'
import LivePage from "./pages/LivePage.jsx";
import LiveDetailPage from "./pages/LiveDetailPage.jsx";
import {
    Dropdown,
    DropdownItem,
    DropdownMenu,
    DropdownTrigger,
    Link,
    Navbar,
    NavbarContent,
    NavbarItem
} from "@heroui/react";


import ListPage from "./pages/ListPage";
import LiverPage from "./pages/LiverPage";
import StatusPage from "./pages/StatusPage";
import RankDialog, {MoneyRankDialog} from "./components/RankDialog";
import UserPage from "./pages/UserPage";
import {AnimatePresence, motion} from "framer-motion";
import NoticeDialog from "./components/NoticeDialog";
import axios from "axios";
import SearchPage from "./pages/SearchPage";
import RawPage from "./pages/RawPage";
import ComparePage from "./pages/ComparePage";
import GeoPage from "./pages/GeoPage";
import ReactionPage from "./pages/ReactionPage";
import SettingDialog from "./components/SettingDialog";
import DemoPage from "./pages/DemoPage";
import {useTheme} from "next-themes";

const calcHeight = () => {
    const vh = window.innerHeight;
    const rem = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const result = vh - 4 * rem;
    return result;
}

function BasicLayout(props) {


    const { theme, setTheme } = useTheme()

    const menu = [{
        Name: 'Overview',
        Path: '/'
    }, {
        Name: 'List', Path: '/list'
    }, {
        Name: 'Search', Path: '/search'
    }, {
        Name: 'Status', Path: '/stat'
    }, {Name: "Raw", Path: '/raw'}, {Name: 'PK', Path: '/pk'}
    ]

    const [ind, setInd] = React.useState(0);

    const redirect = useNavigate()

    const [showDownload, setShowDownload] = React.useState(false);

    const [showRank, setShowRank] = React.useState(false);

    const [showNotice, setShowNotice] = React.useState(false);

    const [content, setContent] = React.useState("");

    const [showSetting,setShowSettings] = React.useState(false)


    const [showMoneyRank, setShowMoneyRank] = React.useState(false)

    const [opacity,setOpacity] = React.useState(window.getOpacity())

    useEffect(() => {
        axios.get("/about.md").then((response) => {
            setContent(response.data)

            if (navigator.userAgent !== 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36') {
                var last = localStorage.getItem("news")
                if (last === null || last !== response.data) {
                    localStorage.setItem("news",response.data)
                    fetchGuild(true)
                    fetchMoney(true)
                    setTimeout(() => {
                        setShowNotice(true)
                    },3000)
                }
            }


        })
    }, [])

    const hide = location.href.includes("hide")
    return (

        <div>
            {showNotice && <NoticeDialog onClose={() => {
                setShowNotice(false);
            }} content={content}></NoticeDialog>}
            {showRank && <RankDialog open={showRank} onClose={() => {
                setShowRank(false)
            }}/>}
            {showMoneyRank && <MoneyRankDialog open={showMoneyRank} onClose={() => {
                setShowMoneyRank(false)
            }}/>}
            {showSetting && <SettingDialog onOpacityChange={(e) => {
                setOpacity(e)
            }} onClose={() => {
                setShowSettings(false)
            }}/>}
            {!hide && <Navbar style={{}}>
                <NavbarContent style={{display: "flex", justifyContent: "center", "overflow": "scroll",width:vwToPx(100)}}
                               className={'scrollbar-hide '}>
                    {
                        menu.map((item, index) => (
                            <NavbarItem isActive={index === ind} key={index}>
                                <Link color="foreground" onPress={() => {
                                    setInd(index);
                                    redirect(item.Path);

                                }}>
                                    {item.Name}
                                </Link>
                            </NavbarItem>
                        ))

                    }
                    <Dropdown>
                        <DropdownTrigger>
                            <Link>
                                Misc
                            </Link>
                        </DropdownTrigger>
                        <DropdownMenu>
                            <DropdownItem key="view" onClick={() => {
                                setShowDownload(true);
                            }}>Bili Downloader</DropdownItem>
                            <DropdownItem key="rank" onClick={() => {
                                setShowRank(true);
                            }}>Level Rank</DropdownItem>
                            <DropdownItem key="money-rank" onClick={() => {

                                setShowMoneyRank(true);
                            }}>Money Rank</DropdownItem>
                            <DropdownItem key="notice" onClick={() => {
                                setShowNotice(true);
                            }}>Notice & Changelog</DropdownItem>
                            <DropdownItem key="setting" onClick={() => {
                                setShowSettings(true);
                            }}>Setting</DropdownItem>
                            <DropdownItem key="reaction" onClick={() => {
                                redirect("/reactions")
                            }}>Reaction</DropdownItem>
                            <DropdownItem key="geo" onClick={() => {
                                redirect("/geo")
                            }}>Geo</DropdownItem>
                        </DropdownMenu>
                    </Dropdown>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        className="w-6 h-6 text-black dark:text-white"
                        fill="currentColor"
                        onClick={() => {
                            setTheme(theme === 'light'?'dark':'light')
                            if (theme === 'dark') {
                                document.documentElement.style.setProperty("--heroui-content1", "240 25% 95%");
                            } else {
                                document.documentElement.style.setProperty("--heroui-content1", "0 0% 0%");
                            }
                        }}
                    >
                        <path d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5M2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1m18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1M11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1m0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1M5.99 4.58c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0s.39-1.03 0-1.41zm12.37 12.37c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0 .39-.39.39-1.03 0-1.41zm1.06-10.96c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0zM7.05 18.36c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0z"></path>
                    </svg>
                </NavbarContent>
            </Navbar>}
            <div className={`site-layout-background`} style={{padding: 24, width: '100%', height: `${calcHeight()}px`,opacity:parseInt(opacity)/100}}>
                <AnimatePresence mode="wait">
                    <Routes location={location} key={location.pathname} >
                        <Route path="/" element={<PageWrapper><LivePage/></PageWrapper>}/>
                        <Route path="/lives" element={<PageWrapper><LivePage/></PageWrapper>}/>
                        <Route path="/search" element={<PageWrapper><SearchPage/></PageWrapper>}/>
                        <Route path="/lives/:id" element={<PageWrapper><LiveDetailPage/></PageWrapper>}/>
                        {/*<Route path="/chat" element={<PageWrapper><ChatPage/></PageWrapper>}/>*/}
                        <Route path="/list" element={<PageWrapper><ListPage/></PageWrapper>}/>
                        <Route path="/stat" element={<PageWrapper><StatusPage/></PageWrapper>}/>
                        <Route path="/liver/:id" element={<PageWrapper><LiverPage/></PageWrapper>}/>
                        <Route path="/user/:id" element={<PageWrapper><UserPage/></PageWrapper>}/>
                        <Route path="/raw" element={<PageWrapper><RawPage/></PageWrapper>}/>
                        <Route path="/pk" element={<PageWrapper><ComparePage/></PageWrapper>}/>
                        <Route path="/geo" element={<PageWrapper><GeoPage/></PageWrapper>}/>
                        <Route path="/reactions" element={<PageWrapper><ReactionPage/></PageWrapper>}/>
                        <Route path="/demo" element={<PageWrapper><DemoPage/></PageWrapper>}/>
                    </Routes>
                </AnimatePresence>

            </div>
        </div>
    )

}

function PageWrapper({children}) {
    return (
     <motion.div
            initial={{opacity: 0, x: 20}}
            animate={{opacity: 1, x: 0}}
            exit={{opacity: 0, x: -20}}
            transition={{duration: 0.3}}
            className="h-full">

         {children}
        </motion.div>
    );
}

export default BasicLayout;