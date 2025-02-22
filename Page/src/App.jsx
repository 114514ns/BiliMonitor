import React from 'react';
import {Link, Route, Routes} from 'react-router-dom';
import {Layout, Menu} from 'antd';
import {DesktopOutlined, PieChartOutlined} from '@ant-design/icons';
import Monitor from "./pages/Monitor.jsx";
import './App.css'
import { redirect } from "react-router";
import LivePage from "./pages/LivePage.jsx";
import Charts from "./pages/Charts.jsx";
import LiveDetailPage from "./pages/LiveDetailPage.jsx";

const {Header, Content, Footer, Sider} = Layout;
const {SubMenu} = Menu;

class BasicLayout extends React.Component {
    state = {
        collapsed: false,
    };

    onCollapse = collapsed => {
        console.log(collapsed);
        this.setState({collapsed});
    };

    render() {
        return (
            <Layout style={{minHeight: '100vh', width: '100%', height: '100%'}}>
                <Sider collapsible collapsed={this.state.collapsed} onCollapse={this.onCollapse}>
                    <div className="logo"/>
                    <Menu theme="dark" defaultSelectedKeys={['/']} mode="inline">
                        <Menu.Item key="/" icon={<PieChartOutlined/>}>
                            Home
                            <Link to="/"></Link>
                        </Menu.Item>
                        <Menu.Item key="/lives" icon={<DesktopOutlined/>}>
                            Lives
                            <Link to="/lives"> </Link>
                        </Menu.Item>
                        <Menu.Item key="/charts" icon={<DesktopOutlined/>}>
                            Trending
                            <Link to="/charts"> </Link>
                        </Menu.Item>
                    </Menu>
                </Sider>
                <div className="site-layout">
                    <Content style={{margin: '0 16px'}}>
                        <div className="site-layout-background" style={{padding: 24, width: '100%', height: '100%'}}>
                            <Routes>
                                <Route path="/" element={<Monitor/>}>
                                    {/* 页面主体 */}
                                </Route>
                                <Route path="/lives" element={<LivePage/>}>

                                </Route>
                                <Route path={'/charts'} element={<Charts/>}>

                                </Route>
                                <Route path={'/lives/:id'} element={<LiveDetailPage/>}>

                                </Route>
                            </Routes>
                        </div>
                    </Content>
                </div>
            </Layout>
        );
    }
}

export default BasicLayout;