import React, {useEffect, useState} from 'react';
import {Route, Routes, useLocation, useNavigate} from 'react-router-dom';
import './App.css'
import LivePage from "./pages/LivePage.jsx";
import LiveDetailPage from "./pages/LiveDetailPage.jsx";
import {
    Alert, Badge,
    Button,
    Dropdown,
    DropdownItem,
    DropdownMenu,
    DropdownTrigger,
    Link, ModalHeader,
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
import BoxPage from "./pages/BoxPage";
import HighLightPage from "./pages/HighLightPage";
import BlackListPage from "./pages/BlackListPage";
import CommentForm from "./components/CommentForm";
import TracePage from "./pages/TracePage";
import FansPage from "./pages/FansPage";
import IndexPage from "./pages/IndexPage";
import {useTheme} from "next-themes";
import DocsDialog from "./components/DocsDialog";

import mitt from 'mitt';
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import DynamicCard from "./components/DynamicCard";

export const eventBus = mitt();

const calcHeight = () => {
    const vh = window.innerHeight;
    const rem = parseFloat(getComputedStyle(document.documentElement).fontSize);
    const result = vh - 4 * rem;
    return result;
}
const MessageIcon = (props) => (
    <svg xmlns="http://www.w3.org/2000/svg" className="icon" viewBox="0 0 1024 1024"
         style={{ width: '32px', height: '32px' }}>
        <path
            d="M464 512a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm200 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm-400 0a48 48 0 1 0 96 0 48 48 0 1 0-96 0zm661.2-173.6c-22.6-53.7-55-101.9-96.3-143.3a444.35 444.35 0 0 0-143.3-96.3C630.6 75.7 572.2 64 512 64h-2c-60.6.3-119.3 12.3-174.5 35.9a445.35 445.35 0 0 0-142 96.5c-40.9 41.3-73 89.3-95.2 142.8-23 55.4-34.6 114.3-34.3 174.9A449.4 449.4 0 0 0 112 714v152a46 46 0 0 0 46 46h152.1A449.4 449.4 0 0 0 510 960h2.1c59.9 0 118-11.6 172.7-34.3a444.48 444.48 0 0 0 142.8-95.2c41.3-40.9 73.8-88.7 96.5-142 23.6-55.2 35.6-113.9 35.9-174.5.3-60.9-11.5-120-34.8-175.6zm-151.1 438C704 845.8 611 884 512 884h-1.7c-60.3-.3-120.2-15.3-173.1-43.5l-8.4-4.5H188V695.2l-4.5-8.4C155.3 633.9 140.3 574 140 513.7c-.4-99.7 37.7-193.3 107.6-263.8 69.8-70.5 163.1-109.5 262.8-109.9h1.7c50 0 98.5 9.7 144.2 28.9 44.6 18.7 84.6 45.6 119 80 34.3 34.3 61.3 74.4 80 119 19.4 46.2 29.1 95.2 28.9 145.8-.6 99.6-39.7 192.9-110.1 262.7z" />
    </svg>
);

const HelpIcon = () => {
    return (

        <svg xmlns="http://www.w3.org/2000/svg" height="32px" viewBox="0 -960 960 960" width="32px" fill="#1f1f1f"><path d="M513.5-254.5Q528-269 528-290t-14.5-35.5Q499-340 478-340t-35.5 14.5Q428-311 428-290t14.5 35.5Q457-240 478-240t35.5-14.5ZM442-394h74q0-33 7.5-52t42.5-52q26-26 41-49.5t15-56.5q0-56-41-86t-97-30q-57 0-92.5 30T342-618l66 26q5-18 22.5-39t53.5-21q32 0 48 17.5t16 38.5q0 20-12 37.5T506-526q-44 39-54 59t-10 73Zm38 314q-83 0-156-31.5T197-197q-54-54-85.5-127T80-480q0-83 31.5-156T197-763q54-54 127-85.5T480-880q83 0 156 31.5T763-763q54 54 85.5 127T880-480q0 83-31.5 156T763-197q-54 54-127 85.5T480-80Zm0-80q134 0 227-93t93-227q0-134-93-227t-227-93q-134 0-227 93t-93 227q0 134 93 227t227 93Zm0-320Z"/></svg>
    )
}

const RefreshIcon = () => {
    return (
        <svg xmlns="http://www.w3.org/2000/svg" height="32px" viewBox="0 -960 960 960" width="32px" fill="#1f1f1f">
            <path d="M480-160q-134 0-227-93t-93-227q0-134 93-227t227-93q69 0 132 28.5T720-690v-110h80v280H520v-80h168q-32-56-87.5-88T480-720q-100 0-170 70t-70 170q0 100 70 170t170 70q77 0 139-44t87-116h84q-28 106-114 173t-196 67Z"/>
        </svg>
    )
}

const AlertIcon = () => {
    return (
        <svg xmlns="http://www.w3.org/2000/svg" height="24px" viewBox="0 -960 960 960" width="24px" fill="#1f1f1f">
            <path d="M400-380v-440h160v440h-160Zm0 220v-160h160v160h-160Z"/>
        </svg>
    )
}
function BasicLayout() {


    const menu = [{
        Name: 'Overview',
        Path: '/'
    }, {
        Name: 'List', Path: '/list'
    }, {
        Name: 'Live', Path: '/live'
    }, {Name: "Raw", Path: '/raw'}, {Name: 'PK', Path: '/pk'}
    ]

    const [ind, setInd] = React.useState(0);

    const redirect = useNavigate()

    const [showDownload, setShowDownload] = React.useState(false);

    const [showRank, setShowRank] = React.useState(false);

    const [showNotice, setShowNotice] = React.useState(false);

    const [content, setContent] = React.useState("");

    const [showSetting,setShowSettings] = React.useState(false)

    const { theme, setTheme } = useTheme()

    const [showMoneyRank, setShowMoneyRank] = React.useState(false)

    const [opacity,setOpacity] = React.useState(window.getOpacity())

    const [showDoc,setShowDoc] = React.useState(false)

    const [showBox,setShowBox] = React.useState(false)

    useEffect(() => {
        const id = "github-markdown-css";

        document.getElementById(id)?.remove();

        const link = document.createElement("link");
        link.id = id;
        link.rel = "stylesheet";
        link.href = `https://unpkg.shop.jd.com/github-markdown-css/github-markdown-${theme}.css`;
        document.head.appendChild(link);
    }, [theme]);

    useEffect(() => {
        axios.get("/about.md").then((response) => {
            setContent(response.data)

            if (navigator.userAgent !== 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/137.0.0.0 Safari/537.36') {
                var last = localStorage.getItem("news")
                if (last === null || last !== response.data) {

                    fetchMoney(true)
                    fetchGuild(true)
                    setShowNotice(true)
                    localStorage.setItem("news", response.data)

                }
            }


        })
    }, [])
    const [commentOpen,setCommentOpen] = useState(false);
    const hide = location.href.includes("hide") || window.parent !== window
    const loc = useLocation();
    const [showRefresh,setShowRefresh] = useState(false)
    useEffect(()=>{
        if( loc.pathname.includes("lives/")){
            setShowRefresh(true)
        } else {
            setShowRefresh(false)
        }
    },[loc.pathname])

    const [docName,setDocName] = useState()

    const [showDocPoint,setShowDocPoint] = useState(false)

    useEffect(() => {
        var p = loc.pathname
        var fName = ''
        if (p === '/') {
            fName = 'index.md'
        }
        if (p === '/list') {
            fName = 'list.md'
        }

        if (p === '/raw') {
            fName = 'raw.md'
        }

        if (p === '/pk') {
            fName = 'pk.md'
        }

        if (p === '/traces') {
            fName = 'traces.md'
        }

        if (p === '/reactions') {
            fName = 'reactions.md'
        }

        if (p === '/relation') {
            fName = 'relation.md'
        }

        if (p === '/fans') {
            fName = 'fans.md'
        }
        if (p === '/feeds') {
            fName = 'dynamics.md'
        }
        if (p === '/highlight') {
            fName = 'highlight.md'
        }
        if (p.includes( '/liver/')) {
            fName = 'liver.md'
        }
        if (p.includes( '/user/')) {
            fName = 'user.md'
        }
        if (p.includes( '/lives/')) {
            fName = 'detail.md'
        }
        if (fName !== '') {
            setDocName(fName)
            axios.get('/docs/' + fName).then((res) => {
                var str = localStorage.getItem('docs')
                var map = null
                if (str === null || JSON.parse(str) === undefined) {
                    map = new Map()
                } else {
                    map = JSON.parse(str)
                }
                if (res.data !== map[fName]) {
                    map[fName] = res.data
                    localStorage.setItem('docs', JSON.stringify(map))
                    setShowDocPoint(true)
                } else {
                    setShowDocPoint(false)
                }
            })
        } else {
            setDocName('')
        }
    }, [loc.pathname]);
    return (

        <div>
            <CommentForm isOpen={commentOpen} onChange={() => setCommentOpen(!commentOpen)} onClose={() => setCommentOpen(false)} />
            <DocsDialog isOpen={showDoc} onClose={() => setShowDoc(false)} fName={docName}/>
            <div className={'fixed right-[3vw] bottom-[3vw] z-40 flex flex-col'}>
                {showRefresh &&                 <Button
                    isIconOnly
                    startContent={<RefreshIcon/>}
                    onClick={() => {
                        eventBus.emit("refresh")
                    }}
                />}

                <Badge color={showDocPoint ? 'danger' : 'default'} size={'sm'} content="" placement="bottom-right" shape={showDocPoint?'circle':undefined}>
                    <Button
                        isIconOnly
                        className={'mt-2'}
                        startContent={<HelpIcon/>}
                        onClick={() => {
                            setShowDoc(true)
                            setShowDocPoint(false)
                        }}
                    />
                </Badge>
                <Button
                    isIconOnly
                    className={'mt-2'}
                    startContent={<MessageIcon/>}
                    onClick={() => {
                        setCommentOpen(true)
                    }}
                />
            </div>
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
                               className={'scrollbar-hide overflow-hidden'}>
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
                            <DropdownItem key="trace" onClick={() => {
                                redirect("/traces")
                            }}>添加主播</DropdownItem>
                            <DropdownItem key="rank" onClick={() => {
                                setShowRank(true);
                            }}>粉丝牌排行</DropdownItem>
                            <DropdownItem key="money-rank" onClick={() => {
                                setShowMoneyRank(true);
                            }}>打米排行</DropdownItem>
                            <DropdownItem key="notice" onClick={() => {
                                setShowNotice(true);
                            }}>更新日志</DropdownItem>
                            <DropdownItem key="setting" onClick={() => {
                                setShowSettings(true);
                            }}>设置</DropdownItem>
                            <DropdownItem key="reaction" onClick={() => {
                                redirect("/highlight")
                            }}>管人痴魅力时刻</DropdownItem>
                            <DropdownItem key="geo" onClick={() => {
                                redirect("/geo")
                            }}>Geo</DropdownItem>
                            <DropdownItem key="status" onClick={() => {
                                redirect("/stat")
                            }}>Status</DropdownItem>
                        </DropdownMenu>
                    </Dropdown>
                    <svg
                        xmlns="http://www.w3.org/2000/svg"
                        viewBox="0 0 24 24"
                        className="w-6 h-6 text-black dark:text-white dark:fill-white"
                        fill="currentColor"
                        onClick={() => {
                            document.getElementsByTagName('html')[0].setAttribute('data-theme', theme === 'light' ? 'dark' : 'light')
                            setTheme(theme === 'light'?'dark':'light')
                        }}
                    >
                        <path d="M12 7c-2.76 0-5 2.24-5 5s2.24 5 5 5 5-2.24 5-5-2.24-5-5-5M2 13h2c.55 0 1-.45 1-1s-.45-1-1-1H2c-.55 0-1 .45-1 1s.45 1 1 1m18 0h2c.55 0 1-.45 1-1s-.45-1-1-1h-2c-.55 0-1 .45-1 1s.45 1 1 1M11 2v2c0 .55.45 1 1 1s1-.45 1-1V2c0-.55-.45-1-1-1s-1 .45-1 1m0 18v2c0 .55.45 1 1 1s1-.45 1-1v-2c0-.55-.45-1-1-1s-1 .45-1 1M5.99 4.58c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0s.39-1.03 0-1.41zm12.37 12.37c-.39-.39-1.03-.39-1.41 0-.39.39-.39 1.03 0 1.41l1.06 1.06c.39.39 1.03.39 1.41 0 .39-.39.39-1.03 0-1.41zm1.06-10.96c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0zM7.05 18.36c.39-.39.39-1.03 0-1.41-.39-.39-1.03-.39-1.41 0l-1.06 1.06c-.39.39-.39 1.03 0 1.41s1.03.39 1.41 0z"></path>
                    </svg>
                </NavbarContent>
            </Navbar>}
            <div className={`site-layout-background`}
                 style={{padding: 24, width: '100%', height: `${calcHeight()}px`, opacity: parseInt(opacity) / 100}}>
                <AnimatePresence mode="wait">
                    <Routes location={location} key={location.pathname}>
                        <Route path="/" element={<PageWrapper><IndexPage/></PageWrapper>}/>
                        <Route path="/lives" element={<PageWrapper><LivePage/></PageWrapper>}/>
                        <Route path="/live" element={<PageWrapper><LivePage/></PageWrapper>}/>
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
                        <Route path="/feeds" element={<PageWrapper><ReactionPage type={'feed'}/></PageWrapper>}/>
                        <Route path="/demo" element={<PageWrapper><DemoPage/></PageWrapper>}/>
                        <Route path="/box" element={<PageWrapper><BoxPage/></PageWrapper>}/>
                        <Route path="/relation" element={<PageWrapper><BlackListPage/></PageWrapper>}/>
                        <Route path="/traces" element={<PageWrapper><TracePage/></PageWrapper>}/>
                        <Route path="/fans" element={<PageWrapper><FansPage/></PageWrapper>}/>
                        <Route path="/index" element={<PageWrapper><IndexPage/></PageWrapper>}/>
                        <Route path="/highlight" element={<PageWrapper><HighLightPage/></PageWrapper>}/>
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