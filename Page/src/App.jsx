import React from 'react';
import {Route, Routes, useNavigate} from 'react-router-dom';
import {Layout, Menu} from 'antd';
import Monitor from "./pages/Monitor.jsx";
import './App.css'
import LivePage from "./pages/LivePage.jsx";
import Charts from "./pages/Charts.jsx";
import LiveDetailPage from "./pages/LiveDetailPage.jsx";
import {Avatar, Link, Navbar, NavbarContent, NavbarItem} from "@heroui/react";

const {Header, Content, Footer, Sider} = Layout;
const {SubMenu} = Menu;

function BasicLayout() {


    const menu = [{
        Name: 'Monitor',
        Path: '/'
    }, {
        Name: 'LivePage',
        Path: '/lives'
    }]

    const [ind, setInd] = React.useState(0);

    const redirect = useNavigate()


    return (

        <div>
            <Navbar style={{}}>
                <NavbarContent style={{display: "flex", justifyContent: "center"}}>
                    {
                        menu.map((item, index) => (
                            <NavbarItem isActive={index === ind}>
                                <Link color="foreground" onClick={() => {
                                    setInd(index);
                                    redirect(item.Path);

                                }}>
                                    {item.Name}
                                </Link>
                            </NavbarItem>
                        ))
                    }
                </NavbarContent>
            </Navbar>
            <div className="site-layout-background" style={{padding: 24, width: '100%', height: '100%'}}>
                <Routes>
                    <Route path="/" element={<Monitor/>}>

                    </Route>
                    <Route path="/lives" element={<LivePage/>}>

                    </Route>
                    <Route path={'/charts'} element={<Charts/>}>

                    </Route>
                    <Route path={'/lives/:id'} element={<LiveDetailPage/>}>

                    </Route>
                </Routes>
            </div>
        </div>
    )

}

export default BasicLayout;