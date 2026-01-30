import React, {useEffect} from 'react';
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

const calcHeight = () => {
    const vh = window.innerHeight;
    const rem = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const result = vh - 4 * rem;
    return result;
}

function BasicLayout() {


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
                <NavbarContent style={{display: "flex", justifyContent: "center", "overflow": "scroll"}}
                               className={'scrollbar-hide'}>
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
                 {/*       <Route path="/reactions" element={<PageWrapper><ReactionPage/></PageWrapper>}/>*/}
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
            className="h-full"
        >
            {children}
        </motion.div>
    );
}

export default BasicLayout;