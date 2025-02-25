import classes from "./Monitor.module.css";
import {useEffect, useState} from "react";
import axios from "axios"
import {Button, Card, Input, message} from "antd";
import LiveCard from "../components/LiveCard.jsx";

const Monitor = () => {
    const [monitor,setMonitor] = useState([])

    const [messageApi, contextHolder] = message.useMessage();
    const [text,setText] = useState("")

    const host = location.hostname;

    const port = location.port;

    const protocol = location.protocol.replace(":","")

    let init = {}

    const refresh = () => {
        axios.get(`${protocol}://${host}:${port}/monitor`).then(res => {
            //setMonitor(res.data.lives)
            if (JSON.stringify(init) === "{}") {
                var j = res.data.lives.sort((a, b) => {
                    // 首先按 Live 排序，Live 为 true 的排在前面
                    if (a.Live === b.Live) {
                        // 如果 Live 相同，再根据 UName 排序
                        return a.UID > b.UID ? 1 : -1;
                    }
                    // Live 为 true 的排在前面，Live 为 false 的排在后面
                    return a.Live ? -1 : 1;
                });
                setMonitor(j);
                init = j

            } else {

                init.forEach((live) => {
                    var id = live.UID
                    res.data.lives.forEach(live0 => {
                        if (live0.UID === id) {
                            if (live.Live != live0.Live) {
                                live.Live = live0.Live
                            }
                        }
                    })
                })
                setMonitor(init)

            }
            const sort = res.data.lives.sort((a, b) => {
                // 首先按 Live 排序，Live 为 true 的排在前面
                if (a.Live === b.Live) {
                    // 如果 Live 相同，再根据 UName 排序
                    return a.UID > b.UID ? 1 : -1;
                }
                // Live 为 true 的排在前面，Live 为 false 的排在后面
                return a.Live ? -1 : 1;
            });
            setMonitor(sort)
        })
    }

    useEffect(() => {
        refresh();

        const intervalId = setInterval(() => {
            refresh();
        }, 2000);

        return () => clearInterval(intervalId);
    }, []);





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
                        axios.get(`http://${host}:${port}/add/` + text).then(res => {
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