import React from 'react';
import {Route, Routes, useNavigate} from 'react-router-dom';
import Monitor from "./pages/Monitor.jsx";
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


function BasicLayout() {


    const menu = [{
        Name: 'Overview',
        Path: '/'
    }, {
        Name: 'LivePage',
        Path: '/lives'
    }, {
        Name: 'Danmakus', Path: '/chat'
    }, {
        Name: 'List', Path: '/list'},{
        Name: 'Status', Path: '/stat'}
    ]

    const [ind, setInd] = React.useState(0);

    const redirect = useNavigate()

    const [showDownload, setShowDownload] = React.useState(false);

    PubSub.subscribe('DownloadDialog', (msg, data) => {
        console.log(msg, data);
        setShowDownload(false);
    });
    return (

        <div>
            <DownloadDialog isOpen={showDownload}/>
            <Navbar style={{}}>
                <NavbarContent style={{display: "flex", justifyContent: "center"}}>
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
                                Toolkit
                            </Link>
                        </DropdownTrigger>
                        <DropdownMenu>
                            <DropdownItem key="view" onClick={() => {
                                setShowDownload(true);
                            }}>Bili Downloader</DropdownItem>
                        </DropdownMenu>
                    </Dropdown>
                </NavbarContent>
            </Navbar>
            <div className="site-layout-background" style={{padding: 24, width: '100%', height: '100vh'}}>
                <Routes>
                    <Route path="/" element={<Monitor/>}>

                    </Route>
                    <Route path="/lives" element={<LivePage/>}>

                    </Route>
                    <Route path={'/lives/:id'} element={<LiveDetailPage/>}>


                    </Route>
                    <Route path={'chat/'} element={<ChatPage/>}/>
                    <Route path={'/list'} element={<ListPage/>}/>
                    <Route path={'/stat'} element={<StatusPage/>}/>
                    <Route path={'/liver/:id'} element={<LiverPage/>}/>

                </Routes>
            </div>
        </div>
    )

}

export default BasicLayout;