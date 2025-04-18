import React from 'react';
import {Route, Routes, useNavigate} from 'react-router-dom';
import Monitor from "./pages/Monitor.jsx";
import './App.css'
import LivePage from "./pages/LivePage.jsx";
import LiveDetailPage from "./pages/LiveDetailPage.jsx";
import {
    Avatar,
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
import Danmaku from "./pages/Danmaku";
import ChatPage from "./pages/ChatPage";


function BasicLayout() {


    const menu = [{
        Name: 'Monitor',
        Path: '/'
    }, {
        Name: 'LivePage',
        Path: '/lives'
    },{
        Name:'Danmakus',Path :'/chat'
    }]

    const [ind, setInd] = React.useState(0);

    const redirect = useNavigate()

    const [showDownload, setShowDownload] = React.useState(false);

    PubSub.subscribe('DownloadDialog', (msg,data) => {
        console.log(msg,data);
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
            <div className="site-layout-background" style={{padding: 24, width: '100%', height: '100%'}}>
                <Routes>
                    <Route path="/" element={<Monitor/>}>

                    </Route>
                    <Route path="/lives" element={<LivePage/>}>

                    </Route>
                    <Route path={'/lives/:id'} element={<LiveDetailPage/>}>

                    </Route>
                    <Route path={'/danmaku'} element={<Danmaku/>}>

                    </Route>
                    <Route path={'chat/'} element={<ChatPage/>}/>
                </Routes>
            </div>
        </div>
    )

}

export default BasicLayout;