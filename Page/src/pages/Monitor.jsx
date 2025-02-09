import classes from "./Monitor.module.css";
import {useEffect, useState} from "react";
import axios from "axios"
import {Button, Card, Input, message} from "antd";
import LiveCard from "../components/LiveCard.jsx";

const Monitor = () => {
    const [monitor,setMonitor] = useState([])

    const [messageApi, contextHolder] = message.useMessage();
    const [text,setText] = useState("")

    useEffect(() => {
        axios.get("http://localhost:8080/monitor").then(res => {
            //setMonitor(res.data.lives)
            res.data.lives.sort((a, b) => {
                if(a.Live === b.Live) return 0; // 当前后两个元素的isOnline相同时, 不改变顺序
                return a.Live ? -1 : 1; // 如果a的isOnline为true, 那么让a排在b前面, 否则排在后面
            });
            setMonitor(res.data.lives)
        })
    }, [])

    return (
        <div className={classes.container}>
            {contextHolder}
            {monitor.map((item, index) => {
                return <LiveCard key={index} liveData={item} />
            })}
            <Card  style={{ width: 300, marginRight: '20px' ,margin:'15px'}}>
                <div style={{display: 'flex', justifyContent: 'space-between',}}>
                    <Input placeholder="请输入房间号" onChange={(e) => {
                        setText(e.target.value)
                    }}/>
                    <Button onClick={() => {
                        axios.get("http://localhost:8080/add/" + text).then(res => {
                            if (res.data.message === "success") {
                                messageApi.info('添加成功');
                            } else {
                                messageApi.error('直播间已存在')
                            }
                        })
                    }}>确定</Button>
                </div>
            </Card>
        </div>
    );
};

export default Monitor;