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
import DownloadDialog from "./components/DownloadDialog";
import PubSub from 'pubsub-js'

import ChatPage from "./pages/ChatPage";
import ListPage from "./pages/ListPage";
import LiverPage from "./pages/LiverPage";
import StatusPage from "./pages/StatusPage";
import RankDialog from "./components/RankDialog";
import UserPage from "./pages/UserPage";
import {AnimatePresence, motion} from "framer-motion";
import NoticeDialog from "./components/NoticeDialog";
import axios from "axios";
import SearchPage from "./pages/SearchPage";

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
    },{
        Name: 'Search', Path: '/search'
    }, {
        Name: 'Status', Path: '/stat'
    }
    ]

    const [ind, setInd] = React.useState(0);

    const redirect = useNavigate()

    const [showDownload, setShowDownload] = React.useState(false);

    const [showRank, setShowRank] = React.useState(false);

    const [showNotice, setShowNotice] = React.useState(false);

    PubSub.subscribe('DownloadDialog', (msg, data) => {
        console.log(msg, data);
        setShowDownload(false);
    });
    const [content, setContent] = React.useState("");
    useEffect(() => {
        axios.get("/about.md").then((response) => {
            setContent(response.data);
        })
    }, [])

    const hide = location.href.includes("hide")
    return (

        <div>
            {showNotice && <NoticeDialog onClose={() => {
                setShowNotice(false);
            }} content={content}></NoticeDialog>}
            <DownloadDialog isOpen={showDownload}/>
            {showRank && <RankDialog open={showRank} onClose={() => {
                setShowRank(false)
            }} content={content}/>}
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
                            }}>Rank</DropdownItem>
                            <DropdownItem key="notice" onClick={() => {
                                setShowNotice(true);
                            }}>Notice & Changelog</DropdownItem>
                        </DropdownMenu>
                    </Dropdown>
                </NavbarContent>
            </Navbar>}
            <div className="site-layout-background" style={{padding: 24, width: '100%', height: `${calcHeight()}px`}}>
                <AnimatePresence mode="wait">
                    <Routes location={location} key={location.pathname}>
                        <Route path="/" element={<PageWrapper><LivePage/></PageWrapper>}/>
                        <Route path="/lives" element={<PageWrapper><LivePage/></PageWrapper>}/>
                        <Route path="/search" element={<PageWrapper><SearchPage/></PageWrapper>}/>
                        <Route path="/lives/:id" element={<PageWrapper><LiveDetailPage/></PageWrapper>}/>
                        <Route path="/chat" element={<PageWrapper><ChatPage/></PageWrapper>}/>
                        <Route path="/list" element={<PageWrapper><ListPage/></PageWrapper>}/>
                        <Route path="/stat" element={<PageWrapper><StatusPage/></PageWrapper>}/>
                        <Route path="/liver/:id" element={<PageWrapper><LiverPage/></PageWrapper>}/>
                        <Route path="/user/:id" element={<PageWrapper><UserPage/></PageWrapper>}/>
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