// components/Layout.js
import React from 'react';
import { Layout } from 'antd';
import SideBar from './SideBar.jsx';


const { Header, Content, Footer } = Layout;

const CustomLayout = ({ children }) => {
    return (
        <Layout>
            <SideBar />
            <Layout>
                <Header style={{ background: '#fff', padding: 0 }}>My App</Header>
                <Content style={{ margin: '24px 16px 0', overflow: 'initial' }}>
                    <div style={{ padding: 24, background: '#fff' }}>{children}</div>
                </Content>
                <Footer style={{ textAlign: 'center' }}>My App ©2023</Footer>
            </Layout>
        </Layout>
    );
};

export default CustomLayout;